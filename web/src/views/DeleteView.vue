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
          <div v-if="fileList.length > 0" class="file-count">
            共 {{ fileList.length }} 个文件
          </div>
        </el-form-item>

        <el-form-item label="删除方式">
          <el-radio-group v-model="deleteForm.mode">
            <el-radio value="recycle">
              <span>移至回收站</span>
              <el-tag size="small" type="success" class="mode-tag">可恢复</el-tag>
            </el-radio>
            <el-radio value="trash">
              <span>直接删除</span>
              <el-tag size="small" type="warning" class="mode-tag">过期清理</el-tag>
            </el-radio>
            <el-radio value="permanent">
              <span>永久删除</span>
              <el-tag size="small" type="danger" class="mode-tag">不可恢复</el-tag>
            </el-radio>
          </el-radio-group>
        </el-form-item>

        <el-alert
          v-if="deleteForm.mode === 'recycle'"
          title="文件将移至系统回收站，可随时恢复"
          type="success"
          :closable="false"
          show-icon
          class="mode-alert"
        />
        <el-alert
          v-if="deleteForm.mode === 'trash'"
          title="文件将移至 Catch 回收站，到期后自动清理（默认7天）"
          type="warning"
          :closable="false"
          show-icon
          class="mode-alert"
        />
        <el-alert
          v-if="deleteForm.mode === 'permanent'"
          title="文件将被永久删除，无法恢复！请谨慎操作"
          type="error"
          :closable="false"
          show-icon
          class="mode-alert"
        />

        <el-form-item v-if="deleteForm.mode === 'permanent'" label="安全密码">
          <el-input v-model="deleteForm.password" type="password" placeholder="输入安全密码" show-password />
          <div v-if="!hasPassword" class="password-hint">
            <el-text type="warning">未设置安全密码，请先在设置中配置安全密码</el-text>
            <el-button type="primary" link @click="router.push('/settings/security')">前往设置</el-button>
          </div>
        </el-form-item>

        <el-form-item>
          <el-button
            type="danger"
            @click="handleDelete"
            :loading="loading"
            :disabled="deleteForm.mode === 'permanent' && !hasPassword"
          >
            <el-icon><Delete /></el-icon>
            执行删除
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card v-if="opProgress" class="progress-card">
      <div class="progress-content">
        <span class="progress-text">正在删除... {{ opProgress.done }} / {{ opProgress.total }}</span>
      </div>
      <el-progress :percentage="opProgress.total > 0 ? Math.round(opProgress.done / opProgress.total * 100) : 0" :show-text="true" />
    </el-card>

    <el-card v-if="result" class="result-card">
      <template #header>
        <span>删除结果</span>
      </template>
      <el-result
        v-if="result.failed.length === 0"
        icon="success"
        title="全部删除成功"
        :sub-title="`成功删除 ${result.success.length} 个文件`"
      />
      <el-result
        v-else
        icon="warning"
        title="部分删除失败"
        :sub-title="`成功 ${result.success.length} 个，失败 ${result.failed.length} 个`"
      >
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
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { Delete } from '@element-plus/icons-vue'
import { deleteFiles } from '../api/files'
import { getConfig } from '../api/config'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const result = ref(null)
const filesInput = ref('')
const hasPassword = ref(false)
const opProgress = ref(null)

let ws = null
let clientId = ''

const deleteForm = reactive({
  mode: 'recycle',
  password: '',
})

const fileList = computed(() => {
  return filesInput.value.split('\n').map(p => p.trim()).filter(p => p)
})

const connectWebSocket = () => {
  clientId = 'client_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9)
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/api/ws?client_id=${clientId}`
  try {
    ws = new WebSocket(wsUrl)
    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'operation_progress') {
          opProgress.value = msg.payload
        }
      } catch {}
    }
    ws.onerror = () => {}
    ws.onclose = () => { ws = null }
  } catch {}
}

onMounted(async () => {
  connectWebSocket()

  const stored = sessionStorage.getItem('catch_selected_files')
  if (stored) {
    try {
      const files = JSON.parse(stored)
      if (Array.isArray(files) && files.length > 0) {
        filesInput.value = files.join('\n')
      }
    } catch {}
    sessionStorage.removeItem('catch_selected_files')
  }

  try {
    const config = await getConfig()
    hasPassword.value = config.has_password || false
  } catch {}
})

onUnmounted(() => {
  if (ws) {
    ws.close()
    ws = null
  }
})

const handleDelete = async () => {
  const paths = fileList.value
  if (paths.length === 0) {
    ElMessage.warning('请输入要删除的文件路径')
    return
  }

  if (deleteForm.mode === 'permanent' && !deleteForm.password) {
    ElMessage.warning('请输入安全密码')
    return
  }

  const modeText = deleteForm.mode === 'permanent' ? '永久删除不可恢复！' :
    deleteForm.mode === 'trash' ? '文件将在过期后自动清理。' :
    '文件将移至系统回收站。'

  try {
    await ElMessageBox.confirm(
      `确定要删除 ${paths.length} 个文件吗？${modeText}`,
      '确认删除',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return
  }

  loading.value = true
  opProgress.value = { operation: 'delete', done: 0, total: paths.length }
  try {
    const headers = {}
    if (clientId) {
      headers['X-Client-ID'] = clientId
    }
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
    opProgress.value = null
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

.mode-tag {
  margin-left: 8px;
}

.mode-alert {
  margin-bottom: 16px;
}

.file-count {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}

.password-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}

.progress-card {
  margin-top: 20px;
}

.progress-content {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.progress-text {
  font-size: 14px;
  color: #606266;
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
