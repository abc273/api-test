import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  CheckCircle2,
  Clipboard,
  Loader2,
  Plus,
  QrCode,
  RefreshCw,
  ShieldAlert,
  UserCheck,
} from 'lucide-react'
import { toast } from 'sonner'
import { formatTimestampToDate } from '@/lib/format'
import { SectionPageLayout } from '@/components/layout'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { createPortraitAssetJob, getPortraitAssets } from './api'
import type { PortraitAssetJob, PortraitAssetStatus } from './types'

const statusLabels: Record<PortraitAssetStatus, string> = {
  pending: '排队中',
  qr_ready: '待扫码',
  waiting_upload: '待上传',
  waiting_accept: '待接收',
  ready: '可用',
  failed: '失败',
  disabled: '停用',
}

function statusVariant(
  status: PortraitAssetStatus
): 'default' | 'secondary' | 'destructive' {
  if (status === 'ready') return 'default'
  if (status === 'failed' || status === 'disabled') return 'destructive'
  return 'secondary'
}

function assetUri(assetId?: string) {
  if (!assetId) return ''
  return assetId.startsWith('asset://') ? assetId : `asset://${assetId}`
}

export function PortraitAssets() {
  const [jobs, setJobs] = useState<PortraitAssetJob[]>([])
  const [loading, setLoading] = useState(true)
  const [creating, setCreating] = useState(false)
  const [name, setName] = useState('')

  const readyJobs = useMemo(
    () => jobs.filter((job) => job.status === 'ready' && job.asset_id),
    [jobs]
  )

  const latestQRJob = useMemo(
    () => jobs.find((job) => job.qr_image || job.invite_url),
    [jobs]
  )

  const fetchJobs = useCallback(async () => {
    try {
      setLoading(true)
      const res = await getPortraitAssets({ p: 1, page_size: 50 })
      if (res.success && res.data) {
        setJobs(res.data.items || [])
      }
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchJobs()
  }, [fetchJobs])

  async function handleCreate() {
    try {
      setCreating(true)
      const res = await createPortraitAssetJob(name)
      if (res.success) {
        toast.success('已创建真人资产任务')
        setName('')
        await fetchJobs()
      } else {
        toast.error(res.message || '创建失败')
      }
    } finally {
      setCreating(false)
    }
  }

  async function handleCopy(value: string, label: string) {
    if (!value) return
    await navigator.clipboard.writeText(value)
    toast.success(`${label} 已复制`)
  }

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>真人资产</SectionPageLayout.Title>
      <SectionPageLayout.Description>
        管理已绑定的火山方舟真人人像资产
      </SectionPageLayout.Description>
      <SectionPageLayout.Actions>
        <Button variant='outline' onClick={fetchJobs} disabled={loading}>
          <RefreshCw className={loading ? 'animate-spin' : ''} />
          刷新
        </Button>
      </SectionPageLayout.Actions>
      <SectionPageLayout.Content>
        <div className='grid gap-4 lg:grid-cols-[minmax(0,1fr)_360px]'>
          <div className='space-y-4'>
            <Card className='rounded-lg'>
              <CardHeader>
                <CardTitle className='flex items-center gap-2 text-base'>
                  <Plus className='size-4' />
                  创建资产任务
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className='flex flex-col gap-3 sm:flex-row'>
                  <Input
                    value={name}
                    onChange={(event) => setName(event.target.value)}
                    placeholder='资产名称'
                    maxLength={50}
                  />
                  <Button onClick={handleCreate} disabled={creating}>
                    {creating ? <Loader2 className='animate-spin' /> : <Plus />}
                    创建
                  </Button>
                </div>
              </CardContent>
            </Card>

            <Card className='rounded-lg'>
              <CardHeader>
                <CardTitle className='flex items-center gap-2 text-base'>
                  <UserCheck className='size-4' />
                  资产列表
                </CardTitle>
              </CardHeader>
              <CardContent>
                {loading ? (
                  <div className='space-y-2'>
                    <Skeleton className='h-10 w-full' />
                    <Skeleton className='h-10 w-full' />
                    <Skeleton className='h-10 w-full' />
                  </div>
                ) : jobs.length === 0 ? (
                  <div className='text-muted-foreground flex h-28 items-center justify-center text-sm'>
                    暂无真人资产
                  </div>
                ) : (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>名称</TableHead>
                        <TableHead>状态</TableHead>
                        <TableHead>Asset ID</TableHead>
                        <TableHead>更新时间</TableHead>
                        <TableHead className='text-right'>操作</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {jobs.map((job) => (
                        <TableRow key={job.id}>
                          <TableCell>{job.name}</TableCell>
                          <TableCell>
                            <Badge variant={statusVariant(job.status)}>
                              {statusLabels[job.status] || job.status}
                            </Badge>
                          </TableCell>
                          <TableCell className='max-w-[260px] truncate font-mono text-xs'>
                            {assetUri(job.asset_id) || '-'}
                          </TableCell>
                          <TableCell>
                            {formatTimestampToDate(job.updated_time)}
                          </TableCell>
                          <TableCell className='text-right'>
                            <Button
                              size='icon-sm'
                              variant='ghost'
                              disabled={!job.asset_id}
                              onClick={() =>
                                handleCopy(assetUri(job.asset_id), 'Asset ID')
                              }
                            >
                              <Clipboard />
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                )}
              </CardContent>
            </Card>
          </div>

          <div className='space-y-4'>
            <Card className='rounded-lg'>
              <CardHeader>
                <CardTitle className='flex items-center gap-2 text-base'>
                  <QrCode className='size-4' />
                  授权二维码
                </CardTitle>
              </CardHeader>
              <CardContent>
                {latestQRJob?.qr_image ? (
                  <div className='space-y-3'>
                    <img
                      src={latestQRJob.qr_image}
                      alt='portrait asset invitation QR code'
                      className='bg-background aspect-square w-full rounded-md border object-contain p-3'
                    />
                    <div className='flex items-center justify-between gap-2'>
                      <Badge variant={statusVariant(latestQRJob.status)}>
                        {statusLabels[latestQRJob.status]}
                      </Badge>
                      <Button
                        size='sm'
                        variant='outline'
                        disabled={!latestQRJob.invite_url}
                        onClick={() =>
                          handleCopy(latestQRJob.invite_url || '', '邀请链接')
                        }
                      >
                        <Clipboard />
                        复制链接
                      </Button>
                    </div>
                  </div>
                ) : (
                  <div className='text-muted-foreground flex aspect-square items-center justify-center rounded-md border text-sm'>
                    等待 RPA 回写二维码
                  </div>
                )}
              </CardContent>
            </Card>

            <Card className='rounded-lg'>
              <CardHeader>
                <CardTitle className='flex items-center gap-2 text-base'>
                  <CheckCircle2 className='size-4' />
                  调用状态
                </CardTitle>
              </CardHeader>
              <CardContent className='space-y-3 text-sm'>
                <div className='flex items-center justify-between'>
                  <span className='text-muted-foreground'>可用资产</span>
                  <span className='font-medium'>{readyJobs.length}</span>
                </div>
                <div className='flex items-center justify-between'>
                  <span className='text-muted-foreground'>最近任务</span>
                  <span className='font-medium'>
                    {jobs[0] ? statusLabels[jobs[0].status] : '-'}
                  </span>
                </div>
                {jobs.some((job) => job.error_message) && (
                  <div className='text-destructive flex gap-2 rounded-md border p-3 text-xs'>
                    <ShieldAlert className='mt-0.5 size-4 shrink-0' />
                    <span>
                      {jobs.find((job) => job.error_message)?.error_message}
                    </span>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
