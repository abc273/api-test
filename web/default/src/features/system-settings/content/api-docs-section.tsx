import * as z from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'
import { RotateCcw } from 'lucide-react'
import { useTranslation } from 'react-i18next'
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
import { Textarea } from '@/components/ui/textarea'
import { FormDirtyIndicator } from '../components/form-dirty-indicator'
import { FormNavigationGuard } from '../components/form-navigation-guard'
import { SettingsSection } from '../components/settings-section'
import { useSettingsForm } from '../hooks/use-settings-form'
import { useUpdateOption } from '../hooks/use-update-option'

const apiDocsSchema = z.object({
  ApiDocs: z.string().optional(),
})

type ApiDocsFormValues = z.infer<typeof apiDocsSchema>

type ApiDocsSectionProps = {
  defaultValue: string
}

export function ApiDocsSection({ defaultValue }: ApiDocsSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()

  const normalizedDefaults: ApiDocsFormValues = {
    ApiDocs: defaultValue ?? '',
  }

  const { form, handleSubmit, handleReset, isDirty, isSubmitting } =
    useSettingsForm<ApiDocsFormValues>({
    resolver: zodResolver(apiDocsSchema),
    defaultValues: normalizedDefaults,
    onSubmit: async (_values, changedFields) => {
      const nextValue = String(changedFields.ApiDocs ?? '')
      await updateOption.mutateAsync({
        key: 'ApiDocs',
        value: nextValue,
      })
    },
  })

  return (
    <>
      <FormNavigationGuard when={isDirty} />

      <SettingsSection
        title={t('API Docs')}
        description={t('Configure the public API documentation page')}
      >
        <Form {...form}>
          <form onSubmit={handleSubmit} className='space-y-6'>
            <FormDirtyIndicator isDirty={isDirty} />

            <FormField
              control={form.control}
              name='ApiDocs'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('API documentation content')}</FormLabel>
                  <FormControl>
                    <Textarea
                      rows={18}
                      placeholder={t(
                        'Enter Markdown, HTML, or a full URL for your API docs'
                      )}
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Publish the public API documentation shown at /docs. Markdown, HTML, or an external URL is supported.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className='flex gap-2'>
              <Button
                type='button'
                onClick={handleSubmit}
                disabled={
                  !isDirty || isSubmitting || updateOption.isPending
                }
              >
                {updateOption.isPending ? t('Saving...') : t('Save Changes')}
              </Button>
              <Button
                type='button'
                variant='outline'
                onClick={handleReset}
                disabled={
                  !isDirty || isSubmitting || updateOption.isPending
                }
              >
                <RotateCcw className='mr-2 h-4 w-4' />
                {t('Reset')}
              </Button>
            </div>
          </form>
        </Form>
      </SettingsSection>
    </>
  )
}
