<template>
  <div class="delete-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-icon><Delete /></el-icon>
          <span>文件删除</span>
        </div>
      </template>

      <el-form :model="deleteForm" label-width="100px" label-position="left">
        <el-form-item label="文件路径">
          <el-input
            v-model="filesInput"
            type="textarea"
            :rows="4"
            placeholder="输入要删除的文件路径，每行一个"
          />
        </el-form-item>

        <el-form-item label="删除方式">
          <el-radio-group v-model="deleteForm.mode">
            <el-radio value="recycle">移至回收站（可恢复）</el-radio>
            <el-radio value="trash">直接删除（过期清理）</el-radio>
            <el-radio value="permanent">永久删除（需密码）</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item v-if="deleteForm.mode === 'permanent'" label="安全密码">
          <el-input v-model="deleteForm.password" type="password" placeholder="输入安全密码" show-password />
        </el-form-item>

        <el-form-item>
          <el-button type="danger" @click="handleDelete" :loading="loading">
            <el-icon><Delete /></el-icon>
            执行删除
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card v-if="result" class="result-card">
      <template #header>
        <span>删除结果</span>
      </template>
      <el-result v-if="result.failed.length === 0" icon="success" title="全部删除成功" :sub-title="`成功删除 ${result.success.length} 个文件`" />
      <el-result v-else icon="warning" title="部分删除失败" :sub-title="`成功 ${result.success.length} 个，失败 ${result.failed.length} 个`">
        <template #extra>
          <div class="failed-list">
            <p v-for="item in result.failed" :key="item" class="failed-item">{{ item }}</p>
          </div>
        </template>
      </el-result>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Delete } from '@element-plus/icons-vue'
import { deleteFiles } from '../api/files'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRoute } from 'vue-router'

const route = useRoute()
const loading = ref(false)
const result = ref(null)
const filesInput = ref('')

const deleteForm = reactive({
  mode: 'recycle',
  password: '',
})

onMounted(() => {
  if (route.query.files) {
    const files = Array.isArray(route.query.files) ? route.query.files : [route.query.files]
    filesInput.value = files.join('\n')
  }
})

const handleDelete = async () => {
  const paths = filesInput.value.split('\n').map(p => p.trim()).filter(p => p)
  if (paths.length === 0) {
    ElMessage.warning('请输入要删除的文件路径')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除 ${paths.length} 个文件吗？${
        deleteForm.mode === 'permanent' ? '永久删除不可恢复！' :
        deleteForm.mode === 'trash' ? '文件将在过期后自动清理。' :
        '文件将移至系统回收站。'
      }`,
      '确认删除',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return
  }

  loading.value = true
  try {
    const data = await deleteFiles({
      paths,
      mode: deleteForm.mode,
      password: deleteForm.mode === 'permanent' ? deleteForm.password : '',
    })
    result.value = data
    ElMessage.success(`删除完成：成功 ${data.success?.length || 0} 个`)
  } catch (err) {
    ElMessage.error(err.message || '删除失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.delete-view {
  max-width: 800px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
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
