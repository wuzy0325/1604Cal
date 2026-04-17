<template>
  <section class="module-page">
    <div class="desktop-shell">
      <header class="module-header">
        <div>
          <p class="module-caption">
            Measurement Workbench
          </p>
          <h1>计量工作台</h1>
        </div>

        <nav class="module-switch">
          <RouterLink
            class="switch-btn"
            :to="{ name: 'module-device-management' }"
          >
            设备管理
          </RouterLink>
          <RouterLink
            class="switch-btn active"
            :to="{ name: 'module-measurement' }"
          >
            计量模块
          </RouterLink>
          <RouterLink
            class="switch-btn"
            :to="{ name: 'module-calibration' }"
          >
            标定模块
          </RouterLink>
          <RouterLink
            class="switch-btn"
            :to="{ name: 'module-multi-pressure' }"
          >
            多设备打压
          </RouterLink>
          <RouterLink
            class="switch-btn switch-btn-ghost"
            :to="{ name: 'module-hub' }"
          >
            <el-icon><ArrowLeft /></el-icon>
            返回
          </RouterLink>
        </nav>
      </header>

      <div class="measurement-layout">
        <!-- 可折叠侧边栏 -->
        <aside
          class="sidebar"
          :class="{ collapsed: sidebarCollapsed }"
        >
          <div
            class="sidebar-toggle"
            role="button"
            tabindex="0"
            aria-label="切换侧边栏"
            @click="sidebarCollapsed = !sidebarCollapsed"
            @keydown.enter="sidebarCollapsed = !sidebarCollapsed"
            @keydown.space.prevent="sidebarCollapsed = !sidebarCollapsed"
          >
            <el-icon>
              <ArrowRight v-if="sidebarCollapsed" />
              <ArrowLeft v-else />
            </el-icon>
          </div>
          <div
            v-show="!sidebarCollapsed"
            class="sidebar-content"
          >
            <!-- 1604 计量设备 -->
            <div class="sidebar-section">
              <h3 class="sidebar-title">
                <el-icon><Monitor /></el-icon>
                1604 计量设备
              </h3>
              <Device1604Panel
                @connect="calibrationStore.connectDevice1604"
                @disconnect="calibrationStore.disconnectDevice1604"
              />
            </div>

            <!-- 打压设备 -->
            <div class="sidebar-section">
              <h3 class="sidebar-title">
                <el-icon><FirstAidKit /></el-icon>
                打压设备
              </h3>
              <PressDevicePanel
                @connect="handlePressDeviceConnect"
                @disconnect="calibrationStore.disconnectPressDevice"
                @set-pressure="handleSetPressure"
              />
            </div>

            <!-- 单位一致性检查 -->
            <div
              v-if="showUnitCheck"
              class="sidebar-section"
            >
              <h3 class="sidebar-title">
                <el-icon><ScaleToOriginal /></el-icon>
                单位检查
              </h3>
              <div
                class="unit-check-result"
                :class="unitConsistent ? 'unit-ok' : 'unit-warn'"
              >
                <el-icon v-if="unitConsistent">
                  <CircleCheckFilled />
                </el-icon>
                <el-icon v-else>
                  <Warning />
                </el-icon>
                <span>{{ unitConsistent ? '设备单位一致' : '设备单位不一致' }}</span>
              </div>
              <div
                v-if="!unitConsistent && unitConflicts.length > 0"
                class="unit-conflicts"
              >
                <div
                  v-for="(msg, idx) in unitConflicts"
                  :key="idx"
                  class="conflict-item"
                >
                  {{ msg }}
                </div>
              </div>
            </div>

            <!-- 通道选择 -->
            <div class="sidebar-section">
              <h3 class="sidebar-title">
                <el-icon><Grid /></el-icon>
                通道选择
                <span class="channel-count">{{ calibrationStore.selectedChannels.length }}/16</span>
              </h3>
              <ChannelMatrix
                :selected-channels="calibrationStore.selectedChannels"
                @update:selected-channels="calibrationStore.setSelectedChannels"
              />
            </div>

            <!-- 启动前置条件 -->
            <div class="sidebar-section">
              <h3 class="sidebar-title">
                <el-icon><CircleCheckFilled /></el-icon>
                启动条件
              </h3>
              <div class="prerequisites-list">
                <div
                  v-for="(item, index) in prerequisites"
                  :key="index"
                  class="prereq-item"
                  :class="{ satisfied: item.satisfied }"
                >
                  <el-icon v-if="item.satisfied">
                    <CircleCheckFilled />
                  </el-icon>
                  <el-icon v-else>
                    <CircleClose />
                  </el-icon>
                  <span>{{ item.label }}</span>
                </div>
              </div>
            </div>
          </div>
        </aside>

        <!-- 主工作区 -->
        <main class="workbench">
          <!-- 控制面板：参数配置 + 模式/进度/报警/操作 -->
          <div class="control-panel">
            <!-- 上部：参数配置 -->
            <div class="control-panel-top">
              <div class="param-group">
                <div class="control-group">
                  <label>最小值</label>
                  <el-input-number
                    v-model="calibrationStore.calibrationParams.minValue"
                    :precision="2"
                    :step="0.1"
                    size="small"
                  />
                </div>
                <div class="control-group">
                  <label>最大值</label>
                  <el-input-number
                    v-model="calibrationStore.calibrationParams.maxValue"
                    :precision="2"
                    :step="0.1"
                    size="small"
                  />
                </div>
                <div class="control-group">
                  <label>测点数</label>
                  <el-input-number
                    v-model="calibrationStore.calibrationParams.points"
                    :min="2"
                    :max="11"
                    size="small"
                  />
                </div>
              </div>

              <div class="param-divider" />

              <div class="param-group">
                <div class="control-group">
                  <label>精度</label>
                  <el-input-number
                    v-model="calibrationStore.calibrationParams.precision"
                    :min="0"
                    :max="4"
                    size="small"
                  />
                </div>
                <div class="control-group">
                  <label>采样数</label>
                  <el-input-number
                    v-model="calibrationStore.calibrationParams.averageCount"
                    :min="1"
                    :max="100"
                    size="small"
                  />
                </div>
              </div>

              <div class="param-divider" />

              <div class="param-group">
                <div class="control-group">
                  <label>稳定时间</label>
                  <el-select
                    v-model="calibrationStore.calibrationParams.stableTime"
                    size="small"
                  >
                    <el-option
                      label="1s"
                      :value="1"
                    />
                    <el-option
                      label="3s"
                      :value="3"
                    />
                    <el-option
                      label="5s"
                      :value="5"
                    />
                    <el-option
                      label="10s"
                      :value="10"
                    />
                  </el-select>
                </div>
                <div class="control-group">
                  <label>精度等级</label>
                  <el-select
                    v-if="!useCustomPrecision"
                    v-model="calibrationStore.calibrationParams.precisionLevel"
                    size="small"
                  >
                    <el-option
                      label="0.01%"
                      value="0.01"
                    />
                    <el-option
                      label="0.05%"
                      value="0.05"
                    />
                    <el-option
                      label="0.1%"
                      value="0.1"
                    />
                    <el-option
                      label="0.2%"
                      value="0.2"
                    />
                  </el-select>
                  <el-input-number
                    v-else
                    v-model="customPrecisionValue"
                    :min="0.0001"
                    :max="0.01"
                    :step="0.0001"
                    :precision="4"
                    size="small"
                    @change="calibrationStore.calibrationParams.precisionLevel = String(customPrecisionValue)"
                  />
                  <el-checkbox
                    v-model="useCustomPrecision"
                    size="small"
                    class="custom-precision-check"
                  >
                    自定义
                  </el-checkbox>
                </div>
              </div>

              <el-button
                type="primary"
                class="generate-btn"
                @click="calibrationStore.generatePressurePoints({ controlMode, pressureMode })"
              >
                <el-icon><List /></el-icon>
                生成测点
              </el-button>
            </div>

            <!-- 下部：模式 + 进度 + 报警 + 操作 -->
            <div class="control-panel-bottom">
              <div class="mode-switches">
                <div class="switch-group">
                  <span>控制模式</span>
                  <el-radio-group
                    v-model="controlMode"
                    size="small"
                  >
                    <el-radio-button label="auto">
                      自动
                    </el-radio-button>
                    <el-radio-button label="manual">
                      手动
                    </el-radio-button>
                  </el-radio-group>
                </div>
                <div class="switch-group">
                  <span>打压模式</span>
                  <el-radio-group
                    v-model="pressureMode"
                    size="small"
                  >
                    <el-radio-button label="single">
                      单程
                    </el-radio-button>
                    <el-radio-button label="return">
                      回程
                    </el-radio-button>
                  </el-radio-group>
                </div>
              </div>

              <div
                v-if="calibrationStore.pressurePoints.length > 0"
                class="progress-section"
              >
                <div class="progress-info">
                  <span>进度 {{ completedCount }}/{{ calibrationStore.pressurePoints.length }}</span>
                  <el-progress
                    :percentage="progressPercent"
                    :stroke-width="8"
                  />
                </div>
                <div class="stable-status">
                  <span
                    v-if="calibrationStore.isRunning"
                    class="realtime-pressure"
                  >
                    压力: {{ calibrationStore.currentPressure?.toFixed(2) || '--' }} kPa
                  </span>
                  <span :class="['stable-badge', calibrationStore.isStable ? 'stable-yes' : 'stable-no']">
                    {{ calibrationStore.isStable ? '已稳定' : '稳定中' }}
                  </span>
                  <span
                    v-if="!calibrationStore.isStable && calibrationStore.isRunning"
                    class="countdown"
                  >{{ stableCountdown }}s</span>
                  <span class="session-state">
                    会话: {{ sessionStateText }}
                  </span>
                </div>
              </div>

              <div
                v-else
                class="session-status-inline"
              >
                <span :class="['status-badge', `status-${sessionState}`]">
                  {{ sessionStateText }}
                </span>
              </div>

              <!-- 报警设置 -->
              <div class="alarm-settings">
                <span class="alarm-label">报警</span>
                <el-switch
                  v-model="alarmEnabled"
                  active-text="启用"
                  size="small"
                />
                <el-switch
                  v-model="alarmSound"
                  :disabled="!alarmEnabled"
                  active-text="声音"
                  size="small"
                />
                <el-switch
                  v-model="alarmConfirm"
                  :disabled="!alarmEnabled"
                  active-text="确认"
                  size="small"
                />
                <el-button
                  :disabled="!alarmEnabled"
                  size="small"
                  @click="showChannelDialog = true"
                >
                  通道 ({{ alarmChannels.length }}/16)
                </el-button>
              </div>

              <div class="action-buttons">
                <!-- 自动模式：开始采集 -->
                <el-button
                  v-if="controlMode === 'auto'"
                  type="success"
                  :disabled="calibrationStore.isRunning"
                  @click="calibrationStore.startCalibration()"
                >
                  <el-icon><VideoPlay /></el-icon>
                  开始采集
                </el-button>
                <!-- 手动模式：手动采集 -->
                <el-button
                  v-else
                  type="success"
                  :disabled="!canManualCollect"
                  @click="handleManualCollect"
                >
                  <el-icon><VideoPlay /></el-icon>
                  采集
                </el-button>
                <el-button
                  :disabled="!calibrationStore.isRunning"
                  @click="calibrationStore.pauseCalibration()"
                >
                  <el-icon><VideoPause /></el-icon>
                  暂停
                </el-button>
                <el-button
                  :disabled="sessionState !== 'paused'"
                  @click="calibrationStore.resumeCalibration()"
                >
                  <el-icon><RefreshRight /></el-icon>
                  恢复
                </el-button>
                <el-button
                  type="danger"
                  :disabled="sessionState === 'idle' || sessionState === 'stopped'"
                  @click="calibrationStore.stopCalibration()"
                >
                  <el-icon><CloseBold /></el-icon>
                  停止
                </el-button>
                <el-button
                  v-if="calibrationStore.hasCollectedData"
                  type="warning"
                  @click="calibrationStore.resetCollection()"
                >
                  <el-icon><RefreshLeft /></el-icon>
                  重置
                </el-button>
                <el-button
                  v-if="canExportReport"
                  type="primary"
                  @click="exportCSV"
                >
                  <el-icon><Download /></el-icon>
                  导出报告
                </el-button>
              </div>
            </div>
          </div>

          <!-- 数据表格 -->
          <div class="data-table-wrapper">
            <div class="table-header">
              <div class="table-title">
                <el-icon><DataLine /></el-icon>
                <h3>采集数据</h3>
                <span
                  v-if="calibrationStore.pressurePoints.length > 0"
                  class="record-count"
                >
                  {{ calibrationStore.pressurePoints.length }} 个测点
                </span>
              </div>
              <div class="table-actions">
                <el-button
                  size="small"
                  @click="handleSelectTemplate"
                >
                  <el-icon><Document /></el-icon>
                  报告模板
                </el-button>
                <el-button
                  size="small"
                  :disabled="calibrationStore.pressurePoints.length === 0"
                  @click="exportCSV"
                >
                  <el-icon><Download /></el-icon>
                  导出CSV
                </el-button>
              </div>
            </div>

            <el-table
              :data="tableData"
              border
              stripe
              class="data-table"
            >
              <el-table-column
                label="序号"
                width="70"
              >
                <template #default="{ row }">
                  <span>{{ row.index }}</span>
                  <el-tag
                    v-if="pressureMode === 'return'"
                    :type="isForwardPoint(row.index) ? 'primary' : 'warning'"
                    size="small"
                    class="trip-badge"
                  >
                    {{ isForwardPoint(row.index) ? '正' : '回' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column
                label="状态"
                width="85"
              >
                <template #default="{ row }">
                  <el-tag
                    :type="getPointStatusType(row.status)"
                    size="small"
                  >
                    {{ getPointStatusText(row.status) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column
                label="目标值"
                width="110"
              >
                <template #default="{ row }">
                  <el-input-number
                    :model-value="row.targetValue"
                    :controls="false"
                    :precision="2"
                    :step="0.1"
                    size="small"
                    :disabled="calibrationStore.isRunning"
                    @change="(val: number | undefined) => updateTargetValue(row.id, val ?? row.targetValue)"
                  />
                </template>
              </el-table-column>
              <el-table-column
                label="实际压力"
                width="90"
              >
                <template #default="{ row }">
                  {{ row.actualPressure?.toFixed(2) || '--' }}
                </template>
              </el-table-column>
              <el-table-column
                v-for="ch in calibrationStore.selectedChannels"
                :key="ch"
                :label="`CH${ch}`"
                width="75"
              >
                <template #default="{ row }">
                  <span :class="getChannelClass(row, ch - 1)">{{ row.channelValues[ch - 1]?.toFixed(3) || '--' }}</span>
                </template>
              </el-table-column>
              <el-table-column
                label="操作"
                width="200"
                fixed="right"
              >
                <template #default="{ row }">
                  <div class="row-actions">
                    <el-button
                      v-if="row.status === 'pending'"
                      type="primary"
                      size="small"
                      @click="calibrationStore.pressurize(row.id)"
                    >
                      打压
                    </el-button>
                    <el-button
                      v-if="row.status === 'stabilizing'"
                      type="success"
                      size="small"
                      @click="calibrationStore.confirmPressure(row.id)"
                    >
                      确认
                    </el-button>
                    <el-button
                      v-if="row.status === 'stabilizing' || row.status === 'pressurizing'"
                      size="small"
                      @click="calibrationStore.collectData(row.id)"
                    >
                      采集
                    </el-button>
                    <el-button
                      v-if="row.status === 'completed'"
                      type="warning"
                      link
                      size="small"
                      @click="calibrationStore.collectData(row.id)"
                    >
                      重新采集
                    </el-button>
                    <span
                      v-if="row.status === 'collecting'"
                      class="collecting-text"
                    >采集中...</span>
                  </div>
                </template>
              </el-table-column>
            </el-table>

            <!-- 空状态 -->
            <div
              v-if="calibrationStore.pressurePoints.length === 0"
              class="empty-state"
            >
              <el-icon class="empty-icon">
                <SetUp />
              </el-icon>
              <p>配置参数后点击「生成测点」开始采集流程</p>
            </div>
          </div>

          <!-- 报告模板选择结果 -->
          <div
            v-if="templateFilename"
            class="template-result-bar"
          >
            <el-icon><DocumentChecked /></el-icon>
            <span>当前报告模板：{{ templateFilename }}</span>
          </div>

          <!-- 错误消息 -->
          <div
            v-if="errorMessage"
            class="error-message"
          >
            <el-icon><Warning /></el-icon>
            {{ errorMessage }}
          </div>
        </main>
      </div>

      <!-- 报警通道选择对话框 -->
      <el-dialog
        v-model="showChannelDialog"
        title="报警检测通道选择"
        width="420px"
      >
        <p class="channel-dialog-tip">
          选择需要参与报警检测的通道，未选择的通道将不进行报警检测
        </p>
        <div class="channel-grid">
          <div
            v-for="ch in 16"
            :key="ch"
            class="channel-tile"
            :class="{ selected: alarmChannels.includes(ch) }"
            @click="toggleAlarmChannel(ch)"
          >
            <span>通道{{ ch }}</span>
            <el-icon v-if="alarmChannels.includes(ch)">
              <Check />
            </el-icon>
          </div>
        </div>
        <template #footer>
          <el-button @click="alarmChannels = Array.from({ length: 16 }, (_, i) => i + 1)">
            全选
          </el-button>
          <el-button @click="alarmChannels = []">
            全不选
          </el-button>
          <el-button @click="showChannelDialog = false">
            取消
          </el-button>
          <el-button
            type="primary"
            @click="showChannelDialog = false"
          >
            确定
          </el-button>
        </template>
      </el-dialog>

      <!-- 报告模板选择对话框 -->
      <el-dialog
        v-model="showTemplateDialog"
        title="选择报告模板"
        width="380px"
      >
        <el-form
          label-width="70px"
        >
          <el-form-item label="测点数">
            <el-input-number
              v-model="templatePoints"
              :min="2"
              :max="11"
            />
          </el-form-item>
          <el-form-item label="模式">
            <el-select v-model="templateMode">
              <el-option
                label="单程"
                value="single"
              />
              <el-option
                label="回程"
                value="return"
              />
            </el-select>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="showTemplateDialog = false">
            取消
          </el-button>
          <el-button
            type="primary"
            @click="confirmTemplate"
          >
            确定
          </el-button>
        </template>
      </el-dialog>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { RouterLink } from 'vue-router'
import {
  ArrowLeft,
  ArrowRight,
  List,
  VideoPlay,
  VideoPause,
  RefreshRight,
  RefreshLeft,
  CloseBold,
  CircleCheckFilled,
  CircleClose,
  DataLine,
  Download,
  SetUp,
  Document,
  DocumentChecked,
  Warning,
  Grid,
  Monitor,
  FirstAidKit,
  ScaleToOriginal,
  Check
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

import Device1604Panel from '@/components/calibration/Device1604Panel.vue'
import PressDevicePanel from '@/components/calibration/PressDevicePanel.vue'
import ChannelMatrix from '@/components/calibration/ChannelMatrix.vue'
import { selectReportTemplate, createEventStream, fetchUnitConsistency, type SessionState, type StreamEventPayload } from '@/services/apiClient'
import { useCalibrationStore } from '@/stores/calibration'
import { useMeasurementDeviceStore } from '@/stores/measurement/deviceStore'

const calibrationStore = useCalibrationStore()
const deviceStore = useMeasurementDeviceStore()

const sidebarCollapsed = ref(false)
const errorMessage = ref('')
const controlMode = ref('auto')
const pressureMode = ref('single')
const useCustomPrecision = ref(false)
const customPrecisionValue = ref(0.05)

// 报警设置（本地 UI 状态）
const alarmEnabled = ref(true)
const alarmSound = ref(true)
const alarmConfirm = ref(true)
const alarmChannels = ref<number[]>(Array.from({ length: 16 }, (_, i) => i + 1))
const showChannelDialog = ref(false)

const toggleAlarmChannel = (ch: number) => {
  const idx = alarmChannels.value.indexOf(ch)
  if (idx >= 0) {
    alarmChannels.value.splice(idx, 1)
  } else {
    alarmChannels.value.push(ch)
  }
}

// 单位一致性检查
const unitConsistent = ref(true)
const unitConflicts = ref<string[]>([])
const showUnitCheck = computed(() =>
  calibrationStore.device1604Connected && calibrationStore.pressDeviceConnected
)

const checkUnitConsistency = async () => {
  try {
    const result = await fetchUnitConsistency()
    unitConsistent.value = result.consistent
    unitConflicts.value = result.conflicts ?? []
  } catch {
    // 静默失败
  }
}

const handlePressDeviceConnect = async (deviceId: string) => {
  await calibrationStore.connectPressDevice(deviceId)
  await checkUnitConsistency()
}

const handleSetPressure = async (_deviceId: string, pressure: number) => {
  // 打压操作已在 PressDevicePanel 内部通过 API 完成
  // 此处用于后续扩展（如更新进度状态）
  console.debug('设定压力:', pressure)
}

// 稳定倒计时
const stableCountdown = ref(0)
let countdownTimer: ReturnType<typeof setInterval> | null = null

const startCountdown = () => {
  if (countdownTimer) return
  countdownTimer = setInterval(() => {
    if (calibrationStore.isStable || !calibrationStore.isRunning) {
      stableCountdown.value = 0
      return
    }
    stableCountdown.value = Math.max(0, stableCountdown.value - 1)
  }, 1000)
}

// 启动前置条件
const prerequisites = computed(() => [
  { label: '设备已选择', satisfied: calibrationStore.device1604Connected },
  { label: '打压设备已连接', satisfied: calibrationStore.pressDeviceConnected },
  { label: '已选择采集通道', satisfied: calibrationStore.channelsSelected },
  { label: '单位一致', satisfied: unitConsistent.value }
])

// 会话状态
const sessionState = computed(() => calibrationStore.sessionState)

// 回程模式：正程/回程判定
const isForwardPoint = (pointIndex: number): boolean => {
  const total = calibrationStore.pressurePoints.length
  const splitIndex = Math.ceil(total / 2)
  return pointIndex <= splitIndex
}

// 更新目标值
const updateTargetValue = (pointId: string, value: number) => {
  const point = calibrationStore.pressurePoints.find(p => p.id === pointId)
  if (point) {
    point.targetPressure = value
  }
}

// 会话状态中文映射
const sessionStateTextMap: Record<SessionState, string> = {
  idle: '空闲',
  ready: '就绪',
  pressurizing: '打压中',
  stabilizing: '稳定中',
  collecting: '采集中',
  point_done: '点完成',
  fitting: '拟合中',
  completed: '已完成',
  paused: '已暂停',
  stopped: '已停止',
  await_manual_collect: '等待手动采集',
  await_alarm_resolution: '等待报警处理',
  recovering: '恢复中',
  error: '错误'
}

const sessionStateText = computed(() => sessionStateTextMap[sessionState.value] || sessionState.value)

// 进度
const completedCount = computed(() =>
  calibrationStore.pressurePoints.filter(p => p.status === 'completed').length
)
const progressPercent = computed(() => {
  const total = calibrationStore.pressurePoints.length
  if (total === 0) return 0
  return Math.round((completedCount.value / total) * 100)
})

// 手动采集条件：手动模式 + 压力已稳定 + 正在运行
const canManualCollect = computed(() =>
  controlMode.value === 'manual' &&
  calibrationStore.isRunning &&
  calibrationStore.isStable
)

// 手动采集：对当前测点执行采集
const handleManualCollect = () => {
  const currentPoint = calibrationStore.pressurePoints.find(
    p => p.status === 'pending' || p.status === 'stabilizing' || p.status === 'pressurizing'
  )
  if (currentPoint) {
    calibrationStore.collectData(currentPoint.id)
  }
}

// 导出报告条件：采集已完成 + 有测点数据 + 设备已连接
const canExportReport = computed(() =>
  sessionState.value === 'completed' &&
  calibrationStore.pressurePoints.length > 0 &&
  calibrationStore.device1604Connected &&
  calibrationStore.pressDeviceConnected
)

// 测点状态
const getPointStatusType = (status: string) => {
  const map: Record<string, string> = {
    pending: 'info',
    pressurizing: 'warning',
    stabilizing: '',
    collecting: 'primary',
    completed: 'success',
    error: 'danger'
  }
  return map[status] || 'info'
}

const getPointStatusText = (status: string) => {
  const map: Record<string, string> = {
    pending: '待执行',
    pressurizing: '打压中',
    stabilizing: '稳定中',
    collecting: '采集中',
    completed: '已完成',
    error: '错误'
  }
  return map[status] || status
}

// 表格数据
interface TableRow {
  id: string
  index: number
  status: string
  targetValue: number
  channelValues: (number | undefined)[]
  actualPressure?: number
}

const tableData = computed<TableRow[]>(() =>
  calibrationStore.pressurePoints.map(point => ({
    id: point.id,
    index: point.index,
    status: point.status,
    targetValue: point.targetPressure,
    channelValues: point.collectedData || [],
    actualPressure: point.actualPressure
  }))
)

const getChannelClass = (row: TableRow, index: number) => {
  const value = row.channelValues[index]
  if (value === undefined) return ''
  const diff = Math.abs(value - row.targetValue)
  if (diff < 0.1) return 'channel-good'
  if (diff < 0.5) return 'channel-warning'
  return 'channel-error'
}

// 报告模板
const showTemplateDialog = ref(false)
const templatePoints = ref(5)
const templateMode = ref<'single' | 'return'>('single')
const templateFilename = ref('')

function handleSelectTemplate() {
  showTemplateDialog.value = true
}

async function confirmTemplate() {
  errorMessage.value = ''
  try {
    const result = await selectReportTemplate(templatePoints.value, templateMode.value)
    templateFilename.value = result.filename
    showTemplateDialog.value = false
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '模板选择失败'
  }
}

// 导出 CSV
const exportCSV = () => {
  const points = calibrationStore.pressurePoints
  if (points.length === 0) {
    ElMessage.warning('没有可导出的数据')
    return
  }

  const channels = calibrationStore.selectedChannels
  const headers = ['序号', '目标压力', '实际压力', '状态', ...channels.map(ch => `CH${ch}`)]
  const rows = points.map(p => [
    p.index,
    p.targetPressure.toFixed(2),
    p.actualPressure?.toFixed(2) || '--',
    getPointStatusText(p.status),
    ...channels.map(ch => p.collectedData?.[ch - 1]?.toFixed(4) || '--')
  ])

  const csvContent = [headers.join(','), ...rows.map(r => r.join(','))].join('\n')
  const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `measurement_data_${new Date().toISOString().split('T')[0]}.csv`
  link.click()
  ElMessage.success('报告已导出')
}

// SSE
let eventSource: EventSource | null = null

function setupSSE() {
  eventSource = createEventStream((payload: StreamEventPayload) => {
    if (payload.type === 'session.state.changed') {
      const data = payload.data as { state: SessionState }
      if (data?.state) {
        calibrationStore.syncSessionState(data.state)
      }
    }
    // 设备变更时刷新设备列表，确保添加/删除设备后下拉框及时更新
    if (payload.type === 'device.status.changed') {
      void deviceStore.loadDevices(true).then(async () => {
        if (calibrationStore.device1604Connected) {
          await calibrationStore.refreshDeviceInfo({ retries: 2, retryDelayMs: 250 })
        }
      })
    }
  })
}

// 轮询
let pollTimer: ReturnType<typeof setInterval> | null = null
let deviceRefreshTimer: ReturnType<typeof setInterval> | null = null

const startPolling = () => {
  if (pollTimer) return
  pollTimer = setInterval(async () => {
    if (calibrationStore.isRunning) {
      await Promise.all([
        calibrationStore.refreshPressure(),
        calibrationStore.refreshStability(),
        calibrationStore.refreshMeasureData()
      ])
    } else {
      // Session 未运行时也刷新已连接打压设备的压力值
      await deviceStore.refreshAllConnectedPressures()
    }
  }, 2000)
}

const startDeviceRefresh = () => {
  if (deviceRefreshTimer) return
  deviceRefreshTimer = setInterval(() => {
    deviceStore.loadDevices(true)
  }, 5000)
}

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  if (deviceRefreshTimer) {
    clearInterval(deviceRefreshTimer)
    deviceRefreshTimer = null
  }
}

onMounted(async () => {
  await deviceStore.loadDevices()
  await calibrationStore.fetchCurrentSessionState()
  await checkUnitConsistency()
  // 如果已有计量设备处于已连接状态但尚未绑定到校准服务，自动补绑定
  const connectedMeasureDev = deviceStore.measureDevices.find(d => d.status === 'connected')
  if (connectedMeasureDev) {
    await calibrationStore.connectDevice1604(connectedMeasureDev.id)
  }
  setupSSE()
  startPolling()
  startDeviceRefresh()
  startCountdown()
})

onUnmounted(() => {
  if (eventSource) {
    eventSource.close()
    eventSource = null
  }
  stopPolling()
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
})
</script>

<style scoped lang="scss">
/* ========================================
   计量工作台 - 布局与间距系统重构
   设计原则：可靠、克制、清晰
   间距层级：xs:4px | sm:8px | md:12px | lg:16px | xl:24px | 2xl:32px | 3xl:48px
   ======================================== */

.module-page {
  padding: var(--spacing-lg);
  box-sizing: border-box;
  height: 100%;
  overflow: hidden;
  background: var(--bg-primary);
}

.desktop-shell {
  max-width: 100%;
  height: 100%;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

/* ===== 页头 ===== */
.module-header {
  align-items: flex-end;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  gap: var(--spacing-lg);
  justify-content: space-between;
  padding-bottom: var(--spacing-lg);
  flex-shrink: 0;
  min-height: 52px;
}

.module-caption {
  color: var(--accent-primary);
  font-size: 11px;
  letter-spacing: 0.08em;
  margin: 0 0 var(--spacing-xs);
  text-transform: uppercase;
  font-weight: 600;
}

.module-header h1 {
  color: var(--text-primary);
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.module-switch {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
}

.switch-btn {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  padding: 6px 14px;
  text-decoration: none;
  font-size: 13px;
  transition: all 0.15s ease;
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);

  &:hover {
    background: var(--bg-quaternary);
    color: var(--text-primary);
  }

  .el-icon {
    font-size: 12px;
  }
}

.switch-btn.active {
  background: var(--accent-primary);
  border-color: var(--accent-primary);
  color: var(--bg-primary);
  font-weight: 600;
}

.switch-btn-ghost {
  background: transparent;
}

/* ===== 主布局: 侧边栏 + 工作台 ===== */
.measurement-layout {
  flex: 1;
  min-height: 0;
  display: flex;
  gap: 0;
  background: var(--bg-primary);
}

.sidebar {
  width: 280px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  position: relative;
  transition: width 0.25s ease;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;

  &.collapsed {
    width: 32px;
  }
}

.sidebar-toggle {
  position: absolute;
  right: -12px;
  top: 50%;
  transform: translateY(-50%);
  width: 12px;
  height: 36px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-left: none;
  border-radius: 0 var(--radius-sm) var(--radius-sm) 0;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 10;

  .el-icon {
    color: var(--text-secondary);
    font-size: 10px;
  }

  &:hover {
    background: var(--bg-quaternary);

    .el-icon {
      color: var(--accent-primary);
    }
  }
}

.sidebar-content {
  padding: var(--spacing-md);
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

.sidebar-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.sidebar-title {
  color: var(--text-secondary);
  font-size: 11px;
  font-weight: 600;
  margin: 0;
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  text-transform: uppercase;
  letter-spacing: 0.06em;

  .el-icon {
    color: var(--accent-primary);
    font-size: 13px;
  }
}

.channel-count {
  margin-left: auto;
  color: var(--text-muted);
  font-weight: 400;
  text-transform: none;
  letter-spacing: 0;
  font-size: 11px;
}

/* ===== 单位一致性检查 ===== */
.unit-check-result {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: 12px;
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--radius-sm);

  &.unit-ok {
    background: var(--status-success-bg);
    color: var(--status-success);
  }

  &.unit-warn {
    background: var(--status-warning-bg);
    color: var(--status-warning);
  }
}

.unit-conflicts {
  display: flex;
  flex-direction: column;
  gap: 2px;

  .conflict-item {
    color: var(--status-error);
    font-size: 11px;
    padding: 2px 0;
  }
}

/* ===== 前置条件列表 ===== */
.prerequisites-list {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: var(--spacing-sm);
  display: flex;
  flex-direction: column;
  gap: 4px;

  .prereq-item {
    display: flex;
    align-items: center;
    gap: var(--spacing-xs);
    padding: 2px 0;
    color: var(--text-muted);
    font-size: 12px;

    .el-icon {
      font-size: 13px;
    }

    &.satisfied {
      color: var(--status-success);
    }
  }
}

/* ===== 工作台 ===== */
.workbench {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
  padding: var(--spacing-md) var(--spacing-lg);
  overflow-y: auto;
  overflow-x: hidden;
}

/* ===== 控制面板（参数+模式+进度+操作聚合） ===== */
.control-panel {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  overflow: hidden;
}

.control-panel-top {
  padding: var(--spacing-md) var(--spacing-lg);
  display: flex;
  align-items: flex-end;
  gap: var(--spacing-xl);
  flex-wrap: wrap;
  border-bottom: 1px solid var(--border-color);
}

.param-group {
  display: flex;
  gap: var(--spacing-md);
  align-items: flex-end;
}

.param-divider {
  width: 1px;
  height: 32px;
  background: var(--border-color-strong);
  flex-shrink: 0;
  align-self: center;
}

.control-group {
  display: flex;
  flex-direction: column;
  gap: 4px;

  label {
    color: var(--text-muted);
    font-size: 11px;
    font-weight: 500;
    white-space: nowrap;
  }

  :deep(.el-input-number) {
    width: 90px;
  }

  :deep(.el-select) {
    width: 85px;
  }

  .custom-precision-check {
    margin-top: 2px;

    :deep(.el-checkbox__label) {
      font-size: 11px;
      color: var(--text-muted);
    }
  }
}

.generate-btn {
  margin-left: auto;
  flex-shrink: 0;
}

/* 控制面板底部 - 模式、进度、报警、操作 */
.control-panel-bottom {
  padding: var(--spacing-sm) var(--spacing-lg);
  background: var(--bg-tertiary);
  display: grid;
  grid-template-columns: auto 1fr auto auto;
  align-items: center;
  gap: var(--spacing-xl);
}

.mode-switches {
  display: flex;
  gap: var(--spacing-lg);
}

.switch-group {
  display: flex;
  flex-direction: column;
  gap: 2px;

  > span {
    color: var(--text-muted);
    font-size: 11px;
    font-weight: 500;
  }
}

.progress-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 180px;
  max-width: 280px;

  .progress-info {
    display: flex;
    align-items: center;
    gap: var(--spacing-sm);
    color: var(--text-secondary);
    font-size: 13px;

    .el-progress {
      flex: 1;
      min-width: 80px;
    }
  }

  .stable-status {
    display: flex;
    gap: var(--spacing-md);
    color: var(--text-muted);
    font-size: 11px;
    align-items: center;

    .realtime-pressure {
      color: var(--text-primary);
      font-weight: 600;
      font-variant-numeric: tabular-nums;
    }

    .stable-badge {
      padding: 2px 6px;
      border-radius: 2px;
      font-weight: 500;
      font-size: 11px;

      &.stable-yes {
        background: var(--status-success-bg);
        color: var(--status-success);
      }

      &.stable-no {
        background: var(--status-warning-bg);
        color: var(--status-warning);
      }
    }

    .countdown {
      color: var(--status-warning);
      font-weight: 600;
      font-variant-numeric: tabular-nums;
    }

    .session-state {
      color: var(--accent-primary);
      font-weight: 500;
    }
  }
}

/* ===== 报警设置 ===== */
.alarm-settings {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding-left: var(--spacing-lg);
  border-left: 1px solid var(--border-color);

  .alarm-label {
    color: var(--text-muted);
    font-size: 11px;
    font-weight: 500;
  }

  :deep(.el-switch) {
    --el-switch-on-color: var(--status-warning);
    --el-switch-off-color: var(--bg-quaternary);

    .el-switch__label {
      font-size: 11px;
      color: var(--text-muted);

      &.is-active {
        color: var(--status-warning);
      }
    }
  }
}

.session-status-inline {
  display: flex;
  align-items: center;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 600;
}

.status-idle { background: var(--bg-quaternary); color: var(--text-secondary); }
.status-ready { background: var(--status-info-bg); color: var(--status-info); }

.status-pressurizing,
.status-stabilizing,
.status-collecting,
.status-point_done,
.status-fitting,
.status-await_manual_collect,
.status-await_alarm_resolution,
.status-recovering { background: var(--status-success-bg); color: var(--status-success); }

.status-paused { background: var(--status-warning-bg); color: var(--status-warning); }
.status-completed { background: var(--status-success-bg); color: var(--status-success); }
.status-stopped { background: var(--bg-quaternary); color: var(--text-secondary); }
.status-error { background: var(--status-error-bg); color: var(--status-error); }

/* ===== 操作按钮组 ===== */
.action-buttons {
  display: flex;
  gap: var(--spacing-xs);
  flex-wrap: wrap;
  align-items: center;

  > :nth-child(4) {
    margin-left: var(--spacing-xs);
  }

  > :nth-child(6) {
    margin-left: var(--spacing-xs);
  }
}

/* ===== 数据表格区 ===== */
.data-table-wrapper {
  flex: 1;
  min-height: 0;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: var(--spacing-md);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-sm);
  flex-shrink: 0;
  padding-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--border-color);
}

.table-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);

  .el-icon {
    color: var(--accent-primary);
    font-size: 16px;
  }

  h3 {
    color: var(--text-primary);
    margin: 0;
    font-size: 14px;
    font-weight: 600;
  }
}

.record-count {
  color: var(--text-muted);
  font-size: 12px;
  margin-left: var(--spacing-xs);
}

.table-actions {
  display: flex;
  gap: var(--spacing-xs);
}

.data-table {
  width: 100%;
  flex: 1;
  min-height: 0;

  :deep(.el-table) {
    --el-table-bg-color: var(--bg-tertiary);
    --el-table-tr-bg-color: var(--bg-tertiary);
    --el-table-header-bg-color: var(--bg-quaternary);
    --el-table-row-hover-bg-color: rgba(255, 215, 0, 0.06);
    --el-table-current-row-bg-color: rgba(255, 215, 0, 0.1);
    --el-table-border-color: var(--border-color-strong);
    --el-table-text-color: var(--text-primary);
    --el-table-header-text-color: var(--text-secondary);
    background-color: var(--bg-tertiary);
  }

  :deep(th.el-table__cell) {
    background: var(--bg-quaternary) !important;
    color: var(--text-secondary) !important;
    font-size: 12px;
    font-weight: 600;
    padding: 8px 0;
    border-bottom: 1px solid var(--border-color-strong);
  }

  :deep(td.el-table__cell) {
    background: var(--bg-tertiary);
    color: var(--text-primary);
    font-size: 13px;
    padding: 6px 0;
    border-bottom: 1px solid var(--border-color);
  }

  :deep(.el-table__row--striped td.el-table__cell) {
    background: var(--bg-secondary) !important;
  }

  :deep(.el-table__row:hover td.el-table__cell) {
    background: rgba(255, 215, 0, 0.06) !important;
  }

  :deep(.el-table--border .el-table__cell) {
    border-right: 1px solid var(--border-color);
  }

  :deep(.el-table__inner-wrapper::before) {
    background-color: var(--bg-quaternary);
  }

  :deep(.el-table__empty-block) {
    background: var(--bg-tertiary);
  }

  .channel-good {
    color: var(--status-success);
    font-family: 'Consolas', monospace;
  }

  .channel-warning {
    color: var(--status-warning);
    font-family: 'Consolas', monospace;
  }

  .channel-error {
    color: var(--status-error);
    font-family: 'Consolas', monospace;
  }
}

.row-actions {
  display: flex;
  gap: var(--spacing-xs);
}

.trip-badge {
  margin-left: var(--spacing-xs);
  font-size: 10px;
  padding: 0 4px;
  height: 18px;
  line-height: 18px;
}

.collecting-text {
  color: var(--text-muted);
  font-size: 12px;
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-xl);
  color: var(--text-muted);

  .empty-icon {
    font-size: 48px;
    color: var(--bg-quaternary);
  }

  p {
    font-size: 13px;
    margin: 0;
  }
}

.template-result-bar {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-xs) var(--spacing-md);
  background: rgba(78, 201, 176, 0.08);
  border: 1px solid rgba(78, 201, 176, 0.3);
  border-radius: var(--radius-sm);
  font-size: 13px;
  flex-shrink: 0;

  .el-icon {
    color: var(--status-success);
  }

  span {
    color: var(--text-secondary);
  }
}

.error-message {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  color: var(--status-error);
  font-size: 13px;
  padding: var(--spacing-sm) var(--spacing-md);
  background: rgba(244, 71, 71, 0.08);
  border-radius: var(--radius-sm);
  flex-shrink: 0;

  .el-icon {
    font-size: 14px;
  }
}

/* ===== 报警通道选择对话框 ===== */
.channel-dialog-tip {
  color: var(--text-secondary);
  font-size: 13px;
  margin: 0 0 var(--spacing-md);
}

.channel-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--spacing-xs);
}

.channel-tile {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-xs) var(--spacing-sm);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  color: var(--text-secondary);
  transition: all 0.15s;

  &:hover {
    border-color: var(--accent-primary);
  }

  &.selected {
    background: rgba(201, 164, 95, 0.12);
    border-color: var(--accent-primary);
    color: var(--text-primary);

    .el-icon {
      color: var(--accent-primary);
    }
  }
}

/* ===== 响应式适配 ===== */
@media (max-width: 1400px) {
  .control-panel-bottom {
    grid-template-columns: auto 1fr auto;
    gap: var(--spacing-lg);
  }

  .alarm-settings {
    grid-column: 1 / -1;
    justify-content: flex-start;
    border-left: none;
    border-top: 1px solid var(--border-color);
    padding-left: 0;
    padding-top: var(--spacing-sm);
  }
}

@media (max-width: 1200px) {
  .control-group {
    :deep(.el-input-number) {
      width: 80px;
    }
  }

  .control-panel-bottom {
    grid-template-columns: 1fr 1fr;
    gap: var(--spacing-md);
  }

  .mode-switches {
    grid-column: 1 / -1;
  }

  .progress-section {
    max-width: none;
  }

  .action-buttons {
    grid-column: 1 / -1;
    justify-content: flex-end;
  }

  .alarm-settings {
    grid-column: 1 / -1;
  }
}

@media (max-width: 900px) {
  .module-page {
    padding: var(--spacing-md);
    overflow: auto;
  }

  .desktop-shell {
    max-width: 100%;
    height: auto;
    gap: var(--spacing-md);
  }

  .module-header {
    flex-direction: column;
    gap: var(--spacing-md);
  }

  .measurement-layout {
    flex-direction: column;
    min-height: auto;
  }

  .sidebar {
    width: 100% !important;
    border-right: none;
    border-bottom: 1px solid var(--border-color);

    .sidebar-toggle {
      display: none;
    }
  }

  .sidebar.collapsed {
    width: 100% !important;
  }

  .workbench {
    overflow: visible;
    padding: var(--spacing-md);
  }

  .control-panel-top {
    flex-direction: column;
    align-items: stretch;
    gap: var(--spacing-md);
  }

  .param-group {
    flex-wrap: wrap;
    justify-content: flex-start;
  }

  .param-divider {
    display: none;
  }

  .generate-btn {
    margin-left: 0;
    width: 100%;
  }

  .control-panel-bottom {
    grid-template-columns: 1fr;
    gap: var(--spacing-md);
  }

  .mode-switches {
    flex-direction: column;
    gap: var(--spacing-sm);
  }

  .progress-section {
    min-width: auto;
  }

  .action-buttons {
    justify-content: center;
    flex-wrap: wrap;
  }

  .alarm-settings {
    flex-wrap: wrap;
    justify-content: center;
  }

  .data-table-wrapper {
    min-height: 300px;
  }
}
</style>
