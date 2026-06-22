# 8liangai.com 接口文档

版本：`2026.06.17.1`

接口地址：`https://8liangai.com`

***

## 0. 阅读说明

### 0.1 先看哪一部分

按接入场景直接看下面对应章节：

| 场景             | 直接看这里       |
| -------------- | ----------- |
| 先看完整接入示例       | `1. 快速开始`   |
| 对接 DeepSeek 对话 | `3.1`、`3.2` |
| 对接视频生成         | `3.3` 到 `7` |
| 对接图片生成         | `8`         |
| 对接真人资产         | `9`         |
| 对接虚拟人像资产       | `10`        |
| 对接资产文件夹        | `11`        |
| 想看控制台操作流程      | `12`        |
| 想做上线前自查        | `13`        |

### 0.2 这份文档怎么用

接入时直接按下面顺序看：

1. 先看 `1. 快速开始` 里的完整示例。
2. 再看 `2. 通用规范`，把请求头、错误结构、分页结构、隔离规则一次看清楚。
3. 最后按具体能力进入对应章节，不要来回跳着翻。

### 0.3 文档范围

这份文档现在统一覆盖下面这些内容：

- OpenAI 兼容对话接口
- 视频生成、状态查询、下载
- 图片生成
- 真人资产
- 虚拟人像资产
- 文件夹分组
- 控制台创建 API Key 和资产的操作流程

## 1. 快速开始

### 1.1 鉴权

所有 OpenAI 兼容接口和资产接口均使用 Bearer Token 鉴权。

```http
Authorization: Bearer sk-你的APIKey
```

请求头填写规则：

| Header          | 必填           | 示例                 | 说明                |
| --------------- | ------------ | ------------------ | ----------------- |
| `Authorization` | 是            | `Bearer sk-xxxx`   | `Bearer` 后必须有一个空格 |
| `Content-Type`  | POST JSON 必填 | `application/json` | JSON 请求体使用        |
| `Accept`        | 建议           | `application/json` | 建议所有 JSON 接口填写    |

### 1.2 常用接口清单

| 能力                  | 方法       | 路径                                         |
| ------------------- | -------- | ------------------------------------------ |
| 对话模型                | `POST`   | `/v1/chat/completions`                     |
| 模型列表                | `GET`    | `/v1/models`                               |
| 视频任务提交              | `POST`   | `/v1/video/generations`                    |
| 视频任务提交，OpenAI 兼容    | `POST`   | `/v1/videos`                               |
| 视频任务状态查询，业务格式       | `GET`    | `/v1/video/generations/{task_id}`          |
| 视频任务状态查询，OpenAI 兼容  | `GET`    | `/v1/videos/{task_id}`                     |
| 视频任务取消或删除，业务格式      | `DELETE` | `/v1/video/generations/{task_id}`          |
| 视频任务取消或删除，OpenAI 兼容 | `DELETE` | `/v1/videos/{task_id}`                     |
| 视频文件下载              | `GET`    | `/v1/videos/{task_id}/content`             |
| 图片生成                | `POST`   | `/v1/images/generations`                   |
| 真人资产配置              | `GET`    | `/api/portrait_assets/official/config`     |
| 真人资产任务列表            | `GET`    | `/api/portrait_assets/official/jobs`       |
| 创建真人资产任务            | `POST`   | `/api/portrait_assets/official/jobs`       |
| 真人资产删除              | `DELETE` | `/api/portrait_assets/official/jobs/{id}`  |
| 虚拟人像配置              | `GET`    | `/api/portrait_assets/virtual/config`      |
| 虚拟人像资产组             | `GET`    | `/api/portrait_assets/virtual/group`       |
| 虚拟人像资产列表            | `GET`    | `/api/portrait_assets/virtual/assets`      |
| 创建虚拟人像资产            | `POST`   | `/api/portrait_assets/virtual/assets`      |
| 虚拟人像资产删除            | `DELETE` | `/api/portrait_assets/virtual/assets/{id}` |
| 资产文件夹列表             | `GET`    | `/api/portrait_assets/folders`             |
| 创建资产文件夹             | `POST`   | `/api/portrait_assets/folders`             |
| 重命名资产文件夹            | `PATCH`  | `/api/portrait_assets/folders/{folder_id}` |
| 删除资产文件夹             | `DELETE` | `/api/portrait_assets/folders/{folder_id}` |
| 移动资产到文件夹            | `POST`   | `/api/portrait_assets/folders/move`        |

### 1.3 对话完整接入示例

```bash
curl https://8liangai.com/v1/chat/completions \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "DeepSeek",
    "messages": [
      {
        "role": "system",
        "content": "You are a concise assistant."
      },
      {
        "role": "user",
        "content": "请只返回：HELLO-8LIANGAI"
      }
    ],
    "temperature": 0,
    "max_tokens": 128,
    "stream": false
  }'
```

接入建议：

- 对话模型统一使用 `DeepSeek`
- 如果接入侧不展示推理过程，可以直接忽略 `reasoning_content`

### 1.4 视频完整接入示例

```JSON
curl -X POST 'https://8liangai.com/v1/video/generations' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "seedance2",
    "prompt": "一只小熊猫在干净的摄影棚里挥手，镜头稳定，动作自然，无文字",
    "resolution": "720p",
    "ratio": "16:9",
    "duration": 5,
    "external_user_id": "sub-account-001",
    "metadata": {
      "watermark": false,
      "generate_audio": false
    }
  }'
```

提交成功后返回 `id` 和 `task_id`，后续查询和下载均使用该任务 ID。

***

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

| 字段              | 类型       | 说明                                                             |
| --------------- | -------- | -------------------------------------------------------------- |
| `error.message` | `string` | 可展示或记录的错误说明                                                    |
| `error.type`    | `string` | 错误类型，例如 `invalid_request_error`、`server_error`、`new_api_error` |
| `error.code`    | `string` | 上游或网关错误码，部分错误不返回该字段                                            |

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

| 字段        | 类型        | 说明                      |
| --------- | --------- | ----------------------- |
| `success` | `boolean` | 业务是否成功                  |
| `message` | `string`  | 成功时为空字符串，失败时为错误说明       |
| `data`    | `any`     | 成功时的数据，可能为对象、数组或 `null` |

### 2.3 分页结构

列表接口统一使用分页包装。

请求参数：

| 参数          | 类型        | 必填 | 默认值   | 最大值   | 说明                             |
| ----------- | --------- | -- | ----- | ----- | ------------------------------ |
| `p`         | `integer` | 否  | `1`   | 无固定上限 | 页码，从 `1` 开始                    |
| `page_size` | `integer` | 否  | 站点默认值 | `100` | 每页数量                           |
| `ps`        | `integer` | 否  | 无     | `100` | 兼容字段，未传 `page_size` 时可用        |
| `size`      | `integer` | 否  | 无     | `100` | 兼容字段，未传 `page_size` 和 `ps` 时可用 |

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

| 字段               | 类型        | 说明     |
| ---------------- | --------- | ------ |
| `data.page`      | `integer` | 当前页码   |
| `data.page_size` | `integer` | 当前每页数量 |
| `data.total`     | `integer` | 总记录数   |
| `data.items`     | `array`   | 当前页数据  |

### 2.4 `external_user_id` 隔离规则

同一个 API Key 下如果需要管理多个下游账号，创建资产和生成视频必须传同一个 `external_user_id`。

| 场景       | 填写位置                                                           | 说明         |
| -------- | -------------------------------------------------------------- | ---------- |
| 创建真人资产任务 | `POST /api/portrait_assets/official/jobs` body                 | 写入该真人资产任务  |
| 查询真人资产任务 | `GET /api/portrait_assets/official/jobs?external_user_id=xxx`  | 只看该隔离空间    |
| 创建虚拟资产   | `POST /api/portrait_assets/virtual/assets` body                | 写入该虚拟资产    |
| 查询虚拟资产   | `GET /api/portrait_assets/virtual/assets?external_user_id=xxx` | 只看该隔离空间    |
| 创建文件夹    | `POST /api/portrait_assets/folders` body                       | 文件夹按隔离空间保存 |
| 视频生成     | 顶层 `external_user_id`，兼容 `metadata.external_user_id`           | 顶层优先       |

规则：

| 规则   | 说明                                               |
| ---- | ------------------------------------------------ |
| 空值   | 代表 API Key 账号级历史资产                               |
| 非空值  | 只能引用同一个 `external_user_id` 下的资产                  |
| 严格隔离 | 传了 `external_user_id` 后，不能引用空值历史资产，也不能引用其他隔离空间资产 |
| 长度限制 | 最长 128 个字符                                       |

### 2.5 计费说明

视频任务一般有预扣、结算、退款差额三个阶段。

| 名称   | 含义                                         |
| ---- | ------------------------------------------ |
| 预扣   | 提交任务时按模型上限预先冻结或扣减额度                        |
| 实扣   | 任务完成后根据上游 `usage.total_tokens` 和模型倍率计算实际额度 |
| 退款差额 | `预扣额度 - 实扣额度`，正数表示退回，负数表示补扣                |

图片任务当前按次计费，成功生成后按模型单价扣费；参数非法未生成时不扣费。

***

## 3. 模型能力

### 3.1 对话模型

DeepSeek 对话接入能力如下：

| 模型名        | 请求接口                   | 说明      |
| ---------- | ---------------------- | ------- |
| `DeepSeek` | `/v1/chat/completions` | 推荐正式接入名 |

接入建议：

- 统一使用 `DeepSeek` 作为模型名
- 使用 `/v1/models` 获取模型列表
- 默认只消费 `choices[].message.content`

### 3.2 DeepSeek 对话接口

接口信息：

```http
POST /v1/chat/completions
```

模型名：

| 字段      | 值          | 说明                     |
| ------- | ---------- | ---------------------- |
| `model` | `DeepSeek` | 当前 DeepSeek 对话模型统一使用该值 |

请求头：

| Header          | 必填 | 示例                 | 说明         |
| --------------- | -- | ------------------ | ---------- |
| `Authorization` | 是  | `Bearer sk-xxxx`   | API Key 鉴权 |
| `Content-Type`  | 是  | `application/json` | JSON 请求体   |

请求体示例：

```json
{
  "model": "DeepSeek",
  "messages": [
    {
      "role": "system",
      "content": "You are a concise assistant."
    },
    {
      "role": "user",
      "content": "请用一句话介绍你自己。"
    }
  ],
  "temperature": 0.3,
  "max_tokens": 256,
  "stream": false
}
```

请求体字段说明：

| 字段                  | 类型                         | 必填 | 示例                | 说明               |
| ------------------- | -------------------------- | -- | ----------------- | ---------------- |
| `model`             | `string`                   | 是  | `DeepSeek`        | 模型名              |
| `messages`          | `array<object>`            | 是  | 见下表               | 对话消息数组           |
| `temperature`       | `number`                   | 否  | `0.3`             | 采样温度，越低越稳定       |
| `max_tokens`        | `integer`                  | 否  | `256`             | 限制本次最大输出 token 数 |
| `stream`            | `boolean`                  | 否  | `false`           | 是否启用流式返回         |
| `top_p`             | `number`                   | 否  | `1`               | 核采样参数            |
| `presence_penalty`  | `number`                   | 否  | `0`               | 提高模型引入新内容的倾向     |
| `frequency_penalty` | `number`                   | 否  | `0`               | 降低重复表达的倾向        |
| `stop`              | `string` 或 `array<string>` | 否  | `["</answer>"]`   | 停止序列             |
| `user`              | `string`                   | 否  | `sub-account-001` | 业务侧自定义标识         |

`messages` 数组元素字段说明：

| 字段        | 类型       | 必填 | 示例                          | 说明   |
| --------- | -------- | -- | --------------------------- | ---- |
| `role`    | `string` | 是  | `system`、`user`、`assistant` | 消息角色 |
| `content` | `string` | 是  | `请生成一段介绍文案`                 | 消息内容 |

参数填写建议：

| 目标      | 推荐参数                      |
| ------- | ------------------------- |
| 稳定输出    | `temperature=0` 到 `0.3`   |
| 日常问答    | `temperature=0.5` 到 `0.8` |
| 限制输出长度  | 显式填写 `max_tokens`         |
| 需要打字机效果 | `stream=true`             |

非流式响应示例：

```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1781080000,
  "model": "DeepSeek",
  "choices": [
    {
      "index": 0,
      "finish_reason": "stop",
      "message": {
        "role": "assistant",
        "content": "你好，我是 DeepSeek 助手。",
        "reasoning_content": "推理内容示例"
      }
    }
  ],
  "usage": {
    "prompt_tokens": 16,
    "completion_tokens": 20,
    "total_tokens": 36,
    "prompt_tokens_details": {
      "cached_tokens": 0
    },
    "completion_tokens_details": {
      "reasoning_tokens": 16
    }
  }
}
```

流式处理规则：

- 使用 SSE 方式接收数据
- 仅拼接 `delta.content` 作为最终正文
- 若存在 `delta.reasoning_content`，按产品需求决定是否展示
- 在最终带 `usage` 的 chunk 后记录 token 消耗
- 收到 `[DONE]` 后关闭流

常见错误：

| 场景                          | 常见原因               |
| --------------------------- | ------------------ |
| `401 Unauthorized`          | API Key 无效、缺失或格式错误 |
| `400 invalid_request_error` | 请求体字段类型错误、缺少必填参数   |
| `finish_reason=length`      | `max_tokens` 设置过小  |

### 3.3 视频模型

| 模型                   | 类型     | 请求接口                    | 状态查询接口                                                   | 下载接口                           |
| -------------------- | ------ | ----------------------- | -------------------------------------------------------- | ------------------------------ |
| `seedance1.5`        | 视频生成   | `/v1/video/generations` | `/v1/video/generations/{task_id}`、`/v1/videos/{task_id}` | `/v1/videos/{task_id}/content` |
| `seedance1.5-sr`     | 视频生成   | 同上                      | 同上                                                       | 同上                             |
| `seedance2`          | 视频生成   | 同上                      | 同上                                                       | 同上                             |
| `sd2.0fast`          | 快速视频生成 | 同上                      | 同上                                                       | 同上                             |
| `seedance2-sr`       | 视频生成   | 同上                      | 同上                                                       | 同上                             |
| `seedance2.0fast-sr` | 快速视频生成 | 同上                      | 同上                                                       | 同上                             |

### 3.4 视频分辨率支持

| 模型                   | 支持的 `resolution`      | 推荐默认值  | 说明            |
| -------------------- | --------------------- | ------ | ------------- |
| `seedance1.5`        | `480p`、`720p`、`1080p` | `720p` | <br />        |
| `seedance1.5-sr`     | `720p`、`1080p`        | `720p` | 高清输出          |
| `seedance2`          | `480p`、`720p`、`1080p` | `720p` | 1080p 使用独立计费档 |
| `sd2.0fast`          | `480p`、`720p`         | `720p` | 快速模型          |
| `seedance2-sr`       | `720p`、`1080p`        | `720p` | 高清输出          |
| `seedance2.0fast-sr` | `720p`、`1080p`        | `720p` | 高清输出          |

### 3.5 视频画幅比例

当前文档按站点已验证能力和火山 Seedance 官方文档整理。常规接入建议显式填写 `ratio`。

| `ratio`    | 含义     | 固定尺寸校验         |
| ---------- | ------ | -------------- |
| `16:9`     | 横版宽屏   | 支持             |
| `4:3`      | 横版传统比例 | 支持             |
| `1:1`      | 正方形    | 支持             |
| `3:4`      | 竖版海报比例 | 支持             |
| `9:16`     | 竖屏短视频  | 支持             |
| `21:9`     | 超宽屏    | 支持             |
| `adaptive` | 自适应比例  | 支持提交，不使用固定宽高断言 |

`adaptive` 说明：

| 场景     | 行为                    |
| ------ | --------------------- |
| 文生视频   | 模型根据提示词自动选择画幅         |
| 图生视频   | 通常优先参考输入图片比例          |
| 视频参考输入 | 通常优先参考输入视频比例          |
| 生产接入   | 建议优先使用明确比例，避免上下游展示不可控 |

官方参考：火山方舟 Seedance 2.0 API 文档说明 `ratio=adaptive` 时会自动适配宽高比，并可在任务查询结果中查看实际 `ratio`。

### 3.6 常见输出像素尺寸

下表为当前线上回归已验证的典型输出尺寸。`adaptive` 不列固定尺寸。

| `resolution` | `ratio` | 输出像素        |
| ------------ | ------- | ----------- |
| `480p`       | `16:9`  | `864x496`   |
| `480p`       | `4:3`   | `752x560`   |
| `480p`       | `1:1`   | `640x640`   |
| `480p`       | `3:4`   | `560x752`   |
| `480p`       | `9:16`  | `496x864`   |
| `480p`       | `21:9`  | `992x432`   |
| `720p`       | `16:9`  | `1280x720`  |
| `720p`       | `4:3`   | `1112x834`  |
| `720p`       | `1:1`   | `960x960`   |
| `720p`       | `3:4`   | `834x1112`  |
| `720p`       | `9:16`  | `720x1280`  |
| `720p`       | `21:9`  | `1470x630`  |
| `1080p`      | `16:9`  | `1920x1080` |
| `1080p`      | `4:3`   | `1664x1248` |
| `1080p`      | `1:1`   | `1440x1440` |
| `1080p`      | `3:4`   | `1248x1664` |
| `1080p`      | `9:16`  | `1080x1920` |
| `1080p`      | `21:9`  | `2206x946`  |

### 3.7 图片模型

| 模型                | 接口                       | 当前有效尺寸      |
| ----------------- | ------------------------ | ----------- |
| `seedream4.5`     | `/v1/images/generations` | `2048x2048` |
| `seedream5.0lite` | `/v1/images/generations` | `2048x2048` |

`1024x1024` 对这两个模型返回参数错误且不扣费。生产接入建议使用 `2048x2048` 起步。

***

## 4. 视频任务提交

### 4.1 接口信息

```http
POST /v1/video/generations
POST /v1/videos
```

推荐使用 `/v1/video/generations`。`/v1/videos` 为 OpenAI 兼容创建入口，当前请求体相同。

### 4.2 请求体字段

| 字段                 | 类型                  | 必填 | 默认值           | 可选值或格式                                            | 说明                                    |
| ------------------ | ------------------- | -- | ------------- | ------------------------------------------------- | ------------------------------------- |
| `model`            | `string`            | 是  | 无             | 见视频模型表                                            | 模型别名                                  |
| `prompt`           | `string`            | 是  | 无             | 自然语言文本                                            | 提示词，会作为 `content` 中的文本发送上游            |
| `resolution`       | `string`            | 否  | 多数模型默认 `720p` | `480p`、`720p`、`1080p`                             | 输出分辨率档位，按模型能力表填写                      |
| `ratio`            | `string`            | 否  | 上游默认值         | `16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive` | 输出画幅比例                                |
| `duration`         | `integer` 或数字字符串    | 否  | 上游默认值         | `5` 等正整数                                          | 视频时长，推荐传整数                            |
| `seconds`          | `string`            | 否  | 无             | `"5"` 等数字字符串                                      | 时长兼容字段；`duration` 优先                  |
| `images`           | `array<string>`     | 否  | 空数组           | 公网可访问图片 URL                                       | 图生视频参考图，推荐使用该字段                       |
| `image`            | `string`            | 否  | 无             | 公网可访问图片 URL                                       | 兼容字段；当前 Doubao 视频链路优先使用 `images`      |
| `size`             | `string`            | 否  | 无             | 兼容字段                                              | 视频模型建议用 `resolution`，不要用 `size` 表达清晰度 |
| `mode`             | `string`            | 否  | 无             | 兼容字段                                              | 当前文档覆盖模型通常不需要填写                       |
| `input_reference`  | `string`            | 否  | 无             | 兼容字段                                              | 当前文档覆盖模型通常不需要填写                       |
| `external_user_id` | `string`            | 否  | 无             | 最长 128 个字符                                        | 多下游账号隔离字段；绑定人像资产时必须和资产创建时一致           |
| `metadata`         | `object` 或 JSON 字符串 | 否  | `{}`          | 见下表                                               | 高级扩展字段                                |

填写建议：

| 场景     | 推荐字段组合                                                     |
| ------ | ---------------------------------------------------------- |
| 文生视频   | `model`、`prompt`、`resolution`、`ratio`、`duration`           |
| 图生视频   | 文生视频字段加 `images`                                           |
| 视频参考输入 | 文生视频字段加 `metadata.content`                                 |
| 真人资产绑定 | 文生视频字段加 `metadata.portrait_asset_id` 或 `metadata.asset_id` |

模型能力差异建议：

| 参数                  | 建议                    |
| ------------------- | --------------------- |
| `generate_audio`    | 仅在支持音频生成的模型上使用        |
| `draft`             | 仅在支持样片模式的模型上使用        |
| `tools`             | 仅在支持工具调用的模型上使用        |
| `frames`            | 仅在支持帧数控制的模型上使用        |
| `1080p`             | 仅在支持该分辨率的模型上使用        |
| `camera_fixed`      | 仅在支持固定机位且不是参考图限制场景时使用 |
| `service_tier=flex` | 仅在支持离线推理的模型上使用        |

### 4.3 `metadata` 字段

`metadata` 会映射到上游视频任务 payload。常用字段如下：

| 字段                        | 类型               | 必填 | 可选值或格式                            | 说明                            |
| ------------------------- | ---------------- | -- | --------------------------------- | ----------------------------- |
| `callback_url`            | `string`         | 否  | `https://your-domain/callback`    | 上游任务回调地址                      |
| `return_last_frame`       | `boolean`        | 否  | `true`、`false`                    | 是否返回最后一帧；可用于连续视频衔接            |
| `service_tier`            | `string`         | 否  | 上游支持的服务层级                         | 通常不填                          |
| `execution_expires_after` | `integer`        | 否  | 秒数                                | 任务执行过期时间                      |
| `generate_audio`          | `boolean`        | 否  | `true`、`false`                    | 是否生成音频，按上游模型能力生效              |
| `draft`                   | `boolean`        | 否  | `true`、`false`                    | 草稿模式                          |
| `tools`                   | `array<object>`  | 否  | `[{ "type": "..." }]`             | 上游工具参数                        |
| `resolution`              | `string`         | 否  | 同顶层 `resolution`                  | metadata 中该字段优先于顶层字段          |
| `ratio`                   | `string`         | 否  | 同顶层 `ratio`                       | metadata 中该字段优先于顶层字段          |
| `duration`                | `integer`        | 否  | 正整数                               | metadata 中该字段可映射上游时长          |
| `frames`                  | `integer`        | 否  | 正整数                               | 指定帧数                          |
| `seed`                    | `integer`        | 否  | 整数                                | 随机种子                          |
| `camera_fixed`            | `boolean`        | 否  | `true`、`false`                    | 是否固定机位                        |
| `watermark`               | `boolean`        | 否  | `true`、`false`                    | 是否添加水印                        |
| `external_user_id`        | `string`         | 否  | 最长 128 个字符                        | 兼容写法；顶层 `external_user_id` 优先 |
| `content`                 | `array<object>`  | 否  | 见下一节                              | 参考媒体内容                        |
| `portrait_asset_id`       | `integer` 或整数字符串 | 否  | 真人资产任务 ID                         | 绑定当前账号已就绪真人资产任务               |
| `asset_id`                | `string`         | 否  | `asset_xxx` 或 `asset://asset_xxx` | 绑定当前账号已就绪真人资产                 |

优先级规则：

| 参数                           | 优先级                    |
| ---------------------------- | ---------------------- |
| `metadata.resolution`        | 高于顶层 `resolution`      |
| `metadata.ratio`             | 高于顶层 `ratio`           |
| 顶层 `duration`                | 高于 `seconds`           |
| `metadata.portrait_asset_id` | 高于 `metadata.asset_id` |

### 4.4 `metadata.content` 字段

`metadata.content` 用于精确传入图片、视频、音频参考内容。

支持的输入组合：

| 组合类型              | 说明                                  |
| ----------------- | ----------------------------------- |
| 文本                | 只传文本提示词                             |
| 文本 + 图片           | 文本配合一个或多个图片参考                       |
| 文本 + 视频           | 文本配合视频参考                            |
| 文本 + 图片 + 音频      | 文本配合图片和音频参考                         |
| 文本 + 图片 + 视频      | 文本配合图片和视频参考                         |
| 文本 + 视频 + 音频      | 文本配合视频和音频参考                         |
| 文本 + 图片 + 视频 + 音频 | 文本配合多种媒体参考                          |
| 样片任务 ID           | 使用已成功生成的样片任务作为参考输入；是否开放以当前模型和上游能力为准 |

填写建议：

- 普通接入优先使用 `prompt` + `metadata.content`。
- 如果需要更稳定地控制人物、场景、镜头和声音引用顺序，建议把所有参考素材显式写进 `metadata.content`。
- 样片任务 ID 属于上游高级能力；如果当前模型映射或站点能力未开放，不建议默认依赖。

兼容规则：

- 只传顶层 `images` 时，网关会自动转成 `reference_image`。
- 同时传顶层 `images` 和 `metadata.content` 时，网关会保留 `metadata.content` 原有顺序，再把顶层 `images` 追加到后面。
- 多人物、多资产、多参考图场景，推荐直接把全部素材按最终顺序写进 `metadata.content`，不要只依赖顶层 `images`。

图序映射规则：

- 提示词里如果写 `图1`、`图2`、`图3`，默认按最终参考素材顺序理解。
- `metadata.content` 里的内容先编号，再继续编号顶层 `images`。
- 也就是说：
  - `metadata.content[0]` = `图1`
  - `metadata.content[1]` = `图2`
  - `metadata.content[2]` = `图3`
  - `images[0]` = `图4`
- 如果提示词里的图序描述和实际传入顺序不一致，最终结果会明显不稳定，容易出现人物、场景、动作引用错位。

内容项字段：

| 字段              | 类型       | 必填                   | 可选值                                                   | 说明                          |
| --------------- | -------- | -------------------- | ----------------------------------------------------- | --------------------------- |
| `type`          | `string` | 是                    | `text`、`image_url`、`video_url`、`audio_url`            | 内容类型                        |
| `role`          | `string` | 按类型                  | `reference_image`、`reference_video`、`reference_audio` | 参考媒体角色                      |
| `text`          | `string` | `type=text` 时使用      | 文本                                                    | 文本内容                        |
| `image_url.url` | `string` | `type=image_url` 时使用 | 图片 URL 或 `asset://<ASSET_ID>`                         | 图片必须可被上游访问；虚拟人像图片资产可填资产 URI |
| `video_url.url` | `string` | `type=video_url` 时使用 | 视频 URL 或 `asset://<ASSET_ID>`                         | 视频必须可被上游访问；虚拟人像视频资产可填资产 URI |
| `audio_url.url` | `string` | `type=audio_url` 时使用 | 音频 URL 或 `asset://<ASSET_ID>`                         | 音频必须可被上游访问；虚拟人像音频资产可填资产 URI |

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

音频参考输入标准写法：

```json
{
  "type": "audio_url",
  "role": "reference_audio",
  "audio_url": {
    "url": "https://example.com/background.mp3"
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

| 资产类型 | `content.type` | URL 字段          | 填法                        |
| ---- | -------------- | --------------- | ------------------------- |
| 图片资产 | `image_url`    | `image_url.url` | `asset://<volc_asset_id>` |
| 视频资产 | `video_url`    | `video_url.url` | `asset://<volc_asset_id>` |
| 音频资产 | `audio_url`    | `audio_url.url` | `asset://<volc_asset_id>` |

注意：`metadata.portrait_asset_id` 是真人资产任务 ID 的绑定方式；虚拟人像资产不要填到 `portrait_asset_id`，应通过 `metadata.content` 使用 `asset://<volc_asset_id>` 引用。

混合引用建议：

- 多人物、多资产、多参考图场景，建议把需要强绑定图序的人像素材全部放进 `metadata.content`。
- 顶层 `images` 更适合补充普通外部图片参考，例如道具图、场景图、动物图、分镜补充图。
- 如果一个请求里同时用了真人资产、虚拟资产、外部图片，建议按最终镜头描述顺序组织素材，再写提示词。

`return_last_frame` 说明：

- `true`：返回生成视频的尾帧图像。
- `false`：不返回尾帧图像。
- 开启后，可在任务查询结果中获取尾帧图信息；尾帧格式通常为 `png`，宽高与生成视频一致，不带水印。
- 这个能力适合做连续视频衔接：上一段视频的尾帧可以作为下一段视频的首帧参考，用来快速拼接多段连续视频。
- 当前站点查询接口里，稳定保证返回的是视频结果 URL；如果上游返回尾帧相关字段，原始上游数据会保留在 `GET /v1/video/generations/{task_id}` 的 `data.data` 中，实际字段名以上游返回为准。
- 如果要对尾帧做程序化依赖，建议先联调一次真实任务，确认当前模型和上游返回体里的尾帧字段结构。

常见模型能力差异：

| 参数                        | 常见限制                                                                   |
| ------------------------- | ---------------------------------------------------------------------- |
| `generate_audio`          | 常见于 `Seedance 2.0`、`Seedance 2.0 fast`、`Seedance 1.5 pro`；其他模型不要默认认为支持 |
| `draft`                   | 常见于 `Seedance 1.5 pro` 的样片模式；其他模型不要默认传                                 |
| `tools`                   | 常见于 `Seedance 2.0`、`Seedance 2.0 fast`；未确认前不要默认传                       |
| `service_tier`            | `flex` 属于离线模式；并非所有模型都支持，`Seedance 2.0`、`Seedance 2.0 fast` 不要默认按离线模式接入 |
| `execution_expires_after` | 建议按上游允许范围填写，超出范围可能被拒绝                                                  |
| `resolution=1080p`        | 并非所有模型和所有参考场景都支持                                                       |
| `ratio=adaptive`          | 不是所有模型、所有输入场景的默认行为都一致                                                  |
| `frames`                  | 并非所有模型都支持；不支持时优先用 `duration`                                           |
| `camera_fixed`            | 参考图场景和部分模型可能不支持                                                        |

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

真人资产 + 虚拟资产 + 外部图片混合示例：

这个示例里：

- `图1` = 第一个虚拟资产
- `图2` = 第二个虚拟资产
- `图3` = 真人资产
- `图4` = 顶层 `images[0]` 里的外部猫图

```json
{
  "model": "sd2.0fast",
  "prompt": "图1和图2中的两个人物正在激烈争吵，先给图1人物一个近景镜头，再切到图2人物一个近景镜头；图3中的人物走到两人中间试图劝阻，给图3人物一个清晰近景镜头；图4中的猫出现在旁边看着他们，给猫一个清晰镜头；镜头在图1、图2、图3、图4之间切换，突出争吵、劝阻和围观关系，人物和猫都要贴合参考素材，动作自然，真人感，8秒，竖屏。",
  "duration": 8,
  "resolution": "480p",
  "ratio": "9:16",
  "metadata": {
    "watermark": false,
    "generate_audio": false,
    "content": [
      {
        "type": "image_url",
        "role": "reference_image",
        "image_url": {
          "url": "asset://asset_virtual_a"
        }
      },
      {
        "type": "image_url",
        "role": "reference_image",
        "image_url": {
          "url": "asset://asset_virtual_b"
        }
      },
      {
        "type": "image_url",
        "role": "reference_image",
        "image_url": {
          "url": "asset://asset_official_ready"
        }
      }
    ]
  },
  "images": [
    "data:image/jpeg;base64,......"
  ]
}
```

这个写法适用的场景：

- 虚拟资产和真人资产一起出现在同一个视频里
- 还要再补一张外部图片参考
- 提示词里需要稳定使用 `图1`、`图2`、`图3`、`图4`

常见错误：

- 把虚拟资产写到 `metadata.portrait_asset_id` 里
- 提示词里写“图3 是猫”，但猫实际上放在 `images[0]`，最终编号是 `图4`
- 把需要强绑定顺序的人像素材散落在顶层 `images` 里，导致图序不稳定

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

| 字段                      | 类型        | 说明                  |
| ----------------------- | --------- | ------------------- |
| `id`                    | `string`  | 对外公开任务 ID           |
| `task_id`               | `string`  | 与 `id` 相同，兼容旧调用端    |
| `object`                | `string`  | 固定为 `video`         |
| `model`                 | `string`  | 请求使用的模型别名           |
| `status`                | `string`  | 初始一般为 `queued`      |
| `progress`              | `integer` | 任务进度，初始通常为 `0`      |
| `created_at`            | `integer` | Unix 秒级时间戳          |
| `completed_at`          | `integer` | 完成时间，提交阶段通常不返回      |
| `expires_at`            | `integer` | 过期时间，存在时返回          |
| `seconds`               | `string`  | 兼容字段，存在时返回          |
| `size`                  | `string`  | 兼容字段，存在时返回          |
| `remixed_from_video_id` | `string`  | Remix 来源视频 ID，存在时返回 |
| `error`                 | `object`  | 失败时返回，提交成功通常为空      |
| `metadata`              | `object`  | 附加信息，存在时返回          |

***

## 5. 视频状态查询

### 5.1 业务格式查询

```http
GET /v1/video/generations/{task_id}
```

路径参数：

| 参数        | 类型       | 必填 | 说明                  |
| --------- | -------- | -- | ------------------- |
| `task_id` | `string` | 是  | 视频提交接口返回的 `task_id` |

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

| 字段        | 类型       | 说明             |
| --------- | -------- | -------------- |
| `code`    | `string` | 成功时为 `success` |
| `message` | `string` | 消息文本，可能为空      |
| `data`    | `object` | 任务记录           |

`data` 字段说明：

| 字段                               | 类型                  | 说明                          |
| -------------------------------- | ------------------- | --------------------------- |
| `id`                             | `integer`           | 本地任务记录 ID                   |
| `created_at`                     | `integer`           | 本地创建时间                      |
| `updated_at`                     | `integer`           | 本地更新时间                      |
| `task_id`                        | `string`            | 对外任务 ID                     |
| `platform`                       | `string` 或 `number` | 任务平台标识                      |
| `user_id`                        | `integer`           | 账号 ID                       |
| `group`                          | `string`            | 计费分组                        |
| `channel_id`                     | `integer`           | 通道 ID，部分响应可能省略              |
| `quota`                          | `integer`           | 当前记录额度                      |
| `action`                         | `string`            | 任务动作，一般为 `generate`         |
| `status`                         | `string`            | 业务状态                        |
| `fail_reason`                    | `string`            | 失败原因或历史兼容结果字段               |
| `result_url`                     | `string`            | 成功后的最终视频 URL                |
| `submit_time`                    | `integer`           | 提交时间                        |
| `start_time`                     | `integer`           | 开始时间                        |
| `finish_time`                    | `integer`           | 完成时间                        |
| `progress`                       | `string`            | 进度字符串，例如 `10%`、`50%`、`100%` |
| `properties.input`               | `string`            | 原始输入摘要                      |
| `properties.upstream_model_name` | `string`            | 上游模型名                       |
| `properties.origin_model_name`   | `string`            | 请求方请求模型名                    |
| `data`                           | `object`            | 上游任务原始数据或后处理包装数据            |

业务状态值：

| 状态            | 含义           |
| ------------- | ------------ |
| `NOT_START`   | 本地任务已创建但尚未提交 |
| `SUBMITTED`   | 已提交到上游       |
| `QUEUED`      | 上游排队中        |
| `IN_PROGRESS` | 上游处理中        |
| `SUCCESS`     | 任务成功         |
| `FAILURE`     | 任务失败         |
| `UNKNOWN`     | 未识别状态        |

尾帧查询说明：

- 如果创建任务时传了 `metadata.return_last_frame=true`，可以优先查询这个业务格式接口。
- 当前站点稳定保证返回的视频结果字段是 `data.result_url` 和 `data.data.content.video_url`。
- 尾帧图如果上游返回，会保留在 `data.data` 这段原始上游任务数据里。
- 由于不同模型和不同上游返回体的尾帧字段名可能不同，当前不要写死单一字段名，建议以实际联调返回为准。
- 如果业务必须稳定消费尾帧字段，建议先固定模型，再拿一条真实任务返回体做字段对齐。

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

| 字段                      | 类型        | 说明                  |
| ----------------------- | --------- | ------------------- |
| `id`                    | `string`  | 任务 ID               |
| `task_id`               | `string`  | 兼容字段，与 `id` 相同      |
| `object`                | `string`  | 固定为 `video`         |
| `model`                 | `string`  | 请求方请求模型别名           |
| `status`                | `string`  | OpenAI 风格状态         |
| `progress`              | `integer` | 数字进度                |
| `created_at`            | `integer` | 创建时间                |
| `completed_at`          | `integer` | 完成时间，完成或失败时返回       |
| `expires_at`            | `integer` | 过期时间，存在时返回          |
| `seconds`               | `string`  | 兼容字段，存在时返回          |
| `size`                  | `string`  | 兼容字段，存在时返回          |
| `remixed_from_video_id` | `string`  | Remix 来源视频 ID，存在时返回 |
| `error`                 | `object`  | 失败时返回               |
| `metadata.url`          | `string`  | 成功后的视频 URL          |

OpenAI 兼容状态值：

| 状态            | 含义    |
| ------------- | ----- |
| `queued`      | 排队中   |
| `in_progress` | 处理中   |
| `completed`   | 已完成   |
| `failed`      | 已失败   |
| `unknown`     | 未识别状态 |

补充说明：

- 当前 OpenAI 兼容查询接口稳定保证的是 `metadata.url` 视频地址。
- 如果要读取尾帧图信息，优先使用 `GET /v1/video/generations/{task_id}` 业务格式查询。

错误对象字段：

| 字段              | 类型       | 说明   |
| --------------- | -------- | ---- |
| `error.message` | `string` | 失败原因 |
| `error.code`    | `string` | 错误码  |

### 5.3 取消或删除视频任务

```http
DELETE /v1/video/generations/{task_id}
DELETE /v1/videos/{task_id}
```

这两个路径作用相同，推荐优先使用 `DELETE /v1/video/generations/{task_id}`。

处理规则：

| 当前任务状态        | 是否支持 | 实际行为        | DELETE 后本地结果 |
| ------------- | ---- | ----------- | ------------ |
| `NOT_START`   | 是    | 取消未提交完成的任务  | `cancelled`  |
| `SUBMITTED`   | 是    | 取消排队中的任务    | `cancelled`  |
| `QUEUED`      | 是    | 取消排队中的任务    | `cancelled`  |
| `IN_PROGRESS` | 否    | 不支持取消运行中的任务 | 返回错误         |
| `SUCCESS`     | 是    | 删除任务记录      | 后续不能再查询      |
| `FAILURE`     | 是    | 删除任务记录      | 后续不能再查询      |
| `CANCELLED`   | 否    | 已取消任务不再重复处理 | 返回错误         |

说明：

- 取消排队成功后，会把本地任务状态改成 `cancelled`。
- 取消排队成功后，会退还这条任务的预扣额度。
- 删除已完成或已失败任务后，本地记录会被删除，后续查询会返回任务不存在。
- 运行中的任务当前不支持取消，这是上游限制。

业务格式完整示例：

```bash
curl -X DELETE 'https://8liangai.com/v1/video/generations/task_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -H 'Accept: application/json'
```

成功响应示例，取消排队：

```json
{
  "id": "task_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "task_id": "task_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "object": "video",
  "deleted": false,
  "status": "cancelled"
}
```

成功响应示例，删除记录：

```json
{
  "id": "task_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "task_id": "task_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "object": "video",
  "deleted": true,
  "status": "deleted"
}
```

失败响应示例，任务正在运行：

```json
{
  "error": {
    "message": "running task does not support delete",
    "type": "new_api_error",
    "code": "task_delete_not_supported_for_running"
  }
}
```

***

## 6. 视频文件下载

### 6.1 接口信息

```http
GET /v1/videos/{task_id}/content
```

路径参数：

| 参数        | 类型       | 必填 | 说明             |
| --------- | -------- | -- | -------------- |
| `task_id` | `string` | 是  | 视频提交接口返回的任务 ID |

### 6.2 成功响应

成功时返回视频二进制流，不是 JSON。

| 项目             | 说明                       |
| -------------- | ------------------------ |
| HTTP 状态码       | `200`                    |
| `Content-Type` | 通常为 `video/mp4`，以实际响应头为准 |
| 响应体            | 视频二进制内容                  |
| 缓存头            | 网关会设置公开视频缓存头             |

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

***

## 7. 各视频模型接入说明

### 7.1 `seedance1.5`

| 项目    | 值                                                                |
| ----- | ---------------------------------------------------------------- |
| 请求接口  | `POST /v1/video/generations`                                     |
| 状态查询  | `GET /v1/video/generations/{task_id}`、`GET /v1/videos/{task_id}` |
| 文件下载  | `GET /v1/videos/{task_id}/content`                               |
| 支持分辨率 | `480p`、`720p`、`1080p`                                            |
| 支持比例  | `16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive`                |

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

| 项目    | 值                                                                |
| ----- | ---------------------------------------------------------------- |
| 请求接口  | `POST /v1/video/generations`                                     |
| 状态查询  | `GET /v1/video/generations/{task_id}`、`GET /v1/videos/{task_id}` |
| 文件下载  | `GET /v1/videos/{task_id}/content`                               |
| 支持分辨率 | `720p`、`1080p`                                                   |
| 支持比例  | `16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive`                |

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

| 项目    | 值                                                                |
| ----- | ---------------------------------------------------------------- |
| 请求接口  | `POST /v1/video/generations`                                     |
| 状态查询  | `GET /v1/video/generations/{task_id}`、`GET /v1/videos/{task_id}` |
| 文件下载  | `GET /v1/videos/{task_id}/content`                               |
| 支持分辨率 | `480p`、`720p`、`1080p`                                            |
| 支持比例  | `16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive`                |

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

| 项目    | 值                                                                |
| ----- | ---------------------------------------------------------------- |
| 请求接口  | `POST /v1/video/generations`                                     |
| 状态查询  | `GET /v1/video/generations/{task_id}`、`GET /v1/videos/{task_id}` |
| 文件下载  | `GET /v1/videos/{task_id}/content`                               |
| 支持分辨率 | `480p`、`720p`                                                    |
| 支持比例  | `16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive`                |

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

| 项目    | 值                                                                |
| ----- | ---------------------------------------------------------------- |
| 请求接口  | `POST /v1/video/generations`                                     |
| 状态查询  | `GET /v1/video/generations/{task_id}`、`GET /v1/videos/{task_id}` |
| 文件下载  | `GET /v1/videos/{task_id}/content`                               |
| 支持分辨率 | `720p`、`1080p`                                                   |
| 支持比例  | `16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive`                |

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

| 项目    | 值                                                                |
| ----- | ---------------------------------------------------------------- |
| 请求接口  | `POST /v1/video/generations`                                     |
| 状态查询  | `GET /v1/video/generations/{task_id}`、`GET /v1/videos/{task_id}` |
| 文件下载  | `GET /v1/videos/{task_id}/content`                               |
| 推荐模型名 | `seedance2.0fast-sr`                                             |
| 兼容模型名 | `sd2.0fast-sr`                                                   |
| 支持分辨率 | `720p`、`1080p`                                                   |
| 支持比例  | `16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive`                |

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

***

## 8. 图片生成

### 8.1 接口信息

```http
POST /v1/images/generations
```

### 8.2 请求体字段

| 字段                   | 类型        | 必填 | 默认值   | 可选值或格式                          | 说明                           |
| -------------------- | --------- | -- | ----- | ------------------------------- | ---------------------------- |
| `model`              | `string`  | 是  | 无     | `seedream4.5`、`seedream5.0lite` | 图片模型别名                       |
| `prompt`             | `string`  | 是  | 无     | 自然语言文本                          | 图片提示词                        |
| `n`                  | `integer` | 否  | `1`   | 正整数                             | 生成图片数量；生产接入建议从 `1` 开始        |
| `size`               | `string`  | 是  | 无     | 当前推荐 `2048x2048`                | 图片尺寸                         |
| `quality`            | `string`  | 否  | 上游默认  | 上游支持值                           | 质量参数，当前两个 Seedream 模型通常不需要填写 |
| `response_format`    | `string`  | 否  | `url` | `url`、`b64_json`                | 返回 URL 或 Base64              |
| `style`              | `any`     | 否  | 无     | 上游支持值                           | 风格参数，透传字段                    |
| `user`               | `any`     | 否  | 无     | 字符串或对象                          | 请求方标识透传                      |
| `extra_fields`       | `any`     | 否  | 无     | 对象                              | 上游扩展字段                       |
| `background`         | `any`     | 否  | 无     | 上游支持值                           | 背景参数                         |
| `moderation`         | `any`     | 否  | 无     | 上游支持值                           | 审核参数                         |
| `output_format`      | `any`     | 否  | 上游默认  | `png`、`jpeg` 等                  | 输出格式，按上游支持值生效                |
| `output_compression` | `any`     | 否  | 上游默认  | 数值                              | 输出压缩参数                       |
| `partial_images`     | `any`     | 否  | 无     | 上游支持值                           | 局部图片参数                       |
| `watermark`          | `boolean` | 否  | 上游默认  | `true`、`false`                  | 是否添加水印                       |
| `watermark_enabled`  | `any`     | 否  | 无     | 上游兼容字段                          | 兼容部分上游                       |
| `user_id`            | `any`     | 否  | 无     | 字符串或数字                          | 上游账号 ID 透传                   |
| `image`              | `any`     | 否  | 无     | URL、Base64 或数组                  | 图片输入兼容字段                     |

### 8.3 `size` 标准填法

图片接口的 `size` 字段会透传到火山图片生成接口。当前站点生产接入推荐使用明确像素值，优先使用已验证的 `2048x2048`。

| 写法     | 示例                                  | 适用模型                            | 当前文档口径          | 说明                         |
| ------ | ----------------------------------- | ------------------------------- | --------------- | -------------------------- |
| 明确像素值  | `2048x2048`                         | `seedream4.5`、`seedream5.0lite` | 已线上验证通过         | 最推荐的生产写法                   |
| 档位写法   | `2K`                                | 按上游模型能力                         | 网关可透传           | 新接入前需先在测试环境验证              |
| 档位写法   | `3K`                                | 主要用于支持 3K 的模型                   | 网关可透传           | `seedream5.0lite` 可按上游能力评估 |
| 档位写法   | `4K`                                | 按上游模型能力                         | 网关可透传           | 成本和耗时通常高于 2K               |
| 宽高像素值  | `2048x1152`、`1152x2048`             | 按上游模型能力                         | 网关可透传           | 用于横版或竖版图片                  |
| 正方形像素值 | `2048x2048`、`3072x3072`、`4096x4096` | 按上游模型能力                         | `2048x2048` 已验证 | 更高像素需按价格页和测试结果确认           |

填写规则：

| 规则   | 说明                                    |
| ---- | ------------------------------------- |
| 推荐首选 | `2048x2048`                           |
| 不推荐  | `1024x1024`，会失败                       |
| 横竖版  | 使用 `宽x高`，例如 `2048x1152` 或 `1152x2048` |
| 档位值  | 使用大写 `K`，例如 `2K`、`3K`、`4K`            |
| 生产上线 | 新尺寸上线前必须先真实生成并核对扣费                    |

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

| 字段                      | 类型              | 说明                                       |
| ----------------------- | --------------- | ---------------------------------------- |
| `created`               | `integer`       | 创建时间戳                                    |
| `data`                  | `array<object>` | 图片结果数组                                   |
| `data[].url`            | `string`        | 图片 URL，`response_format=url` 时优先使用       |
| `data[].b64_json`       | `string`        | Base64 图片，`response_format=b64_json` 时使用 |
| `data[].revised_prompt` | `string`        | 上游改写后的提示词，存在时返回                          |
| `metadata`              | `object`        | 扩展信息，存在时返回                               |

### 8.5 `seedream4.5`

| 项目       | 值                             |
| -------- | ----------------------------- |
| 请求接口     | `POST /v1/images/generations` |
| 当前推荐尺寸   | `2048x2048`                   |
| 当前已验证价格  | `0.3 CNY`                     |
| 当前非法尺寸示例 | `1024x1024`                   |
| 网关可透传尺寸  | 明确像素值、`2K`、`4K`               |

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

| 项目       | 值                             |
| -------- | ----------------------------- |
| 请求接口     | `POST /v1/images/generations` |
| 当前推荐尺寸   | `2048x2048`                   |
| 当前已验证价格  | `0.25 CNY`                    |
| 当前非法尺寸示例 | `1024x1024`                   |
| 网关可透传尺寸  | 明确像素值、`2K`、`3K`、`4K`          |

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

***

## 9. 真人资产接口

真人资产用于创建和管理可绑定到视频生成任务中的真人资产。推荐使用 `/api/portrait_assets/official/*` 官方真人资产接口。

### 9.1 真人资产状态

| 状态                 | 含义              |
| ------------------ | --------------- |
| `pending`          | 任务已创建           |
| `validate_ready`   | 已生成实名校验链接       |
| `validated`        | 实名校验通过，已拿到资产组   |
| `asset_processing` | 素材入库处理中         |
| `qr_ready`         | 历史 RPA 队列二维码已准备 |
| `waiting_upload`   | 历史 RPA 等待上传     |
| `waiting_accept`   | 历史 RPA 等待受理     |
| `pending_confirm`  | 素材已入库，等待账号确认预览  |
| `ready`            | 资产已可用           |
| `failed`           | 失败              |
| `disabled`         | 已禁用             |
| `expired`          | 实名链接或二维码已过期     |

### 9.2 真人资产对象

| 字段                     | 类型        | 说明                                       |
| ---------------------- | --------- | ---------------------------------------- |
| `id`                   | `integer` | 真人资产任务 ID                                |
| `user_id`              | `integer` | 账号 ID                                    |
| `external_user_id`     | `string`  | 多下游账号隔离字段                                |
| `folder_id`            | `integer` | 文件夹 ID，`0` 表示未分组                         |
| `name`                 | `string`  | 资产名称                                     |
| `source`               | `string`  | 来源，官方接口为 `official`                      |
| `status`               | `string`  | 真人资产状态                                   |
| `invite_url`           | `string`  | 实名校验链接                                   |
| `qr_image`             | `string`  | 二维码图片内容或地址                               |
| `validate_result_code` | `string`  | 实名校验结果码，`10000` 表示通过                     |
| `volc_group_id`        | `string`  | 上游资产组 ID                                 |
| `asset_id`             | `string`  | 最终真人资产 ID                                |
| `asset_status`         | `string`  | 上游资产状态，例如 `Processing`、`Active`、`Failed` |
| `asset_preview`        | `string`  | 安全预览地址                                   |
| `asset_url`            | `string`  | 提交的素材 URL                                |
| `asset_type`           | `string`  | `Image`、`Video`、`Audio`                  |
| `project_name`         | `string`  | 上游项目名                                    |
| `error_message`        | `string`  | 错误信息                                     |
| `created_time`         | `integer` | 创建时间                                     |
| `updated_time`         | `integer` | 更新时间                                     |
| `accept_time`          | `integer` | 受理时间                                     |
| `qr_expires_time`      | `integer` | 链接或二维码过期时间                               |
| `ready_time`           | `integer` | 就绪时间                                     |
| `queue_position`       | `integer` | 历史 RPA 队列位置                              |

### 9.3 查询真人资产配置

```http
GET /api/portrait_assets/official/config
```

响应字段：

| 字段                  | 类型        | 说明            |
| ------------------- | --------- | ------------- |
| `data.configured`   | `boolean` | 官方真人资产能力是否已配置 |
| `data.project_name` | `string`  | 当前火山项目名       |

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

查询参数：

| 参数                 | 类型        | 必填 | 说明                              |
| ------------------ | --------- | -- | ------------------------------- |
| `p`                | `integer` | 否  | 页码                              |
| `page_size`        | `integer` | 否  | 每页数量                            |
| `external_user_id` | `string`  | 否  | 按隔离空间查询                         |
| `folder_id`        | `integer` | 否  | 不传返回全部；`0` 返回未分组；大于 `0` 返回指定文件夹 |

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

| 字段                 | 类型        | 必填 | 限制                   | 说明               |
| ------------------ | --------- | -- | -------------------- | ---------------- |
| `name`             | `string`  | 否  | 最长 50 个字符            | 资产名称；为空时系统使用默认名称 |
| `callback_url`     | `string`  | 否  | `http` 或 `https` URL | 真人实名完成后的最终跳转地址   |
| `external_user_id` | `string`  | 否  | 最长 128 个字符           | 多下游账号隔离字段        |
| `folder_id`        | `integer` | 否  | `0` 或已有文件夹 ID        | 创建后放入指定文件夹       |

请求示例：

```json
{
  "name": "主播真人资产",
  "callback_url": "https://example.com/portrait/callback",
  "external_user_id": "account-001",
  "folder_id": 0
}
```

成功后通常返回真人资产对象，并包含 `invite_url`。需要打开 `invite_url` 完成真人实名校验。

### 9.6 刷新真人实名校验链接

```http
POST /api/portrait_assets/official/jobs/{id}/validation
```

路径参数：

| 参数   | 类型        | 必填 | 说明        |
| ---- | --------- | -- | --------- |
| `id` | `integer` | 是  | 真人资产任务 ID |

使用规则：

| 条件                                               | 行为        |
| ------------------------------------------------ | --------- |
| 任务尚未进入素材处理                                       | 可重新生成实名链接 |
| 状态为 `asset_processing`、`pending_confirm`、`ready` | 不允许重新生成   |

### 9.7 上传真人资产素材

```http
POST /api/portrait_assets/official/upload
Content-Type: multipart/form-data
```

表单字段：

| 字段     | 类型     | 必填 | 说明         |
| ------ | ------ | -- | ---------- |
| `file` | `file` | 是  | 图片、视频或音频素材 |

文件限制：

| 类型 | 扩展名                                         | MIME                                                                      |
| -- | ------------------------------------------- | ------------------------------------------------------------------------- |
| 图片 | `.jpg`、`.jpeg`、`.png`、`.webp`、`.gif`、`.bmp` | `image/jpeg`、`image/png`、`image/webp`、`image/gif`、`image/bmp`             |
| 视频 | `.mp4`、`.mov`、`.webm`                       | `video/mp4`、`video/quicktime`、`video/webm`                                |
| 音频 | `.mp3`、`.wav`、`.m4a`、`.ogg`、`.aac`、`.flac`  | `audio/mpeg`、`audio/wav`、`audio/mp4`、`audio/ogg`、`audio/aac`、`audio/flac` |

最大文件大小：`100MB`

响应字段：

| 字段                  | 类型        | 说明                      |
| ------------------- | --------- | ----------------------- |
| `data.url`          | `string`  | 上传后可公开访问的素材 URL         |
| `data.file_name`    | `string`  | 原始文件名                   |
| `data.content_type` | `string`  | 识别出的 MIME               |
| `data.asset_type`   | `string`  | `Image`、`Video`、`Audio` |
| `data.size`         | `integer` | 文件大小，单位字节               |

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

| 参数   | 类型        | 必填 | 说明        |
| ---- | --------- | -- | --------- |
| `id` | `integer` | 是  | 真人资产任务 ID |

请求字段：

| 字段           | 类型        | 必填 | 可选值或格式                  | 说明         |
| ------------ | --------- | -- | ----------------------- | ---------- |
| `asset_url`  | `string`  | 是  | `http` 或 `https` URL    | 必须可被火山访问   |
| `asset_type` | `string`  | 是  | `Image`、`Video`、`Audio` | 素材类型       |
| `name`       | `string`  | 否  | 最长 50 个字符               | 素材名称       |
| `folder_id`  | `integer` | 否  | `0` 或已有文件夹 ID           | 提交后放入指定文件夹 |

请求示例：

```json
{
  "asset_url": "https://8liangai.com/uploads/portrait/20260608/demo.mp4",
  "asset_type": "Video",
  "name": "主播口播素材",
  "folder_id": 0
}
```

前置条件：

| 条件      | 说明                             |
| ------- | ------------------------------ |
| 任务已实名通过 | 状态应为 `validated`               |
| 已有资产组   | `volc_group_id` 非空             |
| 素材可公网访问 | `asset_url` 必须是上游可访问的 HTTP URL |

### 9.9 同步真人资产状态

```http
POST /api/portrait_assets/official/jobs/{id}/sync
```

用途：

| 用途       | 说明                                                    |
| -------- | ----------------------------------------------------- |
| 同步实名结果   | 当回调已完成但本地未刷新时使用                                       |
| 同步素材入库状态 | 将 `asset_processing` 推进到 `pending_confirm` 或 `failed` |
| 更新预览地址   | 刷新 `asset_preview`                                    |

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

### 9.12 删除真人资产

```http
DELETE /api/portrait_assets/official/jobs/{id}
```

路径参数：

| 参数   | 类型        | 必填 | 说明        |
| ---- | --------- | -- | --------- |
| `id` | `integer` | 是  | 真人资产任务 ID |

查询参数：

| 参数                 | 类型       | 必填 | 说明                 |
| ------------------ | -------- | -- | ------------------ |
| `external_user_id` | `string` | 否  | 传入后只允许删除同一隔离空间下的资产 |

完整请求示例：

```bash
curl -X DELETE 'https://8liangai.com/api/portrait_assets/official/jobs/12?external_user_id=account-001' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -H 'Accept: application/json'
```

成功响应示例：

```json
{
  "success": true,
  "message": "",
  "data": null
}
```

失败响应示例：

```json
{
  "success": false,
  "message": "record not found"
}
```

兼容性说明：新增接口，不影响旧调用。

### 9.13 真人资产预览

```http
GET /api/portrait_assets/official/jobs/{id}/preview/{state}
```

该接口是安全预览跳转地址。正常接入不需要自行拼接 `state`，直接使用列表或详情返回的 `asset_preview`。

成功时返回 `302` 跳转到真实预览资源。

### 9.14 真人资产绑定视频

绑定方式：

| 字段                           | 填法                                           | 说明   |
| ---------------------------- | -------------------------------------------- | ---- |
| `metadata.portrait_asset_id` | 真人资产任务 ID，例如 `12`                            | 推荐方式 |
| `metadata.asset_id`          | 最终资产 ID，例如 `asset_xxx` 或 `asset://asset_xxx` | 兼容方式 |

绑定要求：

| 要求                  | 说明              |
| ------------------- | --------------- |
| 资产属于当前 API Key 对应账号 | 不能引用其他账号资产      |
| 资产状态为 `ready`       | 未就绪资产会被拒绝       |
| `asset_id` 非空       | 任务必须已有最终上游资产 ID |

补充说明：

- 真人资产可以和虚拟资产、外部图片一起出现在同一个视频请求里。
- 如果一个请求里既有真人资产又有虚拟资产，建议统一放到 `metadata.content` 的最终顺序里理解图序。
- 混合示例直接看 `4.5 请求示例` 里的“真人资产 + 虚拟资产 + 外部图片混合示例”。

***

## 10. 虚拟人像资产接口

虚拟人像资产用于管理虚拟素材组和素材资产。当前公开接口支持配置查询、资产组查询、列表、上传、创建、同步、预览，以及在 Seedance 视频生成中作为参考资产使用。

虚拟人像资产的视频生成绑定方式是 `asset://<volc_asset_id>`。创建并同步到 `active` 后，取虚拟资产对象中的 `volc_asset_id`，拼成 `asset://asset_xxx`，放入视频生成请求的 `metadata.content` 中。

注意：虚拟人像资产不要使用 `metadata.portrait_asset_id`。`metadata.portrait_asset_id` 是真人资产任务 ID 的绑定字段。

### 10.1 虚拟资产组状态

| 状态         | 含义      |
| ---------- | ------- |
| `creating` | 资产组创建中  |
| `active`   | 资产组可用   |
| `failed`   | 资产组创建失败 |

### 10.2 虚拟资产状态

| 状态           | 含义    |
| ------------ | ----- |
| `processing` | 资产处理中 |
| `active`     | 资产可用  |
| `failed`     | 资产失败  |

### 10.3 虚拟资产组对象

| 字段                 | 类型        | 说明                           |
| ------------------ | --------- | ---------------------------- |
| `id`               | `integer` | 本地资产组 ID                     |
| `user_id`          | `integer` | 账号 ID                        |
| `external_user_id` | `string`  | 多下游账号隔离字段                    |
| `name`             | `string`  | 资产组名称                        |
| `description`      | `string`  | 资产组描述                        |
| `project_name`     | `string`  | 上游项目名                        |
| `volc_group_id`    | `string`  | 上游资产组 ID                     |
| `status`           | `string`  | `creating`、`active`、`failed` |
| `error_message`    | `string`  | 错误信息                         |
| `created_time`     | `integer` | 创建时间                         |
| `updated_time`     | `integer` | 更新时间                         |

### 10.4 虚拟资产对象

| 字段                 | 类型        | 说明                             |
| ------------------ | --------- | ------------------------------ |
| `id`               | `integer` | 本地资产 ID                        |
| `user_id`          | `integer` | 账号 ID                          |
| `external_user_id` | `string`  | 多下游账号隔离字段                      |
| `folder_id`        | `integer` | 文件夹 ID，`0` 表示未分组               |
| `group_id`         | `integer` | 本地资产组 ID                       |
| `name`             | `string`  | 资产名称                           |
| `asset_type`       | `string`  | `Image`、`Video`、`Audio`        |
| `source_url`       | `string`  | 原始素材 URL                       |
| `preview_url`      | `string`  | 安全预览地址                         |
| `project_name`     | `string`  | 上游项目名                          |
| `volc_group_id`    | `string`  | 上游资产组 ID                       |
| `volc_asset_id`    | `string`  | 上游资产 ID                        |
| `status`           | `string`  | `processing`、`active`、`failed` |
| `volc_status`      | `string`  | 上游状态                           |
| `error_message`    | `string`  | 错误信息                           |
| `created_time`     | `integer` | 创建时间                           |
| `updated_time`     | `integer` | 更新时间                           |
| `ready_time`       | `integer` | 就绪时间                           |

### 10.5 查询虚拟人像配置

```http
GET /api/portrait_assets/virtual/config
```

响应字段：

| 字段                  | 类型        | 说明            |
| ------------------- | --------- | ------------- |
| `data.configured`   | `boolean` | 虚拟人像资产能力是否已配置 |
| `data.project_name` | `string`  | 当前火山项目名       |

### 10.6 查询当前账号虚拟资产组

```http
GET /api/portrait_assets/virtual/group
```

查询参数：

| 参数                 | 类型       | 必填 | 说明         |
| ------------------ | -------- | -- | ---------- |
| `external_user_id` | `string` | 否  | 按隔离空间查询资产组 |

响应：

| 情况      | 响应              |
| ------- | --------------- |
| 已有资产组   | `data` 为虚拟资产组对象 |
| 尚未创建资产组 | `data` 为 `null` |

### 10.7 查询虚拟资产列表

```http
GET /api/portrait_assets/virtual/assets
```

查询参数：

| 参数                 | 类型        | 必填 | 说明                              |
| ------------------ | --------- | -- | ------------------------------- |
| `p`                | `integer` | 否  | 页码                              |
| `page_size`        | `integer` | 否  | 每页数量                            |
| `external_user_id` | `string`  | 否  | 按隔离空间查询                         |
| `folder_id`        | `integer` | 否  | 不传返回全部；`0` 返回未分组；大于 `0` 返回指定文件夹 |

注意：列表接口会自动同步当前页里仍处于 `processing` 的虚拟资产。上游已经变成 `Active` 时，本地会同步成 `active`，不用再手动进后台刷新。

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

| 字段     | 类型     | 必填 | 说明         |
| ------ | ------ | -- | ---------- |
| `file` | `file` | 是  | 图片、视频或音频素材 |

限制：

| 项目     | 值                                           |
| ------ | ------------------------------------------- |
| 最大文件大小 | `100MB`                                     |
| 图片格式   | `.jpg`、`.jpeg`、`.png`、`.webp`、`.gif`、`.bmp` |
| 视频格式   | `.mp4`、`.mov`、`.webm`                       |
| 音频格式   | `.mp3`、`.wav`、`.m4a`、`.ogg`、`.aac`、`.flac`  |

响应字段与真人资产上传相同。

### 10.9 创建虚拟资产

```http
POST /api/portrait_assets/virtual/assets
Content-Type: application/json
```

请求字段：

| 字段                 | 类型        | 必填 | 可选值或格式                         | 说明             |
| ------------------ | --------- | -- | ------------------------------ | -------------- |
| `name`             | `string`  | 否  | 最长 50 个字符                      | 资产名称；为空时使用默认名称 |
| `asset_url`        | `string`  | 是  | `http` 或 `https` URL           | 必须可被火山访问       |
| `asset_type`       | `string`  | 是  | `Image`、`Video`、`Audio`，大小写不敏感 | 素材类型           |
| `external_user_id` | `string`  | 否  | 最长 128 个字符                     | 多下游账号隔离字段      |
| `folder_id`        | `integer` | 否  | `0` 或已有文件夹 ID                  | 创建后放入指定文件夹     |

请求示例：

```json
{
  "name": "虚拟口播素材",
  "asset_url": "https://8liangai.com/uploads/portrait/20260608/virtual.mp4",
  "asset_type": "Video",
  "external_user_id": "account-001",
  "folder_id": 0
}
```

创建行为：

| 场景          | 行为                          |
| ----------- | --------------------------- |
| 当前账号没有虚拟资产组 | 系统自动创建账号专属虚拟资产组             |
| 资产组正在初始化    | 返回稍后再试                      |
| 素材提交成功      | 返回虚拟资产对象，初始通常为 `processing` |
| 上游很快完成      | 返回时可能已经是 `active`           |

### 10.10 同步虚拟资产状态

```http
POST /api/portrait_assets/virtual/assets/{id}/sync
```

路径参数：

| 参数   | 类型        | 必填 | 说明      |
| ---- | --------- | -- | ------- |
| `id` | `integer` | 是  | 虚拟资产 ID |

用途：

| 用途     | 说明                                  |
| ------ | ----------------------------------- |
| 拉取上游状态 | 更新 `volc_status`                    |
| 刷新预览地址 | 更新 `preview_url`                    |
| 推进本地状态 | `processing` 变为 `active` 或 `failed` |

### 10.11 虚拟资产预览

```http
GET /api/portrait_assets/virtual/assets/{id}/preview/{state}
```

该接口是安全预览跳转地址。正常接入不需要自行拼接 `state`，直接使用列表返回的 `preview_url`。

成功时返回 `302` 跳转到真实预览资源。

### 10.12 删除虚拟资产

```http
DELETE /api/portrait_assets/virtual/assets/{id}
```

路径参数：

| 参数   | 类型        | 必填 | 说明      |
| ---- | --------- | -- | ------- |
| `id` | `integer` | 是  | 虚拟资产 ID |

查询参数：

| 参数                 | 类型       | 必填 | 说明                 |
| ------------------ | -------- | -- | ------------------ |
| `external_user_id` | `string` | 否  | 传入后只允许删除同一隔离空间下的资产 |

完整请求示例：

```bash
curl -X DELETE 'https://8liangai.com/api/portrait_assets/virtual/assets/31?external_user_id=account-001' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -H 'Accept: application/json'
```

成功响应示例：

```json
{
  "success": true,
  "message": "",
  "data": null
}
```

失败响应示例：

```json
{
  "success": false,
  "message": "record not found"
}
```

兼容性说明：新增接口，不影响旧调用。

### 10.13 虚拟资产用于视频生成

前置条件：

| 条件      | 要求                                                   |
| ------- | ---------------------------------------------------- |
| 虚拟资产状态  | `status=active`                                      |
| 上游资产 ID | `volc_asset_id` 非空                                   |
| 资产 URI  | 使用 `asset://<volc_asset_id>`                         |
| 视频模型    | 推荐 `seedance2`、`seedance2-sr` 或其他支持资产引用的 Seedance 模型 |

字段对应关系：

| 虚拟资产类型  | 视频 `metadata.content` 写法                                                        | 说明                 |
| ------- | ------------------------------------------------------------------------------- | ------------------ |
| `Image` | `type=image_url`，`role=reference_image`，`image_url.url=asset://<volc_asset_id>` | 作为形象或图片参考          |
| `Video` | `type=video_url`，`role=reference_video`，`video_url.url=asset://<volc_asset_id>` | 作为视频参考             |
| `Audio` | `type=audio_url`，`audio_url.url=asset://<volc_asset_id>`                        | 作为音频参考，是否生效取决于模型能力 |

补充说明：

- 虚拟资产可以和真人资产、外部图片一起出现在同一个视频请求里。
- 如果提示词里要稳定写 `图1`、`图2`、`图3`，建议把虚拟资产先按镜头顺序写进 `metadata.content`，再决定是否补顶层 `images`。
- 混合示例直接看 `4.5 请求示例` 里的“真人资产 + 虚拟资产 + 外部图片混合示例”。

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

***

## 11. 人像资产文件夹接口

文件夹只用于八两平台内的资产管理，不同步到火山，不改变 `asset://...` 引用和视频生成逻辑。

固定规则：

| 规则                    | 说明                                        |
| --------------------- | ----------------------------------------- |
| `asset_kind=official` | 管理真人资产文件夹                                 |
| `asset_kind=virtual`  | 管理虚拟人像资产文件夹                               |
| `folder_id=0`         | 固定表示未分组                                   |
| 不传 `folder_id`        | 资产列表返回全部                                  |
| 删除文件夹                 | 文件夹软删除，文件夹内资产移动到未分组，不删除资产                 |
| 隔离字段                  | 传 `external_user_id` 时，只能管理同一隔离空间下的文件夹和资产 |

### 11.1 文件夹对象

| 字段                 | 类型        | 说明                     |
| ------------------ | --------- | ---------------------- |
| `id`               | `integer` | 文件夹 ID                 |
| `user_id`          | `integer` | 账号 ID                  |
| `external_user_id` | `string`  | 多下游账号隔离字段              |
| `asset_kind`       | `string`  | `official` 或 `virtual` |
| `name`             | `string`  | 文件夹名称                  |
| `sort_order`       | `integer` | 排序值，越小越靠前              |
| `created_time`     | `integer` | 创建时间                   |
| `updated_time`     | `integer` | 更新时间                   |

### 11.2 查询文件夹列表

```http
GET /api/portrait_assets/folders
```

Query 参数：

| 参数                 | 类型       | 必填 | 说明                     |
| ------------------ | -------- | -- | ---------------------- |
| `asset_kind`       | `string` | 是  | `official` 或 `virtual` |
| `external_user_id` | `string` | 否  | 按隔离空间查询                |

完整请求示例：

```bash
curl -X GET 'https://8liangai.com/api/portrait_assets/folders?asset_kind=virtual&external_user_id=account-001' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -H 'Accept: application/json'
```

成功响应示例：

```json
{
  "success": true,
  "message": "",
  "data": [
    {
      "id": 101,
      "user_id": 1,
      "external_user_id": "account-001",
      "asset_kind": "virtual",
      "name": "人物A-分镜图",
      "sort_order": 0,
      "created_time": 1780641000,
      "updated_time": 1780641000
    }
  ]
}
```

失败响应示例：

```json
{
  "success": false,
  "message": "asset_kind 仅支持 official 或 virtual"
}
```

兼容性说明：新增接口，不影响旧调用。

### 11.3 创建文件夹

```http
POST /api/portrait_assets/folders
Content-Type: application/json
```

Body 参数：

| 字段                 | 类型        | 必填 | 说明                     |
| ------------------ | --------- | -- | ---------------------- |
| `asset_kind`       | `string`  | 是  | `official` 或 `virtual` |
| `name`             | `string`  | 是  | 文件夹名称，最长 128 个字符       |
| `external_user_id` | `string`  | 否  | 多下游账号隔离字段              |
| `sort_order`       | `integer` | 否  | 排序值，默认 `0`             |

完整请求示例：

```bash
curl -X POST 'https://8liangai.com/api/portrait_assets/folders' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json' \
  -d '{
    "asset_kind": "virtual",
    "name": "人物A-分镜图",
    "external_user_id": "account-001",
    "sort_order": 0
  }'
```

成功响应示例：

```json
{
  "success": true,
  "message": "",
  "data": {
    "id": 101,
    "user_id": 1,
    "external_user_id": "account-001",
    "asset_kind": "virtual",
    "name": "人物A-分镜图",
    "sort_order": 0,
    "created_time": 1780641000,
    "updated_time": 1780641000
  }
}
```

失败响应示例：

```json
{
  "success": false,
  "message": "同名文件夹已存在"
}
```

兼容性说明：新增接口，不影响旧调用。

### 11.4 重命名文件夹

```http
PATCH /api/portrait_assets/folders/{folder_id}
Content-Type: application/json
```

路径参数：

| 参数          | 类型        | 必填 | 说明     |
| ----------- | --------- | -- | ------ |
| `folder_id` | `integer` | 是  | 文件夹 ID |

Body 参数：

| 字段                 | 类型        | 必填 | 说明                |
| ------------------ | --------- | -- | ----------------- |
| `name`             | `string`  | 是  | 新文件夹名称，最长 128 个字符 |
| `external_user_id` | `string`  | 否  | 多下游账号隔离字段         |
| `sort_order`       | `integer` | 否  | 排序值               |

完整请求示例：

```bash
curl -X PATCH 'https://8liangai.com/api/portrait_assets/folders/101' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json' \
  -d '{
    "name": "人物A-口播分镜",
    "external_user_id": "account-001",
    "sort_order": 10
  }'
```

成功响应示例：

```json
{
  "success": true,
  "message": "",
  "data": {
    "id": 101,
    "user_id": 1,
    "external_user_id": "account-001",
    "asset_kind": "virtual",
    "name": "人物A-口播分镜",
    "sort_order": 10,
    "created_time": 1780641000,
    "updated_time": 1780641300
  }
}
```

失败响应示例：

```json
{
  "success": false,
  "message": "文件夹或资产不存在"
}
```

兼容性说明：新增接口，不影响旧调用。

### 11.5 删除文件夹

```http
DELETE /api/portrait_assets/folders/{folder_id}
```

路径参数：

| 参数          | 类型        | 必填 | 说明     |
| ----------- | --------- | -- | ------ |
| `folder_id` | `integer` | 是  | 文件夹 ID |

Query 参数：

| 参数                 | 类型       | 必填 | 说明      |
| ------------------ | -------- | -- | ------- |
| `external_user_id` | `string` | 否  | 按隔离空间校验 |

完整请求示例：

```bash
curl -X DELETE 'https://8liangai.com/api/portrait_assets/folders/101?external_user_id=account-001' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -H 'Accept: application/json'
```

成功响应示例：

```json
{
  "success": true,
  "message": "",
  "data": null
}
```

失败响应示例：

```json
{
  "success": false,
  "message": "文件夹或资产不存在"
}
```

兼容性说明：新增接口，不影响旧调用。删除文件夹不会删除资产，资产会进入未分组。

### 11.6 移动资产到文件夹

```http
POST /api/portrait_assets/folders/move
Content-Type: application/json
```

Body 参数：

| 字段                 | 类型          | 必填 | 说明                     |
| ------------------ | ----------- | -- | ---------------------- |
| `asset_kind`       | `string`    | 是  | `official` 或 `virtual` |
| `asset_ids`        | `integer[]` | 是  | 要移动的资产 ID 列表，最多 1000 个 |
| `folder_id`        | `integer`   | 是  | 目标文件夹 ID；`0` 表示移入未分组   |
| `external_user_id` | `string`    | 否  | 多下游账号隔离字段              |

完整请求示例：

```bash
curl -X POST 'https://8liangai.com/api/portrait_assets/folders/move' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json' \
  -d '{
    "asset_kind": "virtual",
    "asset_ids": [31, 32, 33],
    "folder_id": 101,
    "external_user_id": "account-001"
  }'
```

成功响应示例：

```json
{
  "success": true,
  "message": "",
  "data": {
    "moved": 3
  }
}
```

失败响应示例：

```json
{
  "success": false,
  "message": "文件夹或资产不存在"
}
```

兼容性说明：新增接口，不影响旧调用。

## 12. 兼容接口说明

### Remix 兼容接口

```http
POST /v1/videos/{video_id}/remix
```

该接口用于基于已有视频任务发起 Remix，当前文档不作为主流程展开。使用时 `video_id` 必须属于当前账号，且系统会尽量复用原任务模型和通道。

***

## 13. 网站使用教程

### 13.1 API Key 创建

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

| 字段    | 建议填法                              | 说明                |
| ----- | --------------------------------- | ----------------- |
| 名称    | `video-prod-key`、`image-test-key` | 用于区分用途            |
| 分组    | `default`                         | 本文档覆盖模型建议使用默认分组   |
| 额度    | 按业务预算填写                           | 用于限制该 Key 可消耗金额   |
| 不限额   | 生产内部服务按需开启                        | 开启后不受该 Key 单独额度限制 |
| 过期时间  | 测试 Key 建议设置，生产 Key 按安全策略设置        | 到期后 Key 不可用       |
| 跨分组重试 | 通常关闭                              | 仅在高级路由场景使用        |

安全建议：

| 建议             | 说明          |
| -------------- | ----------- |
| 测试和生产分开 Key    | 便于查账和限额     |
| 不要把 Key 写入前端代码 | 应由服务端调用 API |
| 泄露后立即删除旧 Key   | 重新创建新 Key   |
| 每个业务线单独 Key    | 便于按业务统计成本   |

### 13.2 真人资产创建流程

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

| 条件         | 要求                 |
| ---------- | ------------------ |
| `status`   | 必须为 `ready`        |
| `asset_id` | 必须非空               |
| 所属账号       | 必须为当前 API Key 对应账号 |

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

### 13.3 虚拟人像资产创建流程

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
6. 首次创建时系统会自动创建账号专属虚拟资产组。
7. 资产初始通常为 `processing`。
8. 调用同步接口或等待页面刷新。
9. 状态进入 `active` 后，资产创建完成。
10. 使用 `preview_url` 查看预览。
11. 记录 `volc_asset_id`。
12. 视频生成时将 `volc_asset_id` 拼成 `asset://<volc_asset_id>`，放入 `metadata.content`。

素材类型建议：

| 类型      | 用途             |
| ------- | -------------- |
| `Image` | 角色形象、脸部参考、头像素材 |
| `Video` | 角色动作、口播、动态参考素材 |
| `Audio` | 声音或音频参考素材      |

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

***

## 14. 接入检查清单

### 14.1 视频上线前检查

| 检查项   | 要求                                               |
| ----- | ------------------------------------------------ |
| 模型名   | 必须在 `/v1/models` 可见或后台已配置映射                      |
| 分辨率   | 必须符合模型能力表                                        |
| 比例    | 生产建议显式传 `ratio`                                  |
| 时长    | 建议显式传 `duration`                                 |
| 水印    | 如需无水印，显式传 `metadata.watermark=false`             |
| 状态轮询  | 同时兼容 `queued`、`in_progress`、`completed`、`failed` |
| 下载    | 仅在任务完成后调用 `/content`                             |
| 计费    | 按 `actual_quota / 500000` 核对金额                   |
| SR 模型 | 前端展示请求方请求的 `720p` 或 `1080p`                      |

### 14.2 图片上线前检查

| 检查项  | 要求                                   |
| ---- | ------------------------------------ |
| 模型名  | 使用 `seedream4.5` 或 `seedream5.0lite` |
| 尺寸   | 当前推荐 `2048x2048`                     |
| 数量   | 首版建议 `n=1`                           |
| 返回格式 | 前端展示建议 `response_format=url`         |
| 失败扣费 | 参数非法不应扣费                             |

### 15.3 资产上线前检查

| 检查项    | 要求                                      |
| ------ | --------------------------------------- |
| 真人资产配置 | `official/config` 返回 `configured=true`  |
| 真人实名   | 状态进入 `validated` 后再提交素材                 |
| 真人可用   | `status=ready` 且 `asset_id` 非空          |
| 虚拟资产配置 | `virtual/config` 返回 `configured=true`   |
| 虚拟资产可用 | `status=active`                         |
| 预览地址   | 使用接口返回的 `asset_preview` 或 `preview_url` |

***

## 16. 附录：端到端视频调用流程

### 16.1 提交任务

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

### 16.2 轮询状态

```bash
curl 'https://8liangai.com/v1/videos/task_xxx' \
  -H 'Authorization: Bearer sk-你的APIKey'
```

当 `status=completed` 时进入下载步骤。

### 16.3 下载视频

```bash
curl -L 'https://8liangai.com/v1/videos/task_xxx/content' \
  -H 'Authorization: Bearer sk-你的APIKey' \
  -o result.mp4
```

### 16.4 推荐轮询策略

| 阶段           | 建议                             |
| ------------ | ------------------------------ |
| 提交后 0 到 30 秒 | 每 5 到 10 秒查询一次                 |
| 30 秒后        | 每 10 到 15 秒查询一次                |
| 超时时间         | 建议业务侧设置 20 分钟                  |
| 失败处理         | 记录 `error`、`fail_reason`、完整响应体 |

