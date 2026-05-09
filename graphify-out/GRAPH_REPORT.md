# Graph Report - .  (2026-05-06)

## Corpus Check
- Large corpus: 293 files · ~261,995 words. Semantic extraction will be expensive (many Claude tokens). Consider running on a subfolder, or use --no-semantic to run AST-only.

## Summary
- 1302 nodes · 1906 edges · 136 communities (97 shown, 39 thin omitted)
- Extraction: 83% EXTRACTED · 17% INFERRED · 0% AMBIGUOUS · INFERRED: 324 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Wails Runtime Bridge|Wails Runtime Bridge]]
- [[_COMMUNITY_App Entry Points|App Entry Points]]
- [[_COMMUNITY_Calibration HTTP Handlers|Calibration HTTP Handlers]]
- [[_COMMUNITY_Device Interface Test Doubles|Device Interface Test Doubles]]
- [[_COMMUNITY_Factory & Session Tests|Factory & Session Tests]]
- [[_COMMUNITY_Measurement Core Service|Measurement Core Service]]
- [[_COMMUNITY_Circuit Breaker & Retry|Circuit Breaker & Retry]]
- [[_COMMUNITY_Session Service & Device Store|Session Service & Device Store]]
- [[_COMMUNITY_Calibration API Client|Calibration API Client]]
- [[_COMMUNITY_Report & Excel Generator|Report & Excel Generator]]
- [[_COMMUNITY_Session State Machine|Session State Machine]]
- [[_COMMUNITY_Measurement Components|Measurement Components]]
- [[_COMMUNITY_Calibration Store & Composables|Calibration Store & Composables]]
- [[_COMMUNITY_Device Driver Implementations|Device Driver Implementations]]
- [[_COMMUNITY_TCP Protocol Base|TCP Protocol Base]]
- [[_COMMUNITY_Community 15|Community 15]]
- [[_COMMUNITY_Community 16|Community 16]]
- [[_COMMUNITY_Community 17|Community 17]]
- [[_COMMUNITY_Community 18|Community 18]]
- [[_COMMUNITY_Community 19|Community 19]]
- [[_COMMUNITY_Community 20|Community 20]]
- [[_COMMUNITY_Community 21|Community 21]]
- [[_COMMUNITY_Community 22|Community 22]]
- [[_COMMUNITY_Community 23|Community 23]]
- [[_COMMUNITY_Community 24|Community 24]]
- [[_COMMUNITY_Community 25|Community 25]]
- [[_COMMUNITY_Community 26|Community 26]]
- [[_COMMUNITY_Community 27|Community 27]]
- [[_COMMUNITY_Community 28|Community 28]]
- [[_COMMUNITY_Community 29|Community 29]]
- [[_COMMUNITY_Community 30|Community 30]]
- [[_COMMUNITY_Community 31|Community 31]]
- [[_COMMUNITY_Community 32|Community 32]]
- [[_COMMUNITY_Community 33|Community 33]]
- [[_COMMUNITY_Community 34|Community 34]]
- [[_COMMUNITY_Community 35|Community 35]]
- [[_COMMUNITY_Community 36|Community 36]]
- [[_COMMUNITY_Community 37|Community 37]]
- [[_COMMUNITY_Community 38|Community 38]]
- [[_COMMUNITY_Community 39|Community 39]]
- [[_COMMUNITY_Community 40|Community 40]]
- [[_COMMUNITY_Community 41|Community 41]]
- [[_COMMUNITY_Community 42|Community 42]]
- [[_COMMUNITY_Community 43|Community 43]]
- [[_COMMUNITY_Community 45|Community 45]]
- [[_COMMUNITY_Community 46|Community 46]]
- [[_COMMUNITY_Community 47|Community 47]]
- [[_COMMUNITY_Community 49|Community 49]]
- [[_COMMUNITY_Community 50|Community 50]]
- [[_COMMUNITY_Community 51|Community 51]]
- [[_COMMUNITY_Community 52|Community 52]]
- [[_COMMUNITY_Community 53|Community 53]]
- [[_COMMUNITY_Community 54|Community 54]]
- [[_COMMUNITY_Community 55|Community 55]]
- [[_COMMUNITY_Community 58|Community 58]]
- [[_COMMUNITY_Community 59|Community 59]]
- [[_COMMUNITY_Community 60|Community 60]]
- [[_COMMUNITY_Community 62|Community 62]]
- [[_COMMUNITY_Community 63|Community 63]]
- [[_COMMUNITY_Community 64|Community 64]]
- [[_COMMUNITY_Community 65|Community 65]]
- [[_COMMUNITY_Community 67|Community 67]]
- [[_COMMUNITY_Community 68|Community 68]]
- [[_COMMUNITY_Community 69|Community 69]]
- [[_COMMUNITY_Community 70|Community 70]]
- [[_COMMUNITY_Community 71|Community 71]]
- [[_COMMUNITY_Community 72|Community 72]]
- [[_COMMUNITY_Community 73|Community 73]]
- [[_COMMUNITY_Community 74|Community 74]]
- [[_COMMUNITY_Community 75|Community 75]]
- [[_COMMUNITY_Community 76|Community 76]]
- [[_COMMUNITY_Community 77|Community 77]]
- [[_COMMUNITY_Community 78|Community 78]]
- [[_COMMUNITY_Community 79|Community 79]]
- [[_COMMUNITY_Community 81|Community 81]]
- [[_COMMUNITY_Community 82|Community 82]]
- [[_COMMUNITY_Community 83|Community 83]]
- [[_COMMUNITY_Community 84|Community 84]]
- [[_COMMUNITY_Community 86|Community 86]]
- [[_COMMUNITY_Community 87|Community 87]]
- [[_COMMUNITY_Community 88|Community 88]]

## God Nodes (most connected - your core abstractions)
1. `apiServer` - 74 edges
2. `writeSuccess()` - 69 edges
3. `writeError()` - 62 edges
4. `Service` - 40 edges
5. `Service` - 37 edges
6. `newRouter()` - 28 edges
7. `setupMeasurementService()` - 22 edges
8. `Service` - 22 edges
9. `Service` - 18 edges
10. `WTN1604Driver` - 17 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `NewApp()`  [INFERRED]
  main.go → app.go
- `resolveRuntimeConfig()` --calls--> `Default()`  [INFERRED]
  app.go → internal/config/app_config.go
- `resolveRuntimeConfig()` --calls--> `LoadFromFile()`  [INFERRED]
  app.go → internal/config/app_config.go
- `main()` --calls--> `NewPersistentDeviceManager()`  [INFERRED]
  cmd/server/main.go → internal/device/manager/persistent_device_manager.go
- `main()` --calls--> `NewRouterWithRuntimeConfig()`  [INFERRED]
  cmd/server/main.go → internal/api/http/router.go

## Communities (136 total, 39 thin omitted)

### Community 0 - "Wails Runtime Bridge"
Cohesion: 0.06
Nodes (67): BrowserOpenURL(), CanResolveFilePaths(), CheckNotificationAuthorization(), CleanupNotifications(), ClipboardGetText(), ClipboardSetText(), Environment(), EventsEmit() (+59 more)

### Community 1 - "App Entry Points"
Cohesion: 0.05
Nodes (37): App, NewApp(), resolveRuntimeConfig(), withCORS(), main(), Default(), LoadFromFile(), TestDefaultCalibrationConfigDisablesValveGate() (+29 more)

### Community 2 - "Calibration HTTP Handlers"
Cohesion: 0.09
Nodes (3): apiServer, writeError(), writeSuccess()

### Community 3 - "Device Interface Test Doubles"
Cohesion: 0.05
Nodes (31): embedMD, embedPD, fakeMeasureDriver, fakePressureDriver, fakeStore, mapProvider, reachState(), setupMeasurementService() (+23 more)

### Community 4 - "Factory & Session Tests"
Cohesion: 0.05
Nodes (26): NewFactory(), TestFactoryRejectsUnsupportedModel(), TestFactorySupportsMVPModels(), TestStartWithoutDevice(), embedMeasure, embedPressure, fakeMeasureDriver, fakePressureDriver (+18 more)

### Community 5 - "Measurement Core Service"
Cohesion: 0.06
Nodes (10): CollectedRow, averageMeasurementSamples(), Config, EventPublisher, Point, generatePointsFromConfig(), generatePointsFromCustom(), roundToPrecision() (+2 more)

### Community 6 - "Circuit Breaker & Retry"
Cohesion: 0.05
Nodes (21): DefaultCircuitBreakerConfig(), DefaultRetryConfig(), NewCircuitBreaker(), NewRetryStrategy(), CircuitBreaker, CircuitBreakerConfig, CircuitState, newConST811ADriver() (+13 more)

### Community 7 - "Session Service & Device Store"
Cohesion: 0.06
Nodes (23): Config, DeviceStore, fakeConnectionDriver, fakeDriverFactory, Option, publishRecord, Service, NewService() (+15 more)

### Community 8 - "Calibration API Client"
Cohesion: 0.08
Nodes (21): collectData(), fetchSessionState(), fitData(), generatePressurePoints(), getAlarmConfig(), getCalibrationParamsConfig(), getPressurePoints(), pressurize() (+13 more)

### Community 9 - "Report & Excel Generator"
Cohesion: 0.09
Nodes (18): ChannelBlock, CreateFallbackWorkbook(), FillMeasureData(), FillRoundTripData(), FillStandardValues(), FindChannelBlocks(), LoadTemplate(), collectBackwardData() (+10 more)

### Community 10 - "Session State Machine"
Cohesion: 0.11
Nodes (14): connectDevice(), disconnectDevice(), fetchDeviceConnectConfig(), fetchUnitConsistency(), upsertDevice(), applyDeviceStatusChangedEvent(), generateDeviceId(), isValidIPv4() (+6 more)

### Community 13 - "Device Driver Implementations"
Cohesion: 0.13
Nodes (16): TestConnectAndDisconnectDeviceEndpoints(), TestGetDevicesReturnsEmptyListAtStart(), TestPostDeviceRejectsInvalidHostAndPort(), TestPostDeviceThenList(), TestUnitConsistencyCheck(), TestUpdateDeviceStatus(), TestSSEEndpointStreamsEvents(), handlerFakeDriverFactory (+8 more)

### Community 14 - "TCP Protocol Base"
Cohesion: 0.13
Nodes (4): channelsToBitmap(), unitToCoefficient(), parseCalibrationValues(), WTN1604Driver

### Community 15 - "Community 15"
Cohesion: 0.19
Nodes (17): autoCollectMeasurement(), checkMeasurementAlarmPending(), fetchMeasurementData(), fetchMeasurementPoints(), fetchMeasurementState(), generateMeasurementPoints(), getMeasurementAlarmConfig(), getMeasurementExportUrl() (+9 more)

### Community 18 - "Community 18"
Cohesion: 0.24
Nodes (16): containsStatus(), newValveGateTestService(), TestCollectPointAlarmDecisionRecollect(), TestCollectPointAlarmDecisionSkip(), TestCollectPointAlarmDecisionStop(), TestCollectPublishesPointStatusEvents(), TestCollectUsesCalibrationPointCommandWhenSupported(), TestPauseAutoCollectionStopsRunningLoop() (+8 more)

### Community 19 - "Community 19"
Cohesion: 0.24
Nodes (16): build_tick_segments(), draw_background(), draw_icon(), draw_symbol(), generate_svg(), lerp_color(), main(), parse_args() (+8 more)

### Community 20 - "Community 20"
Cohesion: 0.18
Nodes (8): multipressExhaust(), multipressListDevices(), multipressRegister(), multipressSetPressure(), multipressSetUnit(), multipressStop(), multipressStopAll(), multipressUnregister()

### Community 21 - "Community 21"
Cohesion: 0.19
Nodes (15): TestMeasurementStartRejectsEmptyChannels(), TestMeasurementStartRequiresGeneratedPoints(), newSessionRouterWithMeasureDriver(), newSessionRouterWithMeasureDriverAndFakeDriver(), newSessionRouterWithMeasureDriverAndRuntimeConfig(), TestSessionCalibrateFullScaleEndpoint(), TestSessionCalibrateZeroEndpoint(), TestSessionInitialStateIsIdle() (+7 more)

### Community 22 - "Community 22"
Cohesion: 0.13
Nodes (5): createEventStream(), initDesktopApiBase(), requestJSON(), startEventStream(), bootstrap()

### Community 23 - "Community 23"
Cohesion: 0.16
Nodes (10): Device, DeviceStatus, DeviceType, isValidDeviceStatus(), isValidDeviceType(), deviceConnector, deviceManager, setDeviceStatusRequest (+2 more)

### Community 24 - "Community 24"
Cohesion: 0.19
Nodes (10): NewStabilityAccumulator(), NewStabilityMonitor(), TestStabilityAccumulatorResetsOnDrift(), TestStabilityMonitorPublishesLostEvent(), TestStabilityMonitorPublishesProgressAndAchieved(), StabilityAccumulator, stabilityEvent, StabilityEventPublisher (+2 more)

### Community 25 - "Community 25"
Cohesion: 0.14
Nodes (3): parseSPC4000UnitCode(), pressureUnitToCodeSPC4000(), SPC4000Driver

### Community 26 - "Community 26"
Cohesion: 0.15
Nodes (8): alarmDecisionRequest, decodePointIndexRequest(), channelsResponse, manualPressurizeRequest, pointIndexRequest, setChannelsRequest, setConfigRequest, setDevicesRequest

### Community 27 - "Community 27"
Cohesion: 0.21
Nodes (6): constBaseDriver, approxEqual(), coefficientToUnit(), matchCoefficientToUnit(), parseTargetRange(), parseWTN1604Unit()

### Community 30 - "Community 30"
Cohesion: 0.28
Nodes (11): bindDevices(), bindMeasureDevice(), readDeviceInfo(), readMeasureData(), readMeasureUnit(), readPressure(), readStability(), readValveStatus() (+3 more)

### Community 37 - "Community 37"
Cohesion: 0.18
Nodes (10): calibrateFullScaleRequest, calibrateZeroRequest, normalizeValveStatus(), measureUnitResponse, pressureResponse, setMeasureDeviceRequest, setMeasureUnitRequest, setValveRequest (+2 more)

### Community 38 - "Community 38"
Cohesion: 0.22
Nodes (5): calibrationConfigFromParams(), durationToMilliseconds(), measurementConfigFromParams(), validateMeasurementParams(), deviceConnectConfigPayload

### Community 39 - "Community 39"
Cohesion: 0.2
Nodes (10): CalibrationConfig, CalibrationResult, CalibrationSession, FittingResult, PressurePoint, defaultStartPrerequisiteConfig(), isPointTerminalStatus(), NewService() (+2 more)

### Community 40 - "Community 40"
Cohesion: 0.29
Nodes (7): isValidPressureUnit(), NormalizePressureUnit(), parseConST820Unit(), parseConSTGeneralUnit(), TestNormalizePressureUnit(), TestParseConST820Unit_NormalizesCase(), TestParseConSTGeneralUnit_NormalizesCase()

### Community 42 - "Community 42"
Cohesion: 0.2
Nodes (6): NewAlarmService(), TestAlarmDecisionAllowsSupportedActions(), TestAlarmEvaluateDeviation(), AlarmResult, AlarmService, MultiChannelAlarmResult

### Community 43 - "Community 43"
Cohesion: 0.24
Nodes (9): callMeasurementDataEndpoint(), callMeasurementStateEndpoint(), TestMeasurementStartCreatesWorkflowSession(), TestMeasurementStartRequiresBoundMeasureDevice(), waitForMeasurementData(), measurementDataResponse, measurementPointResponse, measurementRow (+1 more)

### Community 46 - "Community 46"
Cohesion: 0.36
Nodes (8): create_ico(), draw_icon(), lerp_color(), main(), make_dib_from_pil(), radial_gradient(), Convert PIL RGBA image to Windows ICO DIB (bottom-up BGRA + AND mask), save_pngs()

### Community 47 - "Community 47"
Cohesion: 0.31
Nodes (7): CalibrationRuntimeConfig, chainActiveDriverProvider, corsMiddleware(), defaultCalibrationRuntimeConfig(), NewRouterWithConnectConfig(), NewRouterWithDependencies(), NewRouterWithDeviceManager()

### Community 49 - "Community 49"
Cohesion: 0.25
Nodes (3): eventsStreamHandler(), publishEvent(), sseEventPayload

### Community 50 - "Community 50"
Cohesion: 0.36
Nodes (4): create_ico(), draw_icon(), hex_to_rgb(), main()

### Community 52 - "Community 52"
Cohesion: 0.29
Nodes (6): ActiveDriverProvider, ConnectionDriver, ConnectionDriverFactory, DeviceStore, MeasureDriver, PressureDriver

### Community 53 - "Community 53"
Cohesion: 0.52
Nodes (6): _build_request_body(), _call_image_api(), _get_headers(), handle_single_task(), image_generate(), main()

### Community 60 - "Community 60"
Cohesion: 0.4
Nodes (3): deviceEntry, DevicePressureState, StatusPublisher

### Community 63 - "Community 63"
Cohesion: 0.5
Nodes (3): TestCalibrationConfigAcceptsControlAndPressureMode(), TestCalibrationRoutesDoNotExposeMeasurementSessionEndpoints(), calibrationPointResponse

### Community 65 - "Community 65"
Cohesion: 0.5
Nodes (3): multipressDeviceIDRequest, multipressSetPressureRequest, multipressSetUnitRequest

### Community 68 - "Community 68"
Cohesion: 0.83
Nodes (3): get(), main(), post()

### Community 69 - "Community 69"
Cohesion: 0.83
Nodes (3): get(), main(), post()

## Knowledge Gaps
- **97 isolated node(s):** `Response`, `setDevicesRequest`, `setConfigRequest`, `setChannelsRequest`, `pointIndexRequest` (+92 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **39 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `newRouter()` connect `Device Driver Implementations` to `App Entry Points`, `Factory & Session Tests`, `Community 38`, `Session Service & Device Store`, `Community 43`, `Community 47`, `Community 18`, `Community 21`, `Community 63`?**
  _High betweenness centrality (0.225) - this node is a cross-community bridge._
- **Why does `NewFactory()` connect `Factory & Session Tests` to `Device Interface Test Doubles`, `Device Driver Implementations`, `Circuit Breaker & Retry`?**
  _High betweenness centrality (0.194) - this node is a cross-community bridge._
- **Why does `Default()` connect `App Entry Points` to `Device Driver Implementations`, `Community 38`?**
  _High betweenness centrality (0.114) - this node is a cross-community bridge._
- **Are the 68 inferred relationships involving `writeSuccess()` (e.g. with `.calibrationSetDevicesHandler()` and `.calibrationSetConfigHandler()`) actually correct?**
  _`writeSuccess()` has 68 INFERRED edges - model-reasoned connections that need verification._
- **Are the 61 inferred relationships involving `writeError()` (e.g. with `.calibrationSetDevicesHandler()` and `.calibrationSetConfigHandler()`) actually correct?**
  _`writeError()` has 61 INFERRED edges - model-reasoned connections that need verification._
- **What connects `Response`, `setDevicesRequest`, `setConfigRequest` to the rest of the system?**
  _97 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Wails Runtime Bridge` be split into smaller, more focused modules?**
  _Cohesion score 0.06 - nodes in this community are weakly interconnected._