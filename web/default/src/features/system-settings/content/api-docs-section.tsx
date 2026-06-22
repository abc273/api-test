import { useMemo, useState } from 'react'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { RotateCcw } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import * as z from 'zod'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { diffApiDocs, getApiDocsMeta, publishApiDocs } from '@/features/docs/api'
import type { ApiDocDiffResult, PublishApiDocsRequest } from '@/features/docs/types'
import { FormDirtyIndicator } from '../components/form-dirty-indicator'
import { FormNavigationGuard } from '../components/form-navigation-guard'
import { SettingsSection } from '../components/settings-section'
import { useSettingsForm } from '../hooks/use-settings-form'

const apiDocsSchema = z.object({
  version: z.string().min(1),
  title: z.string().min(1),
  summary: z.string().min(1),
  changedSections: z.string().min(1),
  changeItems: z.string().min(1),
  sourceCommit: z.string().optional(),
  ApiDocs: z.string().min(1),
})

type ApiDocsFormValues = z.infer<typeof apiDocsSchema>

type ApiDocsSectionProps = {
  defaultValue: string
}

function buildDefaultVersion() {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${year}.${month}.${day}.1`
}

function formatDocTime(timestamp?: number) {
  if (!timestamp) return ''
  return new Date(timestamp * 1000).toLocaleString()
}

function defaultChangeItemsJson() {
  return JSON.stringify(
    [
      {
        change_type: 'changed',
        endpoint: '',
        method: '',
        section: '接口文档',
        description: '补充本次文档变更说明',
        impact: '兼容旧调用',
      },
    ],
    null,
    2
  )
}

function parsePublishPayload(values: ApiDocsFormValues): PublishApiDocsRequest {
  const forbiddenTerms = ['\u5ba2\u6237', '\u7528\u6237']
  if (forbiddenTerms.some((term) => values.ApiDocs.includes(term))) {
    throw new Error('文档正文包含禁用称呼')
  }

  const changedSections = values.changedSections
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter(Boolean)

  const changeItems = JSON.parse(values.changeItems)
  if (!Array.isArray(changeItems) || changeItems.length === 0) {
    throw new Error('change_items 必须是非空数组')
  }

  return {
    version: values.version.trim(),
    title: values.title.trim(),
    summary: values.summary.trim(),
    changed_sections: changedSections,
    change_items: changeItems,
    content: values.ApiDocs,
    source_commit: values.sourceCommit?.trim() || undefined,
  }
}

function DiffPreview({ diff }: { diff?: ApiDocDiffResult }) {
  const { t } = useTranslation()

  if (!diff) return null

  const visibleLines = diff.lines.slice(0, 160)

  return (
    <Card>
      <CardHeader>
        <CardTitle className='text-base'>{t('API docs diff')}</CardTitle>
      </CardHeader>
      <CardContent className='space-y-3'>
        <div className='flex flex-wrap gap-2'>
          <Badge variant='secondary'>
            {diff.from_version} → {diff.to_version}
          </Badge>
          <Badge variant='outline'>
            +{diff.added_lines} / -{diff.removed_lines}
          </Badge>
        </div>
        <div className='bg-muted max-h-96 overflow-auto rounded-md p-3 font-mono text-xs'>
          {visibleLines.length === 0 ? (
            <div className='text-muted-foreground'>{t('No diff found')}</div>
          ) : (
            visibleLines.map((line, index) => (
              <div
                key={`${line.type}-${index}`}
                className={
                  line.type === 'added'
                    ? 'text-emerald-600'
                    : line.type === 'removed'
                      ? 'text-destructive'
                      : 'text-muted-foreground'
                }
              >
                {line.type === 'added' ? '+ ' : line.type === 'removed' ? '- ' : '  '}
                {line.text}
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  )
}

export function ApiDocsSection({ defaultValue }: ApiDocsSectionProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [diff, setDiff] = useState<ApiDocDiffResult | undefined>()

  const { data: metaData } = useQuery({
    queryKey: ['api-docs-meta'],
    queryFn: getApiDocsMeta,
  })

  const normalizedDefaults: ApiDocsFormValues = useMemo(
    () => ({
      version: buildDefaultVersion(),
      title: '8liangai.com API Docs',
      summary: '',
      changedSections: '',
      changeItems: defaultChangeItemsJson(),
      sourceCommit: '',
      ApiDocs: defaultValue ?? '',
    }),
    [defaultValue]
  )

  const publishMutation = useMutation({
    mutationFn: publishApiDocs,
    onSuccess: (data) => {
      if (!data.success) {
        toast.error(data.message || t('Failed to publish API docs'))
        return
      }
      queryClient.invalidateQueries({ queryKey: ['system-options'] })
      queryClient.invalidateQueries({ queryKey: ['api-docs'] })
      queryClient.invalidateQueries({ queryKey: ['api-docs-meta'] })
      queryClient.invalidateQueries({ queryKey: ['api-docs-changelog'] })
      toast.success(t('API docs published'))
    },
    onError: (error: Error) => {
      toast.error(error.message || t('Failed to publish API docs'))
    },
  })

  const diffMutation = useMutation({
    mutationFn: diffApiDocs,
    onSuccess: (data) => {
      if (!data.success || !data.data) {
        toast.error(data.message || t('Failed to generate diff'))
        return
      }
      setDiff(data.data)
    },
    onError: (error: Error) => {
      toast.error(error.message || t('Failed to generate diff'))
    },
  })

  const { form, handleSubmit, handleReset, isDirty, isSubmitting } =
    useSettingsForm<ApiDocsFormValues>({
      resolver: zodResolver(apiDocsSchema),
      defaultValues: normalizedDefaults,
      onSubmit: async (values) => {
        let payload: PublishApiDocsRequest
        try {
          payload = parsePublishPayload(values)
        } catch (error) {
          toast.error(error instanceof Error ? error.message : t('Invalid JSON'))
          return
        }
        await publishMutation.mutateAsync(payload)
      },
    })

  const buildDiff = () => {
    const values = form.getValues()
    diffMutation.mutate({
      to_version: values.version,
      to_content: values.ApiDocs,
    })
  }

  const current = metaData?.data

  return (
    <>
      <FormNavigationGuard when={isDirty} />

      <SettingsSection
        title={t('API Docs')}
        description={t('Configure the public API documentation page')}
      >
        <div className='space-y-6'>
          {current && (
            <Card>
              <CardContent className='space-y-3 p-4'>
                <div className='flex flex-wrap items-center gap-2'>
                  <Badge variant='secondary'>
                    {current.version || t('No version')}
                  </Badge>
                  {current.published_at > 0 && (
                    <span className='text-muted-foreground text-sm'>
                      {formatDocTime(current.published_at)}
                    </span>
                  )}
                </div>
                {current.summary && (
                  <p className='text-muted-foreground text-sm'>
                    {current.summary}
                  </p>
                )}
              </CardContent>
            </Card>
          )}

          <Form {...form}>
            <form onSubmit={handleSubmit} className='space-y-6'>
              <FormDirtyIndicator isDirty={isDirty} />

              <div className='grid gap-4 md:grid-cols-2'>
                <FormField
                  control={form.control}
                  name='version'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('Version')}</FormLabel>
                      <FormControl>
                        <Input placeholder='2026.06.16.1' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name='title'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('Title')}</FormLabel>
                      <FormControl>
                        <Input placeholder='8liangai.com API Docs' {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <FormField
                control={form.control}
                name='summary'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Update summary')}</FormLabel>
                    <FormControl>
                      <Textarea rows={3} {...field} />
                    </FormControl>
                    <FormDescription>
                      {t('Write what changed in this API docs release.')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='changedSections'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Changed sections')}</FormLabel>
                    <FormControl>
                      <Textarea rows={4} {...field} />
                    </FormControl>
                    <FormDescription>
                      {t('One section per line.')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='changeItems'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Structured changes')}</FormLabel>
                    <FormControl>
                      <Textarea rows={8} className='font-mono' {...field} />
                    </FormControl>
                    <FormDescription>
                      {t('Use JSON array with change_type, endpoint, method, section, description, and impact.')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='sourceCommit'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Source commit')}</FormLabel>
                    <FormControl>
                      <Input placeholder='optional' {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='ApiDocs'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('API documentation content')}</FormLabel>
                    <FormControl>
                      <Textarea
                        rows={22}
                        placeholder={t(
                          'Enter Markdown, HTML, or a full URL for your API docs'
                        )}
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('Forbidden wording is rejected during publish.')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className='flex flex-wrap gap-2'>
                <Button
                  type='submit'
                  disabled={isSubmitting || publishMutation.isPending}
                >
                  {publishMutation.isPending
                    ? t('Publishing...')
                    : t('Publish API docs')}
                </Button>
                <Button
                  type='button'
                  variant='outline'
                  onClick={buildDiff}
                  disabled={diffMutation.isPending}
                >
                  {diffMutation.isPending
                    ? t('Generating diff...')
                    : t('Generate diff')}
                </Button>
                <Button
                  type='button'
                  variant='outline'
                  onClick={handleReset}
                  disabled={!isDirty || isSubmitting || publishMutation.isPending}
                >
                  <RotateCcw className='mr-2 h-4 w-4' />
                  {t('Reset')}
                </Button>
              </div>
            </form>
          </Form>

          <DiffPreview diff={diff} />
        </div>
      </SettingsSection>
    </>
  )
}
