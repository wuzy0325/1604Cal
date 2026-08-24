## 1. 领域数据模型

- [ ] 1.1 在 `internal/domain/pressure_point.go` 新增 `DevicePointData` 结构体（DeviceID / Collected / Status / CollectTime / SkipReason）与 `PressurePoint.CollectedByDevice map[string]DevicePointData` 字段（保留 `CollectedData` 兼容）
- [ ] 1.2 在 `internal/domain/workflow_session.go` 的 `WorkflowSession` 新增 `MeasureDeviceIDs []string` 字段（保留 `MeasureDeviceID` 兼容）
- [ ] 1.3 为 `CollectedByDevice` 增加读取辅助：优先多设备结构、缺失回退单设备 `CollectedData`（供报告 / CSV 统一消费）

## 2. 会话绑定集合化（backend session）

- [ ] 2.1 `internal/application/session/service.go`：`measureDriver`/`measureDevID` 单槽改为 `measureDrivers map[string]device.MeasureDriver` + `measureDevIDs []string`
- [ ] 2.2 `BindingToken.MeasureDeviceID` 新增 `MeasureDeviceIDs []string`；`validateToken` 改为集合成员校验
- [ ] 2.3 `BindDevices` / `BindMeasureDevice` 接受 `[]string` 并逐个 `ResolveMeasureDriver`；绑定冲突按集合判定
- [ ] 2.4 新增 `MeasureDeviceIDs()` 访问器；`MeasureDriver()` 保留（返回首个设备驱动，兼容旧调用方）
- [ ] 2.5 逐设备操作方法加 `deviceID` 参数：`ReadMeasureData`、`ReadValveStatus`、`SetValveStatus`、`CalibrateZero`、`CalibrateFullScale`、`ReadMeasureUnit`、`SetMeasureUnit`、`ReadDeviceInfo`、`ResetDevice`（保留无参版本兼容）
- [ ] 2.6 `EventSessionDeviceBound` payload 增加 `measureDeviceIds`

## 3. API / 事件契约（backend）

- [ ] 3.1 `internal/api/http/device_session_handler.go`、`calibration_handler.go`：`setDevices` / `setMeasureDevice` 请求支持 `measureDeviceIds []string`（保留单设备字段）
- [ ] 3.2 `sessionReadMeasureDataHandler` / `calibrationCollectHandler` 响应增加 `devices map[string]data`（保留 `data`）
- [ ] 3.3 `internal/events/event_types.go` 相关事件 payload 文档 / 常量注释增加 `deviceId`；`internal/application/session`、`calibration`、`measurement` 发布事件时携带 `deviceId`

## 4. 标定采集并行化（backend calibration）

- [ ] 4.1 `calibration/collector.go` `Collect`：按设备并行采集（WTN1604 走 `CollectCalibrationPoint`，否则 `CollectData`），`sync.WaitGroup` + 每设备结果聚合为 `map[deviceID][]float64`；单设备路径复用现有逻辑
- [ ] 4.2 采集结果写入 `pressurePoints[i].CollectedByDevice`；单设备同时回填 `CollectedData`
- [ ] 4.3 `EventDataCollected` / `EventPointCompleted` payload 携带 `deviceId` 与对应设备数据
- [ ] 4.4 `calibration/service.go`：`StartCalibration` / `EndCalibration` / `ValidateStartPrerequisites` 遍历设备集合（阀门门禁、StartCalibration/EndCalibration 逐设备）
- [ ] 4.5 `calibration/service.go` 新增 `skippedDevices map[string]string`（deviceID → reason）；`executePointLoop` 每点过滤被跳过设备
- [ ] 4.6 被跳过设备剩余点标记设备级 `skipped` + 原因，保留已完成数据

## 5. 报警 / 决策设备维度（backend calibration + measurement）

- [ ] 5.1 `calibration/service.go` `checkAlarm`：逐设备评估超限通道，报警事件携带 `deviceId`；决策支持设备级跳过
- [ ] 5.2 `calibration/service.go` `collectPoint` / `handlePointError`：设备失败 → 整批暂停 → `await_alarm_resolution` 等待用户「重试整点 / 跳过该设备 / 停止」
- [ ] 5.3 `measurement/alarm.go` `CheckAlarm`：逐设备评估，报警事件携带 `deviceId`；决策支持设备级跳过
- [ ] 5.4 `measurement/collector.go`：`ManualCollect` / `prepareCollectStep` / `updatePointCollectedData` 按设备并行采集与写入
- [ ] 5.5 前端报警弹窗增加「跳过该设备 + 原因（预设 + 备注）」选项（后端预留 `ResolveSkipDevice(deviceID, reason)`）

## 6. 计量采集并行化（backend measurement）

- [ ] 6.1 `measurement/collector.go` `RunAutoCollection` / `ManualCollect`：并行采集所有未跳过设备，`updatePointCollectedData` 写入 `CollectedByDevice`
- [ ] 6.2 `measurement/service.go`：`Start` / `StartWorkflow` 校验设备集合；`CollectedRow` 增加 `deviceId`
- [ ] 6.3 `EventMeasurementDataUpdated` / `EventMeasurementDataCollected` / `EventMeasurementPointStatus` 携带 `deviceId`
- [ ] 6.4 `measurement/service.go` `WriteCSV` / `rowsFromPoints` 支持设备维度数据

## 7. 报告按设备输出（backend report）

- [ ] 7.1 `internal/report/report_service.go`：`collectChannelData` / `collectBackwardData` / `collectMeasurementChannelData` / `collectMeasurementChannelByTarget` 支持设备维度聚合（回退旧字段）
- [ ] 7.2 `ExportReport` / `ExportMeasurementReport` 按设备 ID 列表分别生成报告文件
- [ ] 7.3 报告设备编号元数据从各设备配置派生

## 8. 前端类型 / store / API

- [ ] 8.1 `web/src/types/device.ts` / `web/src/api/session.ts` / `calibration.ts` / `measurement.ts`：DTO 支持 `measureDeviceIds` 与设备维度数据
- [ ] 8.2 `web/src/stores/deviceStore.ts`：`ModuleDeviceSelection` 增加 `measureDeviceIds`
- [ ] 8.3 `web/src/stores/calibration/`：`pushCalibrationConfigAndStart` 收集所有已勾选 connected 设备；`pressurePoints.ts` `collectData` 支持设备维度；`deviceControl.ts` 多选绑定
- [ ] 8.4 `web/src/stores/measurement/`：`ensureDevicesBound` / `bindDevices` 支持多设备；`channelData` 按设备
- [ ] 8.5 `web/src/shared/events.ts` 事件类型增加 `deviceId`

## 9. 前端 UI

- [ ] 9.1 `DeviceSelectionPanel` / `Device1604Panel` / `MeasurementDevicePanel`：单选改多选（`el-checkbox-group`）；选 1 台行为不变
- [ ] 9.2 `CalibrationDataView` / 计量数据视图：按设备 tab / 卡片展示各设备通道数据与状态（含「已跳过」标记 + 原因）
- [ ] 9.3 报警弹窗增加「跳过该设备」动作与原因输入
- [ ] 9.4 报告导出按设备分别触发 / 展示

## 10. 配置层与回归

- [ ] 10.1 `internal/config/app_config.go` 消费 `MeasureDeviceIDs`：绑定恢复优先读取切片
- [ ] 10.2 单设备回归验证：选 1 台设备时全流程行为与改造前一致
- [ ] 10.3 运行 `make check`（go test + vet、vue typecheck + lint + test）全量通过
- [ ] 10.4 更新 `openspec/specs/` 归档本变更规格（`openspec archive` 或 apply 后归档）