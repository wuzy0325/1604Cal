package measurement

import (
	"context"
	"fmt"
	"math"
	"time"

	"cal1604/internal/application/session"
	"cal1604/internal/device"
)

// RunAutoCollection 按测点顺序执行 measurement 自己的自动采集流程。
func (s *Service) RunAutoCollection(ctx context.Context) error {
	points := s.GetPoints()
	for _, point := range points {
		if err := s.ManualPressurize(ctx, point.Index); err != nil {
			return err
		}
		if err := s.ManualCollect(ctx, point.Index); err != nil {
			return err
		}
	}

	if s.State() != StateCompleted {
		s.mu.Lock()
		if err := s.setStateLocked(StateCompleted); err != nil {
			s.mu.Unlock()
			return fmt.Errorf("complete auto collection: %w", err)
		}
		s.syncSessionStatusLocked(StateCompleted)
		s.mu.Unlock()

		s.publish("measurement.state_changed", map[string]any{"state": string(StateCompleted)})
	}
	return nil
}

// ManualPressurize 对指定 measurement 点执行打压和稳定等待。
func (s *Service) ManualPressurize(ctx context.Context, pointIndex int) error {
	pressureDriver, point, stableWaitMs, err := s.preparePressureStep(pointIndex)
	if err != nil {
		return err
	}

	s.updatePointStatus(pointIndex, "pressurizing")
	if err := s.transitionTo(StatePressuring); err != nil {
		return err
	}

	if err := pressureDriver.SetTargetPressure(ctx, point.TargetPressure); err != nil {
		return fmt.Errorf("set target pressure: %w", err)
	}
	if ctrl, ok := pressureDriver.(interface{ StartControl(context.Context) error }); ok {
		if err := ctrl.StartControl(ctx); err != nil {
			return fmt.Errorf("start pressure control: %w", err)
		}
	}

	s.updatePointStatus(pointIndex, "stabilizing")
	if err := s.transitionTo(StateStabilizing); err != nil {
		return err
	}

	if err := s.waitForMeasurementStability(ctx, pointIndex, pressureDriver, point.TargetPressure, stableWaitMs); err != nil {
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

	s.updatePointStatus(pointIndex, "collecting")
	if err := s.transitionTo(StateCollecting); err != nil {
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

	averaged := averageMeasurementSamples(samples)
	s.updatePointCollectedData(pointIndex, averaged, time.Now())
	s.publish("measurement.data.collected", map[string]any{
		"pointIndex": point.Index,
		"channels":   channels,
		"data":       averaged,
	})

	nextState := StateReady
	if pointIndex >= totalPoints {
		nextState = StateCompleted
	}
	if err := s.transitionTo(nextState); err != nil {
		return err
	}

	return nil
}

func (s *Service) preparePressureStep(pointIndex int) (device.PressureDriver, Point, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pressureDriver := s.sess.PressureDriver()
	if pressureDriver == nil {
		return nil, Point{}, 0, session.ErrPressureDeviceNotSet
	}
	if pointIndex < 1 || pointIndex > len(s.points) {
		return nil, Point{}, 0, fmt.Errorf("invalid point index: %d", pointIndex)
	}

	stableWaitMs := s.config.StableWaitMs
	if stableWaitMs <= 0 {
		stableWaitMs = 2000
	}

	return pressureDriver, s.points[pointIndex-1], stableWaitMs, nil
}

func (s *Service) prepareCollectStep(pointIndex int) (device.MeasureDriver, Point, []int, int, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	measureDriver := s.sess.MeasureDriver()
	if measureDriver == nil {
		return nil, Point{}, nil, 0, 0, session.ErrMeasureDeviceNotSet
	}
	if pointIndex < 1 || pointIndex > len(s.points) {
		return nil, Point{}, nil, 0, 0, fmt.Errorf("invalid point index: %d", pointIndex)
	}

	channels := append([]int(nil), s.channels...)
	if len(channels) == 0 {
		channels = append([]int(nil), s.config.Channels...)
	}
	if len(channels) == 0 {
		return nil, Point{}, nil, 0, 0, fmt.Errorf("no measurement channels configured")
	}

	averageCount := s.config.AverageCount
	if averageCount < 1 {
		averageCount = 1
	}

	return measureDriver, s.points[pointIndex-1], channels, averageCount, len(s.points), nil
}

func (s *Service) transitionTo(state State) error {
	s.mu.Lock()
	if err := s.setStateLocked(state); err != nil {
		s.mu.Unlock()
		return fmt.Errorf("transition to %s: %w", state, err)
	}
	s.syncSessionStatusLocked(state)
	s.mu.Unlock()

	s.publish("measurement.state_changed", map[string]any{"state": string(state)})
	return nil
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

	s.publish("measurement.point.status", point)
}

func (s *Service) updatePointActualPressure(pointIndex int, actualPressure float64) {
	s.mu.Lock()
	if pointIndex < 1 || pointIndex > len(s.points) {
		s.mu.Unlock()
		return
	}
	s.points[pointIndex-1].ActualPressure = actualPressure
	if s.session != nil && pointIndex <= len(s.session.Points) {
		s.session.Points[pointIndex-1].ActualPressure = actualPressure
	}
	point := s.points[pointIndex-1]
	s.mu.Unlock()

	s.publish("measurement.point.status", point)
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
	s.points[pointIndex-1].Status = "completed"
	if s.session != nil && pointIndex <= len(s.session.Points) {
		s.session.Points[pointIndex-1].CollectedData = append([]float64(nil), data...)
		s.session.Points[pointIndex-1].CollectTime = timestamp
		s.session.Points[pointIndex-1].Status = "completed"
	}
	point := s.points[pointIndex-1]
	s.mu.Unlock()

	s.publish("measurement.point.status", point)
}

// waitForMeasurementStability 等待压力稳定，期间通过 SSE 推送稳定进度。
func (s *Service) waitForMeasurementStability(
	ctx context.Context,
	pointIndex int,
	pressureDriver device.PressureDriver,
	targetPressure float64,
	stableWaitMs int,
) error {
	deadline := time.Now().Add(60 * time.Second)
	stableStart := time.Time{}
	wasInRange := false

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				s.publishStabilityUpdate(pointIndex, false, false, 0, 0, stableWaitMs, 0)
				return fmt.Errorf("stability timeout")
			}

			currentVal, valErr := pressureDriver.ReadCurrentPressure(ctx)
			isInRange := false
			if valErr == nil {
				stable, err := pressureDriver.ReadStability(ctx)
				if err == nil {
					isInRange = stable
				} else {
					deviation := math.Abs(currentVal - targetPressure)
					isInRange = deviation <= 0.001
				}
			}

			if isInRange {
				if !wasInRange {
					stableStart = time.Now()
					wasInRange = true
					s.publishStabilityUpdate(pointIndex, false, true, currentVal, 0, stableWaitMs, 0)
				} else {
					elapsed := int(time.Since(stableStart).Milliseconds())
					if elapsed >= stableWaitMs {
						s.publishStabilityUpdate(pointIndex, true, true, currentVal, elapsed, stableWaitMs, 100)
						return nil
					}
					progress := elapsed * 100 / stableWaitMs
					if progress > 100 {
						progress = 100
					}
					s.publishStabilityUpdate(pointIndex, false, true, currentVal, elapsed, stableWaitMs, progress)
				}
			} else {
				if wasInRange {
					wasInRange = false
					s.publishStabilityUpdate(pointIndex, false, false, currentVal, 0, stableWaitMs, 0)
				}
				stableStart = time.Time{}
			}
		}
	}
}

// publishStabilityUpdate 广播稳定进度 SSE 事件。
func (s *Service) publishStabilityUpdate(
	pointIndex int,
	isStable bool,
	isInRange bool,
	currentValue float64,
	stableDuration int,
	requiredDuration int,
	progress int,
) {
	s.publish("measurement.stability.update", map[string]any{
		"pointIndex":       pointIndex,
		"isStable":         isStable,
		"isInRange":        isInRange,
		"currentValue":     currentValue,
		"stableDurationMs": stableDuration,
		"requiredDurationMs": requiredDuration,
		"progress":         progress,
	})
}

func averageMeasurementSamples(samples [][]float64) []float64 {
	if len(samples) == 0 {
		return nil
	}
	width := len(samples[0])
	result := make([]float64, width)
	for _, sample := range samples {
		for i := 0; i < len(sample) && i < width; i++ {
			result[i] += sample[i]
		}
	}
	for i := range result {
		result[i] /= float64(len(samples))
	}
	return result
}
