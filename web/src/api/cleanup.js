import api from './index'

export function getCleanupRules() {
  return api.get('/cleanup/rules')
}

export function scanCleanup(data, clientId) {
  return api.post('/cleanup/scans', data, { params: { client_id: clientId } })
}

export function executeCleanup(data) {
  return api.post('/cleanup/execute', data)
}

export function scanUninstallApps() {
  return api.get('/uninstall/scan')
}

export function analyzeUninstall(data) {
  return api.post('/uninstall/analyze', data)
}

export function executeUninstall(data) {
  return api.post('/uninstall/execute', data)
}
