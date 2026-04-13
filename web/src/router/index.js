import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    redirect: '/search',
  },
  {
    path: '/welcome',
    name: 'Welcome',
    component: () => import('../views/WelcomeView.vue'),
  },
  {
    path: '/tutorial',
    name: 'Tutorial',
    component: () => import('../views/TutorialView.vue'),
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
    path: '/trash',
    name: 'Trash',
    component: () => import('../views/TrashView.vue'),
  },
  {
    path: '/cleanup',
    name: 'Cleanup',
    component: () => import('../views/CleanupView.vue'),
  },
  {
    path: '/uninstall',
    name: 'Uninstall',
    component: () => import('../views/UninstallView.vue'),
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
