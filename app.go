package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	apihttp "cal1604/internal/api/http"
	"cal1604/internal/application/deviceconnect"
	"cal1604/internal/config"
	"cal1604/internal/device/manager"
)

const configPathEnvName = "CAL1604_CONFIG"

// App 是 Wails 桌面应用的核心结构体，持有内嵌 HTTP 服务器生命周期状态。
type App struct {
	ctx    context.Context
	server *http.Server
	port   int
}

// NewApp 创建 App 实例。
func NewApp() *App {
	return &App{}
}

// startup 在 Wails 窗口创建后被调用。
// 启动内嵌 HTTP 服务器，监听 127.0.0.1 上的随机可用端口。
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	runtimeCfg, err := resolveRuntimeConfig(os.Getenv)
	if err != nil {
		log.Fatalf("load runtime config failed: %v", err)
	}
	connectCfg := runtimeCfg.ToDeviceConnectConfig()
	calibrationCfg := apihttp.CalibrationRuntimeConfig{
		EnforceValveCalibrationGate: runtimeCfg.Calibration.EnforceValveCalibrationGate,
	}

	// 使用持久化设备管理器，设备配置会自动保存到本地文件
	deviceManager, err := manager.NewPersistentDeviceManager(manager.StorageConfig{})
	if err != nil {
		log.Fatalf("init persistent device manager failed: %v", err)
	}

	router := apihttp.NewRouterWithRuntimeConfig(deviceManager, connectCfg, calibrationCfg)

	// 为桌面环境添加 CORS 支持。
	// Wails webview 使用 wails:// 协议加载前端页面，
	// 向 http://127.0.0.1 发起的 API 请求需要 CORS 头。
	corsHandler := withCORS(router)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("failed to allocate local port: %v", err)
	}

	a.port = listener.Addr().(*net.TCPAddr).Port
	a.server = &http.Server{Handler: corsHandler}

	go func() {
		if err := a.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("internal API server error: %v", err)
		}
	}()

	log.Printf("internal API server started on 127.0.0.1:%d", a.port)
}

// shutdown 在应用退出时被调用，优雅关闭内嵌 HTTP 服务器。
func (a *App) shutdown(ctx context.Context) {
	if a.server != nil {
		_ = a.server.Shutdown(ctx)
	}
}

// GetAPIPort 返回内嵌 HTTP 服务器端口，供前端通过 Wails 绑定调用。
func (a *App) GetAPIPort() int {
	return a.port
}

// withCORS 包装 handler，添加宽松的 CORS 响应头。
// 仅用于桌面环境：API 服务器监听 127.0.0.1，不暴露到网络。
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// resolveRuntimeConfig 根据环境变量解析运行时配置。
// 当未配置文件路径时，返回内置默认值。
func resolveRuntimeConfig(getenv func(string) string) (config.AppConfig, error) {
	path := strings.TrimSpace(getenv(configPathEnvName))
	if path == "" {
		return config.Default(), nil
	}

	return config.LoadFromFile(path)
}

// resolveConnectConfig 根据环境变量解析连接可靠性配置。
// 当未配置文件路径时，返回内置默认值。
// 兼容已有调用方。
func resolveConnectConfig(getenv func(string) string) (deviceconnect.Config, error) {
	runtimeCfg, err := resolveRuntimeConfig(getenv)
	if err != nil {
		return deviceconnect.Config{}, err
	}
	return runtimeCfg.ToDeviceConnectConfig(), nil
}
