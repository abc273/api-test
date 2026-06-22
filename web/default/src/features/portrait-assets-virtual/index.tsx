import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  Clipboard,
  ExternalLink,
  Folder,
  FolderOpen,
  FolderPlus,
  Loader2,
  RefreshCw,
  ShieldAlert,
  Trash2,
  Upload,
} from 'lucide-react'
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
  createUserVirtualPortraitAsset,
  deleteUserVirtualPortraitAsset,
  getUserVirtualPortraitAssetGroup,
  getUserVirtualPortraitAssets,
  getVirtualPortraitAssetConfig,
  syncUserVirtualPortraitAsset,
  uploadVirtualPortraitAssetMaterial,
} from './api'
import type {
  VirtualPortraitAsset,
  VirtualPortraitAssetConfig,
  VirtualPortraitAssetGroup,
  VirtualPortraitAssetGroupStatus,
  VirtualPortraitAssetStatus,
} from './types'

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

function assetStatusVariant(
  status: VirtualPortraitAssetStatus
): 'default' | 'secondary' | 'destructive' {
  if (status === 'active') return 'default'
  if (status === 'failed') return 'destructive'
  return 'secondary'
}

function groupStatusVariant(
  status: VirtualPortraitAssetGroupStatus
): 'default' | 'secondary' | 'destructive' {
  if (status === 'active') return 'default'
  if (status === 'failed') return 'destructive'
  return 'secondary'
}

export function VirtualPortraitAssets() {
  const { t } = useTranslation()
  const [config, setConfig] = useState<VirtualPortraitAssetConfig | null>(null)
  const [group, setGroup] = useState<VirtualPortraitAssetGroup | null>(null)
  const [assets, setAssets] = useState<VirtualPortraitAsset[]>([])
  const [folders, setFolders] = useState<PortraitAssetFolder[]>([])
  const [selectedFolderId, setSelectedFolderId] = useState<number | 'all'>(
    'all'
  )
  const [loading, setLoading] = useState(true)
  const [uploading, setUploading] = useState(false)
  const [creating, setCreating] = useState(false)
  const [creatingFolder, setCreatingFolder] = useState(false)
  const [deletingFolder, setDeletingFolder] =
    useState<PortraitAssetFolder | null>(null)
  const [deletingAsset, setDeletingAsset] =
    useState<VirtualPortraitAsset | null>(null)
  const [deletingAssetId, setDeletingAssetId] = useState<number | null>(null)
  const [movingAssetId, setMovingAssetId] = useState<number | null>(null)
  const [syncingId, setSyncingId] = useState<number | null>(null)
  const [name, setName] = useState('')
  const [newFolderName, setNewFolderName] = useState('')
  const [assetFolderId, setAssetFolderId] = useState(0)
  const [assetUrl, setAssetUrl] = useState('')
  const [assetType, setAssetType] = useState<'Image' | 'Video' | 'Audio'>(
    'Image'
  )
  const fileInputRef = useRef<HTMLInputElement | null>(null)

  const assetStatusLabels = useMemo<Record<VirtualPortraitAssetStatus, string>>(
    () => ({
      processing: t('Processing'),
      active: t('Ready'),
      failed: t('Failed'),
    }),
    [t]
  )

  const groupStatusLabels = useMemo<
    Record<VirtualPortraitAssetGroupStatus, string>
  >(
    () => ({
      creating: t('Creating group'),
      active: t('Ready'),
      failed: t('Failed'),
    }),
    [t]
  )

  const processingAssets = useMemo(
    () => assets.filter((asset) => asset.status === 'processing'),
    [assets]
  )

  const readyAssets = useMemo(
    () =>
      assets.filter(
        (asset) => asset.status === 'active' && asset.volc_asset_id
      ),
    [assets]
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
      const [configRes, groupRes, assetsRes, foldersRes] = await Promise.all([
        getVirtualPortraitAssetConfig(),
        getUserVirtualPortraitAssetGroup(),
        getUserVirtualPortraitAssets({
          p: 1,
          page_size: 50,
          folder_id: folderId,
        }),
        getPortraitAssetFolders({ asset_kind: 'virtual' }),
      ])
      if (configRes.success && configRes.data) setConfig(configRes.data)
      if (groupRes.success) setGroup(groupRes.data ?? null)
      if (assetsRes.success && assetsRes.data) {
        setAssets(assetsRes.data.items || [])
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
      setAssetFolderId(selectedFolderId)
    }
  }, [selectedFolderId])

  useEffect(() => {
    if (processingAssets.length === 0) return

    let disposed = false
    let running = false

    const syncPendingAssets = async () => {
      if (disposed || running) return
      running = true
      try {
        await Promise.all(
          processingAssets.map((asset) =>
            syncUserVirtualPortraitAsset(asset.id).catch(() => null)
          )
        )
        if (!disposed) {
          await fetchData()
        }
      } finally {
        running = false
      }
    }

    const timer = window.setInterval(syncPendingAssets, 5000)
    return () => {
      disposed = true
      window.clearInterval(timer)
    }
  }, [processingAssets, fetchData])

  async function handleUploadFile(file?: File | null) {
    if (!file) return
    try {
      setUploading(true)
      const res = await uploadVirtualPortraitAssetMaterial(file)
      if (res.success && res.data) {
        setAssetUrl(res.data.url)
        setAssetType(res.data.asset_type)
        setName((current) => current || stripFileExtension(file.name))
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

  async function handleCreate() {
    if (!assetUrl.trim()) {
      toast.error(t('Provide a public asset URL first'))
      return
    }
    try {
      setCreating(true)
      const targetExternalUserID = getFolderExternalUserID(assetFolderId)
      const res = await createUserVirtualPortraitAsset({
        name: name.trim() || undefined,
        asset_url: assetUrl.trim(),
        asset_type: assetType,
        folder_id: assetFolderId,
        external_user_id: targetExternalUserID || undefined,
      })
      if (res.success) {
        toast.success(t('Asset created'))
        setName('')
        setAssetUrl('')
        setAssetType('Image')
        setAssetFolderId(selectedFolderId === 'all' ? 0 : selectedFolderId)
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
        asset_kind: 'virtual',
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

  async function handleDeleteAsset(asset: VirtualPortraitAsset) {
    try {
      setDeletingAssetId(asset.id)
      const res = await deleteUserVirtualPortraitAsset(asset.id, {
        external_user_id:
          normalizeFolderExternalUserID(asset.external_user_id) || undefined,
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
    asset: VirtualPortraitAsset,
    folderId: number
  ) {
    try {
      setMovingAssetId(asset.id)
      const res = await movePortraitAssetsToFolder({
        asset_kind: 'virtual',
        asset_ids: [asset.id],
        folder_id: folderId,
        external_user_id:
          normalizeFolderExternalUserID(asset.external_user_id) || undefined,
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

  async function handleSync(asset: VirtualPortraitAsset) {
    try {
      setSyncingId(asset.id)
      const res = await syncUserVirtualPortraitAsset(asset.id)
      if (res.success) {
        await fetchData()
      } else {
        toast.error(res.message || t('Sync failed'))
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : t('Sync failed'))
    } finally {
      setSyncingId(null)
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
          {t('Virtual Portrait Assets')}
        </SectionPageLayout.Title>
        <SectionPageLayout.Description>
          {t(
            'Assets created by the same user are automatically stored in one VolcEngine asset group.'
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
                    {t('VolcEngine virtual asset API is not configured')}
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
                    <FolderOpen className='size-4' />
                    {t('Current asset group')}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  {loading ? (
                    <div className='space-y-2'>
                      <Skeleton className='h-10 w-full' />
                      <Skeleton className='h-10 w-full' />
                    </div>
                  ) : !group ? (
                    <div className='text-muted-foreground text-sm'>
                      {t(
                        'The asset group is created automatically when you submit the first asset.'
                      )}
                    </div>
                  ) : (
                    <div className='space-y-3 text-sm'>
                      <div className='flex flex-wrap items-center gap-2'>
                        <Badge variant={groupStatusVariant(group.status)}>
                          {groupStatusLabels[group.status]}
                        </Badge>
                        <span className='text-muted-foreground'>
                          {t('Project')}: {group.project_name}
                        </span>
                      </div>
                      <div className='space-y-1'>
                        <div className='text-muted-foreground text-xs'>
                          {t('Group ID')}
                        </div>
                        <div className='font-mono text-xs break-all'>
                          {group.volc_group_id || '-'}
                        </div>
                      </div>
                      <div className='space-y-1'>
                        <div className='text-muted-foreground text-xs'>
                          {t('Group name')}
                        </div>
                        <div>{group.name}</div>
                      </div>
                      {group.error_message && (
                        <div className='text-destructive text-xs'>
                          {group.error_message}
                        </div>
                      )}
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card className='rounded-lg'>
                <CardHeader>
                  <CardTitle className='text-base'>
                    {t('Virtual portrait assets')}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  {loading ? (
                    <div className='space-y-2'>
                      <Skeleton className='h-10 w-full' />
                      <Skeleton className='h-10 w-full' />
                      <Skeleton className='h-10 w-full' />
                    </div>
                  ) : assets.length === 0 ? (
                    <div className='text-muted-foreground flex h-28 items-center justify-center text-sm'>
                      {t('No virtual portrait assets yet')}
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
                        {assets.map((asset) => (
                          <TableRow key={asset.id}>
                            <TableCell>
                              <div className='flex items-center gap-2'>
                                {asset.asset_type === 'Image' &&
                                  asset.preview_url && (
                                    <PreviewImage
                                      src={asset.preview_url}
                                      className='bg-muted size-10 shrink-0 rounded-md border object-cover'
                                    />
                                  )}
                                <div className='min-w-0'>
                                  <div className='truncate'>{asset.name}</div>
                                  <div className='text-muted-foreground truncate text-xs'>
                                    {asset.asset_type}
                                  </div>
                                </div>
                              </div>
                            </TableCell>
                            <TableCell>
                              <div className='space-y-1'>
                                <Badge
                                  variant={assetStatusVariant(asset.status)}
                                >
                                  {assetStatusLabels[asset.status]}
                                </Badge>
                                {asset.volc_status && (
                                  <div className='text-muted-foreground text-xs'>
                                    {asset.volc_status}
                                  </div>
                                )}
                                {asset.error_message && (
                                  <div className='text-destructive max-w-[220px] text-xs'>
                                    {asset.error_message}
                                  </div>
                                )}
                              </div>
                            </TableCell>
                            <TableCell className='max-w-[260px] truncate font-mono text-xs'>
                              {asset.status === 'active'
                                ? assetUri(asset.volc_asset_id) || '-'
                                : '-'}
                            </TableCell>
                            <TableCell>
                              {formatTimestampToDate(asset.updated_time)}
                            </TableCell>
                            <TableCell>
                              <div className='flex justify-end gap-1'>
                                <Button
                                  size='sm'
                                  variant='outline'
                                  disabled={syncingId === asset.id}
                                  onClick={() => handleSync(asset)}
                                >
                                  {syncingId === asset.id ? (
                                    <Loader2 className='animate-spin' />
                                  ) : (
                                    <RefreshCw />
                                  )}
                                  {t('Sync')}
                                </Button>
                                <Select
                                  value={String(asset.folder_id || 0)}
                                  disabled={movingAssetId === asset.id}
                                  onValueChange={(value) =>
                                    handleMoveToFolder(asset, Number(value))
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
                                      asset.external_user_id
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
                                    !asset.volc_asset_id ||
                                    asset.status !== 'active'
                                  }
                                  onClick={() =>
                                    handleCopy(
                                      assetUri(asset.volc_asset_id),
                                      t('Asset ID')
                                    )
                                  }
                                >
                                  <Clipboard />
                                </Button>
                                <Button
                                  size='icon-sm'
                                  variant='ghost'
                                  asChild
                                  disabled={!asset.source_url}
                                >
                                  <a
                                    href={asset.source_url || '#'}
                                    target='_blank'
                                    rel='noreferrer'
                                  >
                                    <ExternalLink />
                                  </a>
                                </Button>
                                <Button
                                  size='icon-sm'
                                  variant='ghost'
                                  disabled={deletingAssetId === asset.id}
                                  onClick={() => setDeletingAsset(asset)}
                                >
                                  {deletingAssetId === asset.id ? (
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
                  <CardTitle className='text-base'>
                    {t('Create virtual portrait asset')}
                  </CardTitle>
                </CardHeader>
                <CardContent className='space-y-3'>
                  <div className='text-muted-foreground flex flex-wrap gap-x-4 gap-y-1 text-xs'>
                    <span>
                      {t('Project')}: {config?.project_name || '-'}
                    </span>
                    <span>
                      {t('Ready assets')}: {readyAssets.length}
                    </span>
                    <span>
                      {t('Current folder')}: {currentFolderLabel}
                    </span>
                  </div>

                  <Input
                    value={name}
                    onChange={(event) => setName(event.target.value)}
                    placeholder={t('Asset name')}
                    maxLength={50}
                  />

                  <Select
                    value={assetType}
                    onValueChange={(value) =>
                      setAssetType(value as 'Image' | 'Video' | 'Audio')
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder={t('Asset type')} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value='Image'>{t('Image')}</SelectItem>
                      <SelectItem value='Video'>{t('Video')}</SelectItem>
                      <SelectItem value='Audio'>{t('Audio')}</SelectItem>
                    </SelectContent>
                  </Select>

                  <Select
                    value={String(assetFolderId)}
                    onValueChange={(value) => setAssetFolderId(Number(value))}
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

                  <Input
                    value={assetUrl}
                    onChange={(event) => setAssetUrl(event.target.value)}
                    placeholder={t('Public asset URL')}
                  />

                  <div className='grid grid-cols-2 gap-2'>
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
                    <Button
                      type='button'
                      onClick={handleCreate}
                      disabled={creating || !config?.configured}
                    >
                      {creating ? (
                        <Loader2 className='animate-spin' />
                      ) : (
                        <Upload />
                      )}
                      {t('Create')}
                    </Button>
                  </div>

                  <input
                    ref={fileInputRef}
                    type='file'
                    className='hidden'
                    accept='image/*,video/*,audio/*'
                    onChange={(event) =>
                      handleUploadFile(event.target.files?.[0])
                    }
                  />
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
