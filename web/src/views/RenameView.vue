<template>
  <div class="rename-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-icon><Edit /></el-icon>
          <span>文件重命名</span>
        </div>
      </template>

      <el-form :model="renameForm" label-width="100px" label-position="left">
        <el-form-item label="文件路径">
          <el-input
            v-model="filesInput"
            type="textarea"
            :rows="4"
            placeholder="输入要重命名的文件路径，每行一个"
          />
          <div v-if="fileList.length > 0" class="file-count">
            共 {{ fileList.length }} 个文件
          </div>
        </el-form-item>

        <el-form-item label="重命名规则">
          <el-select v-model="renameForm.rule" placeholder="选择重命名规则">
            <el-option label="添加前缀" value="prefix" />
            <el-option label="添加后缀" value="suffix" />
            <el-option label="序号编号" value="sequence" />
            <el-option label="替换文本" value="replace" />
            <el-option label="日期时间戳" value="timestamp" />
          </el-select>
        </el-form-item>

        <el-form-item v-if="renameForm.rule === 'prefix'" label="前缀内容">
          <el-input v-model="renameForm.params.prefix" placeholder="输入要添加的前缀" />
        </el-form-item>

        <el-form-item v-if="renameForm.rule === 'suffix'" label="后缀内容">
          <el-input v-model="renameForm.params.suffix" placeholder="输入要添加的后缀" />
        </el-form-item>

        <el-form-item v-if="renameForm.rule === 'sequence'" label="起始编号">
          <el-input-number v-model="renameForm.params.start" :min="0" :max="9999" />
        </el-form-item>

        <el-form-item v-if="renameForm.rule === 'sequence'" label="编号位数">
          <el-input-number v-model="renameForm.params.digits" :min="1" :max="10" />
        </el-form-item>

        <el-form-item v-if="renameForm.rule === 'replace'" label="替换设置">
          <el-row :gutter="10">
            <el-col :span="12">
              <el-input v-model="renameForm.params.old" placeholder="原文本" />
            </el-col>
            <el-col :span="12">
              <el-input v-model="renameForm.params.new" placeholder="替换为" />
            </el-col>
          </el-row>
        </el-form-item>

        <el-alert
          v-if="renameForm.rule === 'timestamp'"
          title="将在文件名后添加当前日期，格式：_YYYYMMDD"
          type="info"
          :closable="false"
          show-icon
          class="rule-hint"
        />

        <el-form-item>
          <el-button type="primary" @click="handlePreview" :loading="previewLoading">
            预览结果
          </el-button>
          <el-button type="success" @click="handleRename" :loading="renameLoading" :disabled="!previewResult">
            执行重命名
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card v-if="previewResult" class="preview-card">
      <template #header>
        <div class="card-header">
          <span>预览结果 ({{ previewResult.previews?.length || 0 }} 个文件)</span>
          <el-tag type="info" size="small">请确认后执行</el-tag>
        </div>
      </template>
      <el-table :data="previewResult.previews" stripe>
        <el-table-column label="原文件名" min-width="300">
          <template #default="{ row }">
            <span class="old-name">{{ getFileName(row.old_path) }}</span>
            <div class="path-hint">{{ row.old_path }}</div>
          </template>
        </el-table-column>
        <el-table-column label="" width="60" align="center">
          <template #default>
            <el-icon><Right /></el-icon>
          </template>
        </el-table-column>
        <el-table-column label="新文件名" min-width="300">
          <template #default="{ row }">
            <span class="new-name">{{ getFileName(row.new_path) }}</span>
            <div class="path-hint">{{ row.new_path }}</div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card v-if="renameResult" class="result-card">
      <template #header>
        <span>执行结果</span>
      </template>
      <el-result
        v-if="renameResult.failed.length === 0"
        icon="success"
        :title="`成功重命名 ${renameResult.success.length} 个文件`"
      />
      <el-result
        v-else
        icon="warning"
        :title="`成功 ${renameResult.success.length} 个，失败 ${renameResult.failed.length} 个`"
      >
        <template #extra>
          <div class="failed-list">
            <p v-for="item in renameResult.failed" :key="item" class="failed-item">{{ item }}</p>
          </div>
        </template>
      </el-result>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Edit, Right } from '@element-plus/icons-vue'
import { renamePreview, renameFiles } from '../api/files'
import { ElMessage } from 'element-plus'
import { useRoute } from 'vue-router'

const route = useRoute()
const previewLoading = ref(false)
const renameLoading = ref(false)
const previewResult = ref(null)
const renameResult = ref(null)
const filesInput = ref('')

const renameForm = reactive({
  rule: 'prefix',
  params: {
    prefix: '',
    suffix: '',
    start: '1',
    digits: '3',
    old: '',
    new: '',
  },
})

const fileList = computed(() => {
  return filesInput.value.split('\n').map(p => p.trim()).filter(p => p)
})

onMounted(() => {
  if (route.query.files) {
    const files = Array.isArray(route.query.files) ? route.query.files : [route.query.files]
    filesInput.value = files.join('\n')
  }
})

const getPaths = () => {
  return filesInput.value.split('\n').map(p => p.trim()).filter(p => p)
}

const getFileName = (path) => {
  const parts = path.replace(/\\/g, '/')
  const idx = parts.lastIndexOf('/')
  return idx >= 0 ? parts.substring(idx + 1) : path
}

const handlePreview = async () => {
  const paths = getPaths()
  if (paths.length === 0) {
    ElMessage.warning('请输入要重命名的文件路径')
    return
  }

  previewLoading.value = true
  try {
    const data = await renamePreview({
      paths,
      rule: renameForm.rule,
      params: renameForm.params,
    })
    previewResult.value = data
    renameResult.value = null
    ElMessage.success('预览生成成功')
  } catch (err) {
    ElMessage.error(err.message || '预览失败')
  } finally {
    previewLoading.value = false
  }
}

const handleRename = async () => {
  const paths = getPaths()
  if (paths.length === 0) {
    ElMessage.warning('请输入要重命名的文件路径')
    return
  }

  renameLoading.value = true
  try {
    const data = await renameFiles({
      paths,
      rule: renameForm.rule,
      params: renameForm.params,
    })
    renameResult.value = data
    ElMessage.success(`重命名完成：成功 ${data.success?.length || 0} 个`)
    previewResult.value = null
  } catch (err) {
    ElMessage.error(err.message || '重命名失败')
  } finally {
    renameLoading.value = false
  }
}
</script>

<style scoped>
.rename-view {
  max-width: 1000px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.preview-card .card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.file-count {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}

.rule-hint {
  margin-bottom: 16px;
}

.old-name {
  color: #909399;
  text-decoration: line-through;
}

.new-name {
  color: #1890ff;
  font-weight: 500;
}

.path-hint {
  font-size: 11px;
  color: #c0c4cc;
  margin-top: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.preview-card {
  margin-top: 20px;
}

.result-card {
  margin-top: 20px;
}

.failed-list {
  text-align: left;
  max-height: 200px;
  overflow-y: auto;
}

.failed-item {
  color: #f5222d;
  font-size: 13px;
  margin: 4px 0;
}
</style>
