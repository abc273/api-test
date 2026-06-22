import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'

type NodeBadgeProps = {
  className?: string
}

export function NodeBadge({ className }: NodeBadgeProps) {
  const { t } = useTranslation()

  return (
    <Badge
      variant='secondary'
      className={cn(
        'h-5 rounded-full border-emerald-500/20 bg-emerald-500/12 px-2 text-[10px] font-semibold whitespace-nowrap text-emerald-700 dark:border-emerald-400/20 dark:bg-emerald-400/12 dark:text-emerald-200',
        className
      )}
    >
      {t('8liang Node')}
    </Badge>
  )
}
