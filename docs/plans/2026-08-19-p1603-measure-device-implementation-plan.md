# Implementation Plan: DAQ-P-1603 计量设备接入

> 状态：执行完成
> 创建：2026-08-19
> 上游 spec：[2026-08-19-p1603-measure-device-spec.md](./2026-08-19-p1603-measure-device-spec.md)（已批准）
> 前置：原 1604Pre spec/plan/代码已删除（方向纠正为 P1603）

---

## 1. 组件分解与依赖顺序

```
T0 删除1604Pre ──→ T1 spec ──→ T2 设备模型扩展 ──→ T3 P1603 adapter ──→ T4 工厂/DLL/前端 ──→ T5 验证
```

## 2. 实现要点决策（与 spec 的偏差记录）

| # | 决策 | 说明 |
|---|---|---|
| P1 | **工程量换算完全复用 device-sdk** | 复核发现 device-sdk readLoop 内部已完成 U16→mA→工程量（用每通道 RangeMin/Max），sink 收到最终工程量。1604Cal adapter 只做配置翻译 + 帧缓存，**无需重写换算公式**（比 spec D3 原设计更优） |
| P2 | **归零语义对齐 WindLabX4** | device-sdk `SetTare` 仅写 profile.TareOffset，readLoop 仍输出原始值，由展示方扣除。1604Cal 是同步 CollectData，故 adapter 本地记录 tareOffsets 并在 CollectData 扣除（展示值 = 原始值 - offset） |
| P3 | **通道号统一 1-based** | domain.ChannelConfig.Index 用 1-based（与计量业务层一致）；adapter 转 device-sdk 时 -1 |
| P4 | **P1603 端口无意义** | DLL 自管端口，前端端口字段仍保留但 P1603 不依赖（默认 9000 不强制） |

## 3. 任务清单

- [x] **T0: 删除 1604Pre 成果**（4 个驱动文件 + 单测 + spec/plan 文档 + 工厂分支 + 前端 1604Pre 选项/阀门适配）
- [x] **T1: spec 更新为 P1603**（DLL FFI + 每通道量程配置 + 仅计量，已批准）
- [x] **T2: 设备模型扩展**（domain.ChannelConfig + Device.Channels + DefaultP1603Channels；DTO 透传；JSON 持久化自动生效；旧数据 nil 回退默认）
- [x] **T3: P1603 驱动 adapter**（device-sdk DAQP1603 复用 + 帧缓存桥接 + 归零 + 阀门/单位桩 + 满量程/复位不支持）
- [x] **T4: 工厂注册 + DLL 初始化 + 前端**（factory 注册 DAQ-P-1603/P1603；main.go ffi.InitWTNDAQ16HFromEnv；DeviceFormDialog 型号选项 + 16 通道量程/单位编辑表 + 校验；MeasurementDevicePanel 无阀门适配）
- [x] **T5: 验证**（P1603 驱动 9 用例含 race + 工厂 3 用例；go build/test/vet 全绿；npm typecheck/lint/build 全绿）

## 4. 验证结果

| 项 | 结果 |
|---|---|
| `go build ./...`（GOWORK=off） | ✅ |
| `go test ./cmd/... ./internal/...` | ✅ 全包通过 |
| `go vet ./...` | ✅ |
| 驱动单测（-race） | ✅ 9 用例（配置翻译/回退/缓存桥接/未连接/归零扣除/阀门桩/单位桩/满量程/复位） |
| 工厂单测 | ✅ P1603 3 型号归一化 |
| 前端 typecheck/lint/build | ✅（lint 6 warning 已 --fix） |

## 5. 遗留（后续迭代）

| 项 | 说明 |
|---|---|
| DLL 安装包分发 | 本期代码就位（可执行文件同目录/环境变量可加载）；`WTNDAQ16H_64.dll` 打进安装包在打包阶段处理 |
| 设备级 Unit 与通道 Unit 联动 | spec Q1：本期设备级 Unit 下拉独立；通道单位仅在配置表编辑，不做联动 |
| 真机联调 | DLL 加载、16 通道采集数值与传感器比对、量程配置影响工程量、归零后读数归零 |
