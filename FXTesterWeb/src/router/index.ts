import { createRouter, createWebHistory } from 'vue-router'
import DashboardLayout from '@/components/templates/DashboardLayout.vue'
import Login from '@/components/views/Login.vue'
import Home from '@/components/views/Home.vue'
import About from '@/components/views/About.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: DashboardLayout,
      children: [
        {
          path: '',
          component: Home
        },
        {
          path: 'about',
          component: About
        }
      ]
    },
    {
      path: '/login',
      component: Login,
      beforeEnter(to, from) {
        console.log("beforeEnter: ", to, from);
        to.meta.previous = to.redirectedFrom?.fullPath ?? from.fullPath
        return true
      }
    },
    {
      path: '/:catchAll(.*)*',
      redirect: '/login'
    }
  ]
})
export default router