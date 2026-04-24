package http

import (
	"net/http"

	"cal1604/internal/application/calibration"
	"cal1604/internal/application/deviceconnect"
	"cal1604/internal/application/measurement"
	"cal1604/internal/application/multipress"
	"cal1604/internal/application/session"
	"cal1604/internal/config"
	"cal1604/internal/device"
	"cal1604/internal/device/manager"
	"cal1604/internal/infrastructure/driver"
	"cal1604/internal/report"
	"cal1604/internal/workflow"
)

// CalibrationRuntimeConfig 定义标定启动门禁的运行时配置。
type CalibrationRuntimeConfig struct {
	EnforceValveCalibrationGate bool
}

func defaultCalibrationRuntimeConfig() CalibrationRuntimeConfig {
	return CalibrationRuntimeConfig{EnforceValveCalibrationGate: false}
}

// chainActiveDriverProvider 顺序查询多个驱动提供者，返回第一个命中的活跃驱动。
type chainActiveDriverProvider struct {
	providers []device.ActiveDriverProvider
}

func (p chainActiveDriverProvider) GetActiveDriver(id string) device.ConnectionDriver {
	for _, provider := range p.providers {
		if provider == nil {
			continue
		}
		if drv := provider.GetActiveDriver(id); drv != nil {
			return drv
		}
	}
	return nil
}

// NewRouterWithDeviceManager 基于指定设备管理器创建路由。
func NewRouterWithDeviceManager(deviceManager deviceManager) http.Handler {
	return newRouter(deviceManager, nil, deviceconnect.DefaultConfig(), defaultCalibrationRuntimeConfig(), nil, "")
}

// NewRouterWithDependencies 基于指定依赖创建路由。
func (s *apiServer) publishEventAdapter(eventType string, data any) {
	publishEvent(eventType, data)
}

// NewRouterWithDependencies 基于指定依赖创建路由。
// 该方法用于生产注入与集成测试注入同一套 HTTP 处理逻辑。
func NewRouterWithDependencies(deviceManager deviceManager, connector deviceConnector) http.Handler {
	return newRouter(deviceManager, connector, deviceconnect.DefaultConfig(), defaultCalibrationRuntimeConfig(), nil, "")
}

// NewRouterWithConnectConfig 基于指定连接配置创建路由。
func NewRouterWithConnectConfig(deviceManager deviceManager, connectConfig deviceconnect.Config) http.Handler {
	return newRouter(deviceManager, nil, connectConfig, defaultCalibrationRuntimeConfig(), nil, "")
}

// NewRouterWithRuntimeConfig 基于连接配置、标定门禁配置和应用配置创建路由。

func NewRouterWithRuntimeConfig(deviceManager deviceManager, connectConfig deviceconnect.Config, calibrationConfig CalibrationRuntimeConfig, configPath string, appCfg ...config.AppConfig) http.Handler {
	var cfg *config.AppConfig
	if len(appCfg) > 0 {
		cfg = &appCfg[0]
	}
	return newRouter(deviceManager, nil, connectConfig, calibrationConfig, cfg, configPath)
}

func newRouter(
	deviceManager deviceManager,
	connector deviceConnector,
	connectConfig deviceconnect.Config,
	calibrationConfig CalibrationRuntimeConfig,
	appCfg *config.AppConfig,
	configPath string,
) http.Handler {
	if deviceManager == nil {
		deviceManager = manager.NewDeviceManager()
	}

	sessionMachine := workflow.NewSessionMachine()
	factory := driver.NewFactory()

	server := &apiServer{
		deviceManager:      deviceManager,
		sessionMachine:     sessionMachine,
		connectConfig:      connectConfig,
		calibrationService: calibration.NewService(sessionMachine, factory, deviceManager, nil, nil, nil),
		appConfig:          appCfg,
		configPath:         configPath,
	}

	// 报告服务（模板目录为空则使用无模板模式）
	server.reportService = report.NewService("")

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

	// 多设备打压控制服务
	server.multipressService = multipress.NewService(
		factory,
		deviceManager,
		server.publishEventAdapter,
	)
	server.multipressService.StartPolling()

	// 聚合驱动提供者：优先复用设备连接服务中的驱动，其次复用 multipress 已注册驱动。
	providers := make([]device.ActiveDriverProvider, 0, 2)
	if dp, ok := server.deviceConnector.(device.ActiveDriverProvider); ok {
		providers = append(providers, dp)
	}
	providers = append(providers, server.multipressService)

	var driverProvider device.ActiveDriverProvider
	if len(providers) > 0 {
		driverProvider = chainActiveDriverProvider{providers: providers}
	}

	// 创建共享设备会话服务
	server.sessionService = session.NewService(
		deviceManager,
		factory,
		server.publishEventAdapter,
		driverProvider,
	)

	// 创建计量服务
	server.measurementService = measurement.NewService(
		server.sessionService,
		server.publishEventAdapter,
	)
	if appCfg != nil {
		server.measurementService.SetConfig(measurementConfigFromParams(appCfg.MeasurementParams))
	} else {
		server.measurementService.SetConfig(measurementConfigFromParams(config.Default().MeasurementParams))
	}

	// 注入事件发布、驱动提供者和 session 服务到校准服务
	server.calibrationService = calibration.NewService(
		sessionMachine,
		factory,
		deviceManager,
		server.publishEventAdapter,
		driverProvider,
		server.sessionService,
	)
	server.calibrationService.SetStartPrerequisiteConfig(calibration.StartPrerequisiteConfig{
		EnforceValveCalibration: calibrationConfig.EnforceValveCalibrationGate,
	})

	mux := http.NewServeMux()

	// 健康检查与事件流
	mux.HandleFunc("/api/v1/health", healthHandler)
	mux.HandleFunc("/api/v1/events/stream", eventsStreamHandler)

	// 配置
	mux.HandleFunc("/api/v1/config/device-connect", server.deviceConnectConfigHandler)
	mux.HandleFunc("/api/v1/config/calibration", server.calibrationConfigHandler)
	mux.HandleFunc("/api/v1/config/measurement", server.measurementConfigHandler)
	mux.HandleFunc("/api/v1/config/alarm", server.alarmConfigHandler)

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

	// 设备会话（共享：设备绑定、读压、读稳定性、读计量数据、阀门、单位）
	mux.HandleFunc("/api/v1/session/devices", server.sessionSetDevicesHandler)
	mux.HandleFunc("/api/v1/session/measure-device", server.sessionSetMeasureDeviceHandler)
	mux.HandleFunc("/api/v1/session/pressure", server.sessionReadPressureHandler)
	mux.HandleFunc("/api/v1/session/stability", server.sessionReadStabilityHandler)
	mux.HandleFunc("/api/v1/session/measure-data", server.sessionReadMeasureDataHandler)
	mux.HandleFunc("/api/v1/session/valve", server.sessionValveHandler)
	mux.HandleFunc("/api/v1/session/calibrate-zero", server.sessionCalibrateZeroHandler)
	mux.HandleFunc("/api/v1/session/calibrate-full-scale", server.sessionCalibrateFullScaleHandler)
	mux.HandleFunc("/api/v1/session/measure-unit", server.sessionMeasureUnitHandler)
	mux.HandleFunc("/api/v1/session/device-info", server.sessionReadDeviceInfoHandler)
	mux.HandleFunc("/api/v1/session/reset", server.sessionResetDeviceHandler)

	// 计量模块
	mux.HandleFunc("/api/v1/measurement/state", server.measurementStateHandler)
	mux.HandleFunc("/api/v1/measurement/start", server.measurementStartHandler)
	mux.HandleFunc("/api/v1/measurement/pause", server.measurementPauseHandler)
	mux.HandleFunc("/api/v1/measurement/stop", server.measurementStopHandler)
	mux.HandleFunc("/api/v1/measurement/points/generate", server.measurementGeneratePointsHandler)
	mux.HandleFunc("/api/v1/measurement/points", server.measurementPointsHandler)
	mux.HandleFunc("/api/v1/measurement/data", server.measurementDataHandler)
	mux.HandleFunc("/api/v1/measurement/export", server.measurementExportHandler)
	mux.HandleFunc("/api/v1/config/measurement-alarm", server.measurementAlarmConfigHandler)
	mux.HandleFunc("/api/v1/measurement/alarm/resolve", server.measurementAlarmResolveHandler)
	mux.HandleFunc("/api/v1/measurement/alarm/pending", server.measurementAlarmPendingHandler)
	mux.HandleFunc("/api/v1/measurement/auto-collect", server.measurementAutoCollectHandler)
	mux.HandleFunc("/api/v1/measurement/manual-pressurize", server.measurementManualPressurizeHandler)
	mux.HandleFunc("/api/v1/measurement/manual-collect", server.measurementManualCollectHandler)

	// 校准流程
	mux.HandleFunc("/api/v1/calibration/devices", server.calibrationSetDevicesHandler)
	mux.HandleFunc("/api/v1/calibration/config", server.calibrationSetConfigHandler)
	mux.HandleFunc("/api/v1/calibration/channels", server.calibrationSetChannelsHandler)
	mux.HandleFunc("/api/v1/calibration/channels/list", server.calibrationGetChannelsHandler)
	mux.HandleFunc("/api/v1/calibration/points/generate", server.calibrationGeneratePointsHandler)
	mux.HandleFunc("/api/v1/calibration/points", server.calibrationGetPointsHandler)
	mux.HandleFunc("/api/v1/calibration/pressurize", server.calibrationPressurizeHandler)
	mux.HandleFunc("/api/v1/calibration/collect", server.calibrationCollectHandler)
	mux.HandleFunc("/api/v1/calibration/fit", server.calibrationFitHandler)
	mux.HandleFunc("/api/v1/calibration/resolve-alarm", server.calibrationResolveAlarmHandler)
	mux.HandleFunc("/api/v1/calibration/retry-point", server.calibrationRetryPointHandler)
	mux.HandleFunc("/api/v1/calibration/alarm-config", server.calibrationGetAlarmConfigHandler)
	mux.HandleFunc("/api/v1/calibration/alarm-config/set", server.calibrationSetAlarmConfigHandler)
	mux.HandleFunc("/api/v1/calibration/session", server.calibrationGetSessionHandler)
	mux.HandleFunc("/api/v1/calibration/manual-pressurize", server.calibrationManualPressurizeHandler)
	mux.HandleFunc("/api/v1/calibration/manual-collect", server.calibrationManualCollectHandler)

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
	mux.HandleFunc("/api/v1/reports/export", server.exportReportHandler)
	mux.HandleFunc("/api/v1/reports/templates", server.listTemplatesHandler)

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
