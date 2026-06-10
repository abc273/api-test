# 8liangai Vidu 标准接口文档

更新时间：2026-06-09

适用域名：`https://8liangai.com`

适用模型：

- `vidu-q1`
- `vidu-q2`
- `vidu-q3-turbo`
- `viduq-3`

## 1. 鉴权

所有接口统一使用 Bearer Token：

```http
Authorization: Bearer sk-你的APIKey
Content-Type: application/json
Accept: application/json
```

## 2. 模型与推荐规格

| 模型              | 接口                           | 推荐规格                           |
| --------------- | ---------------------------- | ------------------------------ |
| `vidu-q1`       | `POST /v1/video/generations` | `duration=5`、`size=1080p`      |
| `vidu-q2`       | `POST /v1/video/generations` | `duration=5`、`size=720p/1080p` |
| `vidu-q3-turbo` | `POST /v1/video/generations` | `duration=5`、`size=720p/1080p` |
| `viduq-3`       | `POST /v1/video/generations` | `duration=5`、`size=720p/1080p` |

## 3. 接口清单

| 能力              | 方法     | 路径                                |
| --------------- | ------ | --------------------------------- |
| 模型列表            | `GET`  | `/v1/models`                      |
| 视频任务提交          | `POST` | `/v1/video/generations`           |
| 视频任务查询          | `GET`  | `/v1/video/generations/{task_id}` |
| OpenAI 兼容视频结果查询 | `GET`  | `/v1/videos/{task_id}`            |
| 视频文件下载          | `GET`  | `/v1/videos/{task_id}/content`    |

## 4. 查询模型列表

### 请求

```http
GET /v1/models
Authorization: Bearer sk-你的APIKey
Accept: application/json
```

### 响应示例

```json
{
  "object": "list",
  "success": true,
  "data": [
    { "id": "vidu-q1", "object": "model", "owned_by": "custom" },
    { "id": "vidu-q2", "object": "model", "owned_by": "custom" },
    { "id": "vidu-q3-turbo", "object": "model", "owned_by": "custom" },
    { "id": "viduq-3", "object": "model", "owned_by": "custom" }
  ]
}
```

## 5. 视频接口

适用模型：

- `vidu-q1`
- `vidu-q2`
- `vidu-q3-turbo`
- `viduq-3`

### 5.1 提交地址

```http
POST /v1/video/generations
```

### 5.2 支持模式

| 模式     | 触发方式                                                    | 上游能力              |
| ------ | ------------------------------------------------------- | ----------------- |
| 文生视频   | 不传 `image` / `images`                                   | `text2video`      |
| 单图生视频  | 传 `image` 或单个 `images`                                  | `img2video`       |
| 首尾帧生视频 | 传 2 张 `images`，或 `metadata.action=firstTailGenerate`    | `start-end2video` |
| 参考图生视频 | 传 3 张及以上 `images`，或 `metadata.action=referenceGenerate` | `reference2video` |

说明：

- 网关会自动把单个 `image` 兼容转换为 `images[0]`。
- 建议显式传入 `size` 为 `720p` 或 `1080p`。
- 如未显式传入，当前适配器默认 `duration=5`、`size=1080p`。

### 5.3 请求体字段

| 字段                            | 类型         | 必填     | 说明                                                                   |
| ----------------------------- | ---------- | ------ | -------------------------------------------------------------------- |
| `model`                       | `string`   | 是      | `vidu-q1`、`vidu-q2`、`vidu-q3-turbo` 或 `viduq-3`                      |
| `prompt`                      | `string`   | 文生视频必填 | 视频提示词                                                                |
| `image`                       | `string`   | 否      | 单图输入，支持 URL 或 Base64                                                 |
| `images`                      | `string[]` | 否      | 多图输入                                                                 |
| `duration`                    | `integer`  | 否      | 推荐 `5`                                                               |
| `size`                        | `string`   | 否      | 推荐 `720p` 或 `1080p`                                                  |
| `metadata.action`             | `string`   | 否      | 可选 `textGenerate`、`generate`、`firstTailGenerate`、`referenceGenerate` |
| `metadata.seed`               | `integer`  | 否      | 随机种子                                                                 |
| `metadata.watermark`          | `boolean`  | 否      | 水印开关                                                                 |
| `metadata.callback_url`       | `string`   | 否      | 任务完成回调地址                                                             |
| `metadata.movement_amplitude` | `string`   | 否      | 运动幅度，常用 `auto`                                                       |
| `metadata.bgm`                | `boolean`  | 否      | 是否带背景音乐                                                              |

### 5.4 文生视频示例

#### `vidu-q1` 1080p

```json
{
  "model": "vidu-q1",
  "prompt": "A glass hummingbird hovers over a quiet garden at sunrise, cinematic lighting, stable camera, no text",
  "duration": 5,
  "size": "1080p",
  "metadata": {
    "watermark": false
  }
}
```

#### `vidu-q2` 720p

```json
{
  "model": "vidu-q2",
  "prompt": "A silver koi fish swims through floating ink clouds, elegant motion, cinematic composition, no text",
  "duration": 5,
  "size": "720p",
  "metadata": {
    "watermark": false
  }
}
```

#### `vidu-q3-turbo` 720p

```json
{
  "model": "vidu-q3-turbo",
  "prompt": "A red paper airplane glides through warm sunset light, cinematic framing, stable camera, no text",
  "duration": 5,
  "size": "720p",
  "metadata": {
    "watermark": false
  }
}
```

#### `viduq-3` 1080p

```json
{
  "model": "viduq-3",
  "prompt": "An orange cat sits by a sunny window, cinematic lighting, stable camera, natural subtle motion, no text",
  "duration": 5,
  "size": "1080p",
  "metadata": {
    "watermark": false
  }
}
```

### 5.5 单图生视频示例

```json
{
  "model": "vidu-q3-turbo",
  "prompt": "Make the character slowly turn the head and blink while preserving the original style",
  "image": "https://example.com/input.png",
  "duration": 5,
  "size": "720p",
  "metadata": {
    "watermark": false
  }
}
```

### 5.6 首尾帧生视频示例

```json
{
  "model": "viduq-3",
  "prompt": "Transition smoothly from the start frame to the end frame with coherent motion and a stable camera",
  "images": [
    "https://example.com/start.png",
    "https://example.com/end.png"
  ],
  "duration": 5,
  "size": "1080p",
  "metadata": {
    "action": "firstTailGenerate",
    "watermark": false
  }
}
```

### 5.7 参考图生视频示例

```json
{
  "model": "viduq-3",
  "prompt": "Generate a unified video that follows the character, outfit, and camera style from the reference images",
  "images": [
    "https://example.com/ref-1.png",
    "https://example.com/ref-2.png",
    "https://example.com/ref-3.png"
  ],
  "duration": 5,
  "size": "1080p",
  "metadata": {
    "action": "referenceGenerate",
    "watermark": false
  }
}
```

### 5.8 提交成功响应

```json
{
  "id": "task_xxx",
  "task_id": "task_xxx",
  "object": "video",
  "model": "vidu-q3-turbo",
  "status": "queued",
  "progress": 0,
  "created_at": 1780000000
}
```

## 6. 查询视频任务

### 请求

```http
GET /v1/video/generations/{task_id}
Authorization: Bearer sk-你的APIKey
Accept: application/json
```

### 响应示例

```json
{
  "code": "success",
  "message": "",
  "data": {
    "task_id": "task_xxx",
    "status": "SUCCESS",
    "progress": "100%",
    "fail_reason": "",
    "properties": {
      "upstream_model_name": "viduq3-turbo",
      "origin_model_name": "vidu-q3-turbo"
    },
    "data": {
      "state": "success",
      "model": "viduq3-turbo",
      "resolution": "720p",
      "type": "text2video",
      "progress": 100,
      "creations": [
        {
          "url": "https://example.com/video.mp4"
        }
      ]
    }
  }
}
```

### 任务状态说明

站内状态 `data.status`：

| 值             | 含义           |
| ------------- | ------------ |
| `NOT_START`   | 任务已创建，尚未开始处理 |
| `SUBMITTED`   | 已提交上游        |
| `IN_PROGRESS` | 处理中          |
| `SUCCESS`     | 已成功完成        |
| `FAILURE`     | 处理失败         |

上游状态 `data.data.state`：

| 值            | 含义  |
| ------------ | --- |
| `created`    | 已创建 |
| `queueing`   | 排队中 |
| `processing` | 处理中 |
| `success`    | 成功  |
| `failed`     | 失败  |

## 7. OpenAI 兼容视频结果接口

### 请求

```http
GET /v1/videos/{task_id}
Authorization: Bearer sk-你的APIKey
Accept: application/json
```

### 响应示例

```json
{
  "id": "task_xxx",
  "object": "video",
  "model": "",
  "status": "completed",
  "progress": 100,
  "created_at": 1780000000,
  "completed_at": 1780000123,
  "metadata": {
    "url": "https://example.com/video.mp4"
  }
}
```

## 8. 下载视频文件

### 请求

```http
GET /v1/videos/{task_id}/content
Authorization: Bearer sk-你的APIKey
```

### curl 示例

```bash
curl -L 'https://8liangai.com/v1/videos/task_xxx/content' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -o result.mp4
```

## 9. 错误响应示例

```json
{
  "error": {
    "message": "openai_error",
    "type": "bad_response_status_code",
    "param": "",
    "code": "bad_response_status_code"
  }
}
```

## 10. 推荐接入顺序

1. 先调用 `GET /v1/models` 确认模型可见。
2. 调用 `POST /v1/video/generations` 提交任务。
3. 轮询 `GET /v1/video/generations/{task_id}` 查询任务状态。
4. 成功后使用 `GET /v1/videos/{task_id}` 取兼容结果，或直接调用 `GET /v1/videos/{task_id}/content` 下载视频。

## 11. 接入建议

1. `vidu-q1` 视频建议优先使用 `1080p / 5s`。
2. `vidu-q2`、`vidu-q3-turbo`、`viduq-3` 视频建议优先使用 `720p` 或 `1080p`。
3. 首次对接建议从最小参数集开始：`duration=5`。
4. 如需首尾帧或参考图视频，建议统一使用 `images` 数组传参，便于前后端接口保持一致。

