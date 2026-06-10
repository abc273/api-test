import { createFileRoute } from '@tanstack/react-router'
import { PortraitAssets } from '@/features/portrait-assets'

export const Route = createFileRoute('/_authenticated/portrait-assets/')({
  component: PortraitAssets,
})
