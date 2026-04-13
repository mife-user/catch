<template>
  <transition name="progress-fade">
    <div v-if="visible" class="progress-bar-wrapper" :class="position">
      <div class="progress-bar-inner">
        <div class="progress-header">
          <div class="progress-info">
            <el-icon v-if="!isPaused" class="progress-spin is-loading"><Loading /></el-icon>
            <el-icon v-else class="progress-pause"><VideoPause /></el-icon>
            <span class="progress-title">{{ title }}</span>
            <span v-if="currentItem" class="progress-current">{{ currentItem }}</span>
          </div>
          <div class="progress-actions">
            <el-button
              v-if="showPause"
              size="small"
              :icon="isPaused ? VideoPlay : VideoPause"
              circle
              @click="$emit('toggle-pause')"
            />
            <el-button
              v-if="showCancel"
              size="small"
              :icon="Close"
              circle
              type="danger"
              @click="$emit('cancel')"
            />
            <el-button
              v-if="showDetail"
              size="small"
              :icon="expanded ? ArrowUp : ArrowDown"
              circle
              @click="expanded = !expanded"
            />
          </div>
        </div>

        <div class="progress-body">
          <el-progress
            :percentage="displayPercentage"
            :status="status"
            :stroke-width="strokeWidth"
            :show-text="true"
            :format="formatProgress"
          />
          <div class="progress-stats">
            <span v-if="done !== undefined && total !== undefined" class="stat-item">
              {{ done }} / {{ total }}
            </span>
            <span v-if="estimatedTime" class="stat-item">
              预计剩余: {{ estimatedTime }}
            </span>
            <span v-if="speed" class="stat-item">
              {{ speed }}
            </span>
          </div>
        </div>

        <transition name="detail-expand">
          <div v-if="expanded && details.length > 0" class="progress-details">
            <div v-for="(detail, idx) in details" :key="idx" class="detail-item" :class="{ 'detail-error': detail.type === 'error', 'detail-important': detail.important }">
              <el-icon v-if="detail.type === 'success'" class="detail-icon success"><CircleCheck /></el-icon>
              <el-icon v-else-if="detail.type === 'error'" class="detail-icon error"><CircleClose /></el-icon>
              <el-icon v-else class="detail-icon info"><InfoFilled /></el-icon>
              <span class="detail-text">{{ detail.message }}</span>
            </div>
          </div>
        </transition>
      </div>
    </div>
  </transition>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import {
  Loading, VideoPause, VideoPlay, Close, ArrowUp, ArrowDown,
  CircleCheck, CircleClose, InfoFilled
} from '@element-plus/icons-vue'

const props = defineProps({
  visible: { type: Boolean, default: false },
  title: { type: String, default: '处理中...' },
  percentage: { type: Number, default: 0 },
  done: { type: Number, default: undefined },
  total: { type: Number, default: undefined },
  currentItem: { type: String, default: '' },
  isPaused: { type: Boolean, default: false },
  position: { type: String, default: 'inline', validator: v => ['inline', 'top', 'side'].includes(v) },
  status: { type: String, default: '' },
  strokeWidth: { type: Number, default: 18 },
  showPause: { type: Boolean, default: true },
  showCancel: { type: Boolean, default: true },
  showDetail: { type: Boolean, default: true },
  details: { type: Array, default: () => [] },
  startTime: { type: Number, default: 0 },
})

defineEmits(['cancel', 'toggle-pause'])

const expanded = ref(false)

const displayPercentage = computed(() => {
  if (props.percentage > 0) return Math.min(props.percentage, 100)
  if (props.done !== undefined && props.total !== undefined && props.total > 0) {
    return Math.round((props.done / props.total) * 100)
  }
  return 0
})

const estimatedTime = computed(() => {
  if (!props.startTime || displayPercentage.value <= 0 || displayPercentage.value >= 100) return ''
  const elapsed = (Date.now() - props.startTime) / 1000
  const remaining = (elapsed / displayPercentage.value) * (100 - displayPercentage.value)
  if (remaining < 60) return Math.round(remaining) + ' 秒'
  if (remaining < 3600) return Math.round(remaining / 60) + ' 分钟'
  return Math.round(remaining / 3600) + ' 小时'
})

const speed = computed(() => {
  if (!props.startTime || !props.done) return ''
  const elapsed = (Date.now() - props.startTime) / 1000
  if (elapsed < 1) return ''
  const rate = props.done / elapsed
  return rate.toFixed(1) + ' 项/秒'
})

const formatProgress = (percentage) => {
  return percentage + '%'
}

watch(() => props.visible, (val) => {
  if (!val) expanded.value = false
})
</script>

<style scoped>
.progress-fade-enter-active,
.progress-fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.progress-fade-enter-from,
.progress-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

.progress-bar-wrapper {
  width: 100%;
}

.progress-bar-wrapper.top {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 2000;
  padding: 12px 24px;
  background: #ffffff;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.progress-bar-wrapper.side {
  position: fixed;
  right: 24px;
  bottom: 24px;
  width: 360px;
  z-index: 2000;
}

.progress-bar-wrapper.inline {
  margin-bottom: 20px;
}

.progress-bar-inner {
  background: #ffffff;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  border: 1px solid #e4e7ed;
}

.progress-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.progress-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.progress-spin {
  color: #1890ff;
  font-size: 16px;
}

.progress-pause {
  color: #faad14;
  font-size: 16px;
}

.progress-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  white-space: nowrap;
}

.progress-current {
  font-size: 12px;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.progress-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.progress-body {
  margin-bottom: 4px;
}

.progress-stats {
  display: flex;
  gap: 16px;
  margin-top: 8px;
}

.stat-item {
  font-size: 12px;
  color: #909399;
}

.detail-expand-enter-active,
.detail-expand-leave-active {
  transition: max-height 0.3s ease;
  overflow: hidden;
}

.detail-expand-enter-from,
.detail-expand-leave-to {
  max-height: 0;
}

.detail-expand-enter-to,
.detail-expand-leave-from {
  max-height: 300px;
}

.progress-details {
  margin-top: 12px;
  border-top: 1px solid #e4e7ed;
  padding-top: 8px;
  max-height: 200px;
  overflow-y: auto;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 0;
  font-size: 12px;
  color: #606266;
}

.detail-item.detail-error {
  color: #f5222d;
}

.detail-item.detail-important {
  color: #f5222d;
  font-weight: 500;
}

.detail-icon {
  font-size: 14px;
  flex-shrink: 0;
}

.detail-icon.success {
  color: #52c41a;
}

.detail-icon.error {
  color: #f5222d;
}

.detail-icon.info {
  color: #1890ff;
}

.detail-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (max-width: 768px) {
  .progress-bar-wrapper.side {
    right: 8px;
    bottom: 8px;
    width: calc(100% - 16px);
  }
}
</style>
