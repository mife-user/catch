import api from './index'

export function getTrashList() {
  return api.get('/trash')
}

export function restoreFiles(data) {
  return api.post('/trash/restore', data)
}

export function cleanExpired() {
  return api.delete('/trash/clean')
}
