import { useMutation, useQueryClient } from '@tanstack/react-query'
import i18next from 'i18next'
import { toast } from 'sonner'
import { updateSystemOption } from '../api'
import type { UpdateOptionRequest } from '../types'

// Configuration keys that require status refresh
const STATUS_RELATED_KEYS = [
  'theme.frontend',
  'HeaderNavModules',
  'SidebarModulesAdmin',
  'Notice',
  'LogConsumeEnabled',
  'QuotaPerUnit',
  'USDExchangeRate',
  'DisplayInCurrencyEnabled',
  'DisplayTokenStatEnabled',
  'general_setting.quota_display_type',
  'general_setting.custom_currency_symbol',
  'general_setting.custom_currency_exchange_rate',
]

const PRICING_RELATED_KEYS = new Set([
  'ModelPrice',
  'ModelRatio',
  'CacheRatio',
  'CreateCacheRatio',
  'CompletionRatio',
  'ImageRatio',
  'AudioRatio',
  'AudioCompletionRatio',
  'GroupRatio',
  'GroupGroupRatio',
  'UserUsableGroups',
  'AutoGroups',
  'DefaultUseAutoGroup',
  'group_ratio_setting.group_special_usable_group',
  'billing_setting.billing_mode',
  'billing_setting.billing_expr',
  'billing_setting.output_tier_pricing',
])

function shouldRefreshPricing(key: string) {
  return (
    PRICING_RELATED_KEYS.has(key) ||
    key.startsWith('billing_setting.') ||
    key.startsWith('group_ratio_setting.')
  )
}

export function useUpdateOption() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: UpdateOptionRequest) => updateSystemOption(request),
    onSuccess: (data, variables) => {
      if (data.success) {
        // Always refresh system-options
        queryClient.invalidateQueries({ queryKey: ['system-options'] })
        if (variables.key === 'ApiDocs') {
          queryClient.invalidateQueries({ queryKey: ['api-docs'] })
        }

        if (shouldRefreshPricing(variables.key)) {
          queryClient.invalidateQueries({ queryKey: ['pricing'] })
        }

        // If updating frontend-display-related config, also refresh status
        if (STATUS_RELATED_KEYS.includes(variables.key)) {
          queryClient.invalidateQueries({ queryKey: ['status'] })
        }

        toast.success(i18next.t('Setting updated successfully'))
      } else {
        toast.error(data.message || i18next.t('Failed to update setting'))
      }
    },
    onError: (error: Error) => {
      toast.error(error.message || i18next.t('Failed to update setting'))
    },
  })
}
