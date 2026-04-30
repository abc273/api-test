# 真人资产 RPA Worker

这个 worker 用于把 New API 的真人资产任务同步到火山方舟控制台：

- 拉取 `pending` 任务并在火山控制台创建资产组邀约二维码
- 把二维码回写到 New API，用户在「真人资产」页面扫码
- 用户扫码/链接上传完成后，在「真人资产」页面点击接收授权
- worker 拉取 `waiting_accept` 任务，按用户点击后的最新授权资产进入「查看资产」
- worker 回写火山资产状态、缩略图；资产可用时进入用户确认，资产失败时回写失败原因
- 缩略图会固化成压缩后的 `data:image/...`，避免火山临时图片链接过期后裂图

默认按公共二维码队列模式运行：`pending` 表示用户排队中，worker 不会自动为每个用户创建新二维码。管理员需要刷新公共二维码时，再显式运行 `PORTRAIT_RPA_PHASE=create_invite`。

## 环境变量

必填：

```bash
export NEW_API_BASE_URL=http://localhost:3000
export PORTRAIT_RPA_SECRET=change-me
```

常用：

```bash
export PORTRAIT_RPA_PROFILE_DIR=./.volc-portrait-profile
export PORTRAIT_RPA_HEADLESS=false
export PORTRAIT_RPA_INTERVAL_MS=15000
export PORTRAIT_RPA_ONCE=false
export PORTRAIT_RPA_PHASE=all
export PORTRAIT_RPA_CREATE_INVITES=false
export PORTRAIT_RPA_REFRESH_PREVIEWS=false
export PORTRAIT_RPA_PREVIEW_MAX_SIZE=240
export PORTRAIT_RPA_PREVIEW_JPEG_QUALITY=0.72
export VOLC_PORTRAIT_ENTRY_URL='https://console.volcengine.com/ark/region:ark+cn-beijing/experience/vision?modelId=doubao-seedance-2-0-260128&tab=GenVideo'
export VOLC_CHROME_EXECUTABLE_PATH='/Applications/Google Chrome.app/Contents/MacOS/Google Chrome'
export VOLC_CHROME_PROFILE_DIRECTORY=Default
export VOLC_CHROME_CDP_URL=http://127.0.0.1:9222
```

只测试二维码生成时，可以设置：

```bash
export PORTRAIT_RPA_ONCE=true
export PORTRAIT_RPA_PHASE=create_invite
```

只测试授权接收时，可以设置：

```bash
export PORTRAIT_RPA_ONCE=true
export PORTRAIT_RPA_PHASE=accept_asset
```

只修复旧的过期缩略图时，可以设置：

```bash
export PORTRAIT_RPA_ONCE=true
export PORTRAIT_RPA_PHASE=refresh_preview
```

火山控制台页面如果调整了 DOM，可以覆盖这些选择器：

```bash
export VOLC_CREATE_GROUP_SELECTOR=''
export VOLC_GROUP_NAME_SELECTOR=''
export VOLC_QR_SELECTOR=''
export VOLC_INVITE_URL_SELECTOR=''
export VOLC_GROUP_SEARCH_SELECTOR=''
export VOLC_ASSET_ID_SELECTOR=''
```

## 运行

```bash
cd scripts/portrait-rpa
npm install
npx playwright install chromium
npm start
```

第一次运行建议 `PORTRAIT_RPA_HEADLESS=false`，在弹出的浏览器中手动登录火山控制台。登录态会保存在 `PORTRAIT_RPA_PROFILE_DIR`。
