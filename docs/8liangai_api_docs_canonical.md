# 8liangAI 接口文档（合并标准稿）

> 用途：这是后台 `ApiDocs` 配置项首次导入用的标准 Markdown 底稿。  
> 来源：`8liangai接口文档v2.docx` 与 `Seedance` 视频超分用户文档。  
> 说明：前台 `/docs` 不自动读取本文件；管理员需要将本文件内容粘贴到后台“接口文档”配置中发布。

## 1. 文档说明

- 站点地址：`https://8liangai.com`
- 适用对象：调用 8liangAI 图片生成、视频生成、任务查询、视频取流、真实资产查询、虚拟资产查询接口的客户开发人员
- 当前确认可用模型：
  - 图片模型：`seedream4.5`、`seedream5.0lite`
  - 视频模型：`seedance2`、`sd2.0fast`、`seedance2-sr`、`sd2.0fast-sr`

## 2. 接入前必读

### 2.1 Base URL

```text
https://8liangai.com
```

### 2.2 认证方式

除公开预览地址外，本文大部分接口都需要 API Key 或登录态。

推荐使用 API Key 调用，在请求头中传：

```http
Authorization: Bearer sk-你的_API_Key
```

说明：

- `Authorization` 必传
- `Bearer` 与密钥之间有一个空格
- 不要把 Key 放在 URL 查询参数中

### 2.3 Content-Type 规则

| 场景 | Content-Type | 说明 |
| --- | --- | --- |
| JSON 请求 | `application/json` | 图片生成、视频生成、任务查询等都用这个 |
| 文件上传 | `multipart/form-data` | 真实资产、虚拟资产上传素材时使用 |

### 2.4 常见鉴权失败返回

如果未携带有效 API Key，或当前登录态无效，受保护接口通常会优先返回 `401 Unauthorized`。

```json
{
  "error": {
    "code": "",
    "message": "Invalid token (request id: xxxxx)",
    "type": "new_api_error"
  }
}
```

建议联调时优先排查：

1. `Authorization` 是否正确传入
2. `Bearer` 前缀是否带空格
3. API Key 是否属于当前环境、当前用户、当前分组
4. 是否把需要登录态的接口误当成公开接口调用

## 3. 常用模型与接口对应

| 模型别名 | 类型 | 调用接口 | 推荐用途 |
| --- | --- | --- | --- |
| `seedream4.5` | 图片生成 | `POST /v1/images/generations` | 高质量图片生成 |
| `seedream5.0lite` | 图片生成 | `POST /v1/images/generations` | 更轻量、更快的图片生成 |
| `seedance2` | 视频生成 | `POST /v1/video/generations` | 标准视频生成 |
| `sd2.0fast` | 视频生成 | `POST /v1/video/generations` | 更快的视频生成 |
| `seedance2-sr` | 视频生成 + 自动超分 | `POST /v1/video/generations` | `seedance2` 专用超分版，适合最终要 `1080p` |
| `sd2.0fast-sr` | 视频生成 + 自动超分 | `POST /v1/video/generations` | `sd2.0fast` 专用超分版，适合最终要 `720p` |

## 4. 返回格式说明

本项目不是所有接口都返回同一种 JSON 结构，请特别注意：

1. `POST /v1/images/generations`  
   直接返回图片结果对象，风格接近 OpenAI 图片接口。
2. `POST /v1/video/generations`  
   直接返回视频任务对象，包含 `task_id`。
3. `GET /v1/video/generations/{task_id}`  
   返回包装结构：`code`、`message`、`data`。
4. `/api/portrait_assets/...`  
   返回包装结构：`success`、`message`、`data`。

接入时不要把所有接口都按同一种返回体解析。

## 5. 调用顺序建议

### 5.1 图片生成

1. 调用 `POST /v1/images/generations`
2. 直接从返回里的 `url` 或 `b64_json` 取图

说明：当前图片生成是同步返回结果，不需要额外轮询图片状态接口。

### 5.2 视频生成

1. 调用 `POST /v1/video/generations` 提交任务
2. 从返回中拿到 `task_id`
3. 调用 `GET /v1/video/generations/{task_id}` 查询任务状态
4. 任务成功后：
   - 可直接使用查询结果里的 `result_url`
   - 或调用 `GET /v1/videos/{task_id}/content` 取视频流

### 5.3 真实资产 / 虚拟资产使用

推荐顺序如下：

1. 先查询真实资产或虚拟资产列表
2. 找到可用资产对应的 `asset_id` / `volc_asset_id`
3. 在视频生成接口中通过 `metadata.portrait_asset_id` 或 `metadata.asset_id` 引用

## 6. 图片生成接口

### 6.1 生成图片

```text
POST /v1/images/generations
```

完整调用地址：

```text
https://8liangai.com/v1/images/generations
```

适用模型：

- `seedream4.5`
- `seedream5.0lite`

最小请求示例：

```json
{
  "model": "seedream4.5",
  "prompt": "一张写实风格的高端商业人像，柔和布光，皮肤纹理自然，细节丰富",
  "n": 1,
  "size": "2048x2048",
  "response_format": "url"
}
```

返回示例：

```json
{
  "created": 1779547000,
  "data": [
    {
      "url": "https://8liangai.com/generated/example-image.png",
      "b64_json": "",
      "revised_prompt": ""
    }
  ]
}
```

接入提示：

- 推荐显式传 `size`
- 推荐 `response_format` 使用 `url`
- 如需图生图，直接传 `image`
- `seedream4.5` 与 `seedream5.0lite` 都支持单图、多图参考

## 7. 视频生成接口

### 7.1 提交视频生成任务

```text
POST /v1/video/generations
```

完整调用地址：

```text
https://8liangai.com/v1/video/generations
```

适用模型：

- `seedance2`
- `sd2.0fast`
- `seedance2-sr`
- `sd2.0fast-sr`

基础请求示例：

```json
{
  "model": "seedance2",
  "prompt": "人物缓慢转头并向镜头挥手，写实电影感，光影自然",
  "duration": 5,
  "metadata": {
    "resolution": "1080p",
    "ratio": "16:9",
    "watermark": false
  }
}
```

参考图请求示例：

```json
{
  "model": "sd2.0fast-sr",
  "prompt": "基于参考图保持服装颜色和材质不变，人物缓慢向前走一步并微微转头",
  "duration": 5,
  "images": [
    "https://example.com/reference-image.jpg"
  ],
  "metadata": {
    "resolution": "720p",
    "ratio": "16:9",
    "watermark": false
  }
}
```

接入提示：

- 建议始终显式传 `metadata.resolution`
- `seedance2` 推荐优先传 `720p` 或 `1080p`
- `sd2.0fast` 推荐优先传 `480p` 或 `720p`，不要传 `1080p`
- 如需生成更可控，建议同时传 `duration`、`ratio`、`watermark`

## 8. 视频超分说明

### 8.1 功能规则

视频超分已改为专用模型触发，不再通过 `metadata.enable_video_super_resolution` 开关触发。

请直接使用 `seedance2-sr` 或 `sd2.0fast-sr`。调用这两个模型时，系统会在原始视频生成完成后自动继续做超分处理。

普通模型仍保持原样：

- `seedance2`：普通版，不自动超分
- `sd2.0fast`：普通版，不自动超分

### 8.2 分辨率提升规则

| 提交分辨率 | 最终返回分辨率 |
| --- | --- |
| `480p` | `720p` |
| `720p` | `1080p` |
| `1080p` | 仍为 `1080p`，不再继续超分 |

### 8.3 客户需要关注的行为

1. 超分属于付费增值能力，但现在由专用模型决定是否启用
2. 调用普通模型时，返回的是原始生成视频
3. 调用专用超分模型时，最终 `result_url` 返回的是超分后成品视频
4. 调用专用超分模型时，`GET /v1/videos/{task_id}/content` 返回的也是超分后成品视频
5. 专用超分模型整体完成时间通常会比普通视频任务更长
6. 建议始终显式传 `metadata.resolution`

### 8.4 推荐传参

想最终拿到 `720p`：

```json
{
  "model": "sd2.0fast",
  "prompt": "人物自然抬头并微笑",
  "duration": 5,
  "metadata": {
    "resolution": "480p",
    "ratio": "16:9",
    "watermark": false
  }
}
```

想最终拿到 `1080p`：

```json
{
  "model": "seedance2-sr",
  "prompt": "人物缓慢转头并向镜头挥手",
  "duration": 5,
  "metadata": {
    "resolution": "720p",
    "ratio": "16:9",
    "watermark": false
  }
}
```

### 8.5 超分 FAQ

**为什么任务比以前慢？**  
因为专用超分模型会在原始视频生成完成后继续执行一段超分处理。

**我拿到的是原始视频还是超分后视频？**  
- 普通模型：拿到原始生成视频  
- 专用超分模型：`result_url` 和 `/content` 默认都返回超分后视频

**为什么我传了 `1080p` 还没有继续变更高？**  
当前自动超分只处理 `480p -> 720p` 与 `720p -> 1080p`。

## 9. 资产接口使用建议

- 真实资产、虚拟资产建议先查列表，再在视频生成时引用
- 如果已经拿到平台资产 ID，优先通过 `metadata.asset_id` 或 `metadata.portrait_asset_id` 传入
- 资产需属于当前账号且处于可用状态

## 10. 联调建议

1. 图片接口按同步返回处理
2. 视频接口按异步任务处理
3. 视频接口联调时显式传 `metadata.resolution`
4. 需要超分时请直接改用 `seedance2-sr` 或 `sd2.0fast-sr`
5. 不需要超分时继续使用 `seedance2` 或 `sd2.0fast`
