package measurement

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
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

		// 打压成功后，采集可能因报警需要重新采集，仅重试采集不重新打压。
		if err := s.ManualPressurize(ctx, point.Index); err != nil {
			if errors.Is(err, ErrPointSkipped) {
				log.Printf("[measurement] point %d skipped by user", point.Index)
				continue
			}
			collectErr = err
			break
		}
		// 单点最多重采集 maxRecollectAttempts 次，避免用户连续选择造成无限循环。
		const maxRecollectAttempts = 10
		recollectAttempts := 0
	collectLoop:
		for {
			select {
			case <-ctx.Done():
				collectErr = ctx.Err()
				break pointsLoop
			default:
			}
			if err := s.ManualCollect(ctx, point.Index); err != nil {
				if errors.Is(err, ErrRecollectPoint) {
					recollectAttempts++
					if recollectAttempts > maxRecollectAttempts {
						collectErr = fmt.Errorf("point %d: exceeded max recollect attempts (%d)", point.Index, maxRecollectAttempts)
						break pointsLoop
					}
					log.Printf("[measurement] point %d recollect requested (attempt %d/%d), retrying collect only",
						point.Index, recollectAttempts, maxRecollectAttempts)
					// ManualCollect 在报警时已转到 await_alarm_resolution，
					// 再次调用 ManualCollect 可直接迁回 collecting。
					continue collectLoop
				}
				collectErr = err
				break pointsLoop
			}
			break collectLoop
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
		if err := s.coordinator.Machine().Transition(domain.SessionStateCompleted); err != nil {
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

	// 稳定等待：压力首次进入容差范围时才切换到 stabilizing，
	// 避免压力远离目标时就显示"稳定中"。
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
	measureDrivers, point, channels, averageCount, totalPoints, err := s.prepareCollectStep(pointIndex)
	if err != nil {
		return err
	}

	s.updatePointStatus(pointIndex, domain.PointStatusCollecting)
	if err := s.transitionTo(domain.SessionStateCollecting); err != nil {
		return err
	}

	// 单设备路径：保持原逻辑
	if len(measureDrivers) <= 1 {
		for _, measureDriver := range measureDrivers {
			flattened, err := s.collectSamplesFromDriver(ctx, measureDriver, channels, averageCount)
			if err != nil {
				return err
			}
			s.updatePointCollectedData(pointIndex, flattened, time.Now())
			s.publish(events.EventMeasurementDataCollected, map[string]any{
				"pointIndex": point.Index,
				"channels":   channels,
				"data":       flattened,
			})
		}
		return s.finalizePointCollect(ctx, pointIndex, point, totalPoints)
	}

	// 多设备路径：并行采集
	results := make(map[string][]float64, len(measureDrivers))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for devID, drv := range measureDrivers {
		wg.Add(1)
		go func(id string, d device.MeasureDriver) {
			defer wg.Done()
			devCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			flattened, err := s.collectSamplesFromDriver(devCtx, d, channels, averageCount)
			if err != nil {
				s.recordDeviceCollectError(pointIndex, id, err)
				return
			}
			mu.Lock()
			results[id] = flattened
			mu.Unlock()
		}(devID, drv)
	}
	wg.Wait()

	for devID, flattened := range results {
		s.updatePointCollectedDataForDevice(pointIndex, devID, flattened, time.Now())
		s.publish(events.EventMeasurementDataCollected, map[string]any{
			"pointIndex": point.Index,
			"channels":   channels,
			"deviceId":   devID,
			"data":       flattened,
		})
	}

	return s.finalizePointCollect(ctx, pointIndex, point, totalPoints)
}

// collectSamplesFromDriver 从单个计量驱动采集平均样本并平铺。
func (s *Service) collectSamplesFromDriver(ctx context.Context, measureDriver device.MeasureDriver, channels []int, averageCount int) ([]float64, error) {
	samples := make([][]float64, 0, averageCount)
	for i := 0; i < averageCount; i++ {
		data, err := measureDriver.CollectData(ctx, channels)
		if err != nil {
			return nil, fmt.Errorf("collect sample %d: %w", i+1, err)
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
	return flattened, nil
}

// finalizePointCollect 完成单点采集后的状态迁移与报警检查。
func (s *Service) finalizePointCollect(ctx context.Context, pointIndex int, point domain.PressurePoint, totalPoints int) error {
	// 采集后自动检查报警。
	updatedPoint := s.getPoint(pointIndex)
	if alarm, _ := s.CheckAlarm(updatedPoint); alarm != nil {
		s.publish(events.EventMeasurementAlarmTriggered, alarm)

		// 进入等待报警决策状态，便于 recollect 后能合法重新进入 collecting。
		if err := s.transitionTo(domain.SessionStateAwaitAlarmResolution); err != nil {
			return fmt.Errorf("transition to await_alarm_resolution: %w", err)
		}

		// 阻塞等待用户确认报警决定
		s.mu.Lock()
		s.alarmCh = make(chan string, 1)
		alarmCh := s.alarmCh
		s.mu.Unlock()

		select {
		case decision := <-alarmCh:
			s.mu.Lock()
			s.alarmPending = false
			s.alarmCh = nil
			s.mu.Unlock()

			switch decision {
			case workflow.AlarmDecisionStop:
				return fmt.Errorf("alarm: user stopped after point %d", pointIndex)
			case workflow.AlarmDecisionSkip:
				// 跳过，继续下一个点
			case workflow.AlarmDecisionRecollect:
				// 用户选择重新采集当前点：由上层 collectLoop 再次进入 ManualCollect，
				// 此时状态停留在 await_alarm_resolution，刚好可以合法迁移到 collecting。
				return ErrRecollectPoint
			default:
				// continue：继续流程
			}

			// 用户选择 continue 或 skip 时，需要先从 await_alarm_resolution
			// 迁回 collecting，否则下方 transitionTo(Ready/Completed) 会因为
			// await_alarm_resolution -> ready 非法而失败，导致自动采集中断。
			if err := s.transitionTo(domain.SessionStateCollecting); err != nil {
				return fmt.Errorf("transition back to collecting: %w", err)
			}
		case <-ctx.Done():
			s.mu.Lock()
			s.alarmPending = false
			s.alarmCh = nil
			s.mu.Unlock()
			return ctx.Err()
		}
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

func (s *Service) prepareCollectStep(pointIndex int) (map[string]device.MeasureDriver, domain.PressurePoint, []int, int, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	measureDrivers := s.sess.MeasureDrivers()
	// 过滤已跳过的设备
	for devID := range s.skippedDevices {
		delete(measureDrivers, devID)
	}
	if len(measureDrivers) == 0 {
		return nil, domain.PressurePoint{}, nil, 0, 0, session.ErrMeasureDeviceNotSet
	}
	if pointIndex < 1 || pointIndex > len(s.points) {
		return nil, domain.PressurePoint{}, nil, 0, 0, fmt.Errorf("invalid point index: %d", pointIndex)
	}

	// 始终采集全部16通道，通道选择仅用于报警判定
	channels := allChannels

	averageCount := s.config.AverageCount
	if averageCount < 1 {
		averageCount = 1
	}

	return measureDrivers, s.points[pointIndex-1], channels, averageCount, len(s.points), nil
}

func (s *Service) transitionTo(state domain.SessionState) error {
	s.mu.Lock()
	if err := s.coordinator.Machine().Transition(state); err != nil {
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

// updatePointCollectedDataForDevice 把指定计量设备的采集结果写入压力点的设备维度数据。
// 单设备场景（points 中仅一台设备绑定）同时回填 CollectedData 兼容旧字段。
func (s *Service) updatePointCollectedDataForDevice(pointIndex int, devID string, data []float64, collectedAt time.Time) {
	s.mu.Lock()
	if pointIndex < 1 || pointIndex > len(s.points) {
		s.mu.Unlock()
		return
	}
	timestamp := collectedAt.UTC().Format(time.RFC3339)
	clonedData := append([]float64(nil), data...)

	point := &s.points[pointIndex-1]
	if point.CollectedByDevice == nil {
		point.CollectedByDevice = make(map[string]domain.DevicePointData)
	}
	point.CollectedByDevice[devID] = domain.DevicePointData{
		DeviceID:    devID,
		Collected:   clonedData,
		Status:      domain.PointStatusCompleted,
		CollectTime: timestamp,
	}

	// 单设备兼容：若当前仅绑定一台计量设备，同时写入旧字段。
	if len(s.sess.MeasureDeviceIDs()) <= 1 {
		point.CollectedData = append([]float64(nil), clonedData...)
		point.CollectTime = timestamp
		point.Status = domain.PointStatusCompleted
	}

	if s.session != nil && pointIndex <= len(s.session.Points) {
		sp := &s.session.Points[pointIndex-1]
		if sp.CollectedByDevice == nil {
			sp.CollectedByDevice = make(map[string]domain.DevicePointData)
		}
		sp.CollectedByDevice[devID] = domain.DevicePointData{
			DeviceID:    devID,
			Collected:   append([]float64(nil), clonedData...),
			Status:      domain.PointStatusCompleted,
			CollectTime: timestamp,
		}
		if len(s.sess.MeasureDeviceIDs()) <= 1 {
			sp.CollectedData = append([]float64(nil), clonedData...)
			sp.CollectTime = timestamp
			sp.Status = domain.PointStatusCompleted
		}
	}

	snapshot := *point
	s.mu.Unlock()

	s.publish(events.EventMeasurementPointStatus, snapshot)
}

// recordDeviceCollectError 记录指定设备在压力点的采集失败状态。
func (s *Service) recordDeviceCollectError(pointIndex int, devID string, err error) {
	s.mu.Lock()
	if pointIndex < 1 || pointIndex > len(s.points) {
		s.mu.Unlock()
		return
	}
	point := &s.points[pointIndex-1]
	if point.CollectedByDevice == nil {
		point.CollectedByDevice = make(map[string]domain.DevicePointData)
	}
	point.CollectedByDevice[devID] = domain.DevicePointData{
		DeviceID: devID,
		Status:   domain.PointStatusError,
		Error:    err.Error(),
	}
	s.mu.Unlock()

	s.publish(events.EventPointError, map[string]any{
		"pointIndex": pointIndex,
		"deviceId":   devID,
		"error":      err.Error(),
	})
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

	// 标记是否已从 pressurizing 切换到 stabilizing
	transitionedToStabilizing := false

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

			// 压力首次进入容差范围时，才从 pressurizing 切换到 stabilizing
			if status.IsInRange && !transitionedToStabilizing {
				transitionedToStabilizing = true
				s.updatePointStatus(pointIndex, domain.PointStatusStabilizing)
				if err := s.transitionTo(domain.SessionStateStabilizing); err != nil {
					return err
				}
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
