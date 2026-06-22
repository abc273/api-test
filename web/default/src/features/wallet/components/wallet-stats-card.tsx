import { useTranslation } from 'react-i18next'
import { formatQuota } from '@/lib/format'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import type { UserWalletData, WalletSummary } from '../types'

interface WalletStatsCardProps {
  user: UserWalletData | null
  summary?: WalletSummary | null
  loading?: boolean
}

export function WalletStatsCard(props: WalletStatsCardProps) {
  const { t } = useTranslation()

  if (props.loading) {
    return (
      <Card>
        <CardContent>
          <div className='grid grid-cols-1 gap-6 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-5'>
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className='space-y-2'>
                <Skeleton className='h-5 w-28' />
                <Skeleton className='h-9 w-32' />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  const usedQuota = props.summary?.used_quota ?? props.user?.used_quota ?? 0
  const refundedQuota = props.summary?.refunded_quota ?? 0
  const actualUsedQuota =
    props.summary?.actual_used_quota ?? Math.max(usedQuota - refundedQuota, 0)

  const stats = [
    {
      label: t('Current Balance'),
      value: formatQuota(props.user?.quota ?? 0),
      valueClassName: '',
    },
    {
      label: t('Actual Usage'),
      value: formatQuota(actualUsedQuota),
      valueClassName: 'text-foreground',
    },
    {
      label: t('Refunded'),
      value: formatQuota(refundedQuota),
      valueClassName: 'text-emerald-600 dark:text-emerald-400',
    },
    {
      label: t('Total Usage'),
      value: formatQuota(usedQuota),
      valueClassName: 'text-muted-foreground',
    },
    {
      label: t('API Requests'),
      value: (props.user?.request_count ?? 0).toLocaleString(),
      valueClassName: '',
    },
  ]

  return (
    <Card>
      <CardContent>
        <div className='grid grid-cols-1 gap-6 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-5'>
          {stats.map((stat) => (
            <div key={stat.label} className='min-w-0 space-y-2'>
              <div className='text-muted-foreground text-sm font-medium'>
                {stat.label}
              </div>
              <div
                className={`text-2xl leading-tight font-semibold tracking-tight break-all ${stat.valueClassName}`}
              >
                {stat.value}
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
