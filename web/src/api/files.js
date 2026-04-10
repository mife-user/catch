import api from './index'

export function searchFiles(params) {
  return api.get('/files/search', { params })
}

export function deleteFiles(data) {
  return api.post('/files/delete', data)
}

export function renamePreview(data) {
  return api.post('/files/rename/preview', data)
}

export function renameFiles(data) {
  return api.post('/files/rename', data)
}

export function moveFiles(data) {
  return api.post('/files/move', data)
}

export function copyFiles(data) {
  return api.post('/files/copy', data)
}
