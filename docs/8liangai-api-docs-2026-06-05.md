# 8liangai.com 接口测试与详细接口文档

更新时间：2026-06-05（Asia/Shanghai）

测试站点：`https://8liangai.com`

测试认证方式：`Authorization: Bearer sk-***`

证据口径说明：

- 本文中的“2026-06-05 实测”优先指向这把 Key 在 `2026-06-05` 当天的 `/api/log/token` 日志、对应任务查询、以及当时保留下来的接口返回。
- 本地文件 [tmp_8liangai_api_test_results.json](/C:/Users/A/Desktop/whdj/tmp_8liangai_api_test_results.json) 的生成时间是 `2026-05-25 14:45:22`，它可作为返回结构样例参考，但不能单独当作“2026-06-05 当天重新提交”的唯一证据。
- 因此，凡是涉及“当天是否真的发起过请求、是否真的扣费”的判断，应以 `2026-06-05` 当天该 Key 的 token 日志为准。

本文包含：

1. `sd2.0fast` 请求接口、视频状态查询接口、视频下载接口
2. `seedance1.5` 请求接口、视频状态查询接口、视频下载接口
3. `seedance2` 请求接口、视频状态查询接口、视频下载接口
4. `seedream4.5` 图片生成接口
5. `seedream5.0lite` 图片生成接口
6. 真人资产查询、接入相关接口
7. 虚拟人像资产查询、接入相关接口
8. `seedance2-sr` 请求接口、视频状态查询接口、视频下载接口
9. `seedance2.0fast-sr` 请求接口、视频状态查询接口、视频下载接口
10. `seedance1.5-sr` 请求接口、视频状态查询接口、视频下载接口
11. 网站使用教程：API 密钥创建、虚拟人像资产创建、真人资产创建

---

## 1. 本次测试结论总览

### 1.1 基础配置实测

- `GET /api/status` 成功
- 站点当前计费展示类型：`CNY`
- `quota_per_unit = 500000`
- 换算公式：`实际金额(元) = quota / 500000`
- 当前测试 Token 分组：`default`
- 当前测试分组倍率：`1`

### 1.2 模型可用性实测结论

| 模型 | 请求接口 | 状态查询 | 下载接口 | 2026-06-05 实测结论 | 扣费核对 |
| --- | --- | --- | --- | --- | --- |
| `seedream4.5` | `POST /v1/images/generations` | 不适用 | 返回图片 URL | 成功，`2048x2048` 可用；`1024x1024` 报错 | 对得上 |
| `seedream5.0lite` | `POST /v1/images/generations` | 不适用 | 返回图片 URL | 成功，`2048x2048` 可用；`1024x1024` 报错 | 对得上 |
| `seedance1.5` | `POST /v1/video/generations` | `GET /v1/video/generations/{task_id}`、`GET /v1/videos/{task_id}` | `GET /v1/videos/{task_id}/content` | 2026-06-05 token 日志显示已提交并完成 | 对得上 |
| `seedance1.5-sr` | 同上 | 同上 | 同上 | 2026-06-05 token 日志显示已提交并完成 | 对得上 |
| `seedance2` | 同上 | 同上 | 同上 | 2026-06-05 token 日志显示已提交并完成 | 对得上 |
| `sd2.0fast` | 同上 | 同上 | 同上 | 2026-06-05 token 日志显示已提交并完成 | 对得上 |
| `seedance2.0fast-sr` | 同上 | 同上 | 同上 | 2026-06-05 当天重新提交成功；查询和下载都成功 | 价格对得上，但实际下载分辨率低于用户请求 |
| `seedance2-sr` | 同上 | 同上 | 同上 | 2026-06-05 当天重新提交成功；查询和下载都成功 | 存在“价格页与实际计费倍率不一致”且“实际下载分辨率低于用户请求” |

### 1.3 资产接口实测结论

| 类别 | 接口链路 | 2026-06-05 实测结论 |
| --- | --- | --- |
| 虚拟人像资产 | 配置查询、分组查询、列表查询、上传、创建、同步、预览 | 全部跑通 |
| 真人资产 | 配置查询、列表查询、创建任务限制、刷新实名链接、上传素材、未实名前提交素材拦截、预览跳转 | 全部跑通到“真人实名”前一步；真人实名需要人工完成 |

### 1.4 本轮重新实测的核心任务与结论

本轮不复用旧结论，重新提交并落地的关键任务如下：

| 模型 | 请求类型 | task_id | 最终 token / 计费结论 | 下载结果 |
| --- | --- | --- | --- | --- |
| `seedance1.5` | 720p 文生视频 | `task_QYnSrLQyhZX83GaYNENzZEI67PTQdWIY` | `108900 × 8 = 871200 quota = ¥1.7424` | `1280x720`，`7,473,124 bytes` |
| `seedance1.5-sr` | 720p 文生视频 | `task_scvUXZ75JhSV9uMhyyMEhhvi4MPUsPBY` | `108900 × 10 = 1089000 quota = ¥2.1780` | `1280x720`，`8,897,420 bytes` |
| `seedance2` | 720p 文生视频 | `task_3WFTgFLZs0VydG4pRbQJXq4qRQ0xlCcg` | `108900 × 23 = 2504700 quota = ¥5.0094` | `1280x720`，`2,365,942 bytes` |
| `seedance2` | 1080p 文生视频 | `task_a2c7YrdDp9LFYDZeJQ72uQrdU7fCVL4E` | `245025 × 25.5 = 6248137 quota = ¥12.4963` | `1920x1080`，`4,056,113 bytes` |
| `seedance2` | 720p 视频输入 | `task_7cXrvbzeNGWbkDhk4ICZxscg82PMkHoI` | `216900 × 15.5 = 3361950 quota = ¥6.7239` | `1280x720`，`2,059,794 bytes` |
| `sd2.0fast` | 720p 文生视频 | `task_jK23tZdbjOKl5z5TrisFWhhxTTF1AFYM` | `108900 × 18.5 = 2014650 quota = ¥4.0293` | `1280x720`，`2,890,056 bytes` |
| `seedance2-sr` | 720p SR 别名 | `task_9MyFZ9cQuSaw1UjgDovb2BsPiZtHquU1` | 实际按 `23x` 结算，不是价格页的 `24.5x` | 实际下载 `864x496`，低于用户请求 `720p` |
| `seedance2-sr` | 1080p SR 别名 | `task_0JGkH3QTx0O87huGepGdJuClWIWdviMI` | 实际按 `25.5x` 结算，不是价格页的 `24.5x` | 实际下载 `1280x720`，低于用户请求 `1080p` |
| `seedance2.0fast-sr` | 720p SR 别名 | `task_ZOQvITphQWqiwz7Ice560EGuIVxvAOcf` | `50638 × 23.15 = 1172269 quota = ¥2.3445` | 实际下载 `864x496`，低于用户请求 `720p` |

另外，这轮还补了两个重要边界场景：

- `seedance2` 视频输入如果少了 `role=reference_video`，上游会报 `content` 参数无效。
- `seedance2` 视频输入即使格式正确，如果输入视频被上游判定“可能包含真人”，也会被拦截。

---

## 2. 统一规范

## 2.1 Base URL

```text
https://8liangai.com
```

## 2.2 认证

所有本文涉及的 API 都使用：

```http
Authorization: Bearer sk-你的APIKey
```

## 2.3 常见成功响应包装

### 业务型接口常见包装

```json
{
  "success": true,
  "message": "",
  "data": {}
}
```

### 视频业务查询接口常见包装

```json
{
  "code": "success",
  "message": "",
  "data": {}
}
```

### OpenAI 兼容视频查询接口常见包装

```json
{
  "id": "task_xxx",
  "task_id": "task_xxx",
  "object": "video",
  "model": "seedance1.5",
  "status": "completed",
  "progress": 100,
  "created_at": 1780641005,
  "completed_at": 1780641060,
  "metadata": {
    "url": "https://..."
  }
}
```

## 2.4 金额换算规则

- 站点实测：`quota_display_type = CNY`
- 站点实测：`quota_per_unit = 500000`
- 换算：`1.0000 元 = 500000 quota`
- 例子：
  - `150000 quota = 0.3000 元`
  - `2000000 quota = 4.0000 元`
  - `5750000 quota = 11.5000 元`

---

## 3. 视频接口总说明

## 3.1 提交任务接口

```http
POST /v1/video/generations
Content-Type: application/json
Authorization: Bearer sk-你的APIKey
```

## 3.2 业务状态查询接口

```http
GET /v1/video/generations/{task_id}
Authorization: Bearer sk-你的APIKey
```

## 3.3 OpenAI 兼容状态查询接口

```http
GET /v1/videos/{task_id}
Authorization: Bearer sk-你的APIKey
```

## 3.4 视频下载接口

```http
GET /v1/videos/{task_id}/content
Authorization: Bearer sk-你的APIKey
```

## 3.5 视频提交请求体字段

这些字段来自网关真实解析结构，顶层可直接传：

| 字段 | 类型 | 必填 | 说明 | 怎么填 |
| --- | --- | --- | --- | --- |
| `model` | `string` | 是 | 模型名 | 如 `seedance1.5`、`seedance2`、`sd2.0fast` |
| `prompt` | `string` | 是 | 文生视频提示词 | 直接写中文或英文描述 |
| `resolution` | `string` | 否 | 分辨率 | 常用 `480p`、`720p`、`1080p` |
| `ratio` | `string` | 否 | 画幅比例 | 常用 `16:9`、`9:16`、`1:1` |
| `duration` | `int` | 否 | 时长秒数 | 例如 `5` |
| `seconds` | `string` | 否 | 时长秒数字符串 | 例如 `"5"`；和 `duration` 二选一即可 |
| `images` | `string[]` | 否 | 参考图 URL 列表 | 用公开可访问的图片 URL |
| `image` | `string` | 否 | 单图字段 | 本项目的 Doubao 视频链路主要使用 `images`，不建议只传 `image` |
| `size` | `string` | 否 | 备用尺寸字段 | 个别链路可能兼容，视频模型优先用 `resolution` |
| `mode` | `string` | 否 | 备用模式字段 | 当前这批模型一般不需要 |
| `input_reference` | `string` | 否 | 备用参考输入 | 当前这批模型一般不需要 |
| `metadata` | `object` 或 `JSON string` | 否 | 扩展参数 | 支持直接传对象，也支持传字符串化 JSON |

## 3.6 `metadata` 支持字段

网关会把 `metadata` 反序列化后直接映射到上游请求，下面这些字段是代码中已明确支持的：

| 字段 | 类型 | 说明 | 怎么填 |
| --- | --- | --- | --- |
| `callback_url` | `string` | 上游回调地址 | 填可公网访问 URL |
| `return_last_frame` | `bool` | 是否返回最后一帧 | `true` / `false` |
| `service_tier` | `string` | 服务层级 | 一般留空或 `default` |
| `execution_expires_after` | `int` | 任务结果过期秒数 | 如 `172800` |
| `generate_audio` | `bool` | 是否生成音频 | `true` / `false` |
| `draft` | `bool` | 是否草稿模式 | `true` / `false` |
| `tools` | `array` | 工具配置 | 一般留空 |
| `resolution` | `string` | 与顶层 `resolution` 同义，且优先级更高 | `480p` / `720p` / `1080p` |
| `ratio` | `string` | 与顶层 `ratio` 同义，且优先级更高 | `16:9` / `9:16` / `1:1` |
| `duration` | `int` | 与顶层 `duration` 同义 | 如 `5` |
| `frames` | `int` | 帧数 | 按需填写 |
| `seed` | `int` | 随机种子 | 如 `12345` |
| `camera_fixed` | `bool` | 是否固定机位 | `true` / `false` |
| `watermark` | `bool` | 是否水印 | `true` / `false` |
| `content` | `array` | 原始内容项 | 进阶用法，可传 `video_url` 输入 |
| `portrait_asset_id` | `int/string` | 绑定已就绪真人资产 Job ID | 例如 `7` |
| `asset_id` | `string` | 绑定已就绪真人资产 Volc 资产 ID | 例如 `asset-202605...` |
| `enable_video_super_resolution` | `bool/string/1` | 是否启用视频超分后处理 | `true`、`"true"`、`1` 都可 |

## 3.7 带视频输入时的 `metadata.content` 写法

```json
{
  "model": "seedance2",
  "prompt": "让人物保持主体不变，做平滑运动",
  "resolution": "720p",
  "ratio": "16:9",
  "duration": 5,
  "metadata": {
    "content": [
      {
        "type": "video_url",
        "video_url": {
          "url": "https://example.com/input.mp4"
        }
      }
    ]
  }
}
```

## 3.8 业务查询响应字段说明

`GET /v1/video/generations/{task_id}` 返回的 `data` 常见字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `int` | 本地数据库任务 ID |
| `created_at` | `int64` | 任务创建时间戳 |
| `updated_at` | `int64` | 最近更新时间戳 |
| `task_id` | `string` | 公开任务 ID |
| `platform` | `string` | 平台类型，本批模型为 `45` |
| `user_id` | `int` | 用户 ID |
| `group` | `string` | 使用分组 |
| `channel_id` | `int` | 路由到的通道 ID |
| `quota` | `int` | 预扣 quota，不一定是最终实际扣费 |
| `action` | `string` | 一般为 `generate` |
| `status` | `string` | `QUEUED` / `IN_PROGRESS` / `SUCCESS` / `FAILURE` |
| `fail_reason` | `string` | 失败原因 |
| `result_url` | `string` | 视频结果 URL |
| `submit_time` | `int64` | 提交时间 |
| `start_time` | `int64` | 开始处理时间 |
| `finish_time` | `int64` | 完成时间 |
| `progress` | `string` | 进度，如 `100%` |
| `properties` | `object` | 附加属性，如原模型名、上游模型名、是否超分 |
| `data.content.video_url` | `string` | 上游视频地址 |
| `data.duration` | `int` | 上游记录的视频秒数 |
| `data.framespersecond` | `int` | FPS |
| `data.resolution` | `string` | 上游处理分辨率 |
| `data.ratio` | `string` | 上游画幅 |
| `data.seed` | `int` | 随机种子 |
| `data.usage.total_tokens` | `int` | 用于异步结算的 token 值 |

## 3.9 OpenAI 兼容查询响应字段说明

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `string` | 任务 ID |
| `task_id` | `string` | 兼容保留字段 |
| `object` | `string` | 固定 `video` |
| `model` | `string` | 你请求时的模型别名 |
| `status` | `string` | `queued` / `in_progress` / `completed` / `failed` |
| `progress` | `int` | 0-100 |
| `created_at` | `int64` | 创建时间戳 |
| `completed_at` | `int64` | 完成时间戳 |
| `metadata.url` | `string` | 视频地址 |
| `error` | `object` | 失败时出现，包含 `message`、`code` |

---

## 4. `seedance1.5`

## 4.1 请求接口

```http
POST /v1/video/generations
```

推荐请求体：

```json
{
  "model": "seedance1.5",
  "prompt": "一个女孩在海边缓慢回头，电影感，真实光影",
  "duration": 5,
  "resolution": "720p",
  "ratio": "16:9"
}
```

2026-06-05 实测成功返回：

```json
{
  "id": "task_QYnSrLQyhZX83GaYNENzZEI67PTQdWIY",
  "task_id": "task_QYnSrLQyhZX83GaYNENzZEI67PTQdWIY",
  "object": "video",
  "model": "seedance1.5",
  "status": "queued",
  "progress": 0,
  "created_at": 1780667070
}
```

## 4.2 视频状态查询接口

```http
GET /v1/video/generations/task_QYnSrLQyhZX83GaYNENzZEI67PTQdWIY
GET /v1/videos/task_QYnSrLQyhZX83GaYNENzZEI67PTQdWIY
```

业务查询实测关键结果：

- `status = SUCCESS`
- `duration = 5`
- `resolution = 720p`
- `ratio = 16:9`
- `framespersecond = 24`
- `total_tokens = 108900`
- `upstream_model_name = doubao-seedance-1-5-pro-251215`

## 4.3 视频下载接口

```http
GET /v1/videos/task_QYnSrLQyhZX83GaYNENzZEI67PTQdWIY/content
```

2026-06-05 实测：

- HTTP `200`
- `Content-Type: video/mp4`
- 返回体大小约 `7,473,124 bytes`

## 4.4 扣费核对

- 价格表：`model_ratio = 8`
- 预扣：`2000000 quota = 4.0000 元`
- 最终结算日志中的 `actual_quota = 871200`
- 实际金额：`871200 / 500000 = 1.7424 元`
- 退款：`2000000 - 871200 = 1128800 quota = 2.2576 元`
- 核对公式：`108900 × 8 = 871200`
- 结论：`对得上`

---

## 5. `seedance1.5-sr`

## 5.1 请求接口

```json
{
  "model": "seedance1.5-sr",
  "prompt": "一个女孩在海边缓慢回头，电影感，真实光影",
  "duration": 5,
  "resolution": "720p",
  "ratio": "16:9"
}
```

2026-06-05 token 日志对应任务：

- `task_id = task_scvUXZ75JhSV9uMhyyMEhhvi4MPUsPBY`

## 5.2 视频状态查询接口

关键返回：

- `status = SUCCESS`
- `duration = 5`
- `resolution = 720p`
- `ratio = 16:9`
- `total_tokens = 108900`
- `origin_model_name = seedance1.5-sr`

## 5.3 视频下载接口

2026-06-05 实测：

- HTTP `200`
- `Content-Type: video/mp4`
- 返回体大小约 `8,897,420 bytes`

## 5.4 扣费核对

- 价格表：`model_ratio = 10`
- 预扣：`2500000 quota = 5.0000 元`
- 最终 `actual_quota = 1089000`
- 实际金额：`2.1780 元`
- 退款：`1411000 quota = 2.8220 元`
- 核对公式：`108900 × 10 = 1089000`
- 结论：`对得上`

---

## 6. `seedance2`

## 6.1 请求接口

```json
{
  "model": "seedance2",
  "prompt": "一个机器人从镜头前走过，背景是未来城市，真实运动",
  "duration": 5,
  "resolution": "720p",
  "ratio": "16:9"
}
```

2026-06-05 token 日志对应任务：

- `task_id = task_3WFTgFLZs0VydG4pRbQJXq4qRQ0xlCcg`

## 6.2 视频状态查询接口

关键返回：

- `status = SUCCESS`
- `duration = 5`
- `resolution = 720p`
- `ratio = 16:9`
- `total_tokens = 108900`
- `upstream_model_name = ep-20260410114758-wzw55`

## 6.3 视频下载接口

2026-06-05 实测：

- HTTP `200`
- `Content-Type: video/mp4`
- 返回体大小约 `2,365,942 bytes`

## 6.4 扣费核对

`seedance2` 当前价格页是分档价，不是直接展示单个 `model_ratio`，但网关内部会把分档价换算成倍率结算。

当前价格表：

| 条件 | `input_price` |
| --- | --- |
| `720p` 无视频输入 | `46` |
| `720p` 有视频输入 | `31` |
| `480p` 无视频输入 | `46` |
| `480p` 有视频输入 | `31` |
| `1080p` 无视频输入 | `51` |
| `1080p` 有视频输入 | `34` |

网关换算规则：

- `model_ratio = input_price / 2`
- `720p` 无视频输入时：`46 / 2 = 23`

本次实测：

- 预扣：`5750000 quota = 11.5000 元`
- 最终 `actual_quota = 2504700`
- 实际金额：`5.0094 元`
- 退款：`3245300 quota = 6.4906 元`
- 核对公式：`108900 × 23 = 2504700`
- 结论：`对得上`

## 6.5 `seedance2` 额外分支：`1080p` 无视频输入

重新实测任务：

- `task_id = task_a2c7YrdDp9LFYDZeJQ72uQrdU7fCVL4E`

关键结果：

- 业务查询 `resolution = 1080p`
- `total_tokens = 245025`
- `/content` 下载成功，实际文件为 `1920x1080`
- 文件大小约 `4,056,113 bytes`

计费核对：

- 价格档：`1080p no video = 51`
- 换算倍率：`51 / 2 = 25.5`
- 预扣：`6375000 quota = 12.7500 元`
- 最终 `actual_quota = 6248137`
- 实际金额：`6248137 / 500000 = 12.4963 元`
- 核对公式：`245025 × 25.5 = 6248137.5`，日志按整数 quota 记为 `6248137`
- 结论：`与 1080p 档定价基本对得上`

## 6.6 `seedance2` 额外分支：带视频输入

### 6.6.1 错误写法：少 `role`

如果 `metadata.content` 里只传：

```json
{
  "type": "video_url",
  "video_url": {
    "url": "https://example.com/input.mp4"
  }
}
```

上游会报：

```json
{
  "error": {
    "code": "InvalidParameter",
    "message": "The parameter `content` specified in the request is not valid: reference media mode requires video role to be reference_video.",
    "type": "BadRequest"
  }
}
```

### 6.6.2 格式正确但输入视频被隐私拦截

把 `role` 改成 `reference_video` 之后，如果输入视频被上游判定为“可能包含真人”，仍可能报：

```json
{
  "error": {
    "code": "***.PrivacyInformation",
    "message": "The request failed because the input video may contain real person.",
    "type": "BadRequest"
  }
}
```

### 6.6.3 成功写法与实测结果

成功请求任务：

- `task_id = task_7cXrvbzeNGWbkDhk4ICZxscg82PMkHoI`

成功示例：

```json
{
  "model": "seedance2",
  "prompt": "基于输入视频做镜头延续，保持机器人主体一致",
  "duration": 5,
  "resolution": "720p",
  "ratio": "16:9",
  "metadata": {
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

关键结果：

- 业务查询 `resolution = 720p`
- `total_tokens = 216900`
- `/content` 下载成功，实际文件为 `1280x720`
- 文件大小约 `2,059,794 bytes`

计费核对：

- 价格档：`720p with video = 31`
- 换算倍率：`31 / 2 = 15.5`
- 预扣：`3875000 quota = 7.7500 元`
- 最终 `actual_quota = 3361950`
- 实际金额：`6.7239 元`
- 核对公式：`216900 × 15.5 = 3361950`
- 结论：`与“有视频输入”档定价对得上`

---

## 7. `sd2.0fast`

## 7.1 请求接口

```json
{
  "model": "sd2.0fast",
  "prompt": "一辆跑车穿过雨夜街道，霓虹反射，镜头平滑推进",
  "duration": 5,
  "resolution": "720p",
  "ratio": "16:9"
}
```

2026-06-05 token 日志对应任务：

- `task_id = task_jK23tZdbjOKl5z5TrisFWhhxTTF1AFYM`

## 7.2 视频状态查询接口

关键返回：

- `status = SUCCESS`
- `duration = 5`
- `resolution = 720p`
- `ratio = 16:9`
- `total_tokens = 108900`
- `upstream_model_name = ep-20260430104948-vv5vf`

## 7.3 视频下载接口

2026-06-05 实测：

- HTTP `200`
- `Content-Type: video/mp4`
- 返回体大小约 `2,890,056 bytes`

## 7.4 扣费核对

- 价格表：`model_ratio = 18.5`
- 预扣：`4625000 quota = 9.2500 元`
- 最终 `actual_quota = 2014650`
- 实际金额：`4.0293 元`
- 退款：`2610350 quota = 5.2207 元`
- 核对公式：`108900 × 18.5 = 2014650`
- 结论：`对得上`

---

## 8. `seedance2-sr`

## 8.1 请求接口

推荐请求体：

```json
{
  "model": "seedance2-sr",
  "prompt": "一个机器人从镜头前走过，背景是未来城市，真实运动",
  "duration": 5,
  "resolution": "720p",
  "ratio": "16:9"
}
```

## 8.2 SR 别名特殊逻辑

代码中这类专用 SR 别名会自动把“用户想要的目标清晰度”换成“先生成的源清晰度”：

- 请求 `720p` -> 上游先生成 `480p`
- 请求 `1080p` -> 上游先生成 `720p`

这条规则适用于：

- `seedance2-sr`
- `seedance2.0-sr`
- `seedance2.0fast-sr`
- `sd2.0fast-sr`

## 8.3 2026-06-05 实测结果

这次重新实测把 `720p` 和 `1080p` 两个分支都重新跑通了：

- `720p` 任务：`task_9MyFZ9cQuSaw1UjgDovb2BsPiZtHquU1`
- `1080p` 任务：`task_0JGkH3QTx0O87huGepGdJuClWIWdviMI`

两条任务都满足：

- `POST /v1/video/generations` 成功返回 `queued`
- `GET /v1/video/generations/{task_id}` 成功返回 `SUCCESS`
- `GET /v1/videos/{task_id}` 成功返回 `completed`
- `GET /v1/videos/{task_id}/content` 成功下载 `video/mp4`

## 8.4 扣费说明

- 价格页当前显示：`model_ratio = 24.5`
- 但本轮重新实测发现，实际并不是按 `24.5x` 结算

### 8.4.1 `720p` 分支

- 业务查询显示：`resolution = 480p`
- `total_tokens = 50638`
- 下载接口成功，实际文件大小约 `998,462 bytes`
- 用 `ffprobe` 读到的实际视频尺寸：`864x496`

计费：

- 预扣：`5750000 quota = 11.5000 元`
- 最终 `actual_quota = 1164674`
- 日志里的实际倍率：`23.0`
- 核对公式：`50638 × 23 = 1164674`

### 8.4.2 `1080p` 分支

- 业务查询显示：`resolution = 720p`
- `total_tokens = 108900`
- 下载接口成功，实际文件大小约 `1,941,720 bytes`
- 用 `ffprobe` 读到的实际视频尺寸：`1280x720`

计费：

- 预扣：`6375000 quota = 12.7500 元`
- 最终 `actual_quota = 2776950`
- 日志里的实际倍率：`25.5`
- 核对公式：`108900 × 25.5 = 2776950`

### 8.4.3 重新实测后的明确结论

- `seedance2-sr` 现在已经不是“只能验证到预扣费”了，而是提交、查询、下载都重新跑通了
- 但它的**实际计费倍率**：
  - `720p` 按 `23x`
  - `1080p` 按 `25.5x`
- 这两档都更像 `seedance2` 的分档逻辑，而不是价格页里单独展示的 `24.5x`
- 它的**最终下载分辨率**也没有达到用户请求的目标：
  - 请求 `720p`，下载到 `864x496`
  - 请求 `1080p`，下载到 `1280x720`
- 所以当前结论不是“模型不可用”，而是：
  - `接口可用`
  - `下载可用`
  - `计费与价格页展示存在不一致`
  - `SR 结果分辨率低于用户请求`

---

## 9. `seedance2.0fast-sr`

## 9.1 请求接口

推荐请求体：

```json
{
  "model": "seedance2.0fast-sr",
  "prompt": "一辆跑车穿过雨夜街道，霓虹反射，镜头平滑推进",
  "duration": 5,
  "resolution": "720p",
  "ratio": "16:9"
}
```

## 9.2 2026-06-05 当天新请求结果

重新实测任务：

- `task_id = task_ZOQvITphQWqiwz7Ice560EGuIVxvAOcf`

这次重新实测已经满足：

- 提交成功
- 业务查询成功
- OpenAI 查询成功
- `/content` 下载成功

## 9.3 当前任务状态查询

当前任务关键返回：

- `status = SUCCESS`
- `origin_model_name = seedance2.0fast-sr`
- 上游实际生成 `resolution = 480p`
- `total_tokens = 50638`

OpenAI 兼容查询返回 `completed`。

## 9.4 视频下载接口

对本轮新任务访问：

```http
GET /v1/videos/task_ZOQvITphQWqiwz7Ice560EGuIVxvAOcf/content
```

实测结果：

- HTTP `200`
- `Content-Type: video/mp4`
- 返回体大小约 `1,138,252 bytes`
- 用 `ffprobe` 读到的实际分辨率：`864x496`

这说明本轮新任务的下载链路是好的，但**实际输出清晰度低于用户请求的 `720p`**。

## 9.5 扣费核对

- `model_ratio = 23.15`
- 预扣：`5787500 quota = 11.5750 元`
- 最终 `actual_quota = 1172269`
- 实际金额：`1172269 / 500000 = 2.3445 元`
- 退款：`4615231 quota = 9.2305 元`
- 核对公式：`50638 × 23.15 = 1172269.7`，日志按整数 quota 记为 `1172269`

结论：

- 当前价格页和本轮重新实测的实际计费是一致的
- 但当前下载到的实际视频尺寸只有 `864x496`
- 所以它当前的真实状态应写成：
  - `接口可用`
  - `查询可用`
  - `下载可用`
  - `计费对得上`
  - `输出分辨率低于用户请求`

---

## 10. `seedream4.5`

## 10.1 请求接口

```http
POST /v1/images/generations
Content-Type: application/json
Authorization: Bearer sk-你的APIKey
```

推荐请求体：

```json
{
  "model": "seedream4.5",
  "prompt": "A clean product shot of a ceramic cup on a soft studio background",
  "size": "2048x2048"
}
```

## 10.2 2026-06-05 实测成功响应

```json
{
  "model": "doubao-seedream-4-5-251128",
  "created": 1780667569,
  "data": [
    {
      "url": "https://...",
      "size": "2048x2048"
    }
  ],
  "usage": {
    "generated_images": 1,
    "output_tokens": 16384,
    "total_tokens": 16384
  }
}
```

## 10.3 字段说明

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `model` | `string` | 是 | 固定写 `seedream4.5` |
| `prompt` | `string` | 是 | 生图提示词 |
| `size` | `string` | 强烈建议填 | 当前实测必须满足最小像素要求，`2048x2048` 成功 |

## 10.4 尺寸限制实测

`1024x1024` 当前重新实测报错：

```json
{
  "error": {
    "message": "The parameter `size` specified in the request is not valid: image size must be at least 3686400 pixels.",
    "type": "upstream_error",
    "code": "InvalidParameter"
  }
}
```

因此当前环境下建议直接使用：

- `2048x2048`

## 10.5 扣费核对

- 价格表：`model_price = 0.3`
- 实际日志：`quota = 150000`
- 实际金额：`150000 / 500000 = 0.3000 元`
- 结论：`对得上`

---

## 11. `seedream5.0lite`

## 11.1 请求接口

```http
POST /v1/images/generations
Content-Type: application/json
Authorization: Bearer sk-你的APIKey
```

```json
{
  "model": "seedream5.0lite",
  "prompt": "A clean product shot of a ceramic cup on a soft studio background",
  "size": "2048x2048"
}
```

## 11.2 2026-06-05 实测成功响应

```json
{
  "model": "doubao-seedream-5-0-260128",
  "created": 1780667582,
  "data": [
    {
      "url": "https://...",
      "size": "2048x2048"
    }
  ],
  "usage": {
    "generated_images": 1,
    "output_tokens": 16384,
    "total_tokens": 16384
  }
}
```

## 11.3 字段说明

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `model` | `string` | 是 | 固定写 `seedream5.0lite` |
| `prompt` | `string` | 是 | 生图提示词 |
| `size` | `string` | 强烈建议填 | 当前实测必须满足最小像素要求，`2048x2048` 成功 |

## 11.4 尺寸限制

`1024x1024` 同样当前重新实测报 `400`，原因也是像素不够：

```json
{
  "error": {
    "message": "The parameter `size` specified in the request is not valid: image size must be at least 3686400 pixels.",
    "type": "upstream_error",
    "code": "InvalidParameter"
  }
}
```

建议直接使用：

- `2048x2048`

## 11.5 扣费核对

- 价格表：`model_price = 0.25`
- 实际日志：`quota = 125000`
- 实际金额：`0.2500 元`
- 结论：`对得上`

---

## 12. 真人资产接口

说明：

- 本文把“真人资产”对应为官方真人资产链路，即 `/api/portrait_assets/official/*`
- 它依赖“真人实名校验”
- 认证方式既支持登录态，也支持 API Key；本文用 API Key 已实测通过

## 12.1 配置查询

```http
GET /api/portrait_assets/official/config
```

实测返回：

```json
{
  "success": true,
  "message": "",
  "data": {
    "configured": true,
    "project_name": "default"
  }
}
```

字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `configured` | `bool` | 是否已配置火山真人资产能力 |
| `project_name` | `string` | 当前项目名 |

## 12.2 列表查询

```http
GET /api/portrait_assets/official/jobs?p=1&page_size=5
```

分页字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `page` | `int` | 当前页 |
| `page_size` | `int` | 每页条数 |
| `total` | `int` | 总数 |
| `items` | `array` | 列表 |

单条 Job 字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `int` | 任务 ID |
| `user_id` | `int` | 用户 ID |
| `name` | `string` | 任务名 |
| `source` | `string` | 固定为 `official` |
| `status` | `string` | 见下方状态说明 |
| `invite_url` | `string` | 真人校验 H5 链接 |
| `qr_image` | `string` | 二维码图片地址，当前实测多数为空 |
| `validate_result_code` | `string` | 实名回调码，`10000` 常表示成功 |
| `volc_group_id` | `string` | 火山素材组 ID |
| `asset_id` | `string` | 火山资产 ID |
| `asset_status` | `string` | 上游资产状态，如 `Active`、`Failed` |
| `asset_preview` | `string` | 预览地址 |
| `asset_url` | `string` | 原始上游素材地址 |
| `asset_type` | `string` | `Image` / `Video` / `Audio` |
| `project_name` | `string` | 项目名 |
| `error_message` | `string` | 错误信息 |
| `created_time` | `int64` | 创建时间 |
| `updated_time` | `int64` | 更新时间 |
| `accept_time` | `int64` | 接受时间 |
| `qr_expires_time` | `int64` | 校验链接过期时间 |
| `ready_time` | `int64` | 完成时间 |
| `queue_position` | `int64` | 队列位置 |

## 12.3 状态值说明

| `status` | 含义 |
| --- | --- |
| `pending` | 已创建，待初始化 |
| `validate_ready` | 已生成实名校验链接 |
| `validated` | 真人校验通过 |
| `asset_processing` | 素材入库中 |
| `pending_confirm` | 待人工确认预览 |
| `ready` | 完成，可用 |
| `failed` | 失败 |
| `expired` | 实名链接过期 |

## 12.4 创建真人资产任务

```http
POST /api/portrait_assets/official/jobs
Content-Type: application/json
```

请求体：

```json
{
  "name": "doc-test-official-job"
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | `string` | 否 | 任务名，最多 50 个字符；不填则后端用默认名 |

2026-06-05 当前重新实测补充说明：

- 这个账号当前已经有一个进行中的官方真人资产任务：`id = 16`
- 因此再次创建会返回：`你已有进行中的官方真人资产任务，请完成后再创建`
- 说明这个接口除了“能创建”之外，还带有“同一账号同一时刻只能保留一个进行中任务”的限制

## 12.5 刷新真人校验链接

```http
POST /api/portrait_assets/official/jobs/{id}/validation
```

用途：

- 原链接过期时重新生成
- 当前任务未进入素材处理阶段时可调用

2026-06-05 当前重新实测对 `id=16` 成功，返回新的 `invite_url`

## 12.6 上传真人素材

```http
POST /api/portrait_assets/official/upload
Content-Type: multipart/form-data
```

表单字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `file` | 文件 | 是 | 要上传的素材文件 |

上传限制：

- 最大 `100 MB`

支持格式：

- 图片：`.jpg` `.jpeg` `.png` `.webp` `.gif` `.bmp`
- 视频：`.mp4` `.mov` `.webm`
- 音频：`.mp3` `.wav` `.m4a` `.ogg` `.aac` `.flac`

实测成功返回：

```json
{
  "success": true,
  "data": {
    "url": "https://8liangai.com/portrait-asset-uploads/20260605/fbcf63d6-32b6-4d9d-9ded-adeb8a75d7d0.png",
    "file_name": "icon.png",
    "content_type": "image/png",
    "asset_type": "Image",
    "size": 31262
  }
}
```

## 12.7 提交真人素材入库

```http
POST /api/portrait_assets/official/jobs/{id}/asset
Content-Type: application/json
```

请求体：

```json
{
  "asset_url": "https://8liangai.com/portrait-asset-uploads/20260605/fbcf63d6-32b6-4d9d-9ded-adeb8a75d7d0.png",
  "asset_type": "Image",
  "name": "doc-test-official-asset"
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `asset_url` | `string` | 是 | 必须是可公开访问的 `http/https` URL |
| `asset_type` | `string` | 是 | 推荐填 `Image`、`Video`、`Audio` |
| `name` | `string` | 否 | 上游素材名 |

2026-06-05 实测：

- 在未完成真人实名前调用
- 返回：

```json
{
  "success": false,
  "message": "请先完成真人认证"
}
```

这说明流程顺序必须是：

1. 创建任务
2. 打开 `invite_url` 完成人脸实名
3. 上传素材
4. 提交素材入库
5. 等待 `pending_confirm`
6. 调用确认接口

## 12.8 同步真人资产状态

```http
POST /api/portrait_assets/official/jobs/{id}/sync
```

用途：

- 主动向上游同步任务/素材状态

## 12.9 确认 / 驳回

```http
POST /api/portrait_assets/official/jobs/{id}/confirm
POST /api/portrait_assets/official/jobs/{id}/reject
```

用途：

- 当任务进入 `pending_confirm` 后确认或驳回

## 12.10 预览地址

```http
GET /api/portrait_assets/official/jobs/{id}/preview/{state}
```

说明：

- 这是一个带签名态的预览跳转地址
- 直接访问通常会 `302` 到真实图片/视频资源 URL

---

## 13. 虚拟人像资产接口

说明：

- 本文把“虚拟人像资产”对应为 `/api/portrait_assets/virtual/*`
- 使用现有素材组，不需要真人实名

## 13.1 配置查询

```http
GET /api/portrait_assets/virtual/config
```

实测返回：

```json
{
  "success": true,
  "data": {
    "configured": true,
    "project_name": "default"
  }
}
```

## 13.2 当前用户素材组查询

```http
GET /api/portrait_assets/virtual/group
```

实测返回字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `int` | 本地分组 ID |
| `user_id` | `int` | 用户 ID |
| `name` | `string` | 分组名 |
| `description` | `string` | 描述 |
| `project_name` | `string` | 项目名 |
| `volc_group_id` | `string` | 火山素材组 ID |
| `status` | `string` | `creating` / `active` / `failed` |
| `error_message` | `string` | 错误信息 |
| `created_time` | `int64` | 创建时间 |
| `updated_time` | `int64` | 更新时间 |

## 13.3 列表查询

```http
GET /api/portrait_assets/virtual/assets?p=1&page_size=5
```

单条资产字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `int` | 资产 ID |
| `user_id` | `int` | 用户 ID |
| `group_id` | `int` | 本地分组 ID |
| `name` | `string` | 资产名 |
| `asset_type` | `string` | `Image` / `Video` / `Audio` |
| `source_url` | `string` | 原始素材 URL |
| `preview_url` | `string` | 预览地址 |
| `project_name` | `string` | 项目名 |
| `volc_group_id` | `string` | 火山素材组 ID |
| `volc_asset_id` | `string` | 火山资产 ID |
| `status` | `string` | `processing` / `active` / `failed` |
| `volc_status` | `string` | 上游状态 |
| `error_message` | `string` | 错误信息 |
| `created_time` | `int64` | 创建时间 |
| `updated_time` | `int64` | 更新时间 |
| `ready_time` | `int64` | 就绪时间 |

## 13.4 上传素材

```http
POST /api/portrait_assets/virtual/upload
Content-Type: multipart/form-data
```

参数和限制与官方真人上传接口一致：

- 字段名：`file`
- 最大：`100 MB`
- 支持图片 / 视频 / 音频格式

2026-06-05 实测成功：

```json
{
  "success": true,
  "data": {
    "url": "https://8liangai.com/portrait-asset-uploads/20260605/8415fb4f-75c3-428d-9462-d748395b6fdc.png",
    "file_name": "icon.png",
    "content_type": "image/png",
    "asset_type": "Image",
    "size": 31262
  }
}
```

## 13.5 创建虚拟人像资产

```http
POST /api/portrait_assets/virtual/assets
Content-Type: application/json
```

请求体：

```json
{
  "name": "doc-test-virtual-asset-512",
  "asset_url": "https://8liangai.com/portrait-asset-uploads/20260605/8415fb4f-75c3-428d-9462-d748395b6fdc.png",
  "asset_type": "Image"
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | `string` | 否 | 资产名，最多 50 字符 |
| `asset_url` | `string` | 是 | 必须是可公开访问的 `http/https` URL |
| `asset_type` | `string` | 是 | 只接受 `Image`、`Video`、`Audio`，大小写不敏感，后端会规范化 |

### 创建成功示例

```json
{
  "success": true,
  "data": {
    "id": 11,
    "name": "doc-test-virtual-asset-512",
    "asset_type": "Image",
    "source_url": "https://8liangai.com/portrait-asset-uploads/20260605/8415fb4f-75c3-428d-9462-d748395b6fdc.png",
    "preview_url": "https://8liangai.com/api/portrait_assets/virtual/assets/11/preview/40a12e8a985ebef43f71a8482db2e1edf1b35438fe2cfba7a5227ac3fcb77bbd",
    "volc_asset_id": "asset-20260605210916-rmtdr",
    "status": "processing",
    "volc_status": "Processing"
  }
}
```

### 创建失败示例

2026-06-05 用过小图片创建时实测失败：

```json
{
  "success": false,
  "message": "volc portrait CreateAsset failed: HTTP 400: ... Width must be between 300px and 6000px."
}
```

这说明：

- 虚拟人像上传成功不代表素材一定能入库成功
- 至少要满足上游图像宽度要求

## 13.6 同步状态

```http
POST /api/portrait_assets/virtual/assets/{id}/sync
```

2026-06-05 当前重新实测对 `id=12` 连续同步：

- 第一次同步就返回 `active`

最终成功结果：

```json
{
  "success": true,
  "data": {
    "id": 12,
    "status": "active",
    "volc_status": "Active",
    "ready_time": 1780667649
  }
}
```

## 13.7 预览地址

```http
GET /api/portrait_assets/virtual/assets/{id}/preview/{state}
```

2026-06-05 当前重新实测对 `id=12`：

- 返回 HTTP `302`
- 跳转到真实的火山素材图片地址

---

## 14. 网站使用教程：API 密钥怎么创建

前端路由：

- `/_authenticated/keys/`

建议从页面路径理解为：

- 登录网站
- 进入“控制台 / API Keys / 密钥管理”

## 14.1 页面表单字段

前端表单当前包含这些字段：

| 字段 | 含义 | 怎么填 |
| --- | --- | --- |
| `Name` | 密钥名称 | 必填，例如 `project-a-prod` |
| `Group` | 分组 | 一般用 `default`；如果你启用了自动熔断可选 `auto` |
| `Cross-group retry` | 跨组重试 | 只有 `group=auto` 时出现；开启后当前组失败会尝试下一组 |
| `Unlimited Quota` | 是否无限额度 | 开启后不限制单 Key 额度 |
| `Quota` | 额度 | 关闭无限额度后填写；站点当前按人民币展示 |
| `Expiration Time` | 过期时间 | 可为空表示永不过期，也可快速选 `1 Hour`、`1 Day`、`1 Month` |
| `Quantity` | 一次创建多少个 Key | 新建时才有；大于 1 时会自动给名称加随机后缀 |
| `Model Limits` | 模型限制 | 为空表示允许全部模型；选中后只允许该 Key 调这些模型 |
| `IP Whitelist` | IP 白名单 | 一行一个 IP 或 CIDR；为空表示不限制 |

## 14.2 创建流程

1. 登录网站。
2. 进入密钥管理页。
3. 点击 `Create API Key`。
4. 填 `Name`。
5. 选 `Group`，普通用户通常选 `default`。
6. 选择是否 `Unlimited Quota`。
7. 如果关闭无限额度，填写 `Quota`。
8. 选 `Expiration Time`，不填就是永不过期。
9. 如果要限制模型，在 `Advanced Options` 中选择 `Model Limits`。
10. 如果要限制来源 IP，在 `IP Whitelist` 中一行填一个 IP。
11. 点 `Save changes`。

## 14.3 后端提交字段

前端最终会提交这些字段：

```json
{
  "name": "project-a-prod",
  "remain_quota": 0,
  "expired_time": -1,
  "unlimited_quota": true,
  "model_limits_enabled": false,
  "model_limits": "",
  "allow_ips": "",
  "group": "default",
  "cross_group_retry": false
}
```

字段说明：

- `remain_quota`：后端用的 quota 单位，不是元
- `expired_time = -1`：表示永不过期
- `model_limits_enabled = true` 时，`model_limits` 为逗号分隔字符串

---

## 15. 网站使用教程：虚拟人像资产怎么创建

前端路由：

- `/_authenticated/portrait-assets-virtual/`

## 15.1 页面操作顺序

1. 登录网站。
2. 打开“虚拟人像资产”页面。
3. 先上传素材文件。
4. 拿到上传后的公网 URL。
5. 填写资产名称。
6. 选择素材类型：`Image` / `Video` / `Audio`。
7. 提交创建。
8. 若状态还是 `processing`，点击刷新或调用同步接口。
9. 等到状态变成 `active` 后即可用于后续业务。

## 15.2 素材要求建议

- 素材 URL 必须可公网访问
- 图片尽量宽度不小于 `300px`
- 当前实测 `512x512 PNG` 成功
- 当前实测 `200x80 PNG` 被上游拒绝

## 15.3 如果要在视频接口里引用官方真人资产

视频 `metadata` 中可用：

```json
{
  "portrait_asset_id": 7
}
```

或者：

```json
{
  "asset_id": "asset-20260506153436-8dz8q"
}
```

注意：

- 这里只是“官方真人资产”的引用方式，不是“虚拟人像资产”的引用方式
- 只能引用“当前用户名下且已经 ready/active 的官方真人资产”
- `portrait_asset_id` 对应本地真人资产 Job ID
- `asset_id` 对应上游火山真人资产 ID

## 15.4 当前没有确认到“虚拟人像资产直绑视频模型”的专用字段

截至 2026-06-05，这个项目代码里已经明确能确认到的视频请求绑定字段，是上面这两个“官方真人资产”字段：

- `metadata.portrait_asset_id`
- `metadata.asset_id`

而对于“虚拟人像资产”：

- 本次已经完整验证了它的配置、分组、上传、创建、同步、预览接口
- 但在当前视频模型请求结构里，没有找到一个已经明确定义、并且与虚拟人像资产 `id` / `volc_asset_id` 直接绑定的专用字段
- 所以文档不能把虚拟人像资产的 `id`、`volc_asset_id`、`volc_group_id` 猜测性地写成视频生成入参

如果后续你希望把“虚拟人像资产”也直接接入某条生成链路，建议再单独补一轮“下游消费字段”的联调验证。

---

## 16. 网站使用教程：真人资产怎么创建

前端路由：

- `/_authenticated/portrait-assets-official/`

## 16.1 页面操作顺序

1. 登录网站。
2. 打开“真人资产”页面。
3. 点击创建任务。
4. 系统返回 `invite_url`。
5. 在手机或浏览器打开 `invite_url`，完成真人实名校验。
6. 校验成功后，回到网站上传图片 / 视频 / 音频素材。
7. 提交素材入库。
8. 调用刷新/同步，等待状态变成 `pending_confirm`。
9. 预览无误后点击确认。
10. 最终状态变成 `ready`。

## 16.2 失败与注意点

- `invite_url` 会过期，过期后调用 `/validation` 重新生成
- 未完成真人实名前，提交素材会返回：`请先完成真人认证`
- 预览地址本质上是带状态签名的跳转 URL

---

## 17. 建议补测项

这轮重新测试已经把“之前没真正落地的 SR 请求”补齐了，所以建议补测项也变了。现在最值得继续盯的是 4 个问题：

1. `seedance2-sr` 的价格页展示为 `24.5x`，但 `720p` 实测按 `23x`、`1080p` 实测按 `25.5x` 结算，建议后台重点核对价格配置。
2. `seedance2-sr` 请求 `720p/1080p` 时，最终下载分辨率分别只有 `864x496` / `1280x720`，建议确认 SR 后处理是否真的启用。
3. `seedance2.0fast-sr` 请求 `720p` 时最终下载分辨率只有 `864x496`，建议确认它当前是否只做了“低分生成”，没有真正超分到目标清晰度。
4. `seedance2` 的视频输入分支对 `metadata.content` 的格式要求比较严格，建议在文档或前端明确写出必须携带 `role=reference_video`，并提示“可能包含真人”的输入会被隐私策略拦截。

---

## 18. 本次测试中最重要的结论

- `seedream4.5`、`seedream5.0lite` 可用，且扣费与价格完全对得上。
- `seedance1.5`、`seedance1.5-sr`、`seedance2`、`sd2.0fast` 本轮重新提交、查询、下载、结算全部成功，价格也都能对上。
- `seedance2` 的三个分支这次都实际测到了：
  - `720p` 无视频输入：按 `23x` 结算
  - `1080p` 无视频输入：按 `25.5x` 结算
  - `720p` 有视频输入：按 `15.5x` 结算
- `seedance2-sr` 现在已经确认“能生成、能查、能下”，但它有两个明显问题：
  - 价格页写 `24.5x`，实际却按 `23x/25.5x` 结算
  - 下载到的实际分辨率低于用户请求
- `seedance2.0fast-sr` 当前价格 `23.15x` 的计费是对得上的，但下载到的实际分辨率同样低于用户请求。
- 虚拟人像资产链路已经完整跑通。
- 真人资产链路已经跑通到“真人实名前一步”，并确认了：
  - 同一账号存在进行中的官方真人资产任务时，不能重复创建
  - 未实名前提交素材会被正常拦截
