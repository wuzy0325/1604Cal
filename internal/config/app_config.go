package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"cal1604/internal/application/deviceconnect"
)

// AppConfig 定义应用运行时配置。
type AppConfig struct {
	DeviceConnect DeviceConnectFileConfig `json:"deviceConnect"`
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
