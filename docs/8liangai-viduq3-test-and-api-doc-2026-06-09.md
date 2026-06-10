# 8liangai `viduq3` 实测与接口文档

更新时间：2026-06-09

Base URL：`https://8liangai.com`

## 1. 实测结论

2026-06-09 我使用你提供的 API Key 对线上站点做了真实调用测试，当前结论是：

- `viduq3` 现在还不能在 `https://8liangai.com` 上正常使用。
- `GET /v1/models` 当前返回 12 个模型，结果中没有 `viduq3`。
- `POST /v1/video/generations` 使用 `model=viduq3` 会直接返回 `503`。
- 错误类型是 `model_not_found`，错误信息明确指出：`group default` 下没有可用通道。

也就是说，当前不是“任务提交后失败”，而是“分发阶段就没有找到可用模型通道”。

## 2. 实测证据

### 2.1 模型列表

请求：

```http
GET /v1/models
Authorization: Bearer sk-****
```

实测结果摘要：

```json
{
  "success": true,
  "model_count": 12,
  "model_ids": [
    "gpt-4",
    "seedance1.5-sr",
    "DeepSeek-v4-pro",
    "seedance2",
    "seedance1.5",
    "seedream4.5",
    "seedream5.0lite",
    "gpt-5",
    "sd2.0fast",
    "seedance2-sr",
    "GLM-4",
    "seedance2.0fast-sr"
  ]
}
```

结论：`viduq3` 未暴露在当前线上模型列表中。

### 2.2 真实提交测试

请求：

```bash
curl -X POST 'https://8liangai.com/v1/video/generations' \
  -H 'Authorization: Bearer sk-****' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "viduq3",
    "prompt": "A calm orange cat sitting by a sunny window, cinematic, natural motion, stable camera, no text",
    "resolution": "720p",
    "duration": 5,
    "metadata": {
      "watermark": false
    }
  }'
```

响应：

```json
{
  "error": {
    "code": "model_not_found",
    "message": "No available channel for model viduq3 under group default (distributor) (request id: 202606090201026885997918268d9d6qAkGJrDI)",
    "type": "new_api_error"
  }
}
```

HTTP 状态码：`503`

## 3. 为什么现在不可用

结合线上返回结果和本地代码，可以先重点检查这几项：

1. 后台分发能力没有真正生效。
   `model_not_found` 说明 `default` 分组下没有任何已启用通道可承接 `viduq3`。

2. 线上代码里的 `vidu` 适配器本身还没有把 `viduq3` 列入模型列表。
   当前 [relay/channel/task/vidu/adaptor.go](/C:/Users/A/Desktop/whdj/relay/channel/task/vidu/adaptor.go:214) 的 `GetModelList()` 只返回：
   `viduq2`、`viduq1`、`vidu2.0`、`vidu1.5`

3. 如果你只是后台手工新增了模型名，但没有同步：
   - 通道 `models`
   - ability/distributor 记录
   - 上游模型映射
   - 新代码部署
   
   那么前台能看到模型名，不代表实际请求一定能分发成功。

## 4. `viduq3` 建议接口

假设你准备把 `viduq3` 作为 Vidu 视频模型接入，建议对外仍然使用项目现有统一视频接口：

### 4.1 提交任务

```http
POST /v1/video/generations
Authorization: Bearer sk-你的APIKey
Content-Type: application/json
```

建议请求体：

```json
{
  "model": "viduq3",
  "prompt": "一只橘猫坐在窗边，阳光洒进房间，镜头平稳，动作自然，无字幕",
  "duration": 5,
  "size": "720p",
  "metadata": {
    "watermark": false
  }
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `model` | `string` | 是 | 固定填 `viduq3` |
| `prompt` | `string` | 是 | 文生视频提示词 |
| `duration` | `int` | 否 | 视频时长，建议先用 `5` |
| `size` | `string` | 否 | 建议填 `720p` 或 `1080p` |
| `image` | `string` | 否 | 单图 URL/Base64，图生视频时使用 |
| `images` | `string[]` | 否 | 多图输入时使用 |
| `metadata.watermark` | `bool` | 否 | 是否水印 |
| `metadata.seed` | `int` | 否 | 随机种子 |
| `metadata.callback_url` | `string` | 否 | 回调地址 |
| `metadata.movement_amplitude` | `string` | 否 | 运动幅度，常见可用值可先保留 `auto` |
| `metadata.bgm` | `bool` | 否 | 是否启用背景音乐 |

注意：

- 按当前适配器实现，Vidu 通道最终读取的是 `size` 并转给上游的 `resolution`。
- 代码位置见 [relay/channel/task/vidu/adaptor.go](/C:/Users/A/Desktop/whdj/relay/channel/task/vidu/adaptor.go:226)。
- 网关通用请求结构支持 `resolution`、`ratio`、`size`、`duration`，定义见 [relay/common/relay_info.go](/C:/Users/A/Desktop/whdj/relay/common/relay_info.go:666)。
- 但当前 Vidu 适配器没有直接消费顶层 `ratio`，所以 `viduq3` 接入时建议优先使用 `size`，不要把画幅控制完全依赖在 `ratio` 上。

### 4.2 提交成功响应

成功时通常返回：

```json
{
  "id": "task_xxx",
  "object": "video",
  "created_at": 1760000000,
  "status": "queued",
  "model": "viduq3",
  "task_id": "task_xxx"
}
```

### 4.3 查询任务状态

```http
GET /v1/video/generations/{task_id}
Authorization: Bearer sk-你的APIKey
```

常见状态：

- `queued`
- `in_progress`
- `completed`
- `failed`

### 4.4 下载成品

```http
GET /v1/videos/{task_id}/content
Authorization: Bearer sk-你的APIKey
```

如果任务成功，接口会返回视频文件流。

## 5. 推荐复测步骤

在你修完后台配置或重新部署后，按下面顺序复测最稳：

1. 先调用 `GET /v1/models`
   确认返回值里已经出现 `viduq3`

2. 再调用 `POST /v1/video/generations`
   建议先用最低成本参数：

```json
{
  "model": "viduq3",
  "prompt": "A cat sitting by the window, stable camera, natural motion, no text",
  "duration": 5,
  "size": "720p",
  "metadata": {
    "watermark": false
  }
}
```

3. 如果提交成功，拿到 `task_id` 后轮询：

```bash
curl 'https://8liangai.com/v1/video/generations/{task_id}' \
  -H 'Authorization: Bearer sk-你的APIKey'
```

4. 状态变成 `completed` 后下载：

```bash
curl -L 'https://8liangai.com/v1/videos/{task_id}/content' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -o result.mp4
```

## 6. 上线前检查清单

- 通道已启用
- 通道分组包含 `default`
- 通道 `models` 已包含 `viduq3`
- `abilities` 已生成或已刷新
- 如有模型映射，`viduq3 -> 上游真实模型名` 已配置
- 线上服务已部署包含 `viduq3` 支持的代码
- `GET /v1/models` 能看到 `viduq3`
- 用真实 API Key 提交后不再返回 `model_not_found`

## 7. 当前最终判断

截至 2026-06-09，这个线上地址上的 `viduq3` 还不能对外使用。

如果你愿意，我下一步可以继续直接帮你排查为什么后台新增后没有进入 distributor，或者直接把代码里的 `vidu` 适配层补到支持 `viduq3`。
