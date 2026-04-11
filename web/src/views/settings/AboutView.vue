<template>
  <div class="about-view">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-icon><InfoFilled /></el-icon>
          <span>关于</span>
        </div>
      </template>

      <div class="about-content">
        <div class="about-info">
          <h1>Catch 文件整理工具</h1>
          <p class="version">版本 {{ systemInfo.version }}</p>
          <p class="desc">面向非技术用户的文件整理工具，通过"命令行启动 + 网页界面操作"的混合架构，帮助用户高效管理本地文件。</p>
        </div>

        <el-divider />

        <div class="system-section">
          <h3>系统信息</h3>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="操作系统">{{ systemInfo.os }}</el-descriptions-item>
            <el-descriptions-item label="系统架构">{{ systemInfo.arch }}</el-descriptions-item>
            <el-descriptions-item label="Go版本">{{ systemInfo.goVersion }}</el-descriptions-item>
            <el-descriptions-item label="浏览器">{{ systemInfo.browser }}</el-descriptions-item>
            <el-descriptions-item label="服务端口">{{ systemInfo.port }}</el-descriptions-item>
            <el-descriptions-item label="安全密码">{{ systemInfo.hasPassword ? '已设置' : '未设置' }}</el-descriptions-item>
          </el-descriptions>
        </div>

        <el-divider />

        <div class="feedback-section">
          <h3>意见反馈</h3>
          <el-form :model="feedbackForm" label-width="80px" label-position="left">
            <el-form-item label="反馈类型">
              <el-select v-model="feedbackForm.type" placeholder="选择反馈类型">
                <el-option label="Bug报告" value="Bug报告" />
                <el-option label="功能建议" value="功能建议" />
                <el-option label="其他" value="其他" />
              </el-select>
            </el-form-item>

            <el-form-item label="反馈内容">
              <el-input
                v-model="feedbackForm.content"
                type="textarea"
                :rows="5"
                placeholder="请详细描述您的问题或建议"
              />
            </el-form-item>

            <el-form-item label="附加信息">
              <el-checkbox v-model="feedbackForm.includeSystemInfo">自动附带系统环境信息</el-checkbox>
              <div v-if="feedbackForm.includeSystemInfo" class="system-preview">
                <el-tag size="small" type="info">操作系统: {{ systemInfo.os }} {{ systemInfo.arch }}</el-tag>
                <el-tag size="small" type="info">版本: {{ systemInfo.version }}</el-tag>
                <el-tag size="small" type="info">浏览器: {{ systemInfo.browser }}</el-tag>
              </div>
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="handleSendFeedback" :loading="sending">
                发送反馈
              </el-button>
              <el-button v-if="!hasSMTP" type="warning" link @click="router.push('/settings/smtp')">
                未配置SMTP，点击前往设置
              </el-button>
            </el-form-item>
          </el-form>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { InfoFilled } from '@element-plus/icons-vue'
import { sendFeedback } from '../../api/feedback'
import { getConfig } from '../../api/config'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'

const router = useRouter()
const sending = ref(false)
const hasSMTP = ref(false)

const systemInfo = reactive({
  version: '1.0.0',
  os: '-',
  arch: '-',
  goVersion: '-',
  browser: '-',
  port: '-',
  hasPassword: false,
})

const feedbackForm = reactive({
  type: 'Bug报告',
  content: '',
  includeSystemInfo: true,
})

onMounted(async () => {
  try {
    const config = await getConfig()
    hasSMTP.value = config.has_smtp || false
    systemInfo.version = config.version || '1.0.0'
    systemInfo.os = config.system_info?.os || '-'
    systemInfo.arch = config.system_info?.arch || '-'
    systemInfo.goVersion = config.system_info?.go_version || '-'
    systemInfo.port = config.server?.port || '-'
    systemInfo.hasPassword = config.has_password || false
  } catch {}

  const ua = navigator.userAgent
  if (ua.includes('Chrome') && !ua.includes('Edg')) {
    systemInfo.browser = 'Chrome'
  } else if (ua.includes('Edg')) {
    systemInfo.browser = 'Edge'
  } else if (ua.includes('Firefox')) {
    systemInfo.browser = 'Firefox'
  } else if (ua.includes('Safari') && !ua.includes('Chrome')) {
    systemInfo.browser = 'Safari'
  } else {
    systemInfo.browser = '其他'
  }
})

const handleSendFeedback = async () => {
  if (!feedbackForm.content) {
    ElMessage.warning('请输入反馈内容')
    return
  }

  if (!hasSMTP.value) {
    ElMessage.warning('请先在SMTP设置中配置邮件信息')
    return
  }

  sending.value = true
  try {
    let content = feedbackForm.content
    if (feedbackForm.includeSystemInfo) {
      content += `\n\n[系统信息] OS: ${systemInfo.os} ${systemInfo.arch}, 版本: ${systemInfo.version}, 浏览器: ${systemInfo.browser}`
    }
    const result = await sendFeedback({
      type: feedbackForm.type,
      content,
    })
    if (result.success) {
      ElMessage.success('反馈已发送，感谢您的意见！')
      feedbackForm.content = ''
    } else {
      ElMessage.error(result.message || '发送失败')
    }
  } catch (err) {
    ElMessage.error(err.message || '发送失败')
  } finally {
    sending.value = false
  }
}
</script>

<style scoped>
.about-view {
  max-width: 700px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
}

.about-content {
  text-align: center;
}

.about-info h1 {
  font-size: 24px;
  color: #1890ff;
  margin-bottom: 8px;
}

.version {
  color: #999;
  font-size: 14px;
  margin-bottom: 12px;
}

.desc {
  color: #666;
  font-size: 14px;
  line-height: 1.6;
}

.system-section {
  text-align: left;
}

.system-section h3 {
  margin-bottom: 12px;
  font-size: 16px;
}

.feedback-section {
  text-align: left;
}

.feedback-section h3 {
  margin-bottom: 16px;
  font-size: 16px;
}

.system-preview {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 8px;
}
</style>
