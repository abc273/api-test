import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  CheckCircle2,
  Clipboard,
  Loader2,
  Plus,
  QrCode,
  RefreshCw,
  Send,
  ShieldAlert,
  UserCheck,
} from 'lucide-react'
import { toast } from 'sonner'
import { formatTimestampToDate } from '@/lib/format'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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
import { SectionPageLayout } from '@/components/layout'
import {
  confirmPortraitAsset,
  createPortraitAssetJob,
  getPortraitAssets,
  rejectPortraitAsset,
  requestPortraitAssetAccept,
} from './api'
import type { PortraitAssetJob, PortraitAssetStatus } from './types'

const statusLabels: Record<PortraitAssetStatus, string> = {
  pending: '排队中',
  qr_ready: '待扫码',
  waiting_upload: '待上传',
  waiting_accept: '待接收',
  pending_confirm: '待确认',
  ready: '可用',
  failed: '失败',
  disabled: '停用',
  expired: '已超时',
}

function statusVariant(
  status: PortraitAssetStatus
): 'default' | 'secondary' | 'destructive' {
  if (status === 'ready') return 'default'
  if (status === 'failed' || status === 'disabled' || status === 'expired') {
    return 'destructive'
  }
  return 'secondary'
}

function assetUri(assetId?: string) {
  if (!assetId) return ''
  return assetId.startsWith('asset://') ? assetId : `asset://${assetId}`
}

function canRequestAccept(status: PortraitAssetStatus) {
  return ['qr_ready', 'waiting_upload', 'waiting_accept'].includes(status)
}

function isActiveJob(status: PortraitAssetStatus) {
  return [
    'pending',
    'qr_ready',
    'waiting_upload',
    'waiting_accept',
    'pending_confirm',
  ].includes(status)
}

function formatCountdown(expiresTime?: number, nowSeconds?: number) {
  if (!expiresTime || !nowSeconds) return ''
  const seconds = Math.max(0, expiresTime - nowSeconds)
  const minutes = Math.floor(seconds / 60)
  const rest = seconds % 60
  return `${minutes}:${String(rest).padStart(2, '0')}`
}

export function PortraitAssets() {
  const [jobs, setJobs] = useState<PortraitAssetJob[]>([])
  const [loading, setLoading] = useState(true)
  const [creating, setCreating] = useState(false)
  const [acceptingId, setAcceptingId] = useState<number | null>(null)
  const [confirmingId, setConfirmingId] = useState<number | null>(null)
  const [rejectingId, setRejectingId] = useState<number | null>(null)
  const [nowSeconds, setNowSeconds] = useState(() =>
    Math.floor(Date.now() / 1000)
  )
  const [name, setName] = useState('')

  const readyJobs = useMemo(
    () => jobs.filter((job) => job.status === 'ready' && job.asset_id),
    [jobs]
  )

  const activeJob = useMemo(
    () => jobs.find((job) => isActiveJob(job.status)),
    [jobs]
  )

  const assignedQRJob = useMemo(
    () =>
      jobs.find(
        (job) => job.status === 'qr_ready' && (job.qr_image || job.invite_url)
      ),
    [jobs]
  )

  const confirmJob = useMemo(
    () => jobs.find((job) => job.status === 'pending_confirm'),
    [jobs]
  )

  const qrCountdown = formatCountdown(
    assignedQRJob?.qr_expires_time,
    nowSeconds
  )
  const activeJobId = activeJob?.id
  const activeJobStatus = activeJob?.status

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

  useEffect(() => {
    const timer = window.setInterval(() => {
      setNowSeconds(Math.floor(Date.now() / 1000))
    }, 1000)
    return () => window.clearInterval(timer)
  }, [])

  useEffect(() => {
    if (!assignedQRJob?.qr_expires_time) return
    const currentSeconds = Math.floor(Date.now() / 1000)
    const delay = Math.max(
      0,
      (assignedQRJob.qr_expires_time - currentSeconds + 1) * 1000
    )
    const timer = window.setTimeout(fetchJobs, delay)
    return () => window.clearTimeout(timer)
  }, [assignedQRJob?.qr_expires_time, fetchJobs])

  useEffect(() => {
    if (!activeJobId) return
    const timer = window.setInterval(fetchJobs, 5000)
    return () => window.clearInterval(timer)
  }, [activeJobId, activeJobStatus, fetchJobs])

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

  async function handleRequestAccept(job: PortraitAssetJob) {
    try {
      setAcceptingId(job.id)
      const res = await requestPortraitAssetAccept(job.id)
      if (res.success) {
        toast.success('已提交接收授权，请稍后刷新查看结果')
        await fetchJobs()
      } else {
        toast.error(res.message || '提交失败')
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : '提交失败')
    } finally {
      setAcceptingId(null)
    }
  }

  async function handleConfirm(job: PortraitAssetJob) {
    try {
      setConfirmingId(job.id)
      const res = await confirmPortraitAsset(job.id)
      if (res.success) {
        toast.success('已确认真人资产')
        await fetchJobs()
      } else {
        toast.error(res.message || '确认失败')
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : '确认失败')
    } finally {
      setConfirmingId(null)
    }
  }

  async function handleReject(job: PortraitAssetJob) {
    try {
      setRejectingId(job.id)
      const res = await rejectPortraitAsset(job.id)
      if (res.success) {
        toast.success('已标记为非本人资产')
        await fetchJobs()
      } else {
        toast.error(res.message || '操作失败')
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : '操作失败')
    } finally {
      setRejectingId(null)
    }
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
                  <Button
                    onClick={handleCreate}
                    disabled={creating || !!activeJob}
                  >
                    {creating ? <Loader2 className='animate-spin' /> : <Plus />}
                    创建
                  </Button>
                </div>
                {activeJob && (
                  <div className='text-muted-foreground mt-3 text-xs'>
                    当前已有进行中的任务：
                    {statusLabels[activeJob.status] || activeJob.status}
                    {activeJob.status === 'pending' &&
                      activeJob.queue_position &&
                      `，队列第 ${activeJob.queue_position} 位`}
                  </div>
                )}
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
                          <TableCell>
                            <div className='flex items-center gap-2'>
                              {job.asset_preview && (
                                <img
                                  src={job.asset_preview}
                                  alt=''
                                  className='bg-muted size-10 shrink-0 rounded-md border object-cover'
                                />
                              )}
                              <div className='min-w-0'>
                                <div className='truncate'>{job.name}</div>
                                {job.asset_status && (
                                  <div className='text-muted-foreground text-xs'>
                                    火山状态：{job.asset_status}
                                  </div>
                                )}
                              </div>
                            </div>
                          </TableCell>
                          <TableCell>
                            <div className='space-y-1'>
                              <Badge variant={statusVariant(job.status)}>
                                {statusLabels[job.status] || job.status}
                              </Badge>
                              {job.error_message && (
                                <div className='text-destructive max-w-[220px] text-xs'>
                                  {job.error_message}
                                </div>
                              )}
                            </div>
                          </TableCell>
                          <TableCell className='max-w-[260px] truncate font-mono text-xs'>
                            {job.status === 'ready'
                              ? assetUri(job.asset_id) || '-'
                              : '-'}
                          </TableCell>
                          <TableCell>
                            {formatTimestampToDate(job.updated_time)}
                          </TableCell>
                          <TableCell>
                            <div className='flex justify-end gap-1'>
                              {canRequestAccept(job.status) && (
                                <Button
                                  size='sm'
                                  variant='outline'
                                  disabled={acceptingId === job.id}
                                  onClick={() => handleRequestAccept(job)}
                                >
                                  {acceptingId === job.id ? (
                                    <Loader2 className='animate-spin' />
                                  ) : (
                                    <Send />
                                  )}
                                  接收授权
                                </Button>
                              )}
                              <Button
                                size='icon-sm'
                                variant='ghost'
                                disabled={
                                  !job.asset_id || job.status !== 'ready'
                                }
                                onClick={() =>
                                  handleCopy(assetUri(job.asset_id), 'Asset ID')
                                }
                              >
                                <Clipboard />
                              </Button>
                              {job.status === 'pending_confirm' && (
                                <>
                                  <Button
                                    size='sm'
                                    disabled={confirmingId === job.id}
                                    onClick={() => handleConfirm(job)}
                                  >
                                    {confirmingId === job.id ? (
                                      <Loader2 className='animate-spin' />
                                    ) : (
                                      <CheckCircle2 />
                                    )}
                                    确认是本人
                                  </Button>
                                  <Button
                                    size='sm'
                                    variant='outline'
                                    disabled={rejectingId === job.id}
                                    onClick={() => handleReject(job)}
                                  >
                                    {rejectingId === job.id ? (
                                      <Loader2 className='animate-spin' />
                                    ) : (
                                      <ShieldAlert />
                                    )}
                                    不是本人
                                  </Button>
                                </>
                              )}
                            </div>
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
                {confirmJob ? (
                  <div className='space-y-3'>
                    {confirmJob.asset_preview && (
                      <img
                        src={confirmJob.asset_preview}
                        alt=''
                        className='bg-background aspect-square w-full rounded-md border object-cover'
                      />
                    )}
                    <div className='flex items-center justify-between gap-2'>
                      <Badge variant={statusVariant(confirmJob.status)}>
                        {statusLabels[confirmJob.status]}
                      </Badge>
                      {confirmJob.asset_status && (
                        <span className='text-muted-foreground text-xs'>
                          火山状态：{confirmJob.asset_status}
                        </span>
                      )}
                    </div>
                    <div className='grid grid-cols-2 gap-2'>
                      <Button
                        disabled={confirmingId === confirmJob.id}
                        onClick={() => handleConfirm(confirmJob)}
                      >
                        {confirmingId === confirmJob.id ? (
                          <Loader2 className='animate-spin' />
                        ) : (
                          <CheckCircle2 />
                        )}
                        确认是本人
                      </Button>
                      <Button
                        variant='outline'
                        disabled={rejectingId === confirmJob.id}
                        onClick={() => handleReject(confirmJob)}
                      >
                        {rejectingId === confirmJob.id ? (
                          <Loader2 className='animate-spin' />
                        ) : (
                          <ShieldAlert />
                        )}
                        不是本人
                      </Button>
                    </div>
                  </div>
                ) : assignedQRJob?.qr_image ? (
                  <div className='space-y-3'>
                    <img
                      src={assignedQRJob.qr_image}
                      alt='portrait asset invitation QR code'
                      className='bg-background aspect-square w-full rounded-md border object-contain p-3'
                    />
                    <div className='flex items-center justify-between gap-2'>
                      <Badge variant={statusVariant(assignedQRJob.status)}>
                        {statusLabels[assignedQRJob.status]}
                      </Badge>
                      <Button
                        size='sm'
                        variant='outline'
                        disabled={!assignedQRJob.invite_url}
                        onClick={() =>
                          handleCopy(assignedQRJob.invite_url || '', '邀请链接')
                        }
                      >
                        <Clipboard />
                        复制链接
                      </Button>
                    </div>
                    {qrCountdown && (
                      <div className='text-muted-foreground text-center text-xs'>
                        剩余扫码时间 {qrCountdown}
                      </div>
                    )}
                    {canRequestAccept(assignedQRJob.status) && (
                      <Button
                        className='w-full'
                        disabled={
                          acceptingId === assignedQRJob.id ||
                          Boolean(
                            assignedQRJob.qr_expires_time &&
                            assignedQRJob.qr_expires_time <= nowSeconds
                          )
                        }
                        onClick={() => handleRequestAccept(assignedQRJob)}
                      >
                        {acceptingId === assignedQRJob.id ? (
                          <Loader2 className='animate-spin' />
                        ) : (
                          <Send />
                        )}
                        我已完成扫码，接收授权
                      </Button>
                    )}
                  </div>
                ) : activeJob?.status === 'pending' ? (
                  <div className='text-muted-foreground flex aspect-square flex-col items-center justify-center gap-2 rounded-md border text-sm'>
                    <span>排队中</span>
                    {activeJob.queue_position ? (
                      <span>当前第 {activeJob.queue_position} 位</span>
                    ) : null}
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
