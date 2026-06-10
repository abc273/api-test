import { createFileRoute } from '@tanstack/react-router'
import { OfficialPortraitAssets } from '@/features/portrait-assets-official'

export const Route = createFileRoute(
  '/_authenticated/portrait-assets-official/'
)({
  component: OfficialPortraitAssets,
})
