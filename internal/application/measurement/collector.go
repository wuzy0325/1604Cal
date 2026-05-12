package measurement

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"cal1604/internal/application/session"
	"cal1604/internal/device"
	"cal1604/internal/domain"
	"cal1604/internal/events"
	"cal1604/internal/workflow"
)

// RunAutoCollection 按测点顺序执行 measurement 自己的自动采集流程。
func (s *Service) RunAutoCollection(ctx context.Context) error {
	if st := s.State(); st != domain.SessionStateReady {
		return fmt.Errorf("auto collection requires ready state, got %s", st)
	}

	points := s.GetPoints()
	pressureDriver := s.sess.PressureDriver()

	var collectErr error
pointsLoop:
	for _, point := range points {
		select {
		case <-ctx.Done():
			collectErr = ctx.Err()
			break pointsLoop
		default:
		}
		if err := s.ManualPressurize(ctx, point.Index); err != nil {
			if errors.Is(err, ErrPointSkipped) {
				log.Printf("[measurement] point %d skipped by user", point.Index)
				continue
			}
			collectErr = err
			break
		}
		if err := s.ManualCollect(ctx, point.Index); err != nil {
			collectErr = err
			break
		}
	}

	// 采集结束后停止压力控制。
	if pressureDriver != nil {
		_ = pressureDriver.Stop(ctx)
	}

	if collectErr != nil {
		return collectErr
	}

	if s.State() != domain.SessionStateCompleted {
		s.mu.Lock()
		if err := s.sessionMachine.Transition(domain.SessionStateCompleted); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("complete auto collection: %w", err)
		}
		s.syncSessionStatusLocked(domain.SessionStateCompleted)
		s.mu.Unlock()

		s.publish(events.EventMeasurementStateChanged, map[string]any{"state": string(domain.SessionStateCompleted)})
	}
	return nil
}

// ManualPressurize 对指定 measurement 点执行打压和稳定等待。
func (s *Service) ManualPressurize(ctx context.Context, pointIndex int) error {
	pressureDriver, point, stableWaitMs, stabilityTimeoutMs, err := s.preparePressureStep(pointIndex)
	if err != nil {
		return err
	}

	log.Printf("[measurement] ManualPressurize point=%d target=%.4f stableWait=%dms timeout=%dms",
		pointIndex, point.TargetPressure, stableWaitMs, stabilityTimeoutMs)

	s.updatePointStatus(pointIndex, domain.PointStatusPressurizing)
	if err := s.transitionTo(domain.SessionStatePressurizing); err != nil {
		return err
	}

	if err := pressureDriver.SetTargetPressure(ctx, point.TargetPressure); err != nil {
		return fmt.Errorf("set target pressure: %w", err)
	}
	if ctrl, ok := pressureDriver.(device.PressureControlCapable); ok {
		if err := ctrl.StartControl(ctx); err != nil {
			return fmt.Errorf("start pressure control: %w", err)
		}
	}

	s.updatePointStatus(pointIndex, domain.PointStatusStabilizing)
	if err := s.transitionTo(domain.SessionStateStabilizing); err != nil {
		return err
	}

	if err := s.waitForMeasurementStability(ctx, pointIndex, pressureDriver, point.TargetPressure, stableWaitMs, stabilityTimeoutMs); err != nil {
		return err
	}

	actualPressure, err := pressureDriver.ReadCurrentPressure(ctx)
	if err != nil {
		return fmt.Errorf("read current pressure: %w", err)
	}
	s.updatePointActualPressure(pointIndex, actualPressure)

	return nil
}

// ManualCollect 对指定 measurement 点执行一次按点采集。
func (s *Service) ManualCollect(ctx context.Context, pointIndex int) error {
	measureDriver, point, channels, averageCount, totalPoints, err := s.prepareCollectStep(pointIndex)
	if err != nil {
		return err
	}

	s.updatePointStatus(pointIndex, domain.PointStatusCollecting)
	if err := s.transitionTo(domain.SessionStateCollecting); err != nil {
		return err
	}

	samples := make([][]float64, 0, averageCount)
	for i := 0; i < averageCount; i++ {
		data, err := measureDriver.CollectData(ctx, channels)
		if err != nil {
			return fmt.Errorf("collect sample %d: %w", i+1, err)
		}
		samples = append(samples, append([]float64(nil), data...))
		if i < averageCount-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	flattened := make([]float64, 0, len(samples)*len(channels))
	for _, sample := range samples {
		flattened = append(flattened, sample...)
	}
	s.updatePointCollectedData(pointIndex, flattened, time.Now())
	s.publish(events.EventMeasurementDataCollected, map[string]any{
		"pointIndex": point.Index,
		"channels":   channels,
		"data":       flattened,
	})

	// 采集后自动检查报警。
	updatedPoint := s.getPoint(pointIndex)
	if alarm, _ := s.CheckAlarm(updatedPoint); alarm != nil {
		s.publish(events.EventMeasurementAlarmTriggered, alarm)
	}

	nextState := domain.SessionStateReady
	if pointIndex >= totalPoints {
		nextState = domain.SessionStateCompleted
	}
	if err := s.transitionTo(nextState); err != nil {
		return err
	}

	return nil
}

func (s *Service) preparePressureStep(pointIndex int) (device.PressureDriver, domain.PressurePoint, int, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pressureDriver := s.sess.PressureDriver()
	if pressureDriver == nil {
		return nil, domain.PressurePoint{}, 0, 0, session.ErrPressureDeviceNotSet
	}
	if pointIndex < 1 || pointIndex > len(s.points) {
		return nil, domain.PressurePoint{}, 0, 0, fmt.Errorf("invalid point index: %d", pointIndex)
	}

	stableWaitMs := s.config.StableWaitMs
	if stableWaitMs <= 0 {
		stableWaitMs = 5000
	}

	stabilityTimeoutMs := s.config.StabilityTimeoutMs
	if stabilityTimeoutMs <= 0 {
		stabilityTimeoutMs = 120000
	}

	return pressureDriver, s.points[pointIndex-1], stableWaitMs, stabilityTimeoutMs, nil
}

func (s *Service) prepareCollectStep(pointIndex int) (device.MeasureDriver, domain.PressurePoint, []int, int, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	measureDriver := s.sess.MeasureDriver()
	if measureDriver == nil {
		return nil, domain.PressurePoint{}, nil, 0, 0, session.ErrMeasureDeviceNotSet
	}
	if pointIndex < 1 || pointIndex > len(s.points) {
		return nil, domain.PressurePoint{}, nil, 0, 0, fmt.Errorf("invalid point index: %d", pointIndex)
	}

	channels := append([]int(nil), s.channels...)
	if len(channels) == 0 {
		channels = append([]int(nil), s.config.Channels...)
	}
	if len(channels) == 0 {
		return nil, domain.PressurePoint{}, nil, 0, 0, fmt.Errorf("no measurement channels configured")
	}

	averageCount := s.config.AverageCount
	if averageCount < 1 {
		averageCount = 1
	}

	return measureDriver, s.points[pointIndex-1], channels, averageCount, len(s.points), nil
}

func (s *Service) transitionTo(state domain.SessionState) error {
	s.mu.Lock()
	if err := s.sessionMachine.Transition(state); err != nil {
		s.mu.Unlock()
		return fmt.Errorf("transition to %s: %w", state, err)
	}
	s.syncSessionStatusLocked(state)
	s.mu.Unlock()

	s.publish(events.EventMeasurementStateChanged, map[string]any{"state": string(state)})
	return nil
}

func (s *Service) getPoint(pointIndex int) domain.PressurePoint {
	s.mu.Lock()
	defer s.mu.Unlock()
	if pointIndex < 1 || pointIndex > len(s.points) {
		return domain.PressurePoint{}
	}
	return s.points[pointIndex-1]
}

func (s *Service) updatePointStatus(pointIndex int, status string) {
	s.mu.Lock()
	if pointIndex < 1 || pointIndex > len(s.points) {
		s.mu.Unlock()
		return
	}
	s.points[pointIndex-1].Status = status
	if s.session != nil && pointIndex <= len(s.session.Points) {
		s.session.Points[pointIndex-1].Status = status
	}
	point := s.points[pointIndex-1]
	s.mu.Unlock()

	s.publish(events.EventMeasurementPointStatus, point)
}

func (s *Service) updatePointActualPressure(pointIndex int, actualPressure float64) {
	s.mu.Lock()
	if pointIndex < 1 || pointIndex > len(s.points) {
		s.mu.Unlock()
		return
	}
	s.points[pointIndex-1].ActualPressure = &actualPressure
	if s.session != nil && pointIndex <= len(s.session.Points) {
		s.session.Points[pointIndex-1].ActualPressure = &actualPressure
	}
	point := s.points[pointIndex-1]
	s.mu.Unlock()

	s.publish(events.EventMeasurementPointStatus, point)
}

func (s *Service) updatePointCollectedData(pointIndex int, data []float64, collectedAt time.Time) {
	s.mu.Lock()
	if pointIndex < 1 || pointIndex > len(s.points) {
		s.mu.Unlock()
		return
	}
	timestamp := collectedAt.UTC().Format(time.RFC3339)
	clonedData := append([]float64(nil), data...)
	s.points[pointIndex-1].CollectedData = clonedData
	s.points[pointIndex-1].CollectTime = timestamp
	s.points[pointIndex-1].Status = domain.PointStatusCompleted
	if s.session != nil && pointIndex <= len(s.session.Points) {
		s.session.Points[pointIndex-1].CollectedData = append([]float64(nil), data...)
		s.session.Points[pointIndex-1].CollectTime = timestamp
		s.session.Points[pointIndex-1].Status = domain.PointStatusCompleted
	}
	point := s.points[pointIndex-1]
	s.mu.Unlock()

	s.publish(events.EventMeasurementPointStatus, point)
}

// waitForMeasurementStability 等待压力稳定，通过 StabilityMonitor 判定并发布 SSE 进度事件。
// 超时时不直接失败，而是发布超时事件等待前端用户决定：继续等待或跳过当前点。
func (s *Service) waitForMeasurementStability(
	ctx context.Context,
	pointIndex int,
	pressureDriver device.PressureDriver,
	targetPressure float64,
	stableWaitMs int,
	stabilityTimeoutMs int,
) error {
	deadline := time.Now().Add(time.Duration(stabilityTimeoutMs) * time.Millisecond)
	stableDuration := time.Duration(stableWaitMs) * time.Millisecond
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	// 检查打压设备是否支持硬件级判稳（SCPI 设备优先使用设备返回的稳定标志）
	deviceStability, hasDeviceStability := pressureDriver.(device.StabilityStatusProvider)

	// 软件判稳：使用偏差计算
	monitor := workflow.NewStabilityMonitor(0.001, stableDuration, nil)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				// 超时：发布事件并等待前端用户决定
				s.publish(events.EventMeasurementStabilityTimeout, map[string]any{
					"pointIndex": pointIndex,
				})

				select {
				case decision := <-s.stabilityTimeoutCh:
					switch decision {
					case "continue":
						// 用户选择继续等待，重置超时倒计时
						deadline = time.Now().Add(time.Duration(stabilityTimeoutMs) * time.Millisecond)
						continue
					case "skip":
						return ErrPointSkipped
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			var status workflow.StabilityStatus
			if hasDeviceStability {
				// 设备判稳路径：SCPI 设备硬件自行判断压力稳定，软件仅依赖硬件 IsStable 标志。
				// 设备报告稳定时 FeedSample 偏差为 0（累积器继续计时）；
				// 设备报告不稳定时 FeedSample 大偏差（累积器重置）。
				stable, err := deviceStability.IsStable(ctx)
				if err != nil {
					continue
				}
				feedVal := targetPressure
				if !stable {
					feedVal = targetPressure + 1000
				}
				status = monitor.FeedSample(targetPressure, feedVal)
			} else {
				currentVal, valErr := pressureDriver.ReadCurrentPressure(ctx)
				if valErr != nil {
					continue
				}
				status = monitor.FeedSample(targetPressure, currentVal)
			}

			s.publish(events.EventMeasurementStabilityUpdate, map[string]any{
				"pointIndex":         pointIndex,
				"isStable":           status.IsStable,
				"isInRange":          status.IsInRange,
				"currentValue":       status.CurrentValue,
				"stableDurationMs":   status.StableDurationMs,
				"requiredDurationMs": status.RequiredDurationMs,
				"progress":           status.Progress,
			})

			if status.IsStable {
				return nil
			}
		}
	}
}
