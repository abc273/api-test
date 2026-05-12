export type OutputTierPricing = {
  label?: string
  resolution?: string
  has_video_input?: boolean
  input_price: number
}

export function sanitizeOutputTierPricing(
  tiers: Array<Partial<OutputTierPricing>> | null | undefined
): OutputTierPricing[] {
  if (!Array.isArray(tiers)) {
    return []
  }

  return tiers
    .map((tier) => {
      const inputPrice = Number(tier.input_price)
      if (!Number.isFinite(inputPrice) || inputPrice <= 0) {
        return null
      }

      const label = tier.label?.trim()
      const resolution = tier.resolution?.trim()
      let hasVideoInput: boolean | undefined
      if (typeof tier.has_video_input === 'boolean') {
        hasVideoInput = tier.has_video_input
      }

      return {
        input_price: inputPrice,
        ...(label ? { label } : {}),
        ...(resolution ? { resolution } : {}),
        ...(hasVideoInput !== undefined
          ? { has_video_input: hasVideoInput }
          : {}),
      }
    })
    .filter((tier): tier is OutputTierPricing => tier !== null)
}

export function hasOutputTierPricing(
  tiers: Array<Partial<OutputTierPricing>> | null | undefined
): boolean {
  return sanitizeOutputTierPricing(tiers).length > 0
}

export function summarizeOutputTierPricing(
  tiers: Array<Partial<OutputTierPricing>> | null | undefined
) {
  const normalized = sanitizeOutputTierPricing(tiers)
  if (normalized.length === 0) {
    return null
  }

  const prices = normalized.map((tier) => tier.input_price)
  const resolutions = Array.from(
    new Set(
      normalized
        .map((tier) => tier.resolution?.trim())
        .filter((resolution): resolution is string => Boolean(resolution))
    )
  )
  const videoStates = Array.from(
    new Set(
      normalized.map((tier) => {
        if (typeof tier.has_video_input !== 'boolean') {
          return 'any'
        }
        return tier.has_video_input ? 'video' : 'no-video'
      })
    )
  )

  return {
    count: normalized.length,
    minPrice: Math.min(...prices),
    maxPrice: Math.max(...prices),
    resolutions,
    videoStates,
  }
}
