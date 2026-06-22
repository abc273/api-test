import { createFileRoute, redirect } from '@tanstack/react-router'
import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'
import { AdminPortraitAssets } from '@/features/admin-portrait-assets'

export const Route = createFileRoute('/_authenticated/portrait-assets-admin/')({
  beforeLoad: () => {
    const { auth } = useAuthStore.getState()

    if (!auth.user || auth.user.role < ROLE.ADMIN) {
      throw redirect({
        to: '/403',
      })
    }
  },
  component: AdminPortraitAssets,
})
