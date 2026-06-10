import type { IntegrationSettings } from '../types'
import { createSectionRegistry } from '../utils/section-registry'
import { EmailSettingsSection } from './email-settings-section'
import { IoNetDeploymentSettingsSection } from './ionet-deployment-settings-section'
import { MonitoringSettingsSection } from './monitoring-settings-section'
import { PaymentSettingsSection } from './payment-settings-section'
import { VideoSuperResolutionSettingsSection } from './video-super-resolution-settings-section'
import { VolcPortraitAssetSettingsSection } from './volc-portrait-asset-settings-section'
import { WorkerSettingsSection } from './worker-settings-section'

const INTEGRATIONS_SECTIONS = [
  {
    id: 'payment',
    titleKey: 'Payment Gateway',
    descriptionKey: 'Configure payment gateway integrations',
    build: (settings: IntegrationSettings) => (
      <PaymentSettingsSection
        defaultValues={{
          PayAddress: settings.PayAddress,
          EpayId: settings.EpayId,
          EpayKey: settings.EpayKey,
          Price: settings.Price,
          MinTopUp: settings.MinTopUp,
          CustomCallbackAddress: settings.CustomCallbackAddress,
          PayMethods: settings.PayMethods,
          AmountOptions: settings['payment_setting.amount_options'],
          AmountDiscount: settings['payment_setting.amount_discount'],
          StripeApiSecret: settings.StripeApiSecret,
          StripeWebhookSecret: settings.StripeWebhookSecret,
          StripePriceId: settings.StripePriceId,
          StripeUnitPrice: settings.StripeUnitPrice,
          StripeMinTopUp: settings.StripeMinTopUp,
          StripePromotionCodesEnabled: settings.StripePromotionCodesEnabled,
          AlipayEnabled: settings.AlipayEnabled,
          AlipayAppId: settings.AlipayAppId,
          AlipayPrivateKey: settings.AlipayPrivateKey,
          AlipayPublicKey: settings.AlipayPublicKey,
          AlipaySandbox: settings.AlipaySandbox,
          AlipayReturnUrl: settings.AlipayReturnUrl,
          AlipaySubscriptionReturnUrl:
            settings.AlipaySubscriptionReturnUrl,
          CreemApiKey: settings.CreemApiKey,
          CreemWebhookSecret: settings.CreemWebhookSecret,
          CreemTestMode: settings.CreemTestMode,
          CreemProducts: settings.CreemProducts,
        }}
        waffoDefaultValues={{
          WaffoEnabled: settings.WaffoEnabled ?? false,
          WaffoApiKey: settings.WaffoApiKey ?? '',
          WaffoPrivateKey: settings.WaffoPrivateKey ?? '',
          WaffoPublicCert: settings.WaffoPublicCert ?? '',
          WaffoSandboxPublicCert: settings.WaffoSandboxPublicCert ?? '',
          WaffoSandboxApiKey: settings.WaffoSandboxApiKey ?? '',
          WaffoSandboxPrivateKey: settings.WaffoSandboxPrivateKey ?? '',
          WaffoSandbox: settings.WaffoSandbox ?? false,
          WaffoMerchantId: settings.WaffoMerchantId ?? '',
          WaffoCurrency: settings.WaffoCurrency ?? 'USD',
          WaffoUnitPrice: settings.WaffoUnitPrice ?? 1,
          WaffoMinTopUp: settings.WaffoMinTopUp ?? 1,
          WaffoNotifyUrl: settings.WaffoNotifyUrl ?? '',
          WaffoReturnUrl: settings.WaffoReturnUrl ?? '',
          WaffoPayMethods: settings.WaffoPayMethods ?? '[]',
        }}
        waffoPancakeDefaultValues={{
          WaffoPancakeEnabled: settings.WaffoPancakeEnabled ?? false,
          WaffoPancakeSandbox: settings.WaffoPancakeSandbox ?? false,
          WaffoPancakeMerchantID: settings.WaffoPancakeMerchantID ?? '',
          WaffoPancakePrivateKey: settings.WaffoPancakePrivateKey ?? '',
          WaffoPancakeWebhookPublicKey:
            settings.WaffoPancakeWebhookPublicKey ?? '',
          WaffoPancakeWebhookTestKey: settings.WaffoPancakeWebhookTestKey ?? '',
          WaffoPancakeStoreID: settings.WaffoPancakeStoreID ?? '',
          WaffoPancakeProductID: settings.WaffoPancakeProductID ?? '',
          WaffoPancakeReturnURL: settings.WaffoPancakeReturnURL ?? '',
          WaffoPancakeCurrency: settings.WaffoPancakeCurrency ?? 'USD',
          WaffoPancakeUnitPrice: settings.WaffoPancakeUnitPrice ?? 1,
          WaffoPancakeMinTopUp: settings.WaffoPancakeMinTopUp ?? 1,
        }}
      />
    ),
  },
  {
    id: 'email',
    titleKey: 'SMTP Email',
    descriptionKey: 'Configure SMTP email settings',
    build: (settings: IntegrationSettings) => (
      <EmailSettingsSection
        defaultValues={{
          SMTPServer: settings.SMTPServer,
          SMTPPort: settings.SMTPPort,
          SMTPAccount: settings.SMTPAccount,
          SMTPFrom: settings.SMTPFrom,
          SMTPToken: settings.SMTPToken,
          SMTPSSLEnabled: settings.SMTPSSLEnabled,
          SMTPForceAuthLogin: settings.SMTPForceAuthLogin,
        }}
      />
    ),
  },
  {
    id: 'worker',
    titleKey: 'Worker Proxy',
    descriptionKey: 'Configure worker service settings',
    build: (settings: IntegrationSettings) => (
      <WorkerSettingsSection
        defaultValues={{
          WorkerUrl: settings.WorkerUrl,
          WorkerValidKey: settings.WorkerValidKey,
          WorkerAllowHttpImageRequestEnabled:
            settings.WorkerAllowHttpImageRequestEnabled,
        }}
      />
    ),
  },
  {
    id: 'ionet',
    titleKey: 'io.net Deployments',
    descriptionKey: 'Configure IoNet model deployment settings',
    build: (settings: IntegrationSettings) => (
      <IoNetDeploymentSettingsSection
        defaultValues={{
          enabled: settings['model_deployment.ionet.enabled'],
          apiKey: settings['model_deployment.ionet.api_key'],
        }}
      />
    ),
  },
  {
    id: 'volc-portrait-assets',
    titleKey: 'VolcEngine Portrait Assets',
    descriptionKey: 'Configure official portrait asset API credentials',
    build: (settings: IntegrationSettings) => (
      <VolcPortraitAssetSettingsSection
        defaultValues={{
          accessKey: settings['portrait_asset.access_key'],
          secretKey: settings['portrait_asset.secret_key'],
          projectName: settings['portrait_asset.project_name'],
          region: settings['portrait_asset.region'],
          callbackBaseURL: settings['portrait_asset.callback_base_url'],
        }}
      />
    ),
  },
  {
    id: 'video-super-resolution',
    titleKey: 'Video Super Resolution',
    descriptionKey:
      'Configure automatic post-processing for Seedance video outputs.',
    build: (settings: IntegrationSettings) => (
      <VideoSuperResolutionSettingsSection
        defaultValues={{
          enabled: settings['video_super_resolution.enabled'],
          baseURL: settings['video_super_resolution.base_url'],
          apiKey: settings['video_super_resolution.api_key'],
          outputTOSPath: settings['video_super_resolution.output_tos_path'],
          operatorID: settings['video_super_resolution.operator_id'],
          operatorVersion: settings['video_super_resolution.operator_version'],
          preserveAudio: settings['video_super_resolution.preserve_audio'],
          outputQualityMode:
            settings['video_super_resolution.output_quality_mode'] === 'master'
              ? 'master'
              : settings['video_super_resolution.output_quality_mode'] ===
                    'compatible'
                ? 'compatible'
                : 'balanced',
          tosPublicBaseURL:
            settings['video_super_resolution.tos_public_base_url'],
          tosEndpoint: settings['video_super_resolution.tos_endpoint'],
          tosRegion: settings['video_super_resolution.tos_region'],
          tosAccessKey: settings['video_super_resolution.tos_access_key'],
          tosSecretKey: settings['video_super_resolution.tos_secret_key'],
          tosSessionToken:
            settings['video_super_resolution.tos_session_token'],
          tosPresignExpires:
            settings['video_super_resolution.tos_presign_expires'],
        }}
      />
    ),
  },
  {
    id: 'monitoring',
    titleKey: 'Monitoring & Alerts',
    descriptionKey: 'Configure channel monitoring and automation',
    build: (settings: IntegrationSettings) => (
      <MonitoringSettingsSection
        defaultValues={{
          ChannelDisableThreshold: settings.ChannelDisableThreshold,
          QuotaRemindThreshold: settings.QuotaRemindThreshold,
          AutomaticDisableChannelEnabled:
            settings.AutomaticDisableChannelEnabled,
          AutomaticEnableChannelEnabled: settings.AutomaticEnableChannelEnabled,
          AutomaticDisableKeywords: settings.AutomaticDisableKeywords,
          AutomaticDisableStatusCodes: settings.AutomaticDisableStatusCodes,
          AutomaticRetryStatusCodes: settings.AutomaticRetryStatusCodes,
          'monitor_setting.auto_test_channel_enabled':
            settings['monitor_setting.auto_test_channel_enabled'],
          'monitor_setting.auto_test_channel_minutes':
            settings['monitor_setting.auto_test_channel_minutes'],
        }}
      />
    ),
  },
] as const

export type IntegrationSectionId = (typeof INTEGRATIONS_SECTIONS)[number]['id']

const integrationsRegistry = createSectionRegistry<
  IntegrationSectionId,
  IntegrationSettings
>({
  sections: INTEGRATIONS_SECTIONS,
  defaultSection: 'payment',
  basePath: '/system-settings/integrations',
  urlStyle: 'path',
})

export const INTEGRATIONS_SECTION_IDS = integrationsRegistry.sectionIds
export const INTEGRATIONS_DEFAULT_SECTION = integrationsRegistry.defaultSection
export const getIntegrationsSectionNavItems =
  integrationsRegistry.getSectionNavItems
export const getIntegrationsSectionContent =
  integrationsRegistry.getSectionContent
