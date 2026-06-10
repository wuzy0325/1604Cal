<template>
  <div class="floating-log-panel" :class="{ expanded }">
    <div class="log-bar" @click="toggleExpand">
      <div class="bar-left">
        <el-icon :size="16"><Monitor /></el-icon>
        <span class="bar-title">通讯日志</span>
        <el-tag size="small" type="info" class="bar-count">{{ store.count }}</el-tag>
        <span v-if="hasNewEntries && !expanded" class="new-dot" />
      </div>
      <div class="bar-right">
        <el-button size="small" text @click.stop="store.clear()">清空</el-button>
        <el-button size="small" text class="bar-toggle">
          <el-icon><ArrowDown v-if="!expanded" /><ArrowUp v-else /></el-icon>
        </el-button>
      </div>
    </div>

    <div v-show="expanded" class="resize-handle" @pointerdown.prevent="startResize" :class="{ active: isResizing }" />

    <div v-show="expanded" class="log-body" ref="logBodyRef">
      <CommLogPanel />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ArrowDown, ArrowUp, Monitor } from '@element-plus/icons-vue'
import { useHardwareLogStore } from '@/stores/hardwareLog'
import CommLogPanel from '@/components/common/CommLogPanel.vue'

const store = useHardwareLogStore()

const expanded = ref(false)
const panelHeight = ref(200)
const isResizing = ref(false)
const logBodyRef = ref<HTMLElement | null>(null)

function toggleExpand() {
  expanded.value = !expanded.value
  if (expanded.value && logBodyRef.value && !logBodyRef.value.style.height) {
    logBodyRef.value.style.height = panelHeight.value + 'px'
  }
}

function startResize(e: PointerEvent) {
  if (isResizing.value) return
  isResizing.value = true
  const startY = e.clientY
  const startHeight = logBodyRef.value?.offsetHeight ?? panelHeight.value

  function onMove(e: MouseEvent) {
    const delta = startY - e.clientY
    const newHeight = Math.max(120, Math.min(startHeight + delta, window.innerHeight * 0.6))
    panelHeight.value = newHeight
    if (logBodyRef.value) {
      logBodyRef.value.style.height = newHeight + 'px'
    }
  }

  function onUp() {
    isResizing.value = false
    document.removeEventListener('pointermove', onMove)
    document.removeEventListener('pointerup', onUp)
  }

  document.addEventListener('pointermove', onMove)
  document.addEventListener('pointerup', onUp)
}

const hasNewEntries = ref(false)
watch(() => store.entries.length, () => {
  if (!expanded.value) {
    hasNewEntries.value = true
  }
})

watch(expanded, (val) => {
  if (val) hasNewEntries.value = false
})
</script>

<style scoped lang="scss">
.floating-log-panel {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: #1e1e1e;
  border-top: 2px solid #10b981;
  position: relative;
  z-index: 100;

  &:not(.expanded) {
    .log-bar {
      border-radius: 8px 8px 0 0;
    }
  }
}

.log-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 32px;
  padding: 0 12px;
  background: #2d2d2d;
  cursor: pointer;
  user-select: none;
  flex-shrink: 0;

  &:hover {
    background: #333;
  }
}

.bar-left {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #e0e0e0;
}

.bar-title {
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.03em;
}

.bar-count {
  font-size: 11px;
}

.new-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #10b981;
  animation: pulse-dot 1.2s ease-in-out infinite;
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(0.8); }
}

.bar-right {
  display: flex;
  align-items: center;
  gap: 2px;
}

.bar-toggle {
  color: #9ca3af;
}

.resize-handle {
  height: 4px;
  cursor: ns-resize;
  background: transparent;
  flex-shrink: 0;
  position: relative;
  z-index: 1;

  &::after {
    content: '';
    position: absolute;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    width: 40px;
    height: 2px;
    background: #3c3c3c;
    border-radius: 1px;
    transition: background 0.15s ease;
  }

  &:hover::after,
  &.active::after {
    background: #10b981;
    width: 60px;
  }
}

.log-body {
  flex: 1;
  min-height: 0;
  display: flex;
  overflow: hidden;
}
</style>
