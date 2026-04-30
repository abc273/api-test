import { chromium } from 'playwright'

const apiBase = env('NEW_API_BASE_URL', 'http://localhost:3000')
const rpaSecret = env('PORTRAIT_RPA_SECRET')
const profileDir = env('PORTRAIT_RPA_PROFILE_DIR', './.volc-portrait-profile')
const headless = env('PORTRAIT_RPA_HEADLESS', 'false') === 'true'
const intervalMs = Number(env('PORTRAIT_RPA_INTERVAL_MS', '15000'))
const chromeExecutablePath = env('VOLC_CHROME_EXECUTABLE_PATH')
const chromeProfileDirectory = env('VOLC_CHROME_PROFILE_DIRECTORY')
const chromeCDPURL = env('VOLC_CHROME_CDP_URL')
const runOnce = env('PORTRAIT_RPA_ONCE', 'false') === 'true'
const phase = env('PORTRAIT_RPA_PHASE', 'all')
const createInviteEnabled = env('PORTRAIT_RPA_CREATE_INVITES', 'false') === 'true'
const entryUrl = env(
  'VOLC_PORTRAIT_ENTRY_URL',
  'https://console.volcengine.com/ark/region:ark+cn-beijing/experience/vision?modelId=doubao-seedance-2-0-260128&tab=GenVideo'
)

if (!rpaSecret) {
  throw new Error('PORTRAIT_RPA_SECRET is required')
}

function env(name, fallback = '') {
  return process.env[name] || fallback
}

async function api(path, options = {}) {
  const res = await fetch(`${apiBase}${path}`, {
    ...options,
    headers: {
      Authorization: `Bearer ${rpaSecret}`,
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
  })
  const body = await res.json().catch(() => ({}))
  if (!res.ok || body.success === false) {
    throw new Error(body.message || `HTTP ${res.status}`)
  }
  return body.data
}

async function updateJob(jobId, patch) {
  return api(`/api/portrait_assets/rpa/jobs/${jobId}`, {
    method: 'PUT',
    body: JSON.stringify(patch),
  })
}

async function listJobs(status, limit = 1) {
  const qs = new URLSearchParams({ status, limit: String(limit) })
  return api(`/api/portrait_assets/rpa/jobs?${qs}`)
}

async function clickByText(page, texts, timeout = 3000) {
  for (const text of texts) {
    for (const exact of [true, false]) {
      const locator = page.getByText(text, { exact }).first()
      try {
        await locator.waitFor({ state: 'visible', timeout })
        await locator.click()
        return true
      } catch {
        // try next match strategy
      }
    }
  }
  return false
}

async function clickByRole(page, names, timeout = 3000) {
  for (const name of names) {
    const locator = page.getByRole('button', { name }).first()
    try {
      await locator.waitFor({ state: 'visible', timeout })
      await locator.click()
      return true
    } catch {
      // try next accessible name
    }
  }
  return false
}

async function fillIfVisible(page, selector, value) {
  if (!selector) return false
  const locator = page.locator(selector).first()
  try {
    await locator.waitFor({ state: 'visible', timeout: 3000 })
    await locator.fill(value)
    return true
  } catch {
    return false
  }
}

async function captureQRCode(page) {
  const visibleQR = await page.evaluate(() => {
    const isVisible = (element) => {
      const rect = element.getBoundingClientRect()
      const style = getComputedStyle(element)
      return rect.width >= 80 &&
        rect.height >= 80 &&
        rect.x > -100 &&
        rect.y > -100 &&
        style.display !== 'none' &&
        style.visibility !== 'hidden'
    }
    const wrappers = Array.from(document.querySelectorAll('[class*="qrWrapper"]'))
    for (const wrapper of wrappers) {
      if (!isVisible(wrapper)) continue
      const image = Array.from(wrapper.querySelectorAll('img')).find(isVisible)
      if (image?.src?.startsWith('data:image')) return image.src
    }
    return ''
  }).catch(() => '')
  if (visibleQR) return visibleQR

  const explicit = env('VOLC_QR_SELECTOR')
  const candidates = [
    explicit,
    '.qrWrapper-eeUTrj img[src^="data:image"]',
    '.qrWrapper-eeUTrj canvas',
    'img[src^="data:image/gif"]',
    'img[src^="data:image/png"]',
    'img[src^="data:image"]',
    'img[src*="qrcode"]',
    'canvas',
  ].filter(Boolean)

  for (const selector of candidates) {
    const locators = page.locator(selector)
    const count = await locators.count().catch(() => 0)
    for (let i = 0; i < count; i += 1) {
      const locator = locators.nth(i)
      try {
        await locator.waitFor({ state: 'visible', timeout: 5000 })
        const box = await locator.boundingBox()
        if (!box || box.width < 80 || box.height < 80) {
          continue
        }
        const src = await locator.getAttribute('src').catch(() => null)
        if (src && src.startsWith('data:image')) return src
        const image = await locator.screenshot({ type: 'png' })
        return `data:image/png;base64,${image.toString('base64')}`
      } catch {
        // try next candidate
      }
    }
  }
  return ''
}

async function readInviteURL(page) {
  const selector = env('VOLC_INVITE_URL_SELECTOR')
  if (selector) {
    const locator = page.locator(selector).first()
    const value = await locator.inputValue().catch(() => '')
    if (value) return value
    return locator.textContent().catch(() => '')
  }
  const bodyText = await page.evaluate(() => {
    const isVisible = (element) => {
      const rect = element.getBoundingClientRect()
      const style = getComputedStyle(element)
      return rect.width > 0 &&
        rect.height > 0 &&
        rect.x > -100 &&
        rect.y > -100 &&
        style.display !== 'none' &&
        style.visibility !== 'hidden'
    }
    const dialog = Array.from(document.querySelectorAll('[role="dialog"]')).find(isVisible)
    return dialog?.innerText || document.body.innerText || ''
  }).catch(() => '')
  return bodyText.match(/https:\/\/ark\.volcengine\.com\/region:[^\s]+\/mobile\/livenees-face-manage\/index\?uuid=[0-9a-f-]+/)?.[0] || ''
}

async function closeVisibleDialog(page) {
  const dialogs = page.locator('[role="dialog"]')
  const count = await dialogs.count().catch(() => 0)
  for (let i = count - 1; i >= 0; i -= 1) {
    const dialog = dialogs.nth(i)
    if (!(await dialog.isVisible().catch(() => false))) continue
    const closeButton = dialog
      .locator('button, [role="button"]')
      .filter({ hasText: /^×$|关闭|取消/ })
      .first()
    if (await closeButton.isVisible().catch(() => false)) {
      await closeButton.click().catch(() => {})
      await page.waitForTimeout(500)
      return true
    }
    await page.keyboard.press('Escape').catch(() => {})
    await page.waitForTimeout(500)
    return true
  }
  return false
}

async function openPortraitAssetManager(page) {
  await page.goto(entryUrl, { waitUntil: 'domcontentloaded' })
  await page.waitForTimeout(2000)
  await closeVisibleDialog(page)
  await clickByText(page, ['我的'], 8000)
  await clickByText(page, ['管理资产'], 8000)
  await clickByText(page, ['真人人像资产'], 5000)
}

async function createInvite(page, job) {
  await openPortraitAssetManager(page)
  const createSelector = env('VOLC_CREATE_GROUP_SELECTOR')
  if (createSelector) {
    await page.locator(createSelector).first().click()
  }
  const opened = createSelector ||
    (await clickByRole(page, ['创建资产组', '新建资产组', '创建邀约'])) ||
    (await clickByText(page, ['创建资产组', '新建资产组', '创建邀约']))
  if (!opened) {
    throw new Error('没有找到“创建资产组/新建资产组”入口，请设置 VOLC_CREATE_GROUP_SELECTOR 或更新默认选择器')
  }

  await fillIfVisible(page, env('VOLC_GROUP_NAME_SELECTOR'), job.name || `new-api-${job.id}`)

  await page.locator('[role="dialog"] input[placeholder="开始日期"]').last().click().catch(() => {})
  await clickByText(page, ['1个月', '30天', '30 天'], 5000)
  const agreement = page.locator('[role="dialog"] .arco-checkbox').last()
  const agreed = await agreement.evaluate((el) => el.className.includes('checked')).catch(() => false)
  if (!agreed) {
    await agreement.click().catch(async () => clickByText(page, ['我已阅读', '请阅读并同意', '同意']))
  }

  const generateButton = page.getByRole('button', { name: '生成二维码' }).last()
  await generateButton.waitFor({ state: 'visible', timeout: 5000 }).catch(() => {})
  const enabled = await generateButton.evaluate((button) => !button.disabled && !button.className.includes('disabled')).catch(() => false)
  const submitted = enabled && await generateButton.click().then(() => true).catch(() => false)
  if (!submitted && !((await clickByRole(page, ['生成二维码', '创建', '确定', '确认'])) || (await clickByText(page, ['生成二维码', '创建', '确定', '确认'])))) {
    throw new Error('没有找到生成二维码的确认按钮')
  }

  await page.waitForFunction(() => {
    const text = document.body.innerText || ''
    return /https:\/\/ark\.volcengine\.com\/region:[^\s]+\/mobile\/livenees-face-manage\/index\?uuid=[0-9a-f-]+/.test(text)
  }, null, { timeout: 20000 }).catch(() => {})

  const qrImage = await captureQRCode(page)
  if (!qrImage) {
    throw new Error('没有捕获到授权二维码，请设置 VOLC_QR_SELECTOR')
  }

  const inviteURL = await readInviteURL(page)
  if (!inviteURL) {
    throw new Error('没有读取到授权二维码链接')
  }
  await updateJob(job.id, {
    status: 'qr_ready',
    qr_image: qrImage,
    invite_url: inviteURL,
  })
}

async function readAssetID(page) {
  const selector = env('VOLC_ASSET_ID_SELECTOR')
  if (selector) {
    const assetID = await page.locator(selector).first().textContent().catch(() => '')
    if (assetID?.trim()) return assetID.trim()
  }

  const text = await page.evaluate(() => {
    const isVisible = (element) => {
      const rect = element.getBoundingClientRect()
      const style = getComputedStyle(element)
      return rect.width > 0 &&
        rect.height > 0 &&
        rect.x > -100 &&
        rect.y > -100 &&
        style.display !== 'none' &&
        style.visibility !== 'hidden'
    }
    return Array.from(document.querySelectorAll('[role="dialog"], body'))
      .filter(isVisible)
      .map((element) => element.innerText || '')
      .join('\n')
  }).catch(() => '')

  return (
    text.match(/(?:Asset\s*ID|AssetID|资产\s*ID|资产ID)\s*[:：]?\s*([A-Za-z0-9_-]+)/i)?.[1] ||
    text.match(/\basset[-_][A-Za-z0-9][A-Za-z0-9_-]{6,}\b/)?.[0] ||
    ''
  )
}

async function findAuthorizedAssetRow(page) {
  const rows = page.locator('tr').filter({ hasText: '查看资产' })
  const count = await rows.count().catch(() => 0)
  let pendingRow = null
  let acceptedRow = null
  let fallbackRow = null

  for (let i = 0; i < count; i += 1) {
    const row = rows.nth(i)
    const text = await row.innerText().catch(() => '')
    if (!fallbackRow) fallbackRow = row
    if (!pendingRow && /待接收|待接受|待确认/.test(text)) pendingRow = row
    if (!acceptedRow && /已接受/.test(text)) acceptedRow = row
  }

  return pendingRow || acceptedRow || fallbackRow
}

async function clickInRow(row, texts) {
  for (const text of texts) {
    const locator = row.getByText(text, { exact: true }).first()
    if (!(await locator.isVisible().catch(() => false))) continue
    if (await locator.click({ timeout: 5000 }).then(() => true).catch(() => false)) return true
    if (await locator.click({ timeout: 5000, force: true }).then(() => true).catch(() => false)) return true
  }
  return false
}

async function openAssetDetailForJob(page) {
  const row = await findAuthorizedAssetRow(page)
  if (!row) return false
  return clickInRow(row, ['查看资产'])
}

async function findAssetDetailRow(page) {
  const rows = page.locator('tr')
  const count = await rows.count().catch(() => 0)
  for (let i = 0; i < count; i += 1) {
    const row = rows.nth(i)
    const text = await row.innerText().catch(() => '')
    if (/\basset[-_][A-Za-z0-9][A-Za-z0-9_-]{6,}\b/.test(text)) return row
  }
  return null
}

function parseAssetTimestamp(assetID) {
  const match = assetID.match(/^asset[-_](\d{14})[-_]/)
  if (!match) return 0
  const value = match[1]
  const year = Number(value.slice(0, 4))
  const month = Number(value.slice(4, 6))
  const day = Number(value.slice(6, 8))
  const hour = Number(value.slice(8, 10))
  const minute = Number(value.slice(10, 12))
  const second = Number(value.slice(12, 14))
  if ([year, month, day, hour, minute, second].some((item) => Number.isNaN(item))) return 0
  return Math.floor(new Date(year, month - 1, day, hour, minute, second).getTime() / 1000)
}

function isAssetInJobWindow(assetID, job) {
  const assetTime = parseAssetTimestamp(assetID)
  if (!assetTime) return true
  const acceptTime = Number(job.accept_time || 0)
  const qrExpireTime = Number(job.qr_expires_time || 0)
  const lowerBound = qrExpireTime
    ? qrExpireTime - 180 - 60
    : acceptTime - 10 * 60
  const upperBound = Math.max(acceptTime, qrExpireTime) + 10 * 60
  return assetTime >= lowerBound && assetTime <= upperBound
}

async function readAssetDetail(page) {
  const assetRow = await findAssetDetailRow(page)
  if (!assetRow) {
    return { assetID: '', assetStatus: '', failureReason: '', assetPreview: '' }
  }
  const text = await assetRow.innerText().catch(() => '')
  let assetPreview = ''
  if (assetRow) {
    const preview = assetRow.locator('img, [haveavatar], [style*="background-image"]').first()
    if (await preview.isVisible().catch(() => false)) {
      const imageBuffer = await preview.screenshot({ type: 'png' }).catch(() => null)
      if (imageBuffer) {
        assetPreview = `data:image/png;base64,${imageBuffer.toString('base64')}`
      } else {
        const previewURL = await preview.getAttribute('src').catch(() => '') ||
          await preview.getAttribute('haveavatar').catch(() => '') ||
          ''
        assetPreview = previewURL
      }
    }
  }
  const lines = text.split('\n').map((line) => line.trim()).filter(Boolean)
  const assetID = text.match(/\basset[-_][A-Za-z0-9][A-Za-z0-9_-]{6,}\b/)?.[0] || ''
  const statusValues = ['失败', '不通过', '不合规', '拒绝', '驳回', '可用', '正常', '成功', '通过', '审核中', '待审核', '处理中']
  const assetIndex = lines.findIndex((line) => line === assetID)
  const statusIndex = lines.findIndex((line, index) =>
    index > assetIndex && statusValues.includes(line)
  )
  const assetStatus = statusIndex >= 0 ? lines[statusIndex] : ''
  const failureReason = statusIndex >= 0 && /失败|不通过|不合规|拒绝|驳回/.test(assetStatus)
    ? lines.slice(statusIndex + 1).find((line) => !/^asset[-_]/.test(line) && !/^group[-_]/.test(line)) || ''
    : ''

  return { assetID, assetStatus, failureReason, assetPreview }
}

async function acceptFinishedAsset(page, job) {
  await openPortraitAssetManager(page)

  await clickByText(page, ['授权给我的', '待接收', '待接受', '待确认', '授权素材'], 5000)
  const row = await findAuthorizedAssetRow(page)
  if (!row) {
    await updateJob(job.id, {
      status: 'waiting_accept',
      error_message: '没有在火山控制台找到授权资产',
    })
    return
  }
  const rowText = await row.innerText().catch(() => '')
  if (/待接收|待接受|待确认/.test(rowText)) {
    const accepted = await clickInRow(row, ['接收', '接受', '接受授权', '确认接收', '确认'])
    if (!accepted) {
      await updateJob(job.id, {
        status: 'waiting_accept',
        error_message: '没有找到火山控制台的接收按钮',
      })
      return
    }
    await page.waitForTimeout(3000)
  }

  if (!(await openAssetDetailForJob(page))) {
    await updateJob(job.id, {
      status: 'waiting_accept',
      error_message: '接收授权后没有找到“查看资产”入口',
    })
    return
  }

  await page.waitForTimeout(3000)
  const detail = await readAssetDetail(page)
  if (detail.assetID && !isAssetInJobWindow(detail.assetID, job)) {
    await updateJob(job.id, {
      status: 'waiting_accept',
      error_message: '未找到本次扫码产生的新资产，请确认上传完成后重试',
    })
    return
  }
  if (/失败|不通过|不合规|拒绝|驳回/.test(detail.assetStatus)) {
    await updateJob(job.id, {
      status: 'failed',
      asset_status: detail.assetStatus,
      asset_preview: detail.assetPreview,
      error_message: detail.failureReason || detail.assetStatus || '资产审核失败',
    })
    return
  }
  if (!detail.assetID || /审核中|待审核|处理中/.test(detail.assetStatus)) {
    await updateJob(job.id, {
      status: 'waiting_accept',
      asset_status: detail.assetStatus,
      asset_preview: detail.assetPreview,
      error_message: detail.assetStatus ? `资产状态：${detail.assetStatus}` : '未找到本次扫码产生的资产详情，请确认上传完成后重试',
    })
    return
  }

  await updateJob(job.id, {
    status: 'pending_confirm',
    asset_id: detail.assetID.trim(),
    asset_status: detail.assetStatus,
    asset_preview: detail.assetPreview,
    error_message: '请确认缩略图是否为本人',
  })
}

async function processOnce(page) {
  if (phase === 'create_invite' || (phase === 'all' && createInviteEnabled)) {
    for (const job of await listJobs('pending', 1)) {
      console.log(`[portrait-rpa] create invite for job ${job.id}`)
      try {
        await createInvite(page, job)
      } catch (error) {
        await updateJob(job.id, {
          status: 'failed',
          error_message: error.message,
        })
      }
    }
  }

  if (phase === 'all' || phase === 'accept_asset') {
    for (const status of ['waiting_accept']) {
      for (const job of await listJobs(status, 5)) {
        console.log(`[portrait-rpa] accept asset for job ${job.id}`)
        try {
          await acceptFinishedAsset(page, job)
        } catch (error) {
          await updateJob(job.id, {
            status: 'failed',
            error_message: error.message,
          })
        }
      }
    }
  }
}

async function main() {
  let browser
  let context
  if (chromeCDPURL) {
    browser = await chromium.connectOverCDP(chromeCDPURL)
    context = browser.contexts()[0] || await browser.newContext({ viewport: { width: 1440, height: 1000 } })
  } else {
    const launchOptions = {
      headless,
      viewport: { width: 1440, height: 1000 },
    }
    if (chromeExecutablePath) {
      launchOptions.executablePath = chromeExecutablePath
    }
    if (chromeProfileDirectory) {
      launchOptions.args = [`--profile-directory=${chromeProfileDirectory}`]
    }

    context = await chromium.launchPersistentContext(profileDir, launchOptions)
  }
  const page = context.pages()[0] || (await context.newPage())

  console.log('[portrait-rpa] started')
  console.log(`[portrait-rpa] profile: ${profileDir}`)
  if (chromeCDPURL) console.log(`[portrait-rpa] cdp: ${chromeCDPURL}`)
  if (chromeExecutablePath) console.log(`[portrait-rpa] chrome: ${chromeExecutablePath}`)
  if (chromeProfileDirectory) console.log(`[portrait-rpa] chrome profile directory: ${chromeProfileDirectory}`)
  console.log(`[portrait-rpa] phase: ${phase}`)
  console.log(`[portrait-rpa] entry: ${entryUrl}`)

  while (true) {
    await processOnce(page).catch((error) => {
      console.error('[portrait-rpa] loop failed:', error)
    })
    if (runOnce) break
    await new Promise((resolve) => setTimeout(resolve, intervalMs))
  }

  if (runOnce) {
    if (!chromeCDPURL) {
      await context?.close().catch(() => {})
      await browser?.close().catch(() => {})
    }
    process.exit(0)
  }
}

main().catch((error) => {
  console.error(error)
  process.exit(1)
})
