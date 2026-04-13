<template>
  <div class="unified-select">
    <div class="select-trigger" @click="openDialog">
      <el-input
        :model-value="displayValue"
        :placeholder="placeholder"
        readonly
        clearable
        @clear="handleClear"
      >
        <template #prefix>
          <el-icon v-if="prefixIcon"><component :is="prefixIcon" /></el-icon>
        </template>
        <template #suffix>
          <el-icon class="select-arrow"><ArrowDown /></el-icon>
        </template>
      </el-input>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="640px"
      :close-on-click-modal="false"
      class="unified-select-dialog"
    >
      <div class="dialog-body">
        <div v-if="searchable" class="search-bar">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索选项..."
            clearable
            prefix-icon="Search"
          />
        </div>

        <div v-if="mode === 'path'" class="path-browser">
          <div class="path-input-row">
            <el-input v-model="pathInput" placeholder="输入路径" @keyup.enter="loadPath">
              <template #prepend>路径</template>
              <template #append>
                <el-button @click="loadPath">前往</el-button>
              </template>
            </el-input>
          </div>
          <div class="path-list">
            <div class="path-item parent-item" @click="goToParent">
              <el-icon><FolderOpened /></el-icon>
              <span>.. (上级目录)</span>
            </div>
            <div
              v-for="item in filteredPathItems"
              :key="item.path"
              class="path-item"
              :class="{ selected: pathInput === item.path }"
              @click="selectPathItem(item)"
            >
              <el-icon><Folder /></el-icon>
              <span>{{ item.name }}</span>
            </div>
            <div v-if="pathItems.length === 0" class="empty-hint">
              该目录下没有子目录
            </div>
          </div>
          <div v-if="recentPaths.length > 0" class="recent-section">
            <div class="section-title">最近使用</div>
            <div class="recent-list">
              <el-tag
                v-for="p in recentPaths"
                :key="p"
                class="recent-tag"
                @click="selectRecentPath(p)"
              >
                {{ p }}
              </el-tag>
            </div>
          </div>
        </div>

        <div v-if="mode === 'fileType'" class="option-browser">
          <div class="option-grid">
            <div
              v-for="opt in filteredOptions"
              :key="opt.value"
              class="option-card"
              :class="{ selected: innerValue === opt.value, important: opt.important }"
              @click="selectOption(opt)"
            >
              <el-icon :size="24" :color="opt.color || '#909399'">
                <component :is="opt.icon || Document" />
              </el-icon>
              <span class="option-label">{{ opt.label }}</span>
              <span v-if="opt.desc" class="option-desc">{{ opt.desc }}</span>
            </div>
          </div>
        </div>

        <div v-if="mode === 'sizeRange'" class="size-browser">
          <div class="preset-section">
            <div class="section-title">常用大小范围</div>
            <div class="preset-grid">
              <div
                v-for="preset in sizePresets"
                :key="preset.label"
                class="preset-card"
                :class="{ selected: isSizePresetSelected(preset) }"
                @click="selectSizePreset(preset)"
              >
                <span class="preset-label">{{ preset.label }}</span>
                <span class="preset-desc">{{ preset.desc }}</span>
              </div>
            </div>
          </div>
          <div class="custom-section">
            <div class="section-title">自定义范围</div>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-input-number
                  v-model="customMinSize"
                  :min="0"
                  placeholder="最小值"
                  controls-position="right"
                  class="size-input"
                />
                <el-select v-model="minSizeUnit" class="size-unit">
                  <el-option label="B" value="B" />
                  <el-option label="KB" value="KB" />
                  <el-option label="MB" value="MB" />
                  <el-option label="GB" value="GB" />
                </el-select>
              </el-col>
              <el-col :span="12">
                <el-input-number
                  v-model="customMaxSize"
                  :min="0"
                  placeholder="最大值"
                  controls-position="right"
                  class="size-input"
                />
                <el-select v-model="maxSizeUnit" class="size-unit">
                  <el-option label="B" value="B" />
                  <el-option label="KB" value="KB" />
                  <el-option label="MB" value="MB" />
                  <el-option label="GB" value="GB" />
                </el-select>
              </el-col>
            </el-row>
          </div>
        </div>

        <div v-if="mode === 'dateRange'" class="date-browser">
          <div class="preset-section">
            <div class="section-title">常用日期范围</div>
            <div class="preset-grid">
              <div
                v-for="preset in datePresets"
                :key="preset.label"
                class="preset-card"
                :class="{ selected: isDatePresetSelected(preset) }"
                @click="selectDatePreset(preset)"
              >
                <span class="preset-label">{{ preset.label }}</span>
                <span class="preset-desc">{{ preset.desc }}</span>
              </div>
            </div>
          </div>
          <div class="custom-section">
            <div class="section-title">自定义范围</div>
            <el-row :gutter="16">
              <el-col :span="12">
                <el-date-picker
                  v-model="customStartDate"
                  type="date"
                  placeholder="开始日期"
                  value-format="YYYY-MM-DD"
                  style="width: 100%"
                />
              </el-col>
              <el-col :span="12">
                <el-date-picker
                  v-model="customEndDate"
                  type="date"
                  placeholder="结束日期"
                  value-format="YYYY-MM-DD"
                  style="width: 100%"
                />
              </el-col>
            </el-row>
          </div>
        </div>

        <div v-if="mode === 'generic'" class="option-browser">
          <div class="option-list">
            <div
              v-for="opt in filteredOptions"
              :key="opt.value"
              class="option-item"
              :class="{ selected: innerValue === opt.value, important: opt.important }"
              @click="selectOption(opt)"
            >
              <el-icon v-if="opt.icon" :color="opt.color || '#909399'">
                <component :is="opt.icon" />
              </el-icon>
              <span class="option-label">{{ opt.label }}</span>
              <span v-if="opt.desc" class="option-desc">{{ opt.desc }}</span>
              <el-icon v-if="innerValue === opt.value" class="check-icon"><Check /></el-icon>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="confirmSelection">确认选择</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import {
  ArrowDown, Folder, FolderOpened, Document, Picture,
  VideoCamera, Headset, Check, Search
} from '@element-plus/icons-vue'
import { browsePath as fetchBrowsePath } from '../api/files'
import { ElMessage } from 'element-plus'

const props = defineProps({
  modelValue: { type: [String, Number, Array], default: '' },
  mode: { type: String, default: 'generic', validator: v => ['path', 'fileType', 'sizeRange', 'dateRange', 'generic'].includes(v) },
  placeholder: { type: String, default: '请选择' },
  dialogTitle: { type: String, default: '选择' },
  options: { type: Array, default: () => [] },
  searchable: { type: Boolean, default: true },
  prefixIcon: { type: Object, default: null },
  minSize: { type: Number, default: 0 },
  maxSize: { type: Number, default: 0 },
  modAfter: { type: String, default: '' },
  modBefore: { type: String, default: '' },
})

const emit = defineEmits(['update:modelValue', 'update:minSize', 'update:maxSize', 'update:modAfter', 'update:modBefore', 'change'])

const dialogVisible = ref(false)
const searchKeyword = ref('')
const innerValue = ref(props.modelValue)

const pathInput = ref('')
const pathItems = ref([])
const pathParentPath = ref('')
const recentPaths = ref([])

const customMinSize = ref(0)
const customMaxSize = ref(0)
const minSizeUnit = ref('MB')
const maxSizeUnit = ref('MB')

const customStartDate = ref('')
const customEndDate = ref('')

const selectedDatePreset = ref('')

const sizePresets = [
  { label: '小于 1MB', desc: '小型文件', min: 0, max: 1048576, minUnit: 'B', maxUnit: 'MB' },
  { label: '1MB - 100MB', desc: '中等文件', min: 1048576, max: 104857600, minUnit: 'MB', maxUnit: 'MB' },
  { label: '100MB - 1GB', desc: '大文件', min: 104857600, max: 1073741824, minUnit: 'MB', maxUnit: 'GB' },
  { label: '大于 1GB', desc: '超大文件', min: 1073741824, max: 0, minUnit: 'GB', maxUnit: 'GB' },
]

const datePresets = [
  { label: '最近7天', desc: '一周内修改', value: '7d' },
  { label: '最近30天', desc: '一个月内修改', value: '30d' },
  { label: '最近90天', desc: '三个月内修改', value: '90d' },
  { label: '最近一年', desc: '一年内修改', value: '365d' },
]

const fileTypeOptions = [
  { label: '全部文件', value: 'all', icon: Document, color: '#909399', desc: '搜索所有文件类型' },
  { label: '文档', value: 'document', icon: Document, color: '#1890ff', desc: 'txt, doc, pdf, xls 等' },
  { label: '图片', value: 'image', icon: Picture, color: '#52c41a', desc: 'jpg, png, gif, svg 等' },
  { label: '视频', value: 'video', icon: VideoCamera, color: '#faad14', desc: 'mp4, avi, mkv 等' },
  { label: '音频', value: 'audio', icon: Headset, color: '#722ed1', desc: 'mp3, wav, flac 等' },
  { label: '自定义', value: 'custom', icon: Search, color: '#eb2f96', desc: '指定文件扩展名' },
]

const displayValue = computed(() => {
  if (props.mode === 'path') return props.modelValue || ''
  if (props.mode === 'fileType') {
    const opt = fileTypeOptions.find(o => o.value === props.modelValue)
    return opt ? opt.label : props.modelValue
  }
  if (props.mode === 'sizeRange') {
    if (!props.minSize && !props.maxSize) return ''
    const fmt = (v) => {
      if (v >= 1073741824) return (v / 1073741824).toFixed(1) + ' GB'
      if (v >= 1048576) return (v / 1048576).toFixed(1) + ' MB'
      if (v >= 1024) return (v / 1024).toFixed(1) + ' KB'
      return v + ' B'
    }
    if (props.minSize && props.maxSize) return fmt(props.minSize) + ' - ' + fmt(props.maxSize)
    if (props.minSize) return '≥ ' + fmt(props.minSize)
    return '≤ ' + fmt(props.maxSize)
  }
  if (props.mode === 'dateRange') {
    if (props.modAfter && props.modBefore) return props.modAfter + ' 至 ' + props.modBefore
    if (props.modAfter) return props.modAfter + ' 起'
    if (props.modBefore) return '至 ' + props.modBefore
    return ''
  }
  if (props.mode === 'generic') {
    const opt = props.options.find(o => o.value === props.modelValue)
    return opt ? opt.label : props.modelValue
  }
  return props.modelValue || ''
})

const filteredOptions = computed(() => {
  let list = props.mode === 'fileType' ? fileTypeOptions : props.options
  if (searchKeyword.value) {
    const kw = searchKeyword.value.toLowerCase()
    list = list.filter(o => o.label.toLowerCase().includes(kw) || (o.desc && o.desc.toLowerCase().includes(kw)))
  }
  return list
})

const filteredPathItems = computed(() => {
  if (!searchKeyword.value) return pathItems.value
  const kw = searchKeyword.value.toLowerCase()
  return pathItems.value.filter(i => i.name.toLowerCase().includes(kw))
})

watch(() => props.modelValue, (val) => {
  innerValue.value = val
})

onMounted(() => {
  loadRecentPaths()
})

const loadRecentPaths = () => {
  try {
    const stored = localStorage.getItem('catch_recent_paths')
    if (stored) recentPaths.value = JSON.parse(stored).slice(0, 5)
  } catch {}
}

const saveRecentPath = (path) => {
  if (!path) return
  const list = recentPaths.value.filter(p => p !== path)
  list.unshift(path)
  recentPaths.value = list.slice(0, 5)
  localStorage.setItem('catch_recent_paths', JSON.stringify(recentPaths.value))
}

const openDialog = () => {
  dialogVisible.value = true
  searchKeyword.value = ''
  if (props.mode === 'path') {
    pathInput.value = props.modelValue || ''
    if (pathInput.value) loadPath()
  }
  if (props.mode === 'sizeRange') {
    customMinSize.value = props.minSize ? convertBytesToUnit(props.minSize, minSizeUnit.value) : 0
    customMaxSize.value = props.maxSize ? convertBytesToUnit(props.maxSize, maxSizeUnit.value) : 0
  }
  if (props.mode === 'dateRange') {
    customStartDate.value = props.modAfter || ''
    customEndDate.value = props.modBefore || ''
    selectedDatePreset.value = ''
  }
}

const convertBytesToUnit = (bytes, unit) => {
  switch (unit) {
    case 'GB': return bytes / 1073741824
    case 'MB': return bytes / 1048576
    case 'KB': return bytes / 1024
    default: return bytes
  }
}

const convertToBytes = (val, unit) => {
  switch (unit) {
    case 'GB': return Math.round(val * 1073741824)
    case 'MB': return Math.round(val * 1048576)
    case 'KB': return Math.round(val * 1024)
    default: return Math.round(val)
  }
}

const loadPath = async () => {
  try {
    const data = await fetchBrowsePath(pathInput.value)
    pathItems.value = data.items || []
    pathInput.value = data.current_path || ''
    pathParentPath.value = data.parent_path || ''
  } catch (err) {
    ElMessage.error(err.message || '无法浏览该路径')
  }
}

const goToParent = () => {
  if (pathParentPath.value) {
    pathInput.value = pathParentPath.value
    loadPath()
  }
}

const selectPathItem = (item) => {
  pathInput.value = item.path
  loadPath()
}

const selectRecentPath = (p) => {
  pathInput.value = p
  loadPath()
}

const selectOption = (opt) => {
  innerValue.value = opt.value
}

const isSizePresetSelected = (preset) => {
  return props.minSize === preset.min && props.maxSize === preset.max
}

const selectSizePreset = (preset) => {
  customMinSize.value = convertBytesToUnit(preset.min, minSizeUnit.value)
  customMaxSize.value = convertBytesToUnit(preset.max, maxSizeUnit.value)
}

const isDatePresetSelected = (preset) => {
  return selectedDatePreset.value === preset.value
}

const selectDatePreset = (preset) => {
  selectedDatePreset.value = preset.value
  const days = parseInt(preset.value)
  const now = new Date()
  const start = new Date(now.getTime() - days * 24 * 60 * 60 * 1000)
  customStartDate.value = start.toISOString().split('T')[0]
  customEndDate.value = now.toISOString().split('T')[0]
}

const handleClear = () => {
  emit('update:modelValue', '')
  emit('change', '')
}

const confirmSelection = () => {
  if (props.mode === 'path') {
    saveRecentPath(pathInput.value)
    emit('update:modelValue', pathInput.value)
    emit('change', pathInput.value)
  } else if (props.mode === 'fileType') {
    emit('update:modelValue', innerValue.value)
    emit('change', innerValue.value)
  } else if (props.mode === 'sizeRange') {
    const minBytes = customMinSize.value ? convertToBytes(customMinSize.value, minSizeUnit.value) : 0
    const maxBytes = customMaxSize.value ? convertToBytes(customMaxSize.value, maxSizeUnit.value) : 0
    emit('update:minSize', minBytes)
    emit('update:maxSize', maxBytes)
    emit('change', { minSize: minBytes, maxSize: maxBytes })
  } else if (props.mode === 'dateRange') {
    emit('update:modAfter', customStartDate.value)
    emit('update:modBefore', customEndDate.value)
    emit('change', { modAfter: customStartDate.value, modBefore: customEndDate.value })
  } else {
    emit('update:modelValue', innerValue.value)
    emit('change', innerValue.value)
  }
  dialogVisible.value = false
}
</script>

<style scoped>
.unified-select {
  width: 100%;
}

.select-trigger {
  cursor: pointer;
}

.select-arrow {
  transition: transform 0.3s;
  cursor: pointer;
}

.dialog-body {
  min-height: 200px;
}

.search-bar {
  margin-bottom: 16px;
}

.path-browser .path-input-row {
  margin-bottom: 12px;
}

.path-list {
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  max-height: 280px;
  overflow-y: auto;
}

.path-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  cursor: pointer;
  transition: background-color 0.2s;
  border-radius: 4px;
}

.path-item:hover {
  background-color: #f5f7fa;
}

.path-item.selected {
  background-color: #ecf5ff;
}

.path-item .el-icon {
  color: #e6a23c;
  font-size: 18px;
}

.parent-item {
  color: #909399;
  border-bottom: 1px solid #e4e7ed;
}

.empty-hint {
  padding: 24px;
  text-align: center;
  color: #909399;
}

.recent-section {
  margin-top: 16px;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #606266;
  margin-bottom: 8px;
}

.recent-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.recent-tag {
  cursor: pointer;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.option-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.option-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 16px 12px;
  border: 2px solid #e4e7ed;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.option-card:hover {
  border-color: #1890ff;
  box-shadow: 0 2px 8px rgba(24, 144, 255, 0.15);
}

.option-card.selected {
  border-color: #1890ff;
  background-color: #ecf5ff;
}

.option-card.important {
  border-color: #f5222d;
}

.option-card.important.selected {
  border-color: #f5222d;
  background-color: #fff1f0;
}

.option-label {
  font-size: 14px;
  font-weight: 500;
}

.option-desc {
  font-size: 11px;
  color: #909399;
  text-align: center;
}

.option-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.option-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.option-item:hover {
  background-color: #f5f7fa;
}

.option-item.selected {
  border-color: #1890ff;
  background-color: #ecf5ff;
}

.option-item.important {
  border-color: #f5222d;
}

.option-item.important.selected {
  border-color: #f5222d;
  background-color: #fff1f0;
}

.check-icon {
  margin-left: auto;
  color: #1890ff;
}

.preset-section {
  margin-bottom: 20px;
}

.preset-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.preset-card {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px 14px;
  border: 2px solid #e4e7ed;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.preset-card:hover {
  border-color: #1890ff;
}

.preset-card.selected {
  border-color: #1890ff;
  background-color: #ecf5ff;
}

.preset-label {
  font-size: 14px;
  font-weight: 500;
}

.preset-desc {
  font-size: 11px;
  color: #909399;
}

.custom-section {
  border-top: 1px solid #e4e7ed;
  padding-top: 16px;
}

.size-input {
  width: 100%;
  margin-bottom: 8px;
}

.size-unit {
  width: 100%;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

@media (max-width: 768px) {
  .option-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .preset-grid {
    grid-template-columns: 1fr;
  }
}
</style>
