import * as z from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
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
import { SettingsSection } from '../components/settings-section'
import { useResetForm } from '../hooks/use-reset-form'
import { useUpdateOption } from '../hooks/use-update-option'

const createSchema = (t: (key: string) => string) =>
  z.object({
    accessKey: z.string(),
    secretKey: z.string(),
    projectName: z.string().min(1, t('Project name is required')),
    region: z.string().min(1, t('Region is required')),
    callbackBaseURL: z.string().refine((value) => {
      const trimmed = value.trim()
      if (!trimmed) return true
      return /^https?:\/\//.test(trimmed)
    }, t('Provide a valid URL starting with http:// or https://')),
  })

type Values = z.infer<ReturnType<typeof createSchema>>

type VolcPortraitAssetSettingsSectionProps = {
  defaultValues: Values
}

export function VolcPortraitAssetSettingsSection({
  defaultValues,
}: VolcPortraitAssetSettingsSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const schema = createSchema(t)

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues,
  })

  useResetForm(form, defaultValues)

  const onSubmit = async (values: Values) => {
    const sanitized = {
      accessKey: values.accessKey.trim(),
      secretKey: values.secretKey.trim(),
      projectName: values.projectName.trim() || 'default',
      region: values.region.trim() || 'cn-beijing',
      callbackBaseURL: values.callbackBaseURL.trim().replace(/\/+$/, ''),
    }
    const initial = {
      projectName: defaultValues.projectName.trim() || 'default',
      region: defaultValues.region.trim() || 'cn-beijing',
      callbackBaseURL: defaultValues.callbackBaseURL.trim().replace(/\/+$/, ''),
    }
    const updates: Array<{ key: string; value: string }> = []

    if (sanitized.accessKey) {
      updates.push({
        key: 'portrait_asset.access_key',
        value: sanitized.accessKey,
      })
    }

    if (sanitized.secretKey) {
      updates.push({
        key: 'portrait_asset.secret_key',
        value: sanitized.secretKey,
      })
    }

    if (sanitized.projectName !== initial.projectName) {
      updates.push({
        key: 'portrait_asset.project_name',
        value: sanitized.projectName,
      })
    }

    if (sanitized.region !== initial.region) {
      updates.push({
        key: 'portrait_asset.region',
        value: sanitized.region,
      })
    }

    if (sanitized.callbackBaseURL !== initial.callbackBaseURL) {
      updates.push({
        key: 'portrait_asset.callback_base_url',
        value: sanitized.callbackBaseURL,
      })
    }

    if (updates.length === 0) {
      toast.info(t('No changes to save'))
      return
    }

    for (const update of updates) {
      await updateOption.mutateAsync(update)
    }

    form.reset({
      accessKey: '',
      secretKey: '',
      projectName: sanitized.projectName,
      region: sanitized.region,
      callbackBaseURL: sanitized.callbackBaseURL,
    })
  }

  return (
    <SettingsSection
      title={t('VolcEngine Portrait Assets')}
      description={t(
        'Configure VolcEngine credentials for official and virtual portrait asset creation.'
      )}
    >
      <Form {...form}>
        <form
          onSubmit={form.handleSubmit(onSubmit)}
          autoComplete='off'
          className='space-y-6'
        >
          <Alert>
            <AlertTitle>{t('Used by ordinary users')}</AlertTitle>
            <AlertDescription>
              {t(
                'After these server-side credentials are configured, ordinary users can create official and virtual portrait assets without seeing the AK/SK.'
              )}
            </AlertDescription>
          </Alert>

          <div className='grid gap-4 md:grid-cols-2'>
            <FormField
              control={form.control}
              name='accessKey'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('VolcEngine Access Key')}</FormLabel>
                  <FormControl>
                    <Input
                      type='password'
                      placeholder={t('Enter new access key to update')}
                      autoComplete='new-password'
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t('Leave blank to keep the existing access key.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='secretKey'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('VolcEngine Secret Key')}</FormLabel>
                  <FormControl>
                    <Input
                      type='password'
                      placeholder={t('Enter new secret key to update')}
                      autoComplete='new-password'
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t('Leave blank to keep the existing secret key.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <div className='grid gap-4 md:grid-cols-2'>
            <FormField
              control={form.control}
              name='projectName'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('VolcEngine Project Name')}</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='default'
                      autoComplete='off'
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Assets and inference endpoints must be in the same project.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='region'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('VolcEngine Region')}</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='cn-beijing'
                      autoComplete='off'
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t('Region used when signing VolcEngine OpenAPI requests.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <FormField
            control={form.control}
            name='callbackBaseURL'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('Portrait callback base URL')}</FormLabel>
                <FormControl>
                  <Input
                    type='url'
                    inputMode='url'
                    placeholder='https://your-domain.example'
                    autoComplete='off'
                    {...field}
                  />
                </FormControl>
                <FormDescription>
                  {t(
                    'Public base URL for the H5 validation callback, without a trailing slash.'
                  )}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <Button
            type='submit'
            disabled={!form.formState.isDirty || updateOption.isPending}
          >
            {updateOption.isPending
              ? t('Saving...')
              : t('Save portrait asset settings')}
          </Button>
        </form>
      </Form>
    </SettingsSection>
  )
}
