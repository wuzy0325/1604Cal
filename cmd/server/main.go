package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	apihttp "cal1604/internal/api/http"
	"cal1604/internal/application/deviceconnect"
	"cal1604/internal/config"
	"cal1604/internal/device/manager"
)

const configPathEnvName = "CAL1604_CONFIG"

func main() {
	addr := ":18080"

	connectCfg, err := resolveConnectConfig(os.Getenv)
	if err != nil {
		log.Fatalf("load runtime config failed: %v", err)
	}

	// 使用持久化设备管理器，设备配置会自动保存到本地文件
	deviceManager, err := manager.NewPersistentDeviceManager(manager.StorageConfig{})
	if err != nil {
		log.Fatalf("init persistent device manager failed: %v", err)
	}

	router := apihttp.NewRouterWithConnectConfig(deviceManager, connectCfg)

	log.Printf("server listening on %s", addr)
	if err = http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server exit: %v", err)
	}
}

// resolveConnectConfig 根据环境变量解析连接可靠性配置。
// 当未配置文件路径时，返回内置默认值。
func resolveConnectConfig(getenv func(string) string) (deviceconnect.Config, error) {
	path := strings.TrimSpace(getenv(configPathEnvName))
	if path == "" {
		return config.Default().ToDeviceConnectConfig(), nil
	}

	runtimeConfig, err := config.LoadFromFile(path)
	if err != nil {
		return deviceconnect.Config{}, err
	}

	return runtimeConfig.ToDeviceConnectConfig(), nil
}
