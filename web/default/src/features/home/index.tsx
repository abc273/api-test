import { useAuthStore } from '@/stores/auth-store'
import { PublicLayout } from '@/components/layout'
import { Footer } from '@/components/layout/components/footer'
import { CTA, Features, Hero, HowItWorks, Stats } from './components'

export function Home() {
  const { auth } = useAuthStore()
  const isAuthenticated = !!auth.user

  return (
    <PublicLayout showMainContainer={false}>
      {/* Edit the landing page directly in web/default/src/features/home/components/sections/*.tsx */}
      <Hero isAuthenticated={isAuthenticated} />
      <Stats />
      <Features />
      <HowItWorks />
      <CTA isAuthenticated={isAuthenticated} />
      <Footer />
    </PublicLayout>
  )
}
