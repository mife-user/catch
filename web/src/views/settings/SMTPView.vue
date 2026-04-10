<template>
  <div class="smtp-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-icon><Message /></el-icon>
          <span>SMTP设置</span>
        </div>
      </template>

      <el-form :model="form" label-width="120px" label-position="left" v-loading="loading">
        <el-form-item label="快速配置">
          <el-select v-model="selectedTemplate" placeholder="选择邮箱服务商" @change="handleTemplateSelect" clearable>
            <el-option v-for="t in templates" :key="t.name" :label="t.name" :value="t.name" />
          </el-select>
          <el-button type="primary" link @click="showHelp = true">如何获取授权码？</el-button>
        </el-form-item>

        <el-form-item label="SMTP服务器">
          <el-input v-model="form.host" placeholder="如 smtp.qq.com" />
        </el-form-item>

        <el-form-item label="SMTP端口">
          <el-input-number v-model="form.port" :min="1" :max="65535" />
        </el-form-item>

        <el-form-item label="发件邮箱">
          <el-input v-model="form.username" placeholder="输入邮箱账号" />
        </el-form-item>

        <el-form-item label="授权码">
          <el-input v-model="form.password" type="password" placeholder="输入邮箱授权码（非登录密码）" show-password />
        </el-form-item>

        <el-form-item label="收件邮箱">
          <el-input v-model="form.to" placeholder="反馈接收邮箱" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSave" :loading="saving">保存设置</el-button>
          <el-button @click="handleTest" :loading="testing">测试连接</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-dialog v-model="showHelp" title="如何获取授权码" width="500px">
      <div class="help-content">
        <h4>QQ邮箱</h4>
        <p>设置 → 账户 → POP3/SMTP服务 → 开启 → 生成授权码</p>
        <h4>163邮箱</h4>
        <p>设置 → POP3/SMTP/IMAP → 开启 → 设置授权密码</p>
        <h4>Gmail</h4>
        <p>Google账户 → 安全性 → 两步验证 → 应用专用密码</p>
        <h4>Outlook</h4>
        <p>设置 → 邮件 → 同步电子邮件 → 应用密码</p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@element-plus/icons-vue'
import { getConfig, updateConfig } from '../../api/config'
import { getSMTPTemplates, testSMTP } from '../../api/feedback'
import { ElMessage } from 'element-plus'

const loading = ref(false)
const saving = ref(false)
const testing = ref(false)
const showHelp = ref(false)
const selectedTemplate = ref('')
const templates = ref([])

const form = reactive({
  host: '',
  port: 465,
  username: '',
  password: '',
  to: '15723556393@163.com',
})

onMounted(async () => {
  loading.value = true
  try {
    const [config, templatesData] = await Promise.all([
      getConfig(),
      getSMTPTemplates(),
    ])
    form.host = config.smtp?.host || ''
    form.port = config.smtp?.port || 465
    form.username = config.smtp?.username || ''
    form.to = config.smtp?.to || '15723556393@163.com'
    templates.value = templatesData.templates || []
  } catch {} finally {
    loading.value = false
  }
})

const handleTemplateSelect = (name) => {
  const template = templates.value.find(t => t.name === name)
  if (template) {
    form.host = template.host
    form.port = template.port
  }
}

const handleSave = async () => {
  if (!form.host) {
    ElMessage.warning('请填写SMTP服务器')
    return
  }
  saving.value = true
  try {
    await updateConfig({
      smtp: { ...form },
    })
    ElMessage.success('SMTP设置已保存')
  } catch (err) {
    ElMessage.error(err.message || '保存失败')
  } finally {
    saving.value = false
  }
}

const handleTest = async () => {
  if (!form.host || !form.username || !form.password) {
    ElMessage.warning('请先填写完整的SMTP配置')
    return
  }
  testing.value = true
  try {
    const result = await testSMTP({ ...form })
    if (result.success) {
      ElMessage.success('SMTP测试成功，邮件已发送')
    } else {
      ElMessage.error(result.message || 'SMTP测试失败')
    }
  } catch (err) {
    ElMessage.error(err.message || '测试失败')
  } finally {
    testing.value = false
  }
}
</script>

<style scoped>
.smtp-view {
  max-width: 700px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.help-content h4 {
  margin: 12px 0 4px;
  color: #1890ff;
}

.help-content p {
  margin: 0 0 8px;
  color: #666;
  font-size: 13px;
}
</style>
