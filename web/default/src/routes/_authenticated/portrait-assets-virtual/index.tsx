import { createFileRoute } from '@tanstack/react-router'
import { VirtualPortraitAssets } from '@/features/portrait-assets-virtual'

export const Route = createFileRoute(
  '/_authenticated/portrait-assets-virtual/'
)({
  component: VirtualPortraitAssets,
})
