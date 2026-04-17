package http

import (
	"net/http"

	"cal1604/internal/application/calibration"
	"cal1604/internal/application/deviceconnect"
	"cal1604/internal/application/multipress"
	"cal1604/internal/device"
	"cal1604/internal/device/manager"
	"cal1604/internal/infrastructure/driver"
	"cal1604/internal/workflow"
)

// NewRouterWithDeviceManager 基于指定设备管理器创建路由。
func NewRouterWithDeviceManager(deviceManager deviceManager) http.Handler {
	return newRouter(deviceManager, nil, deviceconnect.DefaultConfig())
}

// NewRouterWithDependencies 基于指定依赖创建路由。
func (s *apiServer) publishEventAdapter(eventType string, data any) {
	publishEvent(eventType, data)
}

// NewRouterWithDependencies 基于指定依赖创建路由。
// 该方法用于生产注入与集成测试注入同一套 HTTP 处理逻辑。
func NewRouterWithDependencies(deviceManager deviceManager, connector deviceConnector) http.Handler {
	return newRouter(deviceManager, connector, deviceconnect.DefaultConfig())
}

// NewRouterWithConnectConfig 基于指定连接配置创建路由。
func NewRouterWithConnectConfig(deviceManager deviceManager, connectConfig deviceconnect.Config) http.Handler {
	return newRouter(deviceManager, nil, connectConfig)
}

func newRouter(deviceManager deviceManager, connector deviceConnector, connectConfig deviceconnect.Config) http.Handler {
	if deviceManager == nil {
		deviceManager = manager.NewDeviceManager()
	}

	sessionMachine := workflow.NewSessionMachine()
	factory := driver.NewFactory()

	server := &apiServer{
		deviceManager:      deviceManager,
		sessionMachine:     sessionMachine,
		connectConfig:      connectConfig,
		calibrationService: calibration.NewService(sessionMachine, factory, deviceManager, nil, nil),
	}

	if connector == nil {
		server.deviceConnector = deviceconnect.NewService(
			deviceManager,
			factory,
			connectConfig,
			server.publishDeviceStatusChanged,
		)
	} else {
		server.deviceConnector = connector
	}

	// 从 connector 解析驱动提供者（复用已连接驱动）
	var driverProvider device.ActiveDriverProvider
	if dp, ok := server.deviceConnector.(device.ActiveDriverProvider); ok {
		driverProvider = dp
	}

	// 注入事件发布与驱动提供者到校准服务
	server.calibrationService = calibration.NewService(
		sessionMachine,
		factory,
		deviceManager,
		server.publishEventAdapter,
		driverProvider,
	)

	// 多设备打压控制服务
	server.multipressService = multipress.NewService(
		factory,
		deviceManager,
		server.publishEventAdapter,
	)
	server.multipressService.StartPolling()

	mux := http.NewServeMux()

	// 健康检查与事件流
	mux.HandleFunc("/api/v1/health", healthHandler)
	mux.HandleFunc("/api/v1/events/stream", eventsStreamHandler)

	// 配置
	mux.HandleFunc("/api/v1/config/device-connect", server.deviceConnectConfigHandler)

	// 设备管理
	mux.HandleFunc("/api/v1/devices", server.devicesHandler)
	mux.HandleFunc("/api/v1/devices/status", server.deviceStatusHandler)
	mux.HandleFunc("/api/v1/devices/connect", server.deviceConnectHandler)
	mux.HandleFunc("/api/v1/devices/disconnect", server.deviceDisconnectHandler)
	mux.HandleFunc("/api/v1/checks/unit-consistency", server.unitConsistencyHandler)

	// 会话控制
	mux.HandleFunc("/api/v1/sessions/state", server.sessionStateHandler)
	mux.HandleFunc("/api/v1/sessions/start", server.sessionStartHandler)
	mux.HandleFunc("/api/v1/sessions/pause", server.sessionPauseHandler)
	mux.HandleFunc("/api/v1/sessions/resume", server.sessionResumeHandler)
	mux.HandleFunc("/api/v1/sessions/stop", server.sessionStopHandler)

	// 校准流程
	mux.HandleFunc("/api/v1/calibration/devices", server.calibrationSetDevicesHandler)
	mux.HandleFunc("/api/v1/calibration/measure-device", server.calibrationSetMeasureDeviceHandler)
	mux.HandleFunc("/api/v1/calibration/config", server.calibrationSetConfigHandler)
	mux.HandleFunc("/api/v1/calibration/channels", server.calibrationSetChannelsHandler)
	mux.HandleFunc("/api/v1/calibration/channels/list", server.calibrationGetChannelsHandler)
	mux.HandleFunc("/api/v1/calibration/points/generate", server.calibrationGeneratePointsHandler)
	mux.HandleFunc("/api/v1/calibration/points", server.calibrationGetPointsHandler)
	mux.HandleFunc("/api/v1/calibration/pressurize", server.calibrationPressurizeHandler)
	mux.HandleFunc("/api/v1/calibration/collect", server.calibrationCollectHandler)
	mux.HandleFunc("/api/v1/calibration/fit", server.calibrationFitHandler)
	mux.HandleFunc("/api/v1/calibration/pressure", server.calibrationReadPressureHandler)
	mux.HandleFunc("/api/v1/calibration/stability", server.calibrationReadStabilityHandler)
	mux.HandleFunc("/api/v1/calibration/measure-data", server.calibrationReadMeasureDataHandler)
	mux.HandleFunc("/api/v1/calibration/valve", server.calibrationValveHandler)
	mux.HandleFunc("/api/v1/calibration/measure-unit", server.calibrationReadMeasureUnitHandler)
	mux.HandleFunc("/api/v1/calibration/device-info", server.calibrationReadDeviceInfoHandler)
	mux.HandleFunc("/api/v1/calibration/reset", server.calibrationResetDeviceHandler)

	// 多设备打压控制
	mux.HandleFunc("/api/v1/multipress/register", server.multipressRegisterHandler)
	mux.HandleFunc("/api/v1/multipress/unregister", server.multipressUnregisterHandler)
	mux.HandleFunc("/api/v1/multipress/set-pressure", server.multipressSetPressureHandler)
	mux.HandleFunc("/api/v1/multipress/stop", server.multipressStopHandler)
	mux.HandleFunc("/api/v1/multipress/exhaust", server.multipressExhaustHandler)
	mux.HandleFunc("/api/v1/multipress/pressure", server.multipressReadPressureHandler)
	mux.HandleFunc("/api/v1/multipress/stability", server.multipressReadStabilityHandler)
	mux.HandleFunc("/api/v1/multipress/unit", server.multipressUnitHandler)
	mux.HandleFunc("/api/v1/multipress/devices", server.multipressDevicesHandler)
	mux.HandleFunc("/api/v1/multipress/stop-all", server.multipressStopAllHandler)

	// 报告
	mux.HandleFunc("/api/v1/reports/templates/select", server.reportTemplateSelectHandler)

	return corsMiddleware(mux)
}

// corsMiddleware 为所有 API 响应添加 CORS 头，处理 OPTIONS 预检请求。
// 桌面模式下前端从 Vite 开发服务器加载（如 localhost:5179），
// 而 API 运行在不同端口，需要 CORS 支持才能正常通信。
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// NewRouter 返回 API 路由，用于注册系统对外提供的 HTTP 端点。
func NewRouter() http.Handler {
	return NewRouterWithDeviceManager(manager.NewDeviceManager())
}
