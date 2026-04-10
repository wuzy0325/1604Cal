package http

import (
	"net/http"

	"cal1604/internal/application/deviceconnect"
	"cal1604/internal/device/manager"
	"cal1604/internal/infrastructure/driver"
	"cal1604/internal/workflow"
)

// NewRouterWithDeviceManager 基于指定设备管理器创建路由。
func NewRouterWithDeviceManager(deviceManager *manager.DeviceManager) http.Handler {
	return newRouter(deviceManager, nil, deviceconnect.DefaultConfig())
}

// NewRouterWithDependencies 基于指定依赖创建路由。
// 该方法用于生产注入与集成测试注入同一套 HTTP 处理逻辑。
func NewRouterWithDependencies(deviceManager *manager.DeviceManager, connector deviceConnector) http.Handler {
	return newRouter(deviceManager, connector, deviceconnect.DefaultConfig())
}

// NewRouterWithConnectConfig 基于指定连接配置创建路由。
func NewRouterWithConnectConfig(deviceManager *manager.DeviceManager, connectConfig deviceconnect.Config) http.Handler {
	return newRouter(deviceManager, nil, connectConfig)
}

func newRouter(deviceManager *manager.DeviceManager, connector deviceConnector, connectConfig deviceconnect.Config) http.Handler {
	if deviceManager == nil {
		deviceManager = manager.NewDeviceManager()
	}

	server := &apiServer{deviceManager: deviceManager, connectConfig: connectConfig}
	server.sessionMachine = workflow.NewSessionMachine()
	if connector == nil {
		server.deviceConnector = deviceconnect.NewService(
			deviceManager,
			driver.NewFactory(),
			connectConfig,
			server.publishDeviceStatusChanged,
		)
	} else {
		server.deviceConnector = connector
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/health", healthHandler)
	mux.HandleFunc("/api/v1/events/stream", eventsStreamHandler)
	mux.HandleFunc("/api/v1/config/device-connect", server.deviceConnectConfigHandler)
	mux.HandleFunc("/api/v1/devices", server.devicesHandler)
	mux.HandleFunc("/api/v1/devices/status", server.deviceStatusHandler)
	mux.HandleFunc("/api/v1/devices/connect", server.deviceConnectHandler)
	mux.HandleFunc("/api/v1/devices/disconnect", server.deviceDisconnectHandler)
	mux.HandleFunc("/api/v1/checks/unit-consistency", server.unitConsistencyHandler)
	mux.HandleFunc("/api/v1/sessions/state", server.sessionStateHandler)
	mux.HandleFunc("/api/v1/sessions/start", server.sessionStartHandler)
	mux.HandleFunc("/api/v1/sessions/pause", server.sessionPauseHandler)
	mux.HandleFunc("/api/v1/sessions/resume", server.sessionResumeHandler)
	mux.HandleFunc("/api/v1/sessions/stop", server.sessionStopHandler)
	mux.HandleFunc("/api/v1/reports/templates/select", server.reportTemplateSelectHandler)

	return mux
}

// NewRouter 返回 API 路由，用于注册系统对外提供的 HTTP 端点。
func NewRouter() http.Handler {
	return NewRouterWithDeviceManager(manager.NewDeviceManager())
}
