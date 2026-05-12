import { useState } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import {
  type OutputTierPricing,
  sanitizeOutputTierPricing,
} from '@/lib/output-tier-pricing'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

type EditableOutputTierPricing = {
  label: string
  resolution: string
  hasVideoInput: 'any' | 'yes' | 'no'
  inputPrice: string
}

type OutputTierPricingEditorProps = {
  value: OutputTierPricing[]
  onChange: (next: OutputTierPricing[]) => void
}

function toEditableTier(tier?: OutputTierPricing): EditableOutputTierPricing {
  if (!tier) {
    return {
      label: '',
      resolution: '',
      hasVideoInput: 'any',
      inputPrice: '',
    }
  }

  let hasVideoInput: EditableOutputTierPricing['hasVideoInput'] = 'any'
  if (typeof tier.has_video_input === 'boolean') {
    hasVideoInput = tier.has_video_input ? 'yes' : 'no'
  }

  return {
    label: tier.label || '',
    resolution: tier.resolution || '',
    hasVideoInput,
    inputPrice: tier.input_price ? tier.input_price.toString() : '',
  }
}

function fromEditableTier(
  tier: EditableOutputTierPricing
): OutputTierPricing | null {
  const inputPrice = Number(tier.inputPrice)
  if (!Number.isFinite(inputPrice) || inputPrice <= 0) {
    return null
  }

  const result: OutputTierPricing = {
    input_price: inputPrice,
  }

  if (tier.label.trim()) {
    result.label = tier.label.trim()
  }
  if (tier.resolution.trim()) {
    result.resolution = tier.resolution.trim()
  }
  if (tier.hasVideoInput === 'yes') {
    result.has_video_input = true
  }
  if (tier.hasVideoInput === 'no') {
    result.has_video_input = false
  }

  return result
}

export function OutputTierPricingEditor(props: OutputTierPricingEditorProps) {
  const { t } = useTranslation()
  const [editableTiers, setEditableTiers] = useState<
    EditableOutputTierPricing[]
  >(
    props.value.length
      ? props.value.map((tier) => toEditableTier(tier))
      : [toEditableTier()]
  )

  const emitChange = (tiers: EditableOutputTierPricing[]) => {
    props.onChange(
      sanitizeOutputTierPricing(
        tiers
          .map((tier) => fromEditableTier(tier))
          .filter((tier): tier is OutputTierPricing => tier !== null)
      )
    )
  }

  const handleTierChange = (
    index: number,
    field: keyof EditableOutputTierPricing,
    value: string
  ) => {
    const nextTiers = editableTiers.map((tier, tierIndex) =>
      tierIndex === index ? { ...tier, [field]: value } : tier
    )
    setEditableTiers(nextTiers)
    emitChange(nextTiers)
  }

  const handleAddTier = () => {
    setEditableTiers((prev) => [...prev, toEditableTier()])
  }

  const handleRemoveTier = (index: number) => {
    const nextTiers = editableTiers.filter(
      (_, tierIndex) => tierIndex !== index
    )
    if (nextTiers.length === 0) {
      setEditableTiers([toEditableTier()])
      props.onChange([])
      return
    }
    setEditableTiers(nextTiers)
    emitChange(nextTiers)
  }

  return (
    <div className='space-y-4 rounded-lg border p-4'>
      <div className='space-y-1'>
        <Label>{t('Tiered input pricing')}</Label>
        <p className='text-muted-foreground text-sm'>
          {t(
            'Match different output types to different input prices. Prices use USD per 1M tokens in the current token pricing mode.'
          )}
        </p>
      </div>

      <div className='space-y-3'>
        {editableTiers.map((tier, index) => (
          <div
            key={index}
            className='grid gap-3 rounded-md border p-3 md:grid-cols-12'
          >
            <div className='space-y-2 md:col-span-3'>
              <Label>{t('Label')}</Label>
              <Input
                value={tier.label}
                placeholder={t('1080p video')}
                onChange={(e) =>
                  handleTierChange(index, 'label', e.target.value)
                }
              />
            </div>

            <div className='space-y-2 md:col-span-3'>
              <Label>{t('Resolution')}</Label>
              <Input
                value={tier.resolution}
                placeholder={t('720p')}
                onChange={(e) =>
                  handleTierChange(index, 'resolution', e.target.value)
                }
              />
            </div>

            <div className='space-y-2 md:col-span-3'>
              <Label>{t('Video input')}</Label>
              <Select
                value={tier.hasVideoInput}
                onValueChange={(value) =>
                  handleTierChange(index, 'hasVideoInput', value)
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='any'>{t('Any')}</SelectItem>
                  <SelectItem value='yes'>{t('Required')}</SelectItem>
                  <SelectItem value='no'>{t('Excluded')}</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className='space-y-2 md:col-span-2'>
              <Label>{t('Input price')}</Label>
              <Input
                value={tier.inputPrice}
                placeholder='31'
                onChange={(e) =>
                  handleTierChange(index, 'inputPrice', e.target.value)
                }
              />
            </div>

            <div className='flex items-end md:col-span-1'>
              <Button
                type='button'
                variant='ghost'
                size='icon'
                onClick={() => handleRemoveTier(index)}
                aria-label={t('Remove tier')}
              >
                <Trash2 className='h-4 w-4' />
              </Button>
            </div>
          </div>
        ))}
      </div>

      <Button type='button' variant='outline' onClick={handleAddTier}>
        <Plus className='mr-2 h-4 w-4' />
        {t('Add tier')}
      </Button>
    </div>
  )
}
