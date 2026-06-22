# 8liangai `viduq-3` 接口文档

更新时间：2026-06-09

Base URL：`https://8liangai.com`

## 1. 当前可用性

截至 2026-06-09，`viduq-3` 已在 `https://8liangai.com` 线上可用。

本轮真实测试结论：

- `GET /v1/models` 可以看到 `viduq-3`
- `POST /v1/video/generations` 可以成功创建任务
- `GET /v1/video/generations/{task_id}` 可以正常轮询任务状态
- `GET /v1/videos/{task_id}` 可以返回 OpenAI 兼容格式结果
- `GET /v1/videos/{task_id}/content` 可以正常下载视频文件

<br />

## 2. 鉴权方式

所有接口都使用 Bearer Token：

```http
Authorization: Bearer sk-你的APIKey
```

## 3. 模型列表

### 请求

```http
GET /v1/models
Authorization: Bearer sk-你的APIKey
Accept: application/json
```

### 说明

如果 `viduq-3` 能在返回列表里看到，说明该模型已经对当前 token 所在分组暴露。

## 4. 推荐调用流程

推荐顺序：

1. 调用 `POST /v1/video/generations` 提交任务
2. 调用 `GET /v1/video/generations/{task_id}` 轮询状态
3. 任务完成后：
   - 调用 `GET /v1/videos/{task_id}` 获取 OpenAI 兼容结果
   - 或调用 `GET /v1/videos/{task_id}/content` 直接下载视频

## 5. 提交任务

### 请求

```http
POST /v1/video/generations
Authorization: Bearer sk-你的APIKey
Content-Type: application/json
Accept: application/json
```

### 推荐请求体

```json
{
  "model": "viduq-3",
  "prompt": "一只橘猫坐在窗边，阳光照进房间，电影感光影，镜头稳定，动作自然，无文字",
  "duration": 5,
  "size": "720p",
  "metadata": {
    "watermark": false
  }
}
```

### 字段说明

| 字段                            | 类型         | 必填 | 说明                   |
| ----------------------------- | ---------- | -- | -------------------- |
| `model`                       | `string`   | 是  | 固定填 `viduq-3`        |
| `prompt`                      | `string`   | 是  | 文生视频提示词              |
| `duration`                    | `int`      | 否  | 建议先用 `5`             |
| `size`                        | `string`   | 否  | 建议用 `720p` 或 `1080p` |
| `image`                       | `string`   | 否  | 单图输入，图生视频时可用         |
| `images`                      | `string[]` | 否  | 多图输入                 |
| `metadata.watermark`          | `bool`     | 否  | 水印开关                 |
| `metadata.seed`               | `int`      | 否  | 随机种子                 |
| `metadata.callback_url`       | `string`   | 否  | 回调地址                 |
| `metadata.movement_amplitude` | `string`   | 否  | 运动幅度                 |
| `metadata.bgm`                | `bool`     | 否  | 背景音乐开关               |

### 成功响应示例

```json
{
  "id": "task_fEa7iyJuGwrAnUhKUTOx1vcr1Qt23XOr",
  "task_id": "task_fEa7iyJuGwrAnUhKUTOx1vcr1Qt23XOr",
  "object": "video",
  "model": "viduq-3",
  "status": "queued",
  "progress": 0,
  "created_at": 1780973490
}
```

## 6. 轮询任务状态

### 请求

```http
GET /v1/video/generations/{task_id}
Authorization: Bearer sk-你的APIKey
Accept: application/json
```

### 成功响应示例

```json
{
  "code": "success",
  "message": "",
  "data": {
    "task_id": "task_fEa7iyJuGwrAnUhKUTOx1vcr1Qt23XOr",
    "status": "SUCCESS",
    "progress": "100%",
    "properties": {
      "upstream_model_name": "viduq3-pro",
      "origin_model_name": "viduq-3"
    },
    "data": {
      "state": "success",
      "model": "viduq3-pro",
      "resolution": "720p",
      "type": "text2video",
      "progress": 100,
      "creations": [
        {
          "url": "https://..."
        }
      ]
    }
  }
}
```

### 常见状态

| 字段                | 可能值                                                          | 说明               |
| ----------------- | ------------------------------------------------------------ | ---------------- |
| `data.status`     | `NOT_START`                                                  | 站内任务已创建，尚未真正提交上游 |
| `data.status`     | `SUBMITTED`                                                  | 已提交上游            |
| `data.status`     | `IN_PROGRESS`                                                | 正在处理             |
| `data.status`     | `SUCCESS`                                                    | 成功完成             |
| `data.status`     | `FAILURE`                                                    | 处理失败             |
| `data.data.state` | `created` / `queueing` / `processing` / `success` / `failed` | 上游任务状态           |

## 7. OpenAI 兼容结果接口

### 请求

```http
GET /v1/videos/{task_id}
Authorization: Bearer sk-你的APIKey
Accept: application/json
```

### 实测响应示例

```json
{
  "id": "task_fEa7iyJuGwrAnUhKUTOx1vcr1Qt23XOr",
  "object": "video",
  "model": "",
  "status": "completed",
  "progress": 100,
  "created_at": 1780973490,
  "completed_at": 1780973622,
  "metadata": {
    "url": "https://..."
  }
}
```

### 说明

- `metadata.url` 是最终视频地址
- 当前实测里这个接口的 `model` 字段为空字符串，这属于当前线上实际表现

## 8. 下载视频文件

### 请求

```http
GET /v1/videos/{task_id}/content
Authorization: Bearer sk-你的APIKey
```

### curl 示例

```bash
curl -L 'https://8liangai.com/v1/videos/task_fEa7iyJuGwrAnUhKUTOx1vcr1Qt23XOr/content' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -o result.mp4
```

### 本轮下载验证

- 下载成功
- 文件大小：`4699204` 字节
- 视频信息：`1280x720`、`24fps`、时长约 `5.041667s`

## 9. 当前已验证行为

本轮 2026-06-09 真实验证到的行为：

- 对外模型名：`viduq-3`
- 上游实际模型名：`viduq3-pro`
- 任务类型：`text2video`
- 请求 `size=720p`，最终输出为 `1280x720`
- 轮询接口与下载接口均可用

## 10. 已知注意事项

1. `viduq-3` 当前对外名和上游名不是同一个值。
   当前实测显示站内名是 `viduq-3`，上游实际提交的是 `viduq3-pro`。
2. 建议生产接入时使用 `size` 字段。
   当前这条链路实测是按 `size` 走通的。
3. `metadata.watermark=false` 在本轮提交中没有完全证明已按上游生效。
   任务能成功，但如果你对去水印要求严格，建议再单独做一次结果画面核验。

## 11. 最小可用示例

```bash
curl -X POST 'https://8liangai.com/v1/video/generations' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "viduq-3",
    "prompt": "A calm orange cat sitting by a sunny window, cinematic lighting, subtle natural movement, stable camera, no text",
    "duration": 5,
    "size": "720p",
    "metadata": {
      "watermark": false
    }
  }'
```

拿到 `task_id` 后继续查询：

```bash
curl 'https://8liangai.com/v1/video/generations/task_fEa7iyJuGwrAnUhKUTOx1vcr1Qt23XOr' \
  -H 'Authorization: Bearer sk-你的APIKey'
```

完成后下载：

```bash
curl -L 'https://8liangai.com/v1/videos/task_fEa7iyJuGwrAnUhKUTOx1vcr1Qt23XOr/content' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -o result.mp4
```

