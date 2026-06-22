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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { SettingsSection } from '../components/settings-section'
import { useResetForm } from '../hooks/use-reset-form'
import { useUpdateOption } from '../hooks/use-update-option'

const createSchema = (t: (key: string) => string) =>
  z.object({
    enabled: z.boolean(),
    baseURL: z.string().refine((value) => {
      const trimmed = value.trim()
      if (!trimmed) return true
      return /^https?:\/\//.test(trimmed)
    }, t('Provide a valid URL starting with http:// or https://')),
    apiKey: z.string(),
    outputTOSPath: z.string().refine((value) => {
      const trimmed = value.trim()
      if (!trimmed) return true
      return trimmed.startsWith('tos://')
    }, t('Output TOS path must start with tos://')),
    operatorID: z.string().min(1, t('Operator ID is required')),
    operatorVersion: z.string().min(1, t('Operator version is required')),
    preserveAudio: z.boolean(),
    outputQualityMode: z.enum(['compatible', 'balanced', 'master']),
    tosPublicBaseURL: z.string().refine((value) => {
      const trimmed = value.trim()
      if (!trimmed) return true
      return /^https?:\/\//.test(trimmed)
    }, t('Provide a valid URL starting with http:// or https://')),
    tosEndpoint: z.string(),
    tosRegion: z.string(),
    tosAccessKey: z.string(),
    tosSecretKey: z.string(),
    tosSessionToken: z.string(),
    tosPresignExpires: z
      .number()
      .int()
      .min(1, t('Value must be at least 1'))
      .max(604800, t('Value must be at most 604800')),
  })

type Values = z.infer<ReturnType<typeof createSchema>>

type VideoSuperResolutionSettingsSectionProps = {
  defaultValues: Values
}

const OPTION_KEYS = {
  enabled: 'video_super_resolution.enabled',
  baseURL: 'video_super_resolution.base_url',
  apiKey: 'video_super_resolution.api_key',
  outputTOSPath: 'video_super_resolution.output_tos_path',
  operatorID: 'video_super_resolution.operator_id',
  operatorVersion: 'video_super_resolution.operator_version',
  preserveAudio: 'video_super_resolution.preserve_audio',
  outputQualityMode: 'video_super_resolution.output_quality_mode',
  tosPublicBaseURL: 'video_super_resolution.tos_public_base_url',
  tosEndpoint: 'video_super_resolution.tos_endpoint',
  tosRegion: 'video_super_resolution.tos_region',
  tosAccessKey: 'video_super_resolution.tos_access_key',
  tosSecretKey: 'video_super_resolution.tos_secret_key',
  tosSessionToken: 'video_super_resolution.tos_session_token',
  tosPresignExpires: 'video_super_resolution.tos_presign_expires',
} as const

export function VideoSuperResolutionSettingsSection({
  defaultValues,
}: VideoSuperResolutionSettingsSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const schema = createSchema(t)

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues,
  })

  useResetForm(form, defaultValues)

  const onSubmit = async (values: Values) => {
    const sanitized: Values = {
      enabled: values.enabled,
      baseURL: values.baseURL.trim().replace(/\/+$/, ''),
      apiKey: values.apiKey.trim(),
      outputTOSPath: values.outputTOSPath.trim().replace(/\/+$/, ''),
      operatorID: values.operatorID.trim(),
      operatorVersion: values.operatorVersion.trim(),
      preserveAudio: values.preserveAudio,
      outputQualityMode: values.outputQualityMode,
      tosPublicBaseURL: values.tosPublicBaseURL.trim().replace(/\/+$/, ''),
      tosEndpoint: values.tosEndpoint.trim(),
      tosRegion: values.tosRegion.trim(),
      tosAccessKey: values.tosAccessKey.trim(),
      tosSecretKey: values.tosSecretKey.trim(),
      tosSessionToken: values.tosSessionToken.trim(),
      tosPresignExpires: values.tosPresignExpires,
    }

    const initial: Values = {
      ...defaultValues,
      baseURL: defaultValues.baseURL.trim().replace(/\/+$/, ''),
      outputTOSPath: defaultValues.outputTOSPath.trim().replace(/\/+$/, ''),
      operatorID: defaultValues.operatorID.trim(),
      operatorVersion: defaultValues.operatorVersion.trim(),
      tosPublicBaseURL: defaultValues.tosPublicBaseURL
        .trim()
        .replace(/\/+$/, ''),
      tosEndpoint: defaultValues.tosEndpoint.trim(),
      tosRegion: defaultValues.tosRegion.trim(),
      tosPresignExpires: defaultValues.tosPresignExpires,
      apiKey: '',
      tosAccessKey: '',
      tosSecretKey: '',
      tosSessionToken: '',
    }

    const updates: Array<{ key: string; value: string | boolean | number }> = []

    if (sanitized.enabled !== defaultValues.enabled) {
      updates.push({ key: OPTION_KEYS.enabled, value: sanitized.enabled })
    }
    if (sanitized.baseURL !== initial.baseURL) {
      updates.push({ key: OPTION_KEYS.baseURL, value: sanitized.baseURL })
    }
    if (sanitized.apiKey) {
      updates.push({ key: OPTION_KEYS.apiKey, value: sanitized.apiKey })
    }
    if (sanitized.outputTOSPath !== initial.outputTOSPath) {
      updates.push({
        key: OPTION_KEYS.outputTOSPath,
        value: sanitized.outputTOSPath,
      })
    }
    if (sanitized.operatorID !== initial.operatorID) {
      updates.push({ key: OPTION_KEYS.operatorID, value: sanitized.operatorID })
    }
    if (sanitized.operatorVersion !== initial.operatorVersion) {
      updates.push({
        key: OPTION_KEYS.operatorVersion,
        value: sanitized.operatorVersion,
      })
    }
    if (sanitized.preserveAudio !== defaultValues.preserveAudio) {
      updates.push({
        key: OPTION_KEYS.preserveAudio,
        value: sanitized.preserveAudio,
      })
    }
    if (sanitized.outputQualityMode !== defaultValues.outputQualityMode) {
      updates.push({
        key: OPTION_KEYS.outputQualityMode,
        value: sanitized.outputQualityMode,
      })
    }
    if (sanitized.tosPublicBaseURL !== initial.tosPublicBaseURL) {
      updates.push({
        key: OPTION_KEYS.tosPublicBaseURL,
        value: sanitized.tosPublicBaseURL,
      })
    }
    if (sanitized.tosEndpoint !== initial.tosEndpoint) {
      updates.push({
        key: OPTION_KEYS.tosEndpoint,
        value: sanitized.tosEndpoint,
      })
    }
    if (sanitized.tosRegion !== initial.tosRegion) {
      updates.push({ key: OPTION_KEYS.tosRegion, value: sanitized.tosRegion })
    }
    if (sanitized.tosAccessKey) {
      updates.push({
        key: OPTION_KEYS.tosAccessKey,
        value: sanitized.tosAccessKey,
      })
    }
    if (sanitized.tosSecretKey) {
      updates.push({
        key: OPTION_KEYS.tosSecretKey,
        value: sanitized.tosSecretKey,
      })
    }
    if (sanitized.tosSessionToken) {
      updates.push({
        key: OPTION_KEYS.tosSessionToken,
        value: sanitized.tosSessionToken,
      })
    }
    if (sanitized.tosPresignExpires !== initial.tosPresignExpires) {
      updates.push({
        key: OPTION_KEYS.tosPresignExpires,
        value: sanitized.tosPresignExpires,
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
      ...sanitized,
      apiKey: '',
      tosAccessKey: '',
      tosSecretKey: '',
      tosSessionToken: '',
    })
  }

  return (
    <SettingsSection
      title={t('Video Super Resolution')}
      description={t(
        'Configure automatic post-processing for Seedance video outputs.'
      )}
    >
      <Form {...form}>
        <form
          onSubmit={form.handleSubmit(onSubmit)}
          autoComplete='off'
          className='space-y-6'
        >
          <Alert>
            <AlertTitle>{t('What this does')}</AlertTitle>
            <AlertDescription>
              {t(
                'When enabled, Seedance 480p videos are upscaled to 720p and 720p videos are upscaled to 1080p before the task is marked complete.'
              )}
            </AlertDescription>
          </Alert>

          <FormField
            control={form.control}
            name='enabled'
            render={({ field }) => (
              <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
                <div className='space-y-0.5'>
                  <FormLabel className='text-base'>
                    {t('Enable automatic super resolution')}
                  </FormLabel>
                  <FormDescription>
                    {t(
                      'Automatically upscale Seedance outputs after generation finishes.'
                    )}
                  </FormDescription>
                </div>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </FormItem>
            )}
          />

          <div className='grid gap-4 md:grid-cols-2'>
            <FormField
              control={form.control}
              name='baseURL'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('LAS API Base URL')}</FormLabel>
                  <FormControl>
                    <Input
                      type='url'
                      inputMode='url'
                      placeholder='https://operator.las.cn-beijing.volces.com'
                      autoComplete='off'
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t('Base URL used for LAS submit and poll requests.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='apiKey'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('LAS API Key')}</FormLabel>
                  <FormControl>
                    <Input
                      type='password'
                      placeholder={t('Enter new API key to update')}
                      autoComplete='new-password'
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t('Leave blank to keep the existing API key.')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <div className='grid gap-4 md:grid-cols-2'>
            <FormField
              control={form.control}
              name='outputTOSPath'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Output TOS Path')}</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='tos://bucket/path/to/results'
                      autoComplete='off'
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'TOS directory where super-resolved videos will be written.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='tosPresignExpires'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('TOS presign expiry (seconds)')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      min={1}
                      max={604800}
                      autoComplete='off'
                      {...field}
                      onChange={(event) =>
                        field.onChange(event.target.valueAsNumber)
                      }
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Used for private buckets when generating temporary download links.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <div className='grid gap-4 md:grid-cols-3'>
            <FormField
              control={form.control}
              name='operatorID'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Operator ID')}</FormLabel>
                  <FormControl>
                    <Input autoComplete='off' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='operatorVersion'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Operator Version')}</FormLabel>
                  <FormControl>
                    <Input autoComplete='off' {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='outputQualityMode'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Output quality mode')}</FormLabel>
                  <Select value={field.value} onValueChange={field.onChange}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value='compatible'>
                        {t('Compatible')}
                      </SelectItem>
                      <SelectItem value='balanced'>{t('Balanced')}</SelectItem>
                      <SelectItem value='master'>{t('Master')}</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormDescription>
                    {t(
                      'Choose how LAS balances compatibility, speed, and output quality.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <FormField
            control={form.control}
            name='preserveAudio'
            render={({ field }) => (
              <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
                <div className='space-y-0.5'>
                  <FormLabel className='text-base'>
                    {t('Preserve audio')}
                  </FormLabel>
                  <FormDescription>
                    {t(
                      'Keep the original audio track when creating the upscaled video.'
                    )}
                  </FormDescription>
                </div>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </FormItem>
            )}
          />

          <Alert>
            <AlertTitle>{t('Public bucket mode')}</AlertTitle>
            <AlertDescription>
              {t(
                'If you only configure a public base URL or endpoint, the system will assume the output objects are publicly readable.'
              )}
            </AlertDescription>
          </Alert>

          <FormField
            control={form.control}
            name='tosPublicBaseURL'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('TOS public base URL')}</FormLabel>
                <FormControl>
                  <Input
                    placeholder='https://{bucket}.example.com/{key}'
                    autoComplete='off'
                    {...field}
                  />
                </FormControl>
                <FormDescription>
                  {t(
                    'Optional public base URL template for direct downloads, for example https://{bucket}.example.com/{key}'
                  )}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <Alert>
            <AlertTitle>{t('Private bucket mode')}</AlertTitle>
            <AlertDescription>
              {t(
                'If you configure TOS endpoint, region, and AK/SK, the system will generate temporary signed download links for private buckets.'
              )}
            </AlertDescription>
          </Alert>

          <div className='grid gap-4 md:grid-cols-2'>
            <FormField
              control={form.control}
              name='tosEndpoint'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('TOS Endpoint')}</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='tos-cn-beijing.volces.com'
                      autoComplete='off'
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Used to build public object URLs and to sign private bucket downloads.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='tosRegion'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('TOS Region')}</FormLabel>
                  <FormControl>
                    <Input
                      placeholder='cn-beijing'
                      autoComplete='off'
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Required when generating signed URLs for private buckets.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <div className='grid gap-4 md:grid-cols-2'>
            <FormField
              control={form.control}
              name='tosAccessKey'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('TOS Access Key')}</FormLabel>
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
              name='tosSecretKey'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('TOS Secret Key')}</FormLabel>
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

          <FormField
            control={form.control}
            name='tosSessionToken'
            render={({ field }) => (
              <FormItem>
                <FormLabel>{t('TOS Session Token')}</FormLabel>
                <FormControl>
                  <Input
                    type='password'
                    placeholder={t('Enter new session token to update')}
                    autoComplete='new-password'
                    {...field}
                  />
                </FormControl>
                <FormDescription>
                  {t('Leave blank to keep the existing session token.')}
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />

          <Alert>
            <AlertDescription>
              {t(
                'Use the same AK/SK or a dedicated read-only AK/SK for the output bucket.'
              )}
            </AlertDescription>
          </Alert>

          <Button
            type='submit'
            disabled={!form.formState.isDirty || updateOption.isPending}
          >
            {updateOption.isPending
              ? t('Saving...')
              : t('Save video super resolution settings')}
          </Button>
        </form>
      </Form>
    </SettingsSection>
  )
}
