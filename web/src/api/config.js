import api from './index'

export function getConfig() {
  return api.get('/config')
}

export function updateConfig(data) {
  return api.put('/config', data)
}

export function setPassword(data) {
  return api.post('/config/password', data)
}

export function verifyPassword(data) {
  return api.post('/config/password/verify', data)
}
