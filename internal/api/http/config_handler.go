package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"cal1604/internal/application/calibration"
	"cal1604/internal/application/measurement"
	"cal1604/internal/config"
	"cal1604/internal/domain"
	apperrors "cal1604/internal/errors"
)

type deviceConnectConfigPayload struct {
	ConnectAttemptTimeoutMs    int `json:"connectAttemptTimeoutMs"`
	ConnectMaxAttempts         int `json:"connectMaxAttempts"`
	ConnectInitialBackoffMs    int `json:"connectInitialBackoffMs"`
	ConnectMaxBackoffMs        int `json:"connectMaxBackoffMs"`
	DisconnectAttemptTimeoutMs int `json:"disconnectAttemptTimeoutMs"`
	DisconnectMaxAttempts      int `json:"disconnectMaxAttempts"`
	DisconnectInitialBackoffMs int `json:"disconnectInitialBackoffMs"`
	DisconnectMaxBackoffMs     int `json:"disconnectMaxBackoffMs"`
}

// deviceConnectConfigHandler 返回当前生效的连接可靠性配置。
// 该接口用于前端设备面板可视化 timeout/retry 策略，便于现场排障与参数核对。
func (s *apiServer) deviceConnectConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	payload := deviceConnectConfigPayload{
		ConnectAttemptTimeoutMs:    durationToMilliseconds(s.connectConfig.ConnectAttemptTimeout),
		ConnectMaxAttempts:         s.connectConfig.ConnectMaxAttempts,
		ConnectInitialBackoffMs:    durationToMilliseconds(s.connectConfig.ConnectInitialBackoff),
		ConnectMaxBackoffMs:        durationToMilliseconds(s.connectConfig.ConnectMaxBackoff),
		DisconnectAttemptTimeoutMs: durationToMilliseconds(s.connectConfig.DisconnectAttemptTimeout),
		DisconnectMaxAttempts:      s.connectConfig.DisconnectMaxAttempts,
		DisconnectInitialBackoffMs: durationToMilliseconds(s.connectConfig.DisconnectInitialBackoff),
		DisconnectMaxBackoffMs:     durationToMilliseconds(s.connectConfig.DisconnectMaxBackoff),
	}

	writeSuccess(w, http.StatusOK, payload)
}

func durationToMilliseconds(value time.Duration) int {
	if value <= 0 {
		return 0
	}
	return int(value / time.Millisecond)
}

// calibrationConfigHandler 处理校准参数配置的读取和更新。
// GET 返回当前校准参数，POST 更新参数并持久化到配置文件。
func (s *apiServer) calibrationConfigHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if s.appConfig != nil {
			writeSuccess(w, http.StatusOK, s.appConfig.CalibrationParams)
			return
		}
		writeSuccess(w, http.StatusOK, config.Default().CalibrationParams)

	case http.MethodPost:
		var params config.CalibrationParamsConfig
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			writeError(w, apperrors.ErrInvalidArgument)
			return
		}
		if s.appConfig != nil {
			s.appConfig.CalibrationParams = params
			s.persistConfig()
		}
		// 同步到校准服务
		s.calibrationService.SetConfig(calibrationConfigFromParams(params))
		writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// measurementConfigHandler 处理计量参数配置的读取和更新。
// GET 返回当前计量参数，POST 更新参数并持久化到配置文件。
func (s *apiServer) measurementConfigHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if s.appConfig != nil {
			writeSuccess(w, http.StatusOK, s.appConfig.MeasurementParams)
			return
		}
		writeSuccess(w, http.StatusOK, config.Default().MeasurementParams)

	case http.MethodPost:
		var params config.MeasurementParamsConfig
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			writeError(w, apperrors.ErrInvalidArgument)
			return
		}
		if err := validateMeasurementParams(params); err != nil {
			writeError(w, err)
			return
		}

		if s.appConfig != nil {
			s.appConfig.MeasurementParams = params
			s.persistConfig()
		}

		s.measurementService.SetConfig(measurementConfigFromParams(params))

		writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// alarmConfigHandler 处理报警配置的读取和更新。
// GET 返回当前报警配置，POST 更新配置并持久化到配置文件。
func (s *apiServer) alarmConfigHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if s.appConfig != nil {
			writeSuccess(w, http.StatusOK, s.appConfig.Alarm)
			return
		}
		writeSuccess(w, http.StatusOK, config.Default().Alarm)

	case http.MethodPost:
		var alarmCfg struct {
			Enabled            bool    `json:"enabled"`
			PrecisionThreshold float64 `json:"precisionThreshold"`
			SoundEnabled       bool    `json:"soundEnabled"`
			ConfirmOnAlarm     bool    `json:"confirmOnAlarm"`
			EnabledChannels    []int   `json:"enabledChannels"`
		}
		if err := json.NewDecoder(r.Body).Decode(&alarmCfg); err != nil {
			writeError(w, apperrors.ErrInvalidArgument)
			return
		}
		cfg := domain.AlarmConfig{
			Enabled:            alarmCfg.Enabled,
			PrecisionThreshold: alarmCfg.PrecisionThreshold,
			SoundEnabled:       alarmCfg.SoundEnabled,
			ConfirmOnAlarm:     alarmCfg.ConfirmOnAlarm,
			EnabledChannels:    alarmCfg.EnabledChannels,
		}
		if s.appConfig != nil {
			s.appConfig.Alarm = cfg
			s.persistConfig()
		}
		// 同步到校准服务
		s.calibrationService.SetAlarmConfig(cfg)
		writeSuccess(w, http.StatusOK, map[string]string{"status": "ok"})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// persistConfig 将当前配置持久化到文件（如果配置了路径）。
func (s *apiServer) persistConfig() {
	if s.appConfig != nil && s.configPath != "" {
		_ = s.appConfig.SaveToFile(s.configPath)
	}
}

// calibrationConfigFromParams 将持久化参数转换为校准服务配置。
func calibrationConfigFromParams(params config.CalibrationParamsConfig) calibration.CalibrationConfig {
	return calibration.CalibrationConfig{
		MinPressure:    params.MinPressure,
		MaxPressure:    params.MaxPressure,
		PressurePoints: params.PointCount,
		Precision:      params.Precision,
		AverageCount:   params.AverageCount,
		StableWaitMs:   params.StableDurationMs,
		PrecisionLevel: params.PrecisionLevel,
		PressureMode:   params.PressureMode,
		ControlMode:    params.ControlMode,
	}
}

// measurementConfigFromParams 将持久化参数转换为计量服务配置。
func measurementConfigFromParams(params config.MeasurementParamsConfig) measurement.Config {
	return measurement.Config{
		MinPressure:    params.MinPressure,
		MaxPressure:    params.MaxPressure,
		PointCount:     params.PointCount,
		Precision:      params.Precision,
		AverageCount:   params.AverageCount,
		StableWaitMs:   params.StableDurationMs,
		PrecisionLevel: params.PrecisionLevel,
		PressureMode:   params.PressureMode,
		ControlMode:    params.ControlMode,
	}
}

func validateMeasurementParams(params config.MeasurementParamsConfig) error {
	if params.PointCount < 2 {
		return fmt.Errorf("%w: pointCount must be at least 2", apperrors.ErrInvalidArgument)
	}
	if params.Precision < 0 {
		return fmt.Errorf("%w: precision must be non-negative", apperrors.ErrInvalidArgument)
	}
	if params.Precision > 6 {
		return fmt.Errorf("%w: precision must be at most 6", apperrors.ErrInvalidArgument)
	}
	if params.AverageCount < 1 {
		return fmt.Errorf("%w: averageCount must be at least 1", apperrors.ErrInvalidArgument)
	}
	if params.StableDurationMs < 0 {
		return fmt.Errorf("%w: stableDurationMs must be non-negative", apperrors.ErrInvalidArgument)
	}
	if params.PressureMode != "single" && params.PressureMode != "roundTrip" {
		return fmt.Errorf("%w: pressureMode must be single or roundTrip", apperrors.ErrInvalidArgument)
	}
	if params.ControlMode != "auto" && params.ControlMode != "manual" {
		return fmt.Errorf("%w: controlMode must be auto or manual", apperrors.ErrInvalidArgument)
	}
	return nil
}
