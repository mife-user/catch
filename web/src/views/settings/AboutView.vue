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
          <p class="version">版本 1.0.0</p>
          <p class="desc">面向非技术用户的文件整理工具，通过"命令行启动 + 网页界面操作"的混合架构，帮助用户高效管理本地文件。</p>
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

            <el-form-item>
              <el-button type="primary" @click="handleSendFeedback" :loading="sending">
                发送反馈
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

const sending = ref(false)
const hasSMTP = ref(false)

const feedbackForm = reactive({
  type: 'Bug报告',
  content: '',
})

onMounted(async () => {
  try {
    const config = await getConfig()
    hasSMTP.value = config.has_smtp || false
  } catch {}
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
    const result = await sendFeedback(feedbackForm)
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

.feedback-section {
  text-align: left;
}

.feedback-section h3 {
  margin-bottom: 16px;
  font-size: 16px;
}
</style>
