import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    {
      path: '/measurement',
      name: 'measurement',
      component: () => import('../views/measurement/CalibrationView.vue')
    },
    {
      path: '/multi-pressure',
      name: 'multi-pressure',
      component: () => import('../views/measurement/PressureWorkbenchView.vue')
    },
    {
      path: '/calibration',
      name: 'calibration',
      component: () => import('../views/calibration/MainView.vue')
    }
  ]
})

export default router