import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    redirect: '/search',
  },
  {
    path: '/search',
    name: 'Search',
    component: () => import('../views/SearchView.vue'),
  },
  {
    path: '/delete',
    name: 'Delete',
    component: () => import('../views/DeleteView.vue'),
  },
  {
    path: '/rename',
    name: 'Rename',
    component: () => import('../views/RenameView.vue'),
  },
  {
    path: '/move',
    name: 'Move',
    component: () => import('../views/MoveView.vue'),
  },
  {
    path: '/settings/basic',
    name: 'SettingsBasic',
    component: () => import('../views/settings/BasicView.vue'),
  },
  {
    path: '/settings/security',
    name: 'SettingsSecurity',
    component: () => import('../views/settings/SecurityView.vue'),
  },
  {
    path: '/settings/smtp',
    name: 'SettingsSMTP',
    component: () => import('../views/settings/SMTPView.vue'),
  },
  {
    path: '/settings/about',
    name: 'SettingsAbout',
    component: () => import('../views/settings/AboutView.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
