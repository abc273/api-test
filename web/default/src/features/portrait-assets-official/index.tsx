import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  CheckCircle2,
  Clipboard,
  ExternalLink,
  Folder,
  FolderOpen,
  FolderPlus,
  Loader2,
  Plus,
  RefreshCw,
  Send,
  ShieldAlert,
  Trash2,
  Upload,
  UserCheck,
} from 'lucide-react'
import { QRCodeSVG } from 'qrcode.react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { formatTimestampToDate } from '@/lib/format'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { SectionPageLayout } from '@/components/layout'
import {
  createPortraitAssetFolder,
  deletePortraitAssetFolder,
  getPortraitAssetFolders,
  movePortraitAssetsToFolder,
  type PortraitAssetFolder,
} from '@/features/portrait-asset-folders/api'
import {
  confirmOfficialPortraitAsset,
  createOfficialPortraitAssetJob,
  deleteOfficialPortraitAssetJob,
  getOfficialPortraitAssetConfig,
  getOfficialPortraitAssets,
  refreshOfficialPortraitValidation,
  rejectOfficialPortraitAsset,
  submitOfficialPortraitAsset,
  syncOfficialPortraitAssetJob,
  uploadOfficialPortraitAssetMaterial,
} from './api'
import type {
  OfficialPortraitAssetConfig,
  OfficialPortraitAssetJob,
  OfficialPortraitAssetStatus,
} from './types'

const activeStatuses: OfficialPortraitAssetStatus[] = [
  'pending',
  'validate_ready',
  'validated',
  'asset_processing',
  'pending_confirm',
]

function statusVariant(
  status: OfficialPortraitAssetStatus
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

function stripFileExtension(fileName: string) {
  const trimmed = fileName.trim()
  const dotIndex = trimmed.lastIndexOf('.')
  if (dotIndex <= 0) return trimmed
  return trimmed.slice(0, dotIndex)
}

function PreviewImage({ src, className }: { src?: string; className: string }) {
  const [failedSrc, setFailedSrc] = useState<string | null>(null)

  if (!src || failedSrc === src) return null

  return (
    <img
      src={src}
      alt=''
      loading='lazy'
      className={className}
      onError={() => setFailedSrc(src)}
    />
  )
}

export function OfficialPortraitAssets() {
  const { t } = useTranslation()
  const [config, setConfig] = useState<OfficialPortraitAssetConfig | null>(null)
  const [jobs, setJobs] = useState<OfficialPortraitAssetJob[]>([])
  const [folders, setFolders] = useState<PortraitAssetFolder[]>([])
  const [selectedFolderId, setSelectedFolderId] = useState<number | 'all'>(
    'all'
  )
  const [loading, setLoading] = useState(true)
  const [creating, setCreating] = useState(false)
  const [creatingFolder, setCreatingFolder] = useState(false)
  const [deletingFolder, setDeletingFolder] =
    useState<PortraitAssetFolder | null>(null)
  const [deletingAsset, setDeletingAsset] =
    useState<OfficialPortraitAssetJob | null>(null)
  const [deletingAssetId, setDeletingAssetId] = useState<number | null>(null)
  const [movingAssetId, setMovingAssetId] = useState<number | null>(null)
  const [refreshingId, setRefreshingId] = useState<number | null>(null)
  const [syncingId, setSyncingId] = useState<number | null>(null)
  const [uploading, setUploading] = useState(false)
  const [submittingId, setSubmittingId] = useState<number | null>(null)
  const [confirmingId, setConfirmingId] = useState<number | null>(null)
  const [rejectingId, setRejectingId] = useState<number | null>(null)
  const [name, setName] = useState('')
  const [newFolderName, setNewFolderName] = useState('')
  const [createFolderId, setCreateFolderId] = useState(0)
  const [callbackUrl, setCallbackUrl] = useState('')
  const [assetUrl, setAssetUrl] = useState('')
  const [assetName, setAssetName] = useState('')
  const [assetType, setAssetType] = useState('Image')
  const [assetFolderId, setAssetFolderId] = useState(0)
  const fileInputRef = useRef<HTMLInputElement | null>(null)

  const statusLabels = useMemo<Record<OfficialPortraitAssetStatus, string>>(
    () => ({
      pending: t('Creating validation link'),
      validate_ready: t('Pending validation'),
      validated: t('Validation passed'),
      asset_processing: t('Asset processing'),
      pending_confirm: t('Pending confirmation'),
      ready: t('Ready'),
      failed: t('Failed'),
      expired: t('Expired'),
      disabled: t('Disabled'),
    }),
    [t]
  )

  const activeJob = useMemo(
    () => jobs.find((job) => activeStatuses.includes(job.status)),
    [jobs]
  )

  const readyJobs = useMemo(
    () => jobs.filter((job) => job.status === 'ready' && job.asset_id),
    [jobs]
  )

  const folderById = useMemo(
    () => new Map(folders.map((folder) => [folder.id, folder])),
    [folders]
  )
  const folderNameById = useMemo(
    () => new Map(folders.map((folder) => [folder.id, folder.name])),
    [folders]
  )

  const currentFolderLabel =
    selectedFolderId === 'all'
      ? t('All assets')
      : selectedFolderId === 0
        ? t('Ungrouped')
        : folderNameById.get(selectedFolderId) || t('Folder')

  const validationExpiresAt =
    activeJob?.status === 'validate_ready'
      ? formatTimestampToDate(activeJob.qr_expires_time)
      : ''

  function normalizeFolderExternalUserID(externalUserID?: string) {
    return externalUserID?.trim() || ''
  }

  function getFolderExternalUserID(folderId: number) {
    if (folderId <= 0) return ''
    return normalizeFolderExternalUserID(
      folderById.get(folderId)?.external_user_id
    )
  }

  function compatibleFoldersForExternalUser(externalUserID?: string) {
    const normalizedExternalUserID =
      normalizeFolderExternalUserID(externalUserID)
    return folders.filter(
      (folder) =>
        normalizeFolderExternalUserID(folder.external_user_id) ===
        normalizedExternalUserID
    )
  }

  const fetchData = useCallback(async () => {
    try {
      setLoading(true)
      const folderId = selectedFolderId === 'all' ? undefined : selectedFolderId
      const [configRes, jobsRes, foldersRes] = await Promise.all([
        getOfficialPortraitAssetConfig(),
        getOfficialPortraitAssets({ p: 1, page_size: 50, folder_id: folderId }),
        getPortraitAssetFolders({ asset_kind: 'official' }),
      ])
      if (configRes.success && configRes.data) setConfig(configRes.data)
      if (jobsRes.success && jobsRes.data) {
        setJobs(jobsRes.data.items || [])
      }
      if (foldersRes.success && foldersRes.data) {
        setFolders(foldersRes.data)
      }
    } finally {
      setLoading(false)
    }
  }, [selectedFolderId])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  useEffect(() => {
    if (selectedFolderId !== 'all') {
      setCreateFolderId(selectedFolderId)
      setAssetFolderId(selectedFolderId)
    }
  }, [selectedFolderId])

  async function handleCreate() {
    try {
      setCreating(true)
      const res = await createOfficialPortraitAssetJob({
        name,
        callback_url: callbackUrl,
        folder_id: createFolderId,
        external_user_id: getFolderExternalUserID(createFolderId) || undefined,
      })
      if (res.success) {
        toast.success(t('Validation link created'))
        setName('')
        setCallbackUrl('')
        await fetchData()
      } else {
        toast.error(res.message || t('Create failed'))
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : t('Create failed'))
    } finally {
      setCreating(false)
    }
  }

  async function handleCreateFolder() {
    const folderName = newFolderName.trim()
    if (!folderName) {
      toast.error(t('Folder name is required'))
      return
    }
    try {
      setCreatingFolder(true)
      const res = await createPortraitAssetFolder({
        asset_kind: 'official',
        name: folderName,
        external_user_id:
          selectedFolderId !== 'all' && selectedFolderId > 0
            ? getFolderExternalUserID(selectedFolderId)
            : undefined,
      })
      if (res.success && res.data) {
        toast.success(t('Folder created'))
        setNewFolderName('')
        setSelectedFolderId(res.data.id)
        await fetchData()
      } else {
        toast.error(res.message || t('Create failed'))
      }
    } finally {
      setCreatingFolder(false)
    }
  }

  async function handleDeleteFolder(folder: PortraitAssetFolder) {
    try {
      const res = await deletePortraitAssetFolder(folder.id, {
        external_user_id:
          normalizeFolderExternalUserID(folder.external_user_id) || undefined,
      })
      if (res.success) {
        toast.success(t('Folder deleted'))
        if (selectedFolderId === folder.id) {
          setSelectedFolderId('all')
        }
        setDeletingFolder(null)
        await fetchData()
      } else {
        toast.error(res.message || t('Delete failed'))
      }
    } finally {
      setDeletingFolder(null)
    }
  }

  async function handleDeleteAsset(job: OfficialPortraitAssetJob) {
    try {
      setDeletingAssetId(job.id)
      const res = await deleteOfficialPortraitAssetJob(job.id, {
        external_user_id:
          normalizeFolderExternalUserID(job.external_user_id) || undefined,
      })
      if (res.success) {
        toast.success(t('Asset deleted'))
        setDeletingAsset(null)
        await fetchData()
      } else {
        toast.error(res.message || t('Delete failed'))
      }
    } finally {
      setDeletingAssetId(null)
    }
  }

  async function handleMoveToFolder(
    job: OfficialPortraitAssetJob,
    folderId: number
  ) {
    try {
      setMovingAssetId(job.id)
      const res = await movePortraitAssetsToFolder({
        asset_kind: 'official',
        asset_ids: [job.id],
        folder_id: folderId,
        external_user_id:
          normalizeFolderExternalUserID(job.external_user_id) || undefined,
      })
      if (res.success) {
        toast.success(t('Asset moved'))
        await fetchData()
      } else {
        toast.error(res.message || t('Move failed'))
      }
    } finally {
      setMovingAssetId(null)
    }
  }

  async function handleRefreshValidation(job: OfficialPortraitAssetJob) {
    try {
      setRefreshingId(job.id)
      const res = await refreshOfficialPortraitValidation(job.id)
      if (res.success) {
        toast.success(t('Validation link refreshed'))
        await fetchData()
      } else {
        toast.error(res.message || t('Refresh failed'))
      }
    } finally {
      setRefreshingId(null)
    }
  }

  async function handleSync(job: OfficialPortraitAssetJob) {
    try {
      setSyncingId(job.id)
      const res = await syncOfficialPortraitAssetJob(job.id)
      if (res.success) {
        await fetchData()
      } else {
        toast.error(res.message || t('Sync failed'))
      }
    } finally {
      setSyncingId(null)
    }
  }

  async function handleSubmitAsset(job: OfficialPortraitAssetJob) {
    try {
      setSubmittingId(job.id)
      const targetFolderId =
        assetFolderId > 0 &&
        getFolderExternalUserID(assetFolderId) !==
          normalizeFolderExternalUserID(job.external_user_id)
          ? 0
          : assetFolderId
      const res = await submitOfficialPortraitAsset(job.id, {
        asset_url: assetUrl,
        asset_type: assetType,
        name: assetName,
        folder_id: targetFolderId,
      })
      if (res.success) {
        toast.success(t('Asset submitted'))
        setAssetUrl('')
        setAssetName('')
        await fetchData()
      } else {
        toast.error(res.message || t('Submit failed'))
      }
    } finally {
      setSubmittingId(null)
    }
  }

  async function handleUploadFile(file?: File | null) {
    if (!file) return
    try {
      setUploading(true)
      const res = await uploadOfficialPortraitAssetMaterial(file)
      if (res.success && res.data) {
        setAssetUrl(res.data.url)
        setAssetType(res.data.asset_type)
        setAssetName((current) => current || stripFileExtension(file.name))
        toast.success(t('Material uploaded'))
      } else {
        toast.error(res.message || t('Upload failed'))
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : t('Upload failed'))
    } finally {
      setUploading(false)
      if (fileInputRef.current) fileInputRef.current.value = ''
    }
  }

  async function handleConfirm(job: OfficialPortraitAssetJob) {
    try {
      setConfirmingId(job.id)
      const res = await confirmOfficialPortraitAsset(job.id)
      if (res.success) {
        toast.success(t('Portrait asset confirmed'))
        await fetchData()
      } else {
        toast.error(res.message || t('Confirm failed'))
      }
    } finally {
      setConfirmingId(null)
    }
  }

  async function handleReject(job: OfficialPortraitAssetJob) {
    try {
      setRejectingId(job.id)
      const res = await rejectOfficialPortraitAsset(job.id)
      if (res.success) {
        toast.success(t('Portrait asset rejected'))
        await fetchData()
      } else {
        toast.error(res.message || t('Operation failed'))
      }
    } finally {
      setRejectingId(null)
    }
  }

  async function handleCopy(value: string, label: string) {
    if (!value) return
    await navigator.clipboard.writeText(value)
    toast.success(t('{{label}} copied', { label }))
  }

  return (
    <>
      <SectionPageLayout>
        <SectionPageLayout.Title>
          {t('Official Portrait Assets')}
        </SectionPageLayout.Title>
        <SectionPageLayout.Description>
          {t(
            'Create VolcEngine portrait assets through the official H5 and Assets API.'
          )}
        </SectionPageLayout.Description>
        <SectionPageLayout.Actions>
          <Button variant='outline' onClick={fetchData} disabled={loading}>
            <RefreshCw className={loading ? 'animate-spin' : ''} />
            {t('Refresh')}
          </Button>
        </SectionPageLayout.Actions>
        <SectionPageLayout.Content>
          <div className='grid gap-4 xl:grid-cols-[260px_minmax(0,1fr)_380px]'>
            <Card className='rounded-lg'>
              <CardHeader>
                <CardTitle className='flex items-center gap-2 text-base'>
                  <FolderOpen className='size-4' />
                  {t('Asset folders')}
                </CardTitle>
              </CardHeader>
              <CardContent className='space-y-3'>
                <div className='space-y-1'>
                  <Button
                    type='button'
                    variant={selectedFolderId === 'all' ? 'secondary' : 'ghost'}
                    className='w-full justify-start'
                    onClick={() => setSelectedFolderId('all')}
                  >
                    <FolderOpen />
                    {t('All assets')}
                  </Button>
                  <Button
                    type='button'
                    variant={selectedFolderId === 0 ? 'secondary' : 'ghost'}
                    className='w-full justify-start'
                    onClick={() => setSelectedFolderId(0)}
                  >
                    <Folder />
                    {t('Ungrouped')}
                  </Button>
                  {folders.map((folder) => (
                    <div key={folder.id} className='flex items-center gap-1'>
                      <Button
                        type='button'
                        variant={
                          selectedFolderId === folder.id ? 'secondary' : 'ghost'
                        }
                        className='min-w-0 flex-1 justify-start'
                        onClick={() => setSelectedFolderId(folder.id)}
                      >
                        <Folder />
                        <span className='truncate'>{folder.name}</span>
                      </Button>
                      <Button
                        type='button'
                        size='icon-sm'
                        variant='ghost'
                        onClick={() => setDeletingFolder(folder)}
                      >
                        <Trash2 />
                      </Button>
                    </div>
                  ))}
                </div>
                <div className='space-y-2 border-t pt-3'>
                  <Input
                    value={newFolderName}
                    onChange={(event) => setNewFolderName(event.target.value)}
                    placeholder={t('Folder name')}
                    maxLength={128}
                  />
                  <Button
                    type='button'
                    className='w-full'
                    variant='outline'
                    disabled={creatingFolder}
                    onClick={handleCreateFolder}
                  >
                    {creatingFolder ? (
                      <Loader2 className='animate-spin' />
                    ) : (
                      <FolderPlus />
                    )}
                    {t('New folder')}
                  </Button>
                </div>
              </CardContent>
            </Card>

            <div className='space-y-4'>
              {!config?.configured && (
                <Alert variant='destructive'>
                  <ShieldAlert />
                  <AlertTitle>
                    {t('VolcEngine portrait API is not configured')}
                  </AlertTitle>
                  <AlertDescription>
                    {t(
                      'Ask an administrator to configure VolcEngine portrait credentials.'
                    )}
                  </AlertDescription>
                </Alert>
              )}

              <Card className='rounded-lg'>
                <CardHeader>
                  <CardTitle className='flex items-center gap-2 text-base'>
                    <Plus className='size-4' />
                    {t('Create official asset task')}
                  </CardTitle>
                </CardHeader>
                <CardContent className='space-y-3'>
                  <div className='flex flex-col gap-3 sm:flex-row'>
                    <Input
                      value={name}
                      onChange={(event) => setName(event.target.value)}
                      placeholder={t('Asset name')}
                      maxLength={50}
                    />
                    <Button
                      onClick={handleCreate}
                      disabled={creating || !!activeJob || !config?.configured}
                    >
                      {creating ? (
                        <Loader2 className='animate-spin' />
                      ) : (
                        <Plus />
                      )}
                      {t('Create')}
                    </Button>
                  </div>
                  <Input
                    value={callbackUrl}
                    onChange={(event) => setCallbackUrl(event.target.value)}
                    placeholder={t('Validation callback URL (optional)')}
                  />
                  <Select
                    value={String(createFolderId)}
                    onValueChange={(value) => setCreateFolderId(Number(value))}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={t('Target folder')} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value='0'>{t('Ungrouped')}</SelectItem>
                      {folders.map((folder) => (
                        <SelectItem key={folder.id} value={String(folder.id)}>
                          {folder.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <div className='text-muted-foreground flex flex-wrap gap-x-4 gap-y-1 text-xs'>
                    <span>
                      {t('Project')}: {config?.project_name || '-'}
                    </span>
                    <span>
                      {t('Current folder')}: {currentFolderLabel}
                    </span>
                    {activeJob && (
                      <span>
                        {t('Current task')}: {statusLabels[activeJob.status]}
                      </span>
                    )}
                  </div>
                </CardContent>
              </Card>

              <Card className='rounded-lg'>
                <CardHeader>
                  <CardTitle className='flex items-center gap-2 text-base'>
                    <UserCheck className='size-4' />
                    {t('Official asset tasks')}
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
                      {t('No official portrait assets yet')}
                    </div>
                  ) : (
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>{t('Name')}</TableHead>
                          <TableHead>{t('Status')}</TableHead>
                          <TableHead>{t('Asset ID')}</TableHead>
                          <TableHead>{t('Updated')}</TableHead>
                          <TableHead className='text-right'>
                            {t('Actions')}
                          </TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {jobs.map((job) => (
                          <TableRow key={job.id}>
                            <TableCell>
                              <div className='flex items-center gap-2'>
                                {job.asset_preview && (
                                  <PreviewImage
                                    src={job.asset_preview}
                                    className='bg-muted size-10 shrink-0 rounded-md border object-cover'
                                  />
                                )}
                                <div className='min-w-0'>
                                  <div className='truncate'>{job.name}</div>
                                  {job.volc_group_id && (
                                    <div className='text-muted-foreground truncate text-xs'>
                                      {job.volc_group_id}
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
                                {job.asset_status && (
                                  <div className='text-muted-foreground text-xs'>
                                    {job.asset_status}
                                  </div>
                                )}
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
                                <Button
                                  size='sm'
                                  variant='outline'
                                  disabled={syncingId === job.id}
                                  onClick={() => handleSync(job)}
                                >
                                  {syncingId === job.id ? (
                                    <Loader2 className='animate-spin' />
                                  ) : (
                                    <RefreshCw />
                                  )}
                                  {t('Sync')}
                                </Button>
                                <Select
                                  value={String(job.folder_id || 0)}
                                  disabled={movingAssetId === job.id}
                                  onValueChange={(value) =>
                                    handleMoveToFolder(job, Number(value))
                                  }
                                >
                                  <SelectTrigger className='h-8 w-[128px]'>
                                    <SelectValue placeholder={t('Move to')} />
                                  </SelectTrigger>
                                  <SelectContent>
                                    <SelectItem value='0'>
                                      {t('Ungrouped')}
                                    </SelectItem>
                                    {compatibleFoldersForExternalUser(
                                      job.external_user_id
                                    ).map((folder) => (
                                      <SelectItem
                                        key={folder.id}
                                        value={String(folder.id)}
                                      >
                                        {folder.name}
                                      </SelectItem>
                                    ))}
                                  </SelectContent>
                                </Select>
                                <Button
                                  size='icon-sm'
                                  variant='ghost'
                                  disabled={
                                    !job.asset_id || job.status !== 'ready'
                                  }
                                  onClick={() =>
                                    handleCopy(
                                      assetUri(job.asset_id),
                                      t('Asset ID')
                                    )
                                  }
                                >
                                  <Clipboard />
                                </Button>
                                <Button
                                  size='icon-sm'
                                  variant='ghost'
                                  disabled={deletingAssetId === job.id}
                                  onClick={() => setDeletingAsset(job)}
                                >
                                  {deletingAssetId === job.id ? (
                                    <Loader2 className='animate-spin' />
                                  ) : (
                                    <Trash2 />
                                  )}
                                </Button>
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
                    <UserCheck className='size-4' />
                    {t('Current official flow')}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  {!activeJob ? (
                    <div className='text-muted-foreground flex aspect-square items-center justify-center rounded-md border text-sm'>
                      {t('No active official task')}
                    </div>
                  ) : activeJob.status === 'validate_ready' &&
                    activeJob.invite_url ? (
                    <div className='space-y-3'>
                      <div className='bg-background flex aspect-square items-center justify-center rounded-md border p-6'>
                        <QRCodeSVG value={activeJob.invite_url} size={220} />
                      </div>
                      <div className='flex items-center justify-between gap-2'>
                        <Badge variant={statusVariant(activeJob.status)}>
                          {statusLabels[activeJob.status]}
                        </Badge>
                        {validationExpiresAt && (
                          <span className='text-muted-foreground text-xs'>
                            {validationExpiresAt}
                          </span>
                        )}
                      </div>
                      <div className='grid grid-cols-2 gap-2'>
                        <Button asChild variant='outline'>
                          <a
                            href={activeJob.invite_url}
                            target='_blank'
                            rel='noreferrer'
                          >
                            <ExternalLink />
                            {t('Open')}
                          </a>
                        </Button>
                        <Button
                          variant='outline'
                          disabled={refreshingId === activeJob.id}
                          onClick={() => handleRefreshValidation(activeJob)}
                        >
                          {refreshingId === activeJob.id ? (
                            <Loader2 className='animate-spin' />
                          ) : (
                            <RefreshCw />
                          )}
                          {t('Regenerate')}
                        </Button>
                      </div>
                    </div>
                  ) : activeJob.status === 'validated' ? (
                    <div className='space-y-3'>
                      <Badge variant={statusVariant(activeJob.status)}>
                        {statusLabels[activeJob.status]}
                      </Badge>
                      <div className='flex flex-col gap-2 sm:flex-row'>
                        <Input
                          value={assetUrl}
                          onChange={(event) => setAssetUrl(event.target.value)}
                          placeholder={t('Public material URL')}
                        />
                        <input
                          ref={fileInputRef}
                          type='file'
                          accept='image/*,video/*,audio/*'
                          className='hidden'
                          onChange={(event) =>
                            void handleUploadFile(
                              event.target.files?.[0] || null
                            )
                          }
                        />
                        <Button
                          type='button'
                          variant='outline'
                          disabled={uploading}
                          onClick={() => fileInputRef.current?.click()}
                        >
                          {uploading ? (
                            <Loader2 className='animate-spin' />
                          ) : (
                            <Upload />
                          )}
                          {t('Upload')}
                        </Button>
                      </div>
                      <Input
                        value={assetName}
                        onChange={(event) => setAssetName(event.target.value)}
                        placeholder={t('Material name')}
                      />
                      <Select value={assetType} onValueChange={setAssetType}>
                        <SelectTrigger className='w-full'>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value='Image'>{t('Image')}</SelectItem>
                          <SelectItem value='Video'>{t('Video')}</SelectItem>
                          <SelectItem value='Audio'>{t('Audio')}</SelectItem>
                        </SelectContent>
                      </Select>
                      <Select
                        value={String(assetFolderId)}
                        onValueChange={(value) =>
                          setAssetFolderId(Number(value))
                        }
                      >
                        <SelectTrigger className='w-full'>
                          <SelectValue placeholder={t('Target folder')} />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value='0'>{t('Ungrouped')}</SelectItem>
                          {compatibleFoldersForExternalUser(
                            activeJob.external_user_id
                          ).map((folder) => (
                            <SelectItem
                              key={folder.id}
                              value={String(folder.id)}
                            >
                              {folder.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <Button
                        className='w-full'
                        disabled={submittingId === activeJob.id || !assetUrl}
                        onClick={() => handleSubmitAsset(activeJob)}
                      >
                        {submittingId === activeJob.id ? (
                          <Loader2 className='animate-spin' />
                        ) : (
                          <Send />
                        )}
                        {t('Submit material')}
                      </Button>
                    </div>
                  ) : activeJob.status === 'asset_processing' ? (
                    <div className='space-y-3'>
                      {activeJob.asset_preview && (
                        <PreviewImage
                          src={activeJob.asset_preview}
                          className='bg-background aspect-square w-full rounded-md border object-cover'
                        />
                      )}
                      <Badge variant={statusVariant(activeJob.status)}>
                        {statusLabels[activeJob.status]}
                      </Badge>
                      <Button
                        className='w-full'
                        variant='outline'
                        disabled={syncingId === activeJob.id}
                        onClick={() => handleSync(activeJob)}
                      >
                        {syncingId === activeJob.id ? (
                          <Loader2 className='animate-spin' />
                        ) : (
                          <RefreshCw />
                        )}
                        {t('Sync')}
                      </Button>
                    </div>
                  ) : activeJob.status === 'pending_confirm' ? (
                    <div className='space-y-3'>
                      {activeJob.asset_preview && (
                        <PreviewImage
                          src={activeJob.asset_preview}
                          className='bg-background aspect-square w-full rounded-md border object-cover'
                        />
                      )}
                      <div className='flex items-center justify-between gap-2'>
                        <Badge variant={statusVariant(activeJob.status)}>
                          {statusLabels[activeJob.status]}
                        </Badge>
                        {activeJob.asset_status && (
                          <span className='text-muted-foreground text-xs'>
                            {activeJob.asset_status}
                          </span>
                        )}
                      </div>
                      <div className='grid grid-cols-2 gap-2'>
                        <Button
                          disabled={confirmingId === activeJob.id}
                          onClick={() => handleConfirm(activeJob)}
                        >
                          {confirmingId === activeJob.id ? (
                            <Loader2 className='animate-spin' />
                          ) : (
                            <CheckCircle2 />
                          )}
                          {t('Confirm')}
                        </Button>
                        <Button
                          variant='outline'
                          disabled={rejectingId === activeJob.id}
                          onClick={() => handleReject(activeJob)}
                        >
                          {rejectingId === activeJob.id ? (
                            <Loader2 className='animate-spin' />
                          ) : (
                            <ShieldAlert />
                          )}
                          {t('Reject')}
                        </Button>
                      </div>
                    </div>
                  ) : (
                    <div className='space-y-3'>
                      <Badge variant={statusVariant(activeJob.status)}>
                        {statusLabels[activeJob.status] || activeJob.status}
                      </Badge>
                      {activeJob.error_message && (
                        <div className='text-destructive rounded-md border p-3 text-sm'>
                          {activeJob.error_message}
                        </div>
                      )}
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card className='rounded-lg'>
                <CardHeader>
                  <CardTitle className='flex items-center gap-2 text-base'>
                    <CheckCircle2 className='size-4' />
                    {t('Asset summary')}
                  </CardTitle>
                </CardHeader>
                <CardContent className='space-y-3 text-sm'>
                  <div className='flex items-center justify-between'>
                    <span className='text-muted-foreground'>
                      {t('Ready assets')}
                    </span>
                    <span className='font-medium'>{readyJobs.length}</span>
                  </div>
                  <div className='flex items-center justify-between'>
                    <span className='text-muted-foreground'>
                      {t('Latest task')}
                    </span>
                    <span className='font-medium'>
                      {jobs[0] ? statusLabels[jobs[0].status] : '-'}
                    </span>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>
      <ConfirmDialog
        open={!!deletingAsset}
        onOpenChange={(open) => {
          if (!open) setDeletingAsset(null)
        }}
        title={t('Delete asset')}
        desc={t(
          'This only removes the asset from the 8liang asset library. It does not delete the upstream VolcEngine asset.'
        )}
        confirmText={t('Delete')}
        destructive
        isLoading={!!deletingAsset && deletingAssetId === deletingAsset.id}
        handleConfirm={() => {
          if (deletingAsset) void handleDeleteAsset(deletingAsset)
        }}
      />
      <ConfirmDialog
        open={!!deletingFolder}
        onOpenChange={(open) => {
          if (!open) setDeletingFolder(null)
        }}
        title={t('Delete folder')}
        desc={t('Assets in this folder will be moved to Ungrouped.')}
        confirmText={t('Delete')}
        destructive
        handleConfirm={() => {
          if (deletingFolder) void handleDeleteFolder(deletingFolder)
        }}
      />
    </>
  )
}
