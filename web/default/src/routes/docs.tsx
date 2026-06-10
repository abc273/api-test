import { createFileRoute } from '@tanstack/react-router'
import { ApiDocsPage } from '@/features/docs'

export const Route = createFileRoute('/docs')({
  component: ApiDocsPage,
})
