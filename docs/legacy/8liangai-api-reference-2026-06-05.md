# 8liangai.com API Reference

版本：`2026-06-08`

Base URL：`https://8liangai.com`

本文档是 8liangai.com 当前线上接口的标准接入文档，覆盖视频生成、图片生成、真人资产、虚拟人像资产、API Key 创建和资产创建教程。

---

## 1. 快速开始

### 1.1 鉴权

所有 OpenAI 兼容接口和资产接口均使用 Bearer Token 鉴权。

```http
Authorization: Bearer sk-你的APIKey
```

请求头填写规则：

| Header | 必填 | 示例 | 说明 |
| --- | --- | --- | --- |
| `Authorization` | 是 | `Bearer sk-xxxx` | `Bearer` 后必须有一个空格 |
| `Content-Type` | POST JSON 必填 | `application/json` | JSON 请求体使用 |
| `Accept` | 建议 | `application/json` | 建议所有 JSON 接口填写 |

### 1.2 常用接口清单

| 能力 | 方法 | 路径 |
| --- | --- | --- |
| 模型列表 | `GET` | `/v1/models` |
| 视频任务提交 | `POST` | `/v1/video/generations` |
| 视频任务提交，OpenAI 兼容 | `POST` | `/v1/videos` |
| 视频任务状态查询，业务格式 | `GET` | `/v1/video/generations/{task_id}` |
| 视频任务状态查询，OpenAI 兼容 | `GET` | `/v1/videos/{task_id}` |
| 视频文件下载 | `GET` | `/v1/videos/{task_id}/content` |
| 图片生成 | `POST` | `/v1/images/generations` |
| 真人资产配置 | `GET` | `/api/portrait_assets/official/config` |
| 真人资产任务列表 | `GET` | `/api/portrait_assets/official/jobs` |
| 创建真人资产任务 | `POST` | `/api/portrait_assets/official/jobs` |
| 虚拟人像配置 | `GET` | `/api/portrait_assets/virtual/config` |
| 虚拟人像资产组 | `GET` | `/api/portrait_assets/virtual/group` |
| 虚拟人像资产列表 | `GET` | `/api/portrait_assets/virtual/assets` |
| 创建虚拟人像资产 | `POST` | `/api/portrait_assets/virtual/assets` |

### 1.3 推荐最小视频请求

```bash
curl -X POST 'https://8liangai.com/v1/video/generations' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "seedance2",
    "prompt": "一只小熊猫在干净的摄影棚里挥手，镜头稳定，动作自然，无文字",
    "resolution": "720p",
    "ratio": "16:9",
    "duration": 5,
    "metadata": {
      "watermark": false
    }
  }'
```

提交成功后返回 `id` 和 `task_id`，后续查询和下载均使用该任务 ID。

---

## 2. 通用规范

### 2.1 OpenAI 风格错误结构

视频、图片等 OpenAI 兼容接口发生错误时通常返回：

```json
{
  "error": {
    "message": "错误说明",
    "type": "invalid_request_error",
    "code": "错误码"
  }
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `error.message` | `string` | 可展示或记录的错误说明 |
| `error.type` | `string` | 错误类型，例如 `invalid_request_error`、`server_error`、`new_api_error` |
| `error.code` | `string` | 上游或网关错误码，部分错误不返回该字段 |

### 2.2 业务包装结构

资产接口使用统一业务包装。

成功响应：

```json
{
  "success": true,
  "message": "",
  "data": {}
}
```

失败响应：

```json
{
  "success": false,
  "message": "错误说明"
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `success` | `boolean` | 业务是否成功 |
| `message` | `string` | 成功时为空字符串，失败时为错误说明 |
| `data` | `any` | 成功时的数据，可能为对象、数组或 `null` |

### 2.3 分页结构

列表接口统一使用分页包装。

请求参数：

| 参数 | 类型 | 必填 | 默认值 | 最大值 | 说明 |
| --- | --- | --- | --- | --- | --- |
| `p` | `integer` | 否 | `1` | 无固定上限 | 页码，从 `1` 开始 |
| `page_size` | `integer` | 否 | 站点默认值 | `100` | 每页数量 |
| `ps` | `integer` | 否 | 无 | `100` | 兼容字段，未传 `page_size` 时可用 |
| `size` | `integer` | 否 | 无 | `100` | 兼容字段，未传 `page_size` 和 `ps` 时可用 |

响应示例：

```json
{
  "success": true,
  "message": "",
  "data": {
    "page": 1,
    "page_size": 20,
    "total": 42,
    "items": []
  }
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `data.page` | `integer` | 当前页码 |
| `data.page_size` | `integer` | 当前每页数量 |
| `data.total` | `integer` | 总记录数 |
| `data.items` | `array` | 当前页数据 |

### 2.4 计费说明

视频任务一般有预扣、结算、退款差额三个阶段。

| 名称 | 含义 |
| --- | --- |
| 预扣 | 提交任务时按模型上限预先冻结或扣减额度 |
| 实扣 | 任务完成后根据上游 `usage.total_tokens` 和模型倍率计算实际额度 |
| 退款差额 | `预扣额度 - 实扣额度`，正数表示退回，负数表示补扣 |

图片任务当前按次计费，成功生成后按模型单价扣费；参数非法未生成时不扣费。

---

## 3. 模型能力

### 3.1 视频模型

| 模型 | 类型 | 请求接口 | 状态查询接口 | 下载接口 |
| --- | --- | --- | --- | --- |
| `seedance1.5` | 视频生成 | `/v1/video/generations` | `/v1/video/generations/{task_id}`、`/v1/videos/{task_id}` | `/v1/videos/{task_id}/content` |
| `seedance1.5-sr` | 视频生成 | 同上 | 同上 | 同上 |
| `seedance2` | 视频生成 | 同上 | 同上 | 同上 |
| `sd2.0fast` | 快速视频生成 | 同上 | 同上 | 同上 |
| `seedance2-sr` | 视频生成 | 同上 | 同上 | 同上 |
| `seedance2.0fast-sr` | 快速视频生成 | 同上 | 同上 | 同上 |

### 3.2 视频分辨率支持

| 模型 | 支持的 `resolution` | 推荐默认值 | 说明 |
| --- | --- | --- | --- |
| `seedance1.5` | `480p`、`720p`、`1080p` | `720p` |  |
| `seedance1.5-sr` | `720p`、`1080p` | `720p` | 高清输出 |
| `seedance2` | `480p`、`720p`、`1080p` | `720p` | 1080p 使用独立计费档 |
| `sd2.0fast` | `480p`、`720p` | `720p` | 快速模型 |
| `seedance2-sr` | `720p`、`1080p` | `720p` | 高清输出 |
| `seedance2.0fast-sr` | `720p`、`1080p` | `720p` | 高清输出 |

### 3.3 视频画幅比例

当前文档按站点已验证能力和火山 Seedance 官方文档整理。常规接入建议显式填写 `ratio`。

| `ratio` | 含义 | 固定尺寸校验 |
| --- | --- | --- |
| `16:9` | 横版宽屏 | 支持 |
| `4:3` | 横版传统比例 | 支持 |
| `1:1` | 正方形 | 支持 |
| `3:4` | 竖版海报比例 | 支持 |
| `9:16` | 竖屏短视频 | 支持 |
| `21:9` | 超宽屏 | 支持 |
| `adaptive` | 自适应比例 | 支持提交，不使用固定宽高断言 |

`adaptive` 说明：

| 场景 | 行为 |
| --- | --- |
| 文生视频 | 模型根据提示词自动选择画幅 |
| 图生视频 | 通常优先参考输入图片比例 |
| 视频参考输入 | 通常优先参考输入视频比例 |
| 生产接入 | 建议优先使用明确比例，避免上下游展示不可控 |

官方参考：火山方舟 Seedance 2.0 API 文档说明 `ratio=adaptive` 时会自动适配宽高比，并可在任务查询结果中查看实际 `ratio`。

### 3.4 常见输出像素尺寸

下表为当前线上回归已验证的典型输出尺寸。`adaptive` 不列固定尺寸。

| `resolution` | `ratio` | 输出像素 |
| --- | --- | --- |
| `480p` | `16:9` | `864x496` |
| `480p` | `4:3` | `752x560` |
| `480p` | `1:1` | `640x640` |
| `480p` | `3:4` | `560x752` |
| `480p` | `9:16` | `496x864` |
| `480p` | `21:9` | `992x432` |
| `720p` | `16:9` | `1280x720` |
| `720p` | `4:3` | `1112x834` |
| `720p` | `1:1` | `960x960` |
| `720p` | `3:4` | `834x1112` |
| `720p` | `9:16` | `720x1280` |
| `720p` | `21:9` | `1470x630` |
| `1080p` | `16:9` | `1920x1080` |
| `1080p` | `4:3` | `1664x1248` |
| `1080p` | `1:1` | `1440x1440` |
| `1080p` | `3:4` | `1248x1664` |
| `1080p` | `9:16` | `1080x1920` |
| `1080p` | `21:9` | `2206x946` |

### 3.5 图片模型

| 模型 | 接口 | 当前有效尺寸 |
| --- | --- | --- |
| `seedream4.5` | `/v1/images/generations` | `2048x2048` |
| `seedream5.0lite` | `/v1/images/generations` | `2048x2048` |

`1024x1024` 对这两个模型返回参数错误且不扣费。生产接入建议使用 `2048x2048` 起步。

---

## 4. 视频任务提交

### 4.1 接口信息

```http
POST /v1/video/generations
POST /v1/videos
```

推荐使用 `/v1/video/generations`。`/v1/videos` 为 OpenAI 兼容创建入口，当前请求体相同。

### 4.2 请求体字段

| 字段 | 类型 | 必填 | 默认值 | 可选值或格式 | 说明 |
| --- | --- | --- | --- | --- | --- |
| `model` | `string` | 是 | 无 | 见视频模型表 | 模型别名 |
| `prompt` | `string` | 是 | 无 | 自然语言文本 | 提示词，会作为 `content` 中的文本发送上游 |
| `resolution` | `string` | 否 | 多数模型默认 `720p` | `480p`、`720p`、`1080p` | 输出分辨率档位，按模型能力表填写 |
| `ratio` | `string` | 否 | 上游默认值 | `16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive` | 输出画幅比例 |
| `duration` | `integer` 或数字字符串 | 否 | 上游默认值 | `5` 等正整数 | 视频时长，推荐传整数 |
| `seconds` | `string` | 否 | 无 | `"5"` 等数字字符串 | 时长兼容字段；`duration` 优先 |
| `images` | `array<string>` | 否 | 空数组 | 公网可访问图片 URL | 图生视频参考图，推荐使用该字段 |
| `image` | `string` | 否 | 无 | 公网可访问图片 URL | 兼容字段；当前 Doubao 视频链路优先使用 `images` |
| `size` | `string` | 否 | 无 | 兼容字段 | 视频模型建议用 `resolution`，不要用 `size` 表达清晰度 |
| `mode` | `string` | 否 | 无 | 兼容字段 | 当前文档覆盖模型通常不需要填写 |
| `input_reference` | `string` | 否 | 无 | 兼容字段 | 当前文档覆盖模型通常不需要填写 |
| `metadata` | `object` 或 JSON 字符串 | 否 | `{}` | 见下表 | 高级扩展字段 |

填写建议：

| 场景 | 推荐字段组合 |
| --- | --- |
| 文生视频 | `model`、`prompt`、`resolution`、`ratio`、`duration` |
| 图生视频 | 文生视频字段加 `images` |
| 视频参考输入 | 文生视频字段加 `metadata.content` |
| 真人资产绑定 | 文生视频字段加 `metadata.portrait_asset_id` 或 `metadata.asset_id` |

### 4.3 `metadata` 字段

`metadata` 会映射到上游视频任务 payload。常用字段如下：

| 字段 | 类型 | 必填 | 可选值或格式 | 说明 |
| --- | --- | --- | --- | --- |
| `callback_url` | `string` | 否 | `https://your-domain/callback` | 上游任务回调地址 |
| `return_last_frame` | `boolean` | 否 | `true`、`false` | 是否返回最后一帧 |
| `service_tier` | `string` | 否 | 上游支持的服务层级 | 通常不填 |
| `execution_expires_after` | `integer` | 否 | 秒数 | 任务执行过期时间 |
| `generate_audio` | `boolean` | 否 | `true`、`false` | 是否生成音频，按上游模型能力生效 |
| `draft` | `boolean` | 否 | `true`、`false` | 草稿模式 |
| `tools` | `array<object>` | 否 | `[{ "type": "..." }]` | 上游工具参数 |
| `resolution` | `string` | 否 | 同顶层 `resolution` | metadata 中该字段优先于顶层字段 |
| `ratio` | `string` | 否 | 同顶层 `ratio` | metadata 中该字段优先于顶层字段 |
| `duration` | `integer` | 否 | 正整数 | metadata 中该字段可映射上游时长 |
| `frames` | `integer` | 否 | 正整数 | 指定帧数 |
| `seed` | `integer` | 否 | 整数 | 随机种子 |
| `camera_fixed` | `boolean` | 否 | `true`、`false` | 是否固定机位 |
| `watermark` | `boolean` | 否 | `true`、`false` | 是否添加水印 |
| `content` | `array<object>` | 否 | 见下一节 | 参考媒体内容 |
| `portrait_asset_id` | `integer` 或整数字符串 | 否 | 真人资产任务 ID | 绑定当前用户已就绪真人资产任务 |
| `asset_id` | `string` | 否 | `asset_xxx` 或 `asset://asset_xxx` | 绑定当前用户已就绪真人资产 |

优先级规则：

| 参数 | 优先级 |
| --- | --- |
| `metadata.resolution` | 高于顶层 `resolution` |
| `metadata.ratio` | 高于顶层 `ratio` |
| 顶层 `duration` | 高于 `seconds` |
| `metadata.portrait_asset_id` | 高于 `metadata.asset_id` |

### 4.4 `metadata.content` 字段

`metadata.content` 用于精确传入图片、视频、音频参考内容。

内容项字段：

| 字段 | 类型 | 必填 | 可选值 | 说明 |
| --- | --- | --- | --- | --- |
| `type` | `string` | 是 | `text`、`image_url`、`video_url`、`audio_url` | 内容类型 |
| `role` | `string` | 按类型 | `reference_image`、`reference_video` | 参考媒体角色 |
| `text` | `string` | `type=text` 时使用 | 文本 | 文本内容 |
| `image_url.url` | `string` | `type=image_url` 时使用 | 图片 URL 或 `asset://<ASSET_ID>` | 图片必须可被上游访问；虚拟人像图片资产可填资产 URI |
| `video_url.url` | `string` | `type=video_url` 时使用 | 视频 URL 或 `asset://<ASSET_ID>` | 视频必须可被上游访问；虚拟人像视频资产可填资产 URI |
| `audio_url.url` | `string` | `type=audio_url` 时使用 | 音频 URL 或 `asset://<ASSET_ID>` | 音频必须可被上游访问；虚拟人像音频资产可填资产 URI |

视频参考输入必须设置：

```json
{
  "type": "video_url",
  "role": "reference_video",
  "video_url": {
    "url": "https://example.com/input.mp4"
  }
}
```

图片参考输入标准写法：

```json
{
  "type": "image_url",
  "role": "reference_image",
  "image_url": {
    "url": "https://example.com/ref.png"
  }
}
```

虚拟人像资产标准写法：

```json
{
  "type": "image_url",
  "role": "reference_image",
  "image_url": {
    "url": "asset://asset_xxx"
  }
}
```

虚拟人像资产引用规则：

| 资产类型 | `content.type` | URL 字段 | 填法 |
| --- | --- | --- | --- |
| 图片资产 | `image_url` | `image_url.url` | `asset://<volc_asset_id>` |
| 视频资产 | `video_url` | `video_url.url` | `asset://<volc_asset_id>` |
| 音频资产 | `audio_url` | `audio_url.url` | `asset://<volc_asset_id>` |

注意：`metadata.portrait_asset_id` 是真人资产任务 ID 的绑定方式；虚拟人像资产不要填到 `portrait_asset_id`，应通过 `metadata.content` 使用 `asset://<volc_asset_id>` 引用。

### 4.5 请求示例

文生视频：

```json
{
  "model": "seedance1.5",
  "prompt": "一位年轻人在海边回头微笑，风吹动头发，电影感布光，镜头缓慢推进",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "watermark": false
  }
}
```

图生视频：

```json
{
  "model": "seedance2",
  "prompt": "让人物轻微眨眼并自然转头，保持服装和背景一致",
  "resolution": "720p",
  "ratio": "9:16",
  "duration": 5,
  "images": [
    "https://example.com/reference.png"
  ],
  "metadata": {
    "watermark": false
  }
}
```

视频参考输入：

```json
{
  "model": "seedance2",
  "prompt": "保持主体不变，让动作更流畅，镜头稳定",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "watermark": false,
    "content": [
      {
        "type": "video_url",
        "role": "reference_video",
        "video_url": {
          "url": "https://example.com/input.mp4"
        }
      }
    ]
  }
}
```

虚拟人像资产生成视频：

```json
{
  "model": "seedance2",
  "prompt": "参考虚拟人像资产中的人物形象，让人物自然挥手并看向镜头",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "watermark": false,
    "content": [
      {
        "type": "image_url",
        "role": "reference_image",
        "image_url": {
          "url": "asset://asset_xxx"
        }
      }
    ]
  }
}
```

真人资产绑定：

```json
{
  "model": "seedance2",
  "prompt": "人物自然挥手并看向镜头，动作真实，背景干净",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "portrait_asset_id": 12,
    "watermark": false
  }
}
```

通过最终资产 ID 绑定：

```json
{
  "model": "seedance2",
  "prompt": "人物自然挥手并看向镜头，动作真实，背景干净",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "asset_id": "asset_xxx",
    "watermark": false
  }
}
```

### 4.6 成功响应

```json
{
  "id": "task_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "task_id": "task_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "object": "video",
  "model": "seedance2",
  "status": "queued",
  "progress": 0,
  "created_at": 1780641005
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `string` | 对外公开任务 ID |
| `task_id` | `string` | 与 `id` 相同，兼容旧客户端 |
| `object` | `string` | 固定为 `video` |
| `model` | `string` | 请求使用的模型别名 |
| `status` | `string` | 初始一般为 `queued` |
| `progress` | `integer` | 任务进度，初始通常为 `0` |
| `created_at` | `integer` | Unix 秒级时间戳 |
| `completed_at` | `integer` | 完成时间，提交阶段通常不返回 |
| `expires_at` | `integer` | 过期时间，存在时返回 |
| `seconds` | `string` | 兼容字段，存在时返回 |
| `size` | `string` | 兼容字段，存在时返回 |
| `remixed_from_video_id` | `string` | Remix 来源视频 ID，存在时返回 |
| `error` | `object` | 失败时返回，提交成功通常为空 |
| `metadata` | `object` | 附加信息，存在时返回 |

---

## 5. 视频状态查询

### 5.1 业务格式查询

```http
GET /v1/video/generations/{task_id}
```

路径参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `task_id` | `string` | 是 | 视频提交接口返回的 `task_id` |

响应示例：

```json
{
  "code": "success",
  "data": {
    "id": 10001,
    "created_at": 1780641005,
    "updated_at": 1780641060,
    "task_id": "task_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "platform": "doubao-video",
    "user_id": 1,
    "group": "default",
    "quota": 1240631,
    "action": "generate",
    "status": "SUCCESS",
    "fail_reason": "",
    "result_url": "https://example.com/video.mp4",
    "submit_time": 1780641005,
    "start_time": 1780641006,
    "finish_time": 1780641060,
    "progress": "100%",
    "properties": {
      "origin_model_name": "seedance2"
    },
    "data": {
      "id": "upstream_task_id",
      "model": "doubao-seedance-2-0-260128",
      "status": "succeeded",
      "content": {
        "video_url": "https://example.com/video.mp4"
      },
      "resolution": "720p",
      "ratio": "16:9",
      "duration": 5,
      "framespersecond": 24,
      "usage": {
        "completion_tokens": 0,
        "total_tokens": 50638
      }
    }
  }
}
```

顶层字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `code` | `string` | 成功时为 `success` |
| `message` | `string` | 消息文本，可能为空 |
| `data` | `object` | 任务记录 |

`data` 字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `integer` | 本地任务记录 ID |
| `created_at` | `integer` | 本地创建时间 |
| `updated_at` | `integer` | 本地更新时间 |
| `task_id` | `string` | 对外任务 ID |
| `platform` | `string` 或 `number` | 任务平台标识 |
| `user_id` | `integer` | 用户 ID |
| `group` | `string` | 计费分组 |
| `channel_id` | `integer` | 通道 ID，部分响应可能省略 |
| `quota` | `integer` | 当前记录额度 |
| `action` | `string` | 任务动作，一般为 `generate` |
| `status` | `string` | 业务状态 |
| `fail_reason` | `string` | 失败原因或历史兼容结果字段 |
| `result_url` | `string` | 成功后的最终视频 URL |
| `submit_time` | `integer` | 提交时间 |
| `start_time` | `integer` | 开始时间 |
| `finish_time` | `integer` | 完成时间 |
| `progress` | `string` | 进度字符串，例如 `10%`、`50%`、`100%` |
| `properties.input` | `string` | 原始输入摘要 |
| `properties.upstream_model_name` | `string` | 上游模型名 |
| `properties.origin_model_name` | `string` | 用户请求模型名 |
| `properties.video_super_resolution_requested` | `boolean` | 是否触发高清视频处理 |
| `data` | `object` | 上游任务原始数据或后处理包装数据 |

业务状态值：

| 状态 | 含义 |
| --- | --- |
| `NOT_START` | 本地任务已创建但尚未提交 |
| `SUBMITTED` | 已提交到上游 |
| `QUEUED` | 上游排队中 |
| `IN_PROGRESS` | 上游处理中 |
| `SUCCESS` | 任务成功 |
| `FAILURE` | 任务失败 |
| `UNKNOWN` | 未识别状态 |

### 5.2 OpenAI 兼容查询

```http
GET /v1/videos/{task_id}
```

响应示例：

```json
{
  "id": "task_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "task_id": "task_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "object": "video",
  "model": "seedance2",
  "status": "completed",
  "progress": 100,
  "created_at": 1780641005,
  "completed_at": 1780641060,
  "metadata": {
    "url": "https://example.com/video.mp4"
  }
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `string` | 任务 ID |
| `task_id` | `string` | 兼容字段，与 `id` 相同 |
| `object` | `string` | 固定为 `video` |
| `model` | `string` | 用户请求模型别名 |
| `status` | `string` | OpenAI 风格状态 |
| `progress` | `integer` | 数字进度 |
| `created_at` | `integer` | 创建时间 |
| `completed_at` | `integer` | 完成时间，完成或失败时返回 |
| `expires_at` | `integer` | 过期时间，存在时返回 |
| `seconds` | `string` | 兼容字段，存在时返回 |
| `size` | `string` | 兼容字段，存在时返回 |
| `remixed_from_video_id` | `string` | Remix 来源视频 ID，存在时返回 |
| `error` | `object` | 失败时返回 |
| `metadata.url` | `string` | 成功后的视频 URL |

OpenAI 兼容状态值：

| 状态 | 含义 |
| --- | --- |
| `queued` | 排队中 |
| `in_progress` | 处理中 |
| `completed` | 已完成 |
| `failed` | 已失败 |
| `unknown` | 未识别状态 |

错误对象字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `error.message` | `string` | 失败原因 |
| `error.code` | `string` | 错误码 |

---

## 6. 视频文件下载

### 6.1 接口信息

```http
GET /v1/videos/{task_id}/content
```

路径参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `task_id` | `string` | 是 | 视频提交接口返回的任务 ID |

### 6.2 成功响应

成功时返回视频二进制流，不是 JSON。

| 项目 | 说明 |
| --- | --- |
| HTTP 状态码 | `200` |
| `Content-Type` | 通常为 `video/mp4`，以实际响应头为准 |
| 响应体 | 视频二进制内容 |
| 缓存头 | 网关会设置公开视频缓存头 |

下载示例：

```bash
curl -L 'https://8liangai.com/v1/videos/task_xxx/content' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -o result.mp4
```

### 6.3 常见错误

任务不存在：

```json
{
  "error": {
    "message": "Task not found",
    "type": "invalid_request_error"
  }
}
```

任务未完成：

```json
{
  "error": {
    "message": "Task is not completed yet, current status: IN_PROGRESS",
    "type": "invalid_request_error"
  }
}
```

上游文件不可取：

```json
{
  "error": {
    "message": "Failed to fetch video content",
    "type": "server_error"
  }
}
```

---

## 7. 各视频模型接入说明

### 7.1 `seedance1.5`

| 项目 | 值 |
| --- | --- |
| 请求接口 | `POST /v1/video/generations` |
| 状态查询 | `GET /v1/video/generations/{task_id}`、`GET /v1/videos/{task_id}` |
| 文件下载 | `GET /v1/videos/{task_id}/content` |
| 支持分辨率 | `480p`、`720p`、`1080p` |
| 支持比例 | `16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive` |

示例：

```json
{
  "model": "seedance1.5",
  "prompt": "一个年轻人在海边回头，风吹动头发，镜头缓慢推进",
  "resolution": "1080p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "watermark": false
  }
}
```

### 7.2 `seedance1.5-sr`

| 项目 | 值 |
| --- | --- |
| 请求接口 | `POST /v1/video/generations` |
| 状态查询 | `GET /v1/video/generations/{task_id}`、`GET /v1/videos/{task_id}` |
| 文件下载 | `GET /v1/videos/{task_id}/content` |
| 支持分辨率 | `720p`、`1080p` |
| 支持比例 | `16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive` |

示例：

```json
{
  "model": "seedance1.5-sr",
  "prompt": "高清商业广告镜头，人物自然走向镜头，动作流畅，画面清晰",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "watermark": false
  }
}
```

### 7.3 `seedance2`

| 项目 | 值 |
| --- | --- |
| 请求接口 | `POST /v1/video/generations` |
| 状态查询 | `GET /v1/video/generations/{task_id}`、`GET /v1/videos/{task_id}` |
| 文件下载 | `GET /v1/videos/{task_id}/content` |
| 支持分辨率 | `480p`、`720p`、`1080p` |
| 支持比例 | `16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive` |

文生视频示例：

```json
{
  "model": "seedance2",
  "prompt": "一只小熊猫在摄影棚里挥手，镜头稳定，动作自然",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "watermark": false
  }
}
```

视频参考输入示例：

```json
{
  "model": "seedance2",
  "prompt": "保持原视频主体和构图，让动作更自然流畅",
  "resolution": "1080p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "watermark": false,
    "content": [
      {
        "type": "video_url",
        "role": "reference_video",
        "video_url": {
          "url": "https://example.com/input.mp4"
        }
      }
    ]
  }
}
```

### 7.4 `sd2.0fast`

| 项目 | 值 |
| --- | --- |
| 请求接口 | `POST /v1/video/generations` |
| 状态查询 | `GET /v1/video/generations/{task_id}`、`GET /v1/videos/{task_id}` |
| 文件下载 | `GET /v1/videos/{task_id}/content` |
| 支持分辨率 | `480p`、`720p` |
| 支持比例 | `16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive` |

示例：

```json
{
  "model": "sd2.0fast",
  "prompt": "产品广告短片，桌面上的陶瓷杯轻微旋转，光线柔和，镜头稳定",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "watermark": false
  }
}
```

### 7.5 `seedance2-sr`

| 项目 | 值 |
| --- | --- |
| 请求接口 | `POST /v1/video/generations` |
| 状态查询 | `GET /v1/video/generations/{task_id}`、`GET /v1/videos/{task_id}` |
| 文件下载 | `GET /v1/videos/{task_id}/content` |
| 支持分辨率 | `720p`、`1080p` |
| 支持比例 | `16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive` |

示例：

```json
{
  "model": "seedance2-sr",
  "prompt": "高清品牌宣传片，人物自然转身并向镜头微笑，质感真实",
  "resolution": "1080p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "watermark": false
  }
}
```

### 7.6 `seedance2.0fast-sr` 与 `sd2.0fast-sr`

| 项目 | 值 |
| --- | --- |
| 请求接口 | `POST /v1/video/generations` |
| 状态查询 | `GET /v1/video/generations/{task_id}`、`GET /v1/videos/{task_id}` |
| 文件下载 | `GET /v1/videos/{task_id}/content` |
| 推荐模型名 | `seedance2.0fast-sr` |
| 兼容模型名 | `sd2.0fast-sr` |
| 支持分辨率 | `720p`、`1080p` |
| 支持比例 | `16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive` |

示例：

```json
{
  "model": "seedance2.0fast-sr",
  "prompt": "快速生成高清短视频，咖啡杯旁边有蒸汽上升，商业摄影风格",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "watermark": false
  }
}
```

---

## 8. 图片生成

### 8.1 接口信息

```http
POST /v1/images/generations
```

### 8.2 请求体字段

| 字段 | 类型 | 必填 | 默认值 | 可选值或格式 | 说明 |
| --- | --- | --- | --- | --- | --- |
| `model` | `string` | 是 | 无 | `seedream4.5`、`seedream5.0lite` | 图片模型别名 |
| `prompt` | `string` | 是 | 无 | 自然语言文本 | 图片提示词 |
| `n` | `integer` | 否 | `1` | 正整数 | 生成图片数量；生产接入建议从 `1` 开始 |
| `size` | `string` | 是 | 无 | 当前推荐 `2048x2048` | 图片尺寸 |
| `quality` | `string` | 否 | 上游默认 | 上游支持值 | 质量参数，当前两个 Seedream 模型通常不需要填写 |
| `response_format` | `string` | 否 | `url` | `url`、`b64_json` | 返回 URL 或 Base64 |
| `style` | `any` | 否 | 无 | 上游支持值 | 风格参数，透传字段 |
| `user` | `any` | 否 | 无 | 字符串或对象 | 用户标识透传 |
| `extra_fields` | `any` | 否 | 无 | 对象 | 上游扩展字段 |
| `background` | `any` | 否 | 无 | 上游支持值 | 背景参数 |
| `moderation` | `any` | 否 | 无 | 上游支持值 | 审核参数 |
| `output_format` | `any` | 否 | 上游默认 | `png`、`jpeg` 等 | 输出格式，按上游支持值生效 |
| `output_compression` | `any` | 否 | 上游默认 | 数值 | 输出压缩参数 |
| `partial_images` | `any` | 否 | 无 | 上游支持值 | 局部图片参数 |
| `watermark` | `boolean` | 否 | 上游默认 | `true`、`false` | 是否添加水印 |
| `watermark_enabled` | `any` | 否 | 无 | 上游兼容字段 | 兼容部分上游 |
| `user_id` | `any` | 否 | 无 | 字符串或数字 | 上游用户 ID 透传 |
| `image` | `any` | 否 | 无 | URL、Base64 或数组 | 图片输入兼容字段 |

### 8.3 `size` 标准填法

图片接口的 `size` 字段会透传到火山图片生成接口。当前站点生产接入推荐使用明确像素值，优先使用已验证的 `2048x2048`。

| 写法 | 示例 | 适用模型 | 当前文档口径 | 说明 |
| --- | --- | --- | --- | --- |
| 明确像素值 | `2048x2048` | `seedream4.5`、`seedream5.0lite` | 已线上验证通过 | 最推荐的生产写法 |
| 档位写法 | `2K` | 按上游模型能力 | 网关可透传 | 新接入前需先在测试环境验证 |
| 档位写法 | `3K` | 主要用于支持 3K 的模型 | 网关可透传 | `seedream5.0lite` 可按上游能力评估 |
| 档位写法 | `4K` | 按上游模型能力 | 网关可透传 | 成本和耗时通常高于 2K |
| 宽高像素值 | `2048x1152`、`1152x2048` | 按上游模型能力 | 网关可透传 | 用于横版或竖版图片 |
| 正方形像素值 | `2048x2048`、`3072x3072`、`4096x4096` | 按上游模型能力 | `2048x2048` 已验证 | 更高像素需按价格页和测试结果确认 |

填写规则：

| 规则 | 说明 |
| --- | --- |
| 推荐首选 | `2048x2048` |
| 不推荐 | `1024x1024`，会失败 |
| 横竖版 | 使用 `宽x高`，例如 `2048x1152` 或 `1152x2048` |
| 档位值 | 使用大写 `K`，例如 `2K`、`3K`、`4K` |
| 生产上线 | 新尺寸上线前必须先真实生成并核对扣费 |

### 8.4 成功响应

```json
{
  "created": 1780641005,
  "data": [
    {
      "url": "https://example.com/result.png",
      "b64_json": "",
      "revised_prompt": ""
    }
  ],
  "metadata": {}
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `created` | `integer` | 创建时间戳 |
| `data` | `array<object>` | 图片结果数组 |
| `data[].url` | `string` | 图片 URL，`response_format=url` 时优先使用 |
| `data[].b64_json` | `string` | Base64 图片，`response_format=b64_json` 时使用 |
| `data[].revised_prompt` | `string` | 上游改写后的提示词，存在时返回 |
| `metadata` | `object` | 扩展信息，存在时返回 |

### 8.5 `seedream4.5`

| 项目 | 值 |
| --- | --- |
| 请求接口 | `POST /v1/images/generations` |
| 当前推荐尺寸 | `2048x2048` |
| 当前已验证价格 | `0.3 CNY` |
| 当前非法尺寸示例 | `1024x1024` |
| 网关可透传尺寸 | 明确像素值、`2K`、`4K` |

请求示例：

```json
{
  "model": "seedream4.5",
  "prompt": "一张高端商业人像，柔和布光，真实皮肤质感，浅景深",
  "n": 1,
  "size": "2048x2048",
  "response_format": "url",
  "watermark": false
}
```

### 8.6 `seedream5.0lite`

| 项目 | 值 |
| --- | --- |
| 请求接口 | `POST /v1/images/generations` |
| 当前推荐尺寸 | `2048x2048` |
| 当前已验证价格 | `0.25 CNY` |
| 当前非法尺寸示例 | `1024x1024` |
| 网关可透传尺寸 | 明确像素值、`2K`、`3K`、`4K` |

请求示例：

```json
{
  "model": "seedream5.0lite",
  "prompt": "服装海报，干净背景，商品细节清晰，商业摄影风格",
  "n": 1,
  "size": "2048x2048",
  "response_format": "url",
  "output_format": "png",
  "watermark": false
}
```

---

## 9. 真人资产接口

真人资产用于创建和管理可绑定到视频生成任务中的真人资产。推荐使用 `/api/portrait_assets/official/*` 官方真人资产接口。

### 9.1 真人资产状态

| 状态 | 含义 |
| --- | --- |
| `pending` | 任务已创建 |
| `validate_ready` | 已生成实名校验链接 |
| `validated` | 实名校验通过，已拿到资产组 |
| `asset_processing` | 素材入库处理中 |
| `qr_ready` | 历史 RPA 队列二维码已准备 |
| `waiting_upload` | 历史 RPA 等待上传 |
| `waiting_accept` | 历史 RPA 等待受理 |
| `pending_confirm` | 素材已入库，等待用户确认预览 |
| `ready` | 资产已可用 |
| `failed` | 失败 |
| `disabled` | 已禁用 |
| `expired` | 实名链接或二维码已过期 |

### 9.2 真人资产对象

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `integer` | 真人资产任务 ID |
| `user_id` | `integer` | 用户 ID |
| `name` | `string` | 资产名称 |
| `source` | `string` | 来源，官方接口为 `official` |
| `status` | `string` | 真人资产状态 |
| `invite_url` | `string` | 实名校验链接 |
| `qr_image` | `string` | 二维码图片内容或地址 |
| `validate_result_code` | `string` | 实名校验结果码，`10000` 表示通过 |
| `volc_group_id` | `string` | 上游资产组 ID |
| `asset_id` | `string` | 最终真人资产 ID |
| `asset_status` | `string` | 上游资产状态，例如 `Processing`、`Active`、`Failed` |
| `asset_preview` | `string` | 安全预览地址 |
| `asset_url` | `string` | 提交的素材 URL |
| `asset_type` | `string` | `Image`、`Video`、`Audio` |
| `project_name` | `string` | 上游项目名 |
| `error_message` | `string` | 错误信息 |
| `created_time` | `integer` | 创建时间 |
| `updated_time` | `integer` | 更新时间 |
| `accept_time` | `integer` | 受理时间 |
| `qr_expires_time` | `integer` | 链接或二维码过期时间 |
| `ready_time` | `integer` | 就绪时间 |
| `queue_position` | `integer` | 历史 RPA 队列位置 |

### 9.3 查询真人资产配置

```http
GET /api/portrait_assets/official/config
```

响应字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `data.configured` | `boolean` | 官方真人资产能力是否已配置 |
| `data.project_name` | `string` | 当前火山项目名 |

示例：

```json
{
  "success": true,
  "message": "",
  "data": {
    "configured": true,
    "project_name": "portrait-project"
  }
}
```

### 9.4 查询真人资产任务列表

```http
GET /api/portrait_assets/official/jobs
```

查询参数见通用分页结构。

响应：

```json
{
  "success": true,
  "message": "",
  "data": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "items": [
      {
        "id": 12,
        "name": "主播真人资产",
        "source": "official",
        "status": "ready",
        "asset_id": "asset_xxx",
        "asset_preview": "https://8liangai.com/api/portrait_assets/official/jobs/12/preview/state",
        "asset_type": "Video",
        "ready_time": 1780641005
      }
    ]
  }
}
```

### 9.5 创建真人资产任务

```http
POST /api/portrait_assets/official/jobs
Content-Type: application/json
```

请求字段：

| 字段 | 类型 | 必填 | 限制 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `string` | 否 | 最长 50 个字符 | 资产名称；为空时系统使用默认名称 |

请求示例：

```json
{
  "name": "主播真人资产"
}
```

成功后通常返回真人资产对象，并包含 `invite_url`。用户需要打开 `invite_url` 完成真人实名校验。

### 9.6 刷新真人实名校验链接

```http
POST /api/portrait_assets/official/jobs/{id}/validation
```

路径参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` | 是 | 真人资产任务 ID |

使用规则：

| 条件 | 行为 |
| --- | --- |
| 任务尚未进入素材处理 | 可重新生成实名链接 |
| 状态为 `asset_processing`、`pending_confirm`、`ready` | 不允许重新生成 |

### 9.7 上传真人资产素材

```http
POST /api/portrait_assets/official/upload
Content-Type: multipart/form-data
```

表单字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `file` | `file` | 是 | 图片、视频或音频素材 |

文件限制：

| 类型 | 扩展名 | MIME |
| --- | --- | --- |
| 图片 | `.jpg`、`.jpeg`、`.png`、`.webp`、`.gif`、`.bmp` | `image/jpeg`、`image/png`、`image/webp`、`image/gif`、`image/bmp` |
| 视频 | `.mp4`、`.mov`、`.webm` | `video/mp4`、`video/quicktime`、`video/webm` |
| 音频 | `.mp3`、`.wav`、`.m4a`、`.ogg`、`.aac`、`.flac` | `audio/mpeg`、`audio/wav`、`audio/mp4`、`audio/ogg`、`audio/aac`、`audio/flac` |

最大文件大小：`100MB`

响应字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `data.url` | `string` | 上传后可公开访问的素材 URL |
| `data.file_name` | `string` | 原始文件名 |
| `data.content_type` | `string` | 识别出的 MIME |
| `data.asset_type` | `string` | `Image`、`Video`、`Audio` |
| `data.size` | `integer` | 文件大小，单位字节 |

示例：

```bash
curl -X POST 'https://8liangai.com/api/portrait_assets/official/upload' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -F 'file=@demo.mp4'
```

### 9.8 提交素材到真人资产任务

```http
POST /api/portrait_assets/official/jobs/{id}/asset
Content-Type: application/json
```

路径参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` | 是 | 真人资产任务 ID |

请求字段：

| 字段 | 类型 | 必填 | 可选值或格式 | 说明 |
| --- | --- | --- | --- | --- |
| `asset_url` | `string` | 是 | `http` 或 `https` URL | 必须可被火山访问 |
| `asset_type` | `string` | 是 | `Image`、`Video`、`Audio` | 素材类型 |
| `name` | `string` | 否 | 最长 50 个字符 | 素材名称 |

请求示例：

```json
{
  "asset_url": "https://8liangai.com/uploads/portrait/20260608/demo.mp4",
  "asset_type": "Video",
  "name": "主播口播素材"
}
```

前置条件：

| 条件 | 说明 |
| --- | --- |
| 任务已实名通过 | 状态应为 `validated` |
| 已有资产组 | `volc_group_id` 非空 |
| 素材可公网访问 | `asset_url` 必须是上游可访问的 HTTP URL |

### 9.9 同步真人资产状态

```http
POST /api/portrait_assets/official/jobs/{id}/sync
```

用途：

| 用途 | 说明 |
| --- | --- |
| 同步实名结果 | 当回调已完成但本地未刷新时使用 |
| 同步素材入库状态 | 将 `asset_processing` 推进到 `pending_confirm` 或 `failed` |
| 更新预览地址 | 刷新 `asset_preview` |

### 9.10 确认真人资产

```http
POST /api/portrait_assets/official/jobs/{id}/confirm
```

当状态为 `pending_confirm` 且预览无误时调用。成功后状态变为 `ready`，该资产可用于视频生成。

### 9.11 拒绝真人资产

```http
POST /api/portrait_assets/official/jobs/{id}/reject
```

当状态为 `pending_confirm` 且预览不符合要求时调用。成功后状态变为 `failed`。

### 9.12 真人资产预览

```http
GET /api/portrait_assets/official/jobs/{id}/preview/{state}
```

该接口是安全预览跳转地址。正常接入不需要自行拼接 `state`，直接使用列表或详情返回的 `asset_preview`。

成功时返回 `302` 跳转到真实预览资源。

### 9.13 真人资产绑定视频

绑定方式：

| 字段 | 填法 | 说明 |
| --- | --- | --- |
| `metadata.portrait_asset_id` | 真人资产任务 ID，例如 `12` | 推荐方式 |
| `metadata.asset_id` | 最终资产 ID，例如 `asset_xxx` 或 `asset://asset_xxx` | 兼容方式 |

绑定要求：

| 要求 | 说明 |
| --- | --- |
| 资产属于当前 API Key 对应用户 | 不能引用其他用户资产 |
| 资产状态为 `ready` | 未就绪资产会被拒绝 |
| `asset_id` 非空 | 任务必须已有最终上游资产 ID |

---

## 10. 虚拟人像资产接口

虚拟人像资产用于管理虚拟素材组和素材资产。当前公开接口支持配置查询、资产组查询、列表、上传、创建、同步、预览，以及在 Seedance 视频生成中作为参考资产使用。

虚拟人像资产的视频生成绑定方式是 `asset://<volc_asset_id>`。创建并同步到 `active` 后，取虚拟资产对象中的 `volc_asset_id`，拼成 `asset://asset_xxx`，放入视频生成请求的 `metadata.content` 中。

注意：虚拟人像资产不要使用 `metadata.portrait_asset_id`。`metadata.portrait_asset_id` 是真人资产任务 ID 的绑定字段。

### 10.1 虚拟资产组状态

| 状态 | 含义 |
| --- | --- |
| `creating` | 资产组创建中 |
| `active` | 资产组可用 |
| `failed` | 资产组创建失败 |

### 10.2 虚拟资产状态

| 状态 | 含义 |
| --- | --- |
| `processing` | 资产处理中 |
| `active` | 资产可用 |
| `failed` | 资产失败 |

### 10.3 虚拟资产组对象

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `integer` | 本地资产组 ID |
| `user_id` | `integer` | 用户 ID |
| `name` | `string` | 资产组名称 |
| `description` | `string` | 资产组描述 |
| `project_name` | `string` | 上游项目名 |
| `volc_group_id` | `string` | 上游资产组 ID |
| `status` | `string` | `creating`、`active`、`failed` |
| `error_message` | `string` | 错误信息 |
| `created_time` | `integer` | 创建时间 |
| `updated_time` | `integer` | 更新时间 |

### 10.4 虚拟资产对象

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `integer` | 本地资产 ID |
| `user_id` | `integer` | 用户 ID |
| `group_id` | `integer` | 本地资产组 ID |
| `name` | `string` | 资产名称 |
| `asset_type` | `string` | `Image`、`Video`、`Audio` |
| `source_url` | `string` | 原始素材 URL |
| `preview_url` | `string` | 安全预览地址 |
| `project_name` | `string` | 上游项目名 |
| `volc_group_id` | `string` | 上游资产组 ID |
| `volc_asset_id` | `string` | 上游资产 ID |
| `status` | `string` | `processing`、`active`、`failed` |
| `volc_status` | `string` | 上游状态 |
| `error_message` | `string` | 错误信息 |
| `created_time` | `integer` | 创建时间 |
| `updated_time` | `integer` | 更新时间 |
| `ready_time` | `integer` | 就绪时间 |

### 10.5 查询虚拟人像配置

```http
GET /api/portrait_assets/virtual/config
```

响应字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `data.configured` | `boolean` | 虚拟人像资产能力是否已配置 |
| `data.project_name` | `string` | 当前火山项目名 |

### 10.6 查询当前用户虚拟资产组

```http
GET /api/portrait_assets/virtual/group
```

响应：

| 情况 | 响应 |
| --- | --- |
| 已有资产组 | `data` 为虚拟资产组对象 |
| 尚未创建资产组 | `data` 为 `null` |

### 10.7 查询虚拟资产列表

```http
GET /api/portrait_assets/virtual/assets
```

查询参数见通用分页结构。

响应示例：

```json
{
  "success": true,
  "message": "",
  "data": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "items": [
      {
        "id": 31,
        "name": "虚拟口播素材",
        "asset_type": "Video",
        "source_url": "https://example.com/virtual.mp4",
        "preview_url": "https://8liangai.com/api/portrait_assets/virtual/assets/31/preview/state",
        "volc_asset_id": "asset_xxx",
        "status": "active",
        "volc_status": "Active"
      }
    ]
  }
}
```

### 10.8 上传虚拟资产素材

```http
POST /api/portrait_assets/virtual/upload
Content-Type: multipart/form-data
```

该接口复用真人资产上传能力。

表单字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `file` | `file` | 是 | 图片、视频或音频素材 |

限制：

| 项目 | 值 |
| --- | --- |
| 最大文件大小 | `100MB` |
| 图片格式 | `.jpg`、`.jpeg`、`.png`、`.webp`、`.gif`、`.bmp` |
| 视频格式 | `.mp4`、`.mov`、`.webm` |
| 音频格式 | `.mp3`、`.wav`、`.m4a`、`.ogg`、`.aac`、`.flac` |

响应字段与真人资产上传相同。

### 10.9 创建虚拟资产

```http
POST /api/portrait_assets/virtual/assets
Content-Type: application/json
```

请求字段：

| 字段 | 类型 | 必填 | 可选值或格式 | 说明 |
| --- | --- | --- | --- | --- |
| `name` | `string` | 否 | 最长 50 个字符 | 资产名称；为空时使用默认名称 |
| `asset_url` | `string` | 是 | `http` 或 `https` URL | 必须可被火山访问 |
| `asset_type` | `string` | 是 | `Image`、`Video`、`Audio`，大小写不敏感 | 素材类型 |

请求示例：

```json
{
  "name": "虚拟口播素材",
  "asset_url": "https://8liangai.com/uploads/portrait/20260608/virtual.mp4",
  "asset_type": "Video"
}
```

创建行为：

| 场景 | 行为 |
| --- | --- |
| 用户没有虚拟资产组 | 系统自动创建用户专属虚拟资产组 |
| 资产组正在初始化 | 返回稍后再试 |
| 素材提交成功 | 返回虚拟资产对象，初始通常为 `processing` |
| 上游很快完成 | 返回时可能已经是 `active` |

### 10.10 同步虚拟资产状态

```http
POST /api/portrait_assets/virtual/assets/{id}/sync
```

路径参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | `integer` | 是 | 虚拟资产 ID |

用途：

| 用途 | 说明 |
| --- | --- |
| 拉取上游状态 | 更新 `volc_status` |
| 刷新预览地址 | 更新 `preview_url` |
| 推进本地状态 | `processing` 变为 `active` 或 `failed` |

### 10.11 虚拟资产预览

```http
GET /api/portrait_assets/virtual/assets/{id}/preview/{state}
```

该接口是安全预览跳转地址。正常接入不需要自行拼接 `state`，直接使用列表返回的 `preview_url`。

成功时返回 `302` 跳转到真实预览资源。

### 10.12 虚拟资产用于视频生成

前置条件：

| 条件 | 要求 |
| --- | --- |
| 虚拟资产状态 | `status=active` |
| 上游资产 ID | `volc_asset_id` 非空 |
| 资产 URI | 使用 `asset://<volc_asset_id>` |
| 视频模型 | 推荐 `seedance2`、`seedance2-sr` 或其他支持资产引用的 Seedance 模型 |

字段对应关系：

| 虚拟资产类型 | 视频 `metadata.content` 写法 | 说明 |
| --- | --- | --- |
| `Image` | `type=image_url`，`role=reference_image`，`image_url.url=asset://<volc_asset_id>` | 作为形象或图片参考 |
| `Video` | `type=video_url`，`role=reference_video`，`video_url.url=asset://<volc_asset_id>` | 作为视频参考 |
| `Audio` | `type=audio_url`，`audio_url.url=asset://<volc_asset_id>` | 作为音频参考，是否生效取决于模型能力 |

图片虚拟资产示例：

```json
{
  "model": "seedance2",
  "prompt": "参考虚拟人像资产中的人物形象，让人物自然挥手并看向镜头",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "watermark": false,
    "content": [
      {
        "type": "image_url",
        "role": "reference_image",
        "image_url": {
          "url": "asset://asset_xxx"
        }
      }
    ]
  }
}
```

视频虚拟资产示例：

```json
{
  "model": "seedance2",
  "prompt": "参考虚拟人像视频资产，让人物保持身份一致并做自然动作",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "watermark": false,
    "content": [
      {
        "type": "video_url",
        "role": "reference_video",
        "video_url": {
          "url": "asset://asset_xxx"
        }
      }
    ]
  }
}
```

---

## 11. 兼容接口说明

### Remix 兼容接口

```http
POST /v1/videos/{video_id}/remix
```

该接口用于基于已有视频任务发起 Remix，当前文档不作为主流程展开。使用时 `video_id` 必须属于当前用户，且系统会尽量复用原任务模型和通道。

---

## 12. 网站使用教程

### 12.1 API Key 创建

页面入口：

```text
/keys
```

操作步骤：

1. 登录 `https://8liangai.com`。
2. 进入控制台的 API Keys 或密钥页面。
3. 点击创建 API Key。
4. 填写密钥名称、额度、分组、过期时间等字段。
5. 创建成功后立即保存完整 `sk-...` Key。

常见页面字段：

| 字段 | 建议填法 | 说明 |
| --- | --- | --- |
| 名称 | `video-prod-key`、`image-test-key` | 用于区分用途 |
| 分组 | `default` | 本文档覆盖模型建议使用默认分组 |
| 额度 | 按业务预算填写 | 用于限制该 Key 可消耗金额 |
| 不限额 | 生产内部服务按需开启 | 开启后不受该 Key 单独额度限制 |
| 过期时间 | 测试 Key 建议设置，生产 Key 按安全策略设置 | 到期后 Key 不可用 |
| 跨分组重试 | 通常关闭 | 仅在高级路由场景使用 |

安全建议：

| 建议 | 说明 |
| --- | --- |
| 测试和生产分开 Key | 便于查账和限额 |
| 不要把 Key 写入前端代码 | 应由服务端调用 API |
| 泄露后立即删除旧 Key | 重新创建新 Key |
| 每个业务线单独 Key | 便于按业务统计成本 |

### 12.2 真人资产创建流程

页面入口：

```text
/portrait-assets-official
```

完整流程：

1. 创建真人资产任务，填写资产名称。
2. 打开系统返回的 `invite_url`。
3. 对应真人完成实名校验。
4. 调用同步接口或等待页面刷新，状态进入 `validated`。
5. 上传素材文件，获取 `asset_url` 和 `asset_type`。
6. 调用提交素材接口，把素材提交到真人资产任务。
7. 任务进入 `asset_processing`。
8. 调用同步接口，直到状态进入 `pending_confirm`。
9. 打开 `asset_preview` 查看预览。
10. 预览无误后调用确认接口。
11. 状态进入 `ready` 后，即可在视频接口绑定。

成功可用标准：

| 条件 | 要求 |
| --- | --- |
| `status` | 必须为 `ready` |
| `asset_id` | 必须非空 |
| 所属用户 | 必须为当前 API Key 对应用户 |

绑定示例：

```json
{
  "model": "seedance2",
  "prompt": "人物自然挥手并看向镜头",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "portrait_asset_id": 12,
    "watermark": false
  }
}
```

### 12.3 虚拟人像资产创建流程

页面入口：

```text
/portrait-assets-virtual
```

完整流程：

1. 进入虚拟人像资产页面。
2. 查看配置状态，确保 `configured=true`。
3. 上传素材文件，获取 `asset_url` 和 `asset_type`。
4. 填写资产名称。
5. 创建虚拟资产。
6. 首次创建时系统会自动创建用户专属虚拟资产组。
7. 资产初始通常为 `processing`。
8. 调用同步接口或等待页面刷新。
9. 状态进入 `active` 后，资产创建完成。
10. 使用 `preview_url` 查看预览。
11. 记录 `volc_asset_id`。
12. 视频生成时将 `volc_asset_id` 拼成 `asset://<volc_asset_id>`，放入 `metadata.content`。

素材类型建议：

| 类型 | 用途 |
| --- | --- |
| `Image` | 角色形象、脸部参考、头像素材 |
| `Video` | 角色动作、口播、动态参考素材 |
| `Audio` | 声音或音频参考素材 |

虚拟人像资产接入视频示例：

```json
{
  "model": "seedance2",
  "prompt": "参考虚拟人像资产中的人物形象，生成自然口播动作",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "watermark": false,
    "content": [
      {
        "type": "image_url",
        "role": "reference_image",
        "image_url": {
          "url": "asset://asset_xxx"
        }
      }
    ]
  }
}
```

---

## 13. 接入检查清单

### 13.1 视频上线前检查

| 检查项 | 要求 |
| --- | --- |
| 模型名 | 必须在 `/v1/models` 可见或后台已配置映射 |
| 分辨率 | 必须符合模型能力表 |
| 比例 | 生产建议显式传 `ratio` |
| 时长 | 建议显式传 `duration` |
| 水印 | 如需无水印，显式传 `metadata.watermark=false` |
| 状态轮询 | 同时兼容 `queued`、`in_progress`、`completed`、`failed` |
| 下载 | 仅在任务完成后调用 `/content` |
| 计费 | 按 `actual_quota / 500000` 核对金额 |
| SR 模型 | 前端展示用户请求的 `720p` 或 `1080p` |

### 13.2 图片上线前检查

| 检查项 | 要求 |
| --- | --- |
| 模型名 | 使用 `seedream4.5` 或 `seedream5.0lite` |
| 尺寸 | 当前推荐 `2048x2048` |
| 数量 | 首版建议 `n=1` |
| 返回格式 | 前端展示建议 `response_format=url` |
| 失败扣费 | 参数非法不应扣费 |

### 13.3 资产上线前检查

| 检查项 | 要求 |
| --- | --- |
| 真人资产配置 | `official/config` 返回 `configured=true` |
| 真人实名 | 状态进入 `validated` 后再提交素材 |
| 真人可用 | `status=ready` 且 `asset_id` 非空 |
| 虚拟资产配置 | `virtual/config` 返回 `configured=true` |
| 虚拟资产可用 | `status=active` |
| 预览地址 | 使用接口返回的 `asset_preview` 或 `preview_url` |

---

## 14. 附录：端到端视频调用流程

### 14.1 提交任务

```bash
curl -X POST 'https://8liangai.com/v1/video/generations' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "seedance2-sr",
    "prompt": "高清品牌短片，人物自然转身，镜头稳定",
    "resolution": "720p",
    "ratio": "16:9",
    "duration": 5,
    "metadata": {
      "watermark": false
    }
  }'
```

### 14.2 轮询状态

```bash
curl 'https://8liangai.com/v1/videos/task_xxx' \
  -H 'Authorization: Bearer sk-你的APIKey'
```

当 `status=completed` 时进入下载步骤。

### 14.3 下载视频

```bash
curl -L 'https://8liangai.com/v1/videos/task_xxx/content' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -o result.mp4
```

### 14.4 推荐轮询策略

| 阶段 | 建议 |
| --- | --- |
| 提交后 0 到 30 秒 | 每 5 到 10 秒查询一次 |
| 30 秒后 | 每 10 到 15 秒查询一次 |
| 超时时间 | 建议业务侧设置 20 分钟 |
| 失败处理 | 记录 `error`、`fail_reason`、完整响应体 |
