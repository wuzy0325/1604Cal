package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	apihttp "cal1604/internal/api/http"
	"cal1604/internal/application/deviceconnect"
	"cal1604/internal/config"
)

const configPathEnvName = "CAL1604_CONFIG"

func main() {
	addr := ":8080"

	connectCfg, err := resolveConnectConfig(os.Getenv)
	if err != nil {
		log.Fatalf("load runtime config failed: %v", err)
	}

	router := apihttp.NewRouterWithConnectConfig(nil, connectCfg)

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
