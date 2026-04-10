import { createRouter, createWebHistory } from 'vue-router'

import CalibrationView from '@/views/CalibrationView.vue'
import DeviceManagementView from '@/views/DeviceManagementView.vue'
import MeasurementView from '@/views/MeasurementView.vue'
import ModuleHubView from '@/views/ModuleHubView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'module-hub',
      component: ModuleHubView
    },
    {
      path: '/workbench',
      redirect: '/module/device-management'
    },
    {
      path: '/module/device-management',
      name: 'module-device-management',
      component: DeviceManagementView
    },
    {
      path: '/module/measurement',
      name: 'module-measurement',
      component: MeasurementView
    },
    {
      path: '/module/calibration',
      name: 'module-calibration',
      component: CalibrationView
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/'
    }
  ]
})

export default router
