import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Clipboard,
  ExternalLink,
  Loader2,
  RefreshCw,
  ShieldAlert,
  Trash2,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { isAxiosError } from 'axios'
import { SectionPageLayout } from '@/components/layout'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { formatTimestampToDate } from '@/lib/format'
import {
  deleteAdminOfficialPortraitAsset,
  deleteAdminVirtualPortraitAsset,
  getAdminOfficialPortraitAssets,
  getAdminVirtualPortraitAssets,
  syncAdminOfficialPortraitAsset,
  syncAdminVirtualPortraitAsset,
  type AdminOfficialPortraitAssetItem,
  type AdminVirtualPortraitAssetItem,
} from './api'

type AssetTab = 'official' | 'virtual'

const officialStatusOptions = [
  'all',
  'pending',
  'validate_ready',
  'validated',
  'asset_processing',
  'pending_confirm',
  'ready',
  'failed',
  'expired',
  'disabled',
] as const

const virtualStatusOptions = ['all', 'processing', 'active', 'failed'] as const

function assetUri(assetId?: string) {
  if (!assetId) return ''
  return assetId.startsWith('asset://') ? assetId : `asset://${assetId}`
}

function getErrorMessage(error: unknown) {
  if (isAxiosError<{ message?: string; error?: { message?: string } }>(error)) {
    return (
      error.response?.data?.message ||
      error.response?.data?.error?.message ||
      error.message
    )
  }
  return error instanceof Error ? error.message : 'Request failed'
}

function PreviewImage({ src }: { src?: string }) {
  const [failed, setFailed] = useState(false)

  if (!src || failed) return null

  return (
    <img
      src={src}
      alt=''
      loading='lazy'
      className='h-12 w-12 rounded-md border object-cover'
      onError={() => setFailed(true)}
    />
  )
}

export function AdminPortraitAssets() {
  const { t } = useTranslation()
  const [tab, setTab] = useState<AssetTab>('official')
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [officialItems, setOfficialItems] = useState<AdminOfficialPortraitAssetItem[]>([])
  const [virtualItems, setVirtualItems] = useState<AdminVirtualPortraitAssetItem[]>([])
  const [officialTotal, setOfficialTotal] = useState(0)
  const [virtualTotal, setVirtualTotal] = useState(0)
  const [userId, setUserId] = useState('')
  const [externalUserID, setExternalUserID] = useState('')
  const [keyword, setKeyword] = useState('')
  const [assetID, setAssetID] = useState('')
  const [officialStatus, setOfficialStatus] =
    useState<(typeof officialStatusOptions)[number]>('all')
  const [virtualStatus, setVirtualStatus] =
    useState<(typeof virtualStatusOptions)[number]>('all')
  const [syncingKey, setSyncingKey] = useState<string | null>(null)
  const [deletingTarget, setDeletingTarget] = useState<{
    id: number
    tab: AssetTab
    name: string
  } | null>(null)

  const params = useMemo(() => {
    const parsedUserID = Number(userId)
    return {
      p: 1,
      page_size: 50,
      user_id: Number.isFinite(parsedUserID) && parsedUserID > 0 ? parsedUserID : undefined,
      external_user_id: externalUserID.trim() || undefined,
      keyword: keyword.trim() || undefined,
      asset_id: assetID.trim() || undefined,
    }
  }, [assetID, externalUserID, keyword, userId])

  const fetchData = useCallback(async () => {
    try {
      setLoading(true)
      const [officialRes, virtualRes] = await Promise.all([
        getAdminOfficialPortraitAssets({
          ...params,
          status: officialStatus === 'all' ? undefined : officialStatus,
        }),
        getAdminVirtualPortraitAssets({
          ...params,
          status: virtualStatus === 'all' ? undefined : virtualStatus,
        }),
      ])
      if (officialRes.success && officialRes.data) {
        setOfficialItems(officialRes.data.items || [])
        setOfficialTotal(officialRes.data.total || 0)
      }
      if (virtualRes.success && virtualRes.data) {
        setVirtualItems(virtualRes.data.items || [])
        setVirtualTotal(virtualRes.data.total || 0)
      }
    } catch (error) {
      toast.error(getErrorMessage(error))
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }, [officialStatus, params, virtualStatus])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  async function handleRefresh() {
    setRefreshing(true)
    await fetchData()
  }

  async function handleCopy(text: string, successText: string) {
    if (!text) return
    await navigator.clipboard.writeText(text)
    toast.success(successText)
  }

  async function handleSync(id: number, currentTab: AssetTab) {
    try {
      setSyncingKey(`${currentTab}:${id}`)
      const res =
        currentTab === 'official'
          ? await syncAdminOfficialPortraitAsset(id)
          : await syncAdminVirtualPortraitAsset(id)
      if (res.success) {
        toast.success(t('Sync completed'))
        await fetchData()
      }
    } finally {
      setSyncingKey(null)
    }
  }

  async function handleDelete() {
    if (!deletingTarget) return
    try {
      if (deletingTarget.tab === 'official') {
        await deleteAdminOfficialPortraitAsset(deletingTarget.id)
      } else {
        await deleteAdminVirtualPortraitAsset(deletingTarget.id)
      }
      toast.success(t('Deleted successfully'))
      await fetchData()
    } finally {
      setDeletingTarget(null)
    }
  }

  return (
    <>
      <SectionPageLayout>
        <SectionPageLayout.Title>
          {t('Portrait Asset Management')}
        </SectionPageLayout.Title>
        <SectionPageLayout.Description>
          {t('View and manage official and virtual portrait assets across all accounts.')}
        </SectionPageLayout.Description>
        <SectionPageLayout.Actions>
          <Button variant='outline' onClick={handleRefresh} disabled={refreshing}>
            {refreshing ? (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            ) : (
              <RefreshCw className='mr-2 h-4 w-4' />
            )}
            {t('Refresh')}
          </Button>
        </SectionPageLayout.Actions>
        <SectionPageLayout.Content>
          <div className='space-y-4'>
            <div className='grid gap-3 md:grid-cols-2 xl:grid-cols-5'>
              <Input
                value={userId}
                onChange={(e) => setUserId(e.target.value)}
                placeholder={t('Filter by user ID')}
              />
              <Input
                value={externalUserID}
                onChange={(e) => setExternalUserID(e.target.value)}
                placeholder={t('Filter by external_user_id')}
              />
              <Input
                value={keyword}
                onChange={(e) => setKeyword(e.target.value)}
                placeholder={t('Filter by name or group')}
              />
              <Input
                value={assetID}
                onChange={(e) => setAssetID(e.target.value)}
                placeholder={t('Filter by asset ID')}
              />
              <div className='flex items-center rounded-lg border px-3 text-sm text-muted-foreground'>
                {t('Official')}: {officialTotal} / {t('Virtual')}: {virtualTotal}
              </div>
            </div>

            <Tabs value={tab} onValueChange={(value) => setTab(value as AssetTab)}>
              <TabsList className='grid w-full grid-cols-2'>
                <TabsTrigger value='official'>
                  {t('Official Portrait Assets')}
                </TabsTrigger>
                <TabsTrigger value='virtual'>
                  {t('Virtual Portrait Assets')}
                </TabsTrigger>
              </TabsList>

              <TabsContent value='official' className='space-y-4'>
                <div className='flex items-center gap-3'>
                  <Select
                    value={officialStatus}
                    onValueChange={(value) =>
                      setOfficialStatus(value as (typeof officialStatusOptions)[number])
                    }
                  >
                    <SelectTrigger className='w-[220px]'>
                      <SelectValue placeholder={t('Filter by status')} />
                    </SelectTrigger>
                    <SelectContent>
                      {officialStatusOptions.map((status) => (
                        <SelectItem key={status} value={status}>
                          {status === 'all' ? t('All statuses') : status}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className='rounded-lg border'>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{t('Preview')}</TableHead>
                        <TableHead>{t('Asset')}</TableHead>
                        <TableHead>{t('Owner')}</TableHead>
                        <TableHead>{t('Status')}</TableHead>
                        <TableHead>{t('Created')}</TableHead>
                        <TableHead className='text-right'>{t('Actions')}</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {officialItems.length === 0 ? (
                        <TableRow>
                          <TableCell colSpan={6} className='py-10 text-center text-muted-foreground'>
                            {loading ? t('Loading...') : t('No official portrait assets found')}
                          </TableCell>
                        </TableRow>
                      ) : (
                        officialItems.map((item) => {
                          const syncKey = `official:${item.id}`
                          return (
                            <TableRow key={item.id}>
                              <TableCell>
                                <PreviewImage src={item.asset_preview} />
                              </TableCell>
                              <TableCell className='space-y-1'>
                                <div className='font-medium'>{item.name}</div>
                                <div className='text-xs text-muted-foreground'>
                                  ID #{item.id}
                                </div>
                                {item.asset_id && (
                                  <div className='text-xs text-muted-foreground'>
                                    {assetUri(item.asset_id)}
                                  </div>
                                )}
                              </TableCell>
                              <TableCell className='space-y-1 text-sm'>
                                <div>{item.owner.username || '-'}</div>
                                <div className='text-xs text-muted-foreground'>
                                  UID {item.owner.user_id}
                                </div>
                                {item.external_user_id && (
                                  <div className='text-xs text-muted-foreground'>
                                    ext: {item.external_user_id}
                                  </div>
                                )}
                              </TableCell>
                              <TableCell className='space-y-2'>
                                <Badge variant={item.status === 'failed' ? 'destructive' : 'secondary'}>
                                  {item.status}
                                </Badge>
                                {item.error_message && (
                                  <div className='max-w-xs text-xs text-muted-foreground'>
                                    {item.error_message}
                                  </div>
                                )}
                              </TableCell>
                              <TableCell className='text-sm text-muted-foreground'>
                                {formatTimestampToDate(item.created_time)}
                              </TableCell>
                              <TableCell>
                                <div className='flex justify-end gap-2'>
                                  {item.asset_id && (
                                    <Button
                                      size='icon'
                                      variant='outline'
                                      onClick={() =>
                                        handleCopy(assetUri(item.asset_id), t('Asset URI copied'))
                                      }
                                    >
                                      <Clipboard className='h-4 w-4' />
                                    </Button>
                                  )}
                                  {item.asset_preview && (
                                    <Button size='icon' variant='outline' asChild>
                                      <a href={item.asset_preview} target='_blank' rel='noreferrer'>
                                        <ExternalLink className='h-4 w-4' />
                                      </a>
                                    </Button>
                                  )}
                                  <Button
                                    size='icon'
                                    variant='outline'
                                    disabled={syncingKey === syncKey}
                                    onClick={() => handleSync(item.id, 'official')}
                                  >
                                    {syncingKey === syncKey ? (
                                      <Loader2 className='h-4 w-4 animate-spin' />
                                    ) : (
                                      <RefreshCw className='h-4 w-4' />
                                    )}
                                  </Button>
                                  <Button
                                    size='icon'
                                    variant='destructive'
                                    onClick={() =>
                                      setDeletingTarget({
                                        id: item.id,
                                        tab: 'official',
                                        name: item.name,
                                      })
                                    }
                                  >
                                    <Trash2 className='h-4 w-4' />
                                  </Button>
                                </div>
                              </TableCell>
                            </TableRow>
                          )
                        })
                      )}
                    </TableBody>
                  </Table>
                </div>
              </TabsContent>

              <TabsContent value='virtual' className='space-y-4'>
                <div className='flex items-center gap-3'>
                  <Select
                    value={virtualStatus}
                    onValueChange={(value) =>
                      setVirtualStatus(value as (typeof virtualStatusOptions)[number])
                    }
                  >
                    <SelectTrigger className='w-[220px]'>
                      <SelectValue placeholder={t('Filter by status')} />
                    </SelectTrigger>
                    <SelectContent>
                      {virtualStatusOptions.map((status) => (
                        <SelectItem key={status} value={status}>
                          {status === 'all' ? t('All statuses') : status}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className='rounded-lg border'>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{t('Preview')}</TableHead>
                        <TableHead>{t('Asset')}</TableHead>
                        <TableHead>{t('Owner')}</TableHead>
                        <TableHead>{t('Status')}</TableHead>
                        <TableHead>{t('Created')}</TableHead>
                        <TableHead className='text-right'>{t('Actions')}</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {virtualItems.length === 0 ? (
                        <TableRow>
                          <TableCell colSpan={6} className='py-10 text-center text-muted-foreground'>
                            {loading ? t('Loading...') : t('No virtual portrait assets found')}
                          </TableCell>
                        </TableRow>
                      ) : (
                        virtualItems.map((item) => {
                          const syncKey = `virtual:${item.id}`
                          return (
                            <TableRow key={item.id}>
                              <TableCell>
                                <PreviewImage src={item.preview_url} />
                              </TableCell>
                              <TableCell className='space-y-1'>
                                <div className='font-medium'>{item.name}</div>
                                <div className='text-xs text-muted-foreground'>
                                  ID #{item.id}
                                </div>
                                {item.volc_asset_id && (
                                  <div className='text-xs text-muted-foreground'>
                                    {assetUri(item.volc_asset_id)}
                                  </div>
                                )}
                              </TableCell>
                              <TableCell className='space-y-1 text-sm'>
                                <div>{item.owner.username || '-'}</div>
                                <div className='text-xs text-muted-foreground'>
                                  UID {item.owner.user_id}
                                </div>
                                {item.external_user_id && (
                                  <div className='text-xs text-muted-foreground'>
                                    ext: {item.external_user_id}
                                  </div>
                                )}
                              </TableCell>
                              <TableCell className='space-y-2'>
                                <Badge variant={item.status === 'failed' ? 'destructive' : 'secondary'}>
                                  {item.status}
                                </Badge>
                                {item.error_message && (
                                  <div className='max-w-xs text-xs text-muted-foreground'>
                                    {item.error_message}
                                  </div>
                                )}
                              </TableCell>
                              <TableCell className='text-sm text-muted-foreground'>
                                {formatTimestampToDate(item.created_time)}
                              </TableCell>
                              <TableCell>
                                <div className='flex justify-end gap-2'>
                                  {item.volc_asset_id && (
                                    <Button
                                      size='icon'
                                      variant='outline'
                                      onClick={() =>
                                        handleCopy(
                                          assetUri(item.volc_asset_id),
                                          t('Asset URI copied')
                                        )
                                      }
                                    >
                                      <Clipboard className='h-4 w-4' />
                                    </Button>
                                  )}
                                  {item.preview_url && (
                                    <Button size='icon' variant='outline' asChild>
                                      <a href={item.preview_url} target='_blank' rel='noreferrer'>
                                        <ExternalLink className='h-4 w-4' />
                                      </a>
                                    </Button>
                                  )}
                                  <Button
                                    size='icon'
                                    variant='outline'
                                    disabled={syncingKey === syncKey}
                                    onClick={() => handleSync(item.id, 'virtual')}
                                  >
                                    {syncingKey === syncKey ? (
                                      <Loader2 className='h-4 w-4 animate-spin' />
                                    ) : (
                                      <RefreshCw className='h-4 w-4' />
                                    )}
                                  </Button>
                                  <Button
                                    size='icon'
                                    variant='destructive'
                                    onClick={() =>
                                      setDeletingTarget({
                                        id: item.id,
                                        tab: 'virtual',
                                        name: item.name,
                                      })
                                    }
                                  >
                                    <Trash2 className='h-4 w-4' />
                                  </Button>
                                </div>
                              </TableCell>
                            </TableRow>
                          )
                        })
                      )}
                    </TableBody>
                  </Table>
                </div>
              </TabsContent>
            </Tabs>

            <div className='flex items-start gap-3 rounded-lg border border-dashed p-4 text-sm text-muted-foreground'>
              <ShieldAlert className='mt-0.5 h-4 w-4 shrink-0' />
              <p>
                {t('This page only manages local records and existing management actions. It does not change video generation rules or asset isolation rules.')}
              </p>
            </div>
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <ConfirmDialog
        destructive
        open={Boolean(deletingTarget)}
        onOpenChange={(open) => !open && setDeletingTarget(null)}
        handleConfirm={handleDelete}
        className='max-w-md'
        title={t('Delete portrait asset')}
        desc={t('This will delete the selected portrait asset record: {{name}}', {
          name: deletingTarget?.name || '',
        })}
        confirmText={t('Delete')}
      />
    </>
  )
}
