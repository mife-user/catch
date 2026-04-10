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
        <span>预览结果</span>
      </template>
      <el-table :data="previewResult.previews" stripe>
        <el-table-column prop="old_path" label="原路径" min-width="300" show-overflow-tooltip />
        <el-table-column prop="new_path" label="新路径" min-width="300" show-overflow-tooltip />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Edit } from '@element-plus/icons-vue'
import { renamePreview, renameFiles } from '../api/files'
import { ElMessage } from 'element-plus'
import { useRoute } from 'vue-router'

const route = useRoute()
const previewLoading = ref(false)
const renameLoading = ref(false)
const previewResult = ref(null)
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

onMounted(() => {
  if (route.query.files) {
    const files = Array.isArray(route.query.files) ? route.query.files : [route.query.files]
    filesInput.value = files.join('\n')
  }
})

const getPaths = () => {
  return filesInput.value.split('\n').map(p => p.trim()).filter(p => p)
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

.preview-card {
  margin-top: 20px;
}
</style>
