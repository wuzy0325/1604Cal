package deviceconnect

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"cal1604/internal/device"
	"cal1604/internal/domain"
	apperrors "cal1604/internal/errors"
)

// DeviceStore 抽象设备读写能力，保持应用层不依赖具体存储实现。
type DeviceStore interface {
	Get(id string) (domain.Device, bool)
	Upsert(dev domain.Device)
}

// StatusPublisher 用于向外广播设备状态变更（通常由 API 层转发到 SSE）。
type StatusPublisher func(dev domain.Device)

// Config 定义连接可靠性策略。
// 说明：
// 1. AttemptTimeout 是单次调用的上限，避免底层 I/O 无限阻塞。
// 2. MaxAttempts + Backoff 形成“重试+退避”容错，兼顾稳定性和响应性。
type Config struct {
	ConnectAttemptTimeout time.Duration
	ConnectMaxAttempts    int
	ConnectInitialBackoff time.Duration
	ConnectMaxBackoff     time.Duration

	DisconnectAttemptTimeout time.Duration
	DisconnectMaxAttempts    int
	DisconnectInitialBackoff time.Duration
	DisconnectMaxBackoff     time.Duration
}

// Option 用于注入时间与睡眠策略，便于测试重试流程。
type Option func(*Service)

// Service 负责编排设备 connect/disconnect 状态机迁移。
// 状态迁移规则：
// disconnected -> connecting -> connected
// disconnected -> connecting -> error
// connected -> disconnected
// connected -> error（断连失败）
type Service struct {
	store   DeviceStore
	factory device.ConnectionDriverFactory
	config  Config
	publish StatusPublisher

	nowFn   func() time.Time
	sleepFn func(ctx context.Context, delay time.Duration) error

	mu            sync.Mutex
	activeDrivers map[string]device.ConnectionDriver
}

// DefaultConfig 返回默认连接可靠性配置。
func DefaultConfig() Config {
	return Config{
		ConnectAttemptTimeout:    600 * time.Millisecond,
		ConnectMaxAttempts:       3,
		ConnectInitialBackoff:    80 * time.Millisecond,
		ConnectMaxBackoff:        300 * time.Millisecond,
		DisconnectAttemptTimeout: 400 * time.Millisecond,
		DisconnectMaxAttempts:    2,
		DisconnectInitialBackoff: 40 * time.Millisecond,
		DisconnectMaxBackoff:     120 * time.Millisecond,
	}
}

// WithNowFunc 覆盖当前时间函数。
func WithNowFunc(nowFn func() time.Time) Option {
	return func(s *Service) {
		if nowFn != nil {
			s.nowFn = nowFn
		}
	}
}

// WithSleepFunc 覆盖退避等待函数。
func WithSleepFunc(sleepFn func(ctx context.Context, delay time.Duration) error) Option {
	return func(s *Service) {
		if sleepFn != nil {
			s.sleepFn = sleepFn
		}
	}
}

// NewService 创建设备连接应用服务。
func NewService(
	store DeviceStore,
	factory device.ConnectionDriverFactory,
	config Config,
	publisher StatusPublisher,
	options ...Option,
) *Service {
	if store == nil {
		panic("deviceconnect: store is nil")
	}
	if factory == nil {
		panic("deviceconnect: factory is nil")
	}

	svc := &Service{
		store:         store,
		factory:       factory,
		config:        normalizeConfig(config),
		publish:       publisher,
		nowFn:         time.Now,
		sleepFn:       defaultSleep,
		activeDrivers: make(map[string]device.ConnectionDriver),
	}

	if svc.publish == nil {
		svc.publish = func(domain.Device) {}
	}

	for _, option := range options {
		if option != nil {
			option(svc)
		}
	}

	return svc
}

// Connect 执行设备连接流程，严格遵循 connecting -> connected/error 状态迁移。
func (s *Service) Connect(ctx context.Context, id string) (domain.Device, error) {
	dev, ok := s.store.Get(id)
	if !ok {
		return domain.Device{}, apperrors.ErrNotFound
	}

	dev.Status = domain.DeviceStatusConnecting
	dev.LastErrorReason = ""
	dev.LastErrorAt = nil
	s.store.Upsert(dev)
	s.publish(dev)

	drv, err := s.factory.Create(dev)
	if err != nil {
		return s.markError(dev, fmt.Errorf("create driver: %w", err))
	}

	err = s.retryWithBackoff(
		ctx,
		s.config.ConnectAttemptTimeout,
		s.config.ConnectMaxAttempts,
		s.config.ConnectInitialBackoff,
		s.config.ConnectMaxBackoff,
		drv.Connect,
	)
	if err != nil {
		return s.markError(dev, fmt.Errorf("connect failed after retries: %w", err))
	}

	s.setActiveDriver(id, drv)

	// 连接成功后从硬件读取实际单位，确保显示单位与设备真实单位一致。
	if reader, ok := drv.(interface{ ReadUnit(context.Context) (string, error) }); ok {
		if unit, err := reader.ReadUnit(ctx); err == nil {
			dev.Unit = unit
		}
	}

	dev.Status = domain.DeviceStatusConnected
	dev.LastErrorReason = ""
	dev.LastErrorAt = nil
	s.store.Upsert(dev)
	s.publish(dev)

	return dev, nil
}

// Disconnect 执行设备断连流程；失败会落入 error 状态并记录原因。
func (s *Service) Disconnect(ctx context.Context, id string) (domain.Device, error) {
	dev, ok := s.store.Get(id)
	if !ok {
		return domain.Device{}, apperrors.ErrNotFound
	}

	drv, err := s.getOrCreateDriver(dev)
	if err != nil {
		return s.markError(dev, fmt.Errorf("create driver: %w", err))
	}

	err = s.retryWithBackoff(
		ctx,
		s.config.DisconnectAttemptTimeout,
		s.config.DisconnectMaxAttempts,
		s.config.DisconnectInitialBackoff,
		s.config.DisconnectMaxBackoff,
		drv.Disconnect,
	)
	if err != nil {
		return s.markError(dev, fmt.Errorf("disconnect failed after retries: %w", err))
	}

	s.clearActiveDriver(id)
	dev.Status = domain.DeviceStatusDisconnected
	dev.LastErrorReason = ""
	dev.LastErrorAt = nil
	s.store.Upsert(dev)
	s.publish(dev)

	return dev, nil
}

func (s *Service) markError(dev domain.Device, err error) (domain.Device, error) {
	now := s.nowFn()
	dev.Status = domain.DeviceStatusError
	dev.LastErrorReason = err.Error()
	dev.LastErrorAt = &now
	s.store.Upsert(dev)
	s.publish(dev)
	return dev, err
}

func (s *Service) getOrCreateDriver(dev domain.Device) (device.ConnectionDriver, error) {
	s.mu.Lock()
	drv, ok := s.activeDrivers[dev.ID]
	s.mu.Unlock()
	if ok {
		return drv, nil
	}

	return s.factory.Create(dev)
}

// GetActiveDriver 返回指定设备的已连接驱动实例；设备未连接时返回 nil。
func (s *Service) GetActiveDriver(id string) device.ConnectionDriver {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.activeDrivers[id]
}

func (s *Service) setActiveDriver(id string, drv device.ConnectionDriver) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeDrivers[id] = drv
}

func (s *Service) clearActiveDriver(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.activeDrivers, id)
}

func (s *Service) retryWithBackoff(
	ctx context.Context,
	attemptTimeout time.Duration,
	maxAttempts int,
	initialBackoff time.Duration,
	maxBackoff time.Duration,
	operation func(context.Context) error,
) error {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	backoff := initialBackoff

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, attemptTimeout)
		err := operation(attemptCtx)
		cancel()

		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		lastErr = err
		if attempt == maxAttempts {
			break
		}

		if backoff > 0 {
			if sleepErr := s.sleepFn(ctx, backoff); sleepErr != nil {
				return sleepErr
			}
			backoff = nextBackoff(backoff, maxBackoff)
		}
	}

	if lastErr == nil {
		return errors.New("operation failed with unknown error")
	}

	return lastErr
}

func nextBackoff(current time.Duration, max time.Duration) time.Duration {
	if current <= 0 {
		return 0
	}

	next := current * 2
	if max > 0 && next > max {
		return max
	}

	return next
}

func defaultSleep(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func normalizeConfig(cfg Config) Config {
	defaults := DefaultConfig()

	if cfg.ConnectAttemptTimeout <= 0 {
		cfg.ConnectAttemptTimeout = defaults.ConnectAttemptTimeout
	}
	if cfg.ConnectMaxAttempts <= 0 {
		cfg.ConnectMaxAttempts = defaults.ConnectMaxAttempts
	}
	if cfg.ConnectInitialBackoff < 0 {
		cfg.ConnectInitialBackoff = 0
	}
	if cfg.ConnectMaxBackoff <= 0 {
		cfg.ConnectMaxBackoff = defaults.ConnectMaxBackoff
	}
	if cfg.ConnectMaxBackoff < cfg.ConnectInitialBackoff {
		cfg.ConnectMaxBackoff = cfg.ConnectInitialBackoff
	}

	if cfg.DisconnectAttemptTimeout <= 0 {
		cfg.DisconnectAttemptTimeout = defaults.DisconnectAttemptTimeout
	}
	if cfg.DisconnectMaxAttempts <= 0 {
		cfg.DisconnectMaxAttempts = defaults.DisconnectMaxAttempts
	}
	if cfg.DisconnectInitialBackoff < 0 {
		cfg.DisconnectInitialBackoff = 0
	}
	if cfg.DisconnectMaxBackoff <= 0 {
		cfg.DisconnectMaxBackoff = defaults.DisconnectMaxBackoff
	}
	if cfg.DisconnectMaxBackoff < cfg.DisconnectInitialBackoff {
		cfg.DisconnectMaxBackoff = cfg.DisconnectInitialBackoff
	}

	return cfg
}
