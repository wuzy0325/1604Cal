import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'module-hub',
      component: () => import('../views/ModuleHubView.vue')
    },
    {
      path: '/device-management',
      name: 'module-device-management',
      component: () => import('../views/DeviceManagementView.vue')
    },
    {
      path: '/measurement',
      name: 'module-measurement',
      component: () => import('../views/MeasurementView.vue')
    },
    {
      path: '/calibration',
      name: 'module-calibration',
      component: () => import('../views/CalibrationView.vue')
    },
    {
      path: '/multi-pressure',
      name: 'multi-pressure',
      component: () => import('../views/measurement/PressureWorkbenchView.vue')
    }
  ]
})

export default router
