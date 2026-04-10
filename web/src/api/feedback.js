import api from './index'

export function sendFeedback(data) {
  return api.post('/feedback', data)
}

export function getSMTPTemplates() {
  return api.get('/smtp/templates')
}

export function testSMTP(data) {
  return api.post('/smtp/test', data)
}
