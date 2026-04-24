package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"cal1604/internal/application/deviceconnect"
	"cal1604/internal/domain"
)

// AppConfig 定义应用运行时配置。
type AppConfig struct {
	DeviceConnect     DeviceConnectFileConfig `json:"deviceConnect"`
	Calibration       CalibrationFileConfig   `json:"calibration"`
	CalibrationParams CalibrationParamsConfig `json:"calibrationParams"`
	MeasurementParams MeasurementParamsConfig `json:"measurementParams"`
	Alarm             domain.AlarmConfig      `json:"alarm"`
	LastDevices       LastDevicesConfig       `json:"lastDevices"`
}

// DeviceConnectFileConfig 定义设备连接可靠性参数（单位：毫秒）。
type DeviceConnectFileConfig struct {
	ConnectAttemptTimeoutMs    int `json:"connectAttemptTimeoutMs"`
	ConnectMaxAttempts         int `json:"connectMaxAttempts"`
	ConnectInitialBackoffMs    int `json:"connectInitialBackoffMs"`
	ConnectMaxBackoffMs        int `json:"connectMaxBackoffMs"`
	DisconnectAttemptTimeoutMs int `json:"disconnectAttemptTimeoutMs"`
	DisconnectMaxAttempts      int `json:"disconnectMaxAttempts"`
	DisconnectInitialBackoffMs int `json:"disconnectInitialBackoffMs"`
	DisconnectMaxBackoffMs     int `json:"disconnectMaxBackoffMs"`
}

// CalibrationFileConfig 定义标定流程运行开关。
type CalibrationFileConfig struct {
	// EnforceValveCalibrationGate 为 true 时，开始标定前必须校验阀门=calibration。
	EnforceValveCalibrationGate bool `json:"enforceValveCalibrationGate"`
}

// CalibrationParamsConfig 校准参数持久化配置。
type CalibrationParamsConfig struct {
	MinPressure      float64 `json:"minPressure"`
	MaxPressure      float64 `json:"maxPressure"`
	PointCount       int     `json:"pointCount"`
	Precision        int     `json:"precision"`
	AverageCount     int     `json:"averageCount"`
	StableDurationMs int     `json:"stableDurationMs"`
	PrecisionLevel   float64 `json:"precisionLevel"`
	PressureMode     string  `json:"pressureMode"`
	ControlMode      string  `json:"controlMode"`
}

// MeasurementParamsConfig 计量模块参数持久化配置。
type MeasurementParamsConfig struct {
	MinPressure      float64 `json:"minPressure"`
	MaxPressure      float64 `json:"maxPressure"`
	PointCount       int     `json:"pointCount"`
	Precision        int     `json:"precision"`
	AverageCount     int     `json:"averageCount"`
	StableDurationMs int     `json:"stableDurationMs"`
	PrecisionLevel   float64 `json:"precisionLevel"`
	PressureMode     string  `json:"pressureMode"`
	ControlMode      string  `json:"controlMode"`
}

// LastDevicesConfig 记录上次使用的设备 ID，用于页面加载时自动绑定。
type LastDevicesConfig struct {
	PressureDeviceID string   `json:"pressureDeviceId"`
	MeasureDeviceIDs []string `json:"measureDeviceIds"`
}

// Default 返回默认配置。
func Default() AppConfig {
	defaults := deviceconnect.DefaultConfig()
	return AppConfig{
		DeviceConnect: DeviceConnectFileConfig{
			ConnectAttemptTimeoutMs:    int(defaults.ConnectAttemptTimeout / time.Millisecond),
			ConnectMaxAttempts:         defaults.ConnectMaxAttempts,
			ConnectInitialBackoffMs:    int(defaults.ConnectInitialBackoff / time.Millisecond),
			ConnectMaxBackoffMs:        int(defaults.ConnectMaxBackoff / time.Millisecond),
			DisconnectAttemptTimeoutMs: int(defaults.DisconnectAttemptTimeout / time.Millisecond),
			DisconnectMaxAttempts:      defaults.DisconnectMaxAttempts,
			DisconnectInitialBackoffMs: int(defaults.DisconnectInitialBackoff / time.Millisecond),
			DisconnectMaxBackoffMs:     int(defaults.DisconnectMaxBackoff / time.Millisecond),
		},
		Calibration: CalibrationFileConfig{
			EnforceValveCalibrationGate: false,
		},
		CalibrationParams: CalibrationParamsConfig{
			MinPressure:      0,
			MaxPressure:      100,
			PointCount:       5,
			Precision:        2,
			AverageCount:     1,
			StableDurationMs: 5000,
			PrecisionLevel:   0.0005,
			PressureMode:     "single",
			ControlMode:      "auto",
		},
		MeasurementParams: MeasurementParamsConfig{
			MinPressure:      0,
			MaxPressure:      100,
			PointCount:       6,
			Precision:        2,
			AverageCount:     1,
			StableDurationMs: 5000,
			PrecisionLevel:   0.05,
			PressureMode:     "single",
			ControlMode:      "auto",
		},
		Alarm: domain.AlarmConfig{
			Enabled:            true,
			PrecisionThreshold: 5.0,
			SoundEnabled:       true,
			ConfirmOnAlarm:     true,
			EnabledChannels:    []int{},
		},
		LastDevices: LastDevicesConfig{
			PressureDeviceID: "",
			MeasureDeviceIDs: []string{},
		},
	}
}

// LoadFromFile 从 JSON 文件读取应用配置。
func LoadFromFile(path string) (AppConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return AppConfig{}, fmt.Errorf("read config file %s: %w", path, err)
	}

	config := Default()
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return AppConfig{}, fmt.Errorf("decode config file %s: %w", path, err)
	}

	return config, nil
}

// SaveToFile 将当前配置写入 JSON 文件。
func (c AppConfig) SaveToFile(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config file %s: %w", path, err)
	}
	return nil
}

// ToDeviceConnectConfig 将文件配置转换为连接服务运行配置。
func (c AppConfig) ToDeviceConnectConfig() deviceconnect.Config {
	defaults := deviceconnect.DefaultConfig()

	cfg := defaults

	if c.DeviceConnect.ConnectAttemptTimeoutMs > 0 {
		cfg.ConnectAttemptTimeout = time.Duration(c.DeviceConnect.ConnectAttemptTimeoutMs) * time.Millisecond
	}
	if c.DeviceConnect.ConnectMaxAttempts > 0 {
		cfg.ConnectMaxAttempts = c.DeviceConnect.ConnectMaxAttempts
	}
	if c.DeviceConnect.ConnectInitialBackoffMs >= 0 {
		cfg.ConnectInitialBackoff = time.Duration(c.DeviceConnect.ConnectInitialBackoffMs) * time.Millisecond
	}
	if c.DeviceConnect.ConnectMaxBackoffMs > 0 {
		cfg.ConnectMaxBackoff = time.Duration(c.DeviceConnect.ConnectMaxBackoffMs) * time.Millisecond
	}

	if c.DeviceConnect.DisconnectAttemptTimeoutMs > 0 {
		cfg.DisconnectAttemptTimeout = time.Duration(c.DeviceConnect.DisconnectAttemptTimeoutMs) * time.Millisecond
	}
	if c.DeviceConnect.DisconnectMaxAttempts > 0 {
		cfg.DisconnectMaxAttempts = c.DeviceConnect.DisconnectMaxAttempts
	}
	if c.DeviceConnect.DisconnectInitialBackoffMs >= 0 {
		cfg.DisconnectInitialBackoff = time.Duration(c.DeviceConnect.DisconnectInitialBackoffMs) * time.Millisecond
	}
	if c.DeviceConnect.DisconnectMaxBackoffMs > 0 {
		cfg.DisconnectMaxBackoff = time.Duration(c.DeviceConnect.DisconnectMaxBackoffMs) * time.Millisecond
	}

	return cfg
}
