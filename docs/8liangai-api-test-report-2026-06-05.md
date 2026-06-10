# 8liangai.com API 重新实测报告

测试时间：2026-06-05（Asia/Shanghai）

测试站点：`https://8liangai.com`

测试认证：`Authorization: Bearer sk-***`

关联接口文档：[8liangai-api-reference-2026-06-05.md](/C:/Users/A/Desktop/whdj/docs/8liangai-api-reference-2026-06-05.md)

说明：

- 本文只记录“本轮重新测试”的真实结果。
- 本文不再混入接口字段说明，不再把“接口规范”和“测试结论”写在一起。
- 本文的“能不能用”和“扣费是否对得上”，统一以本轮重新提交任务、本轮查询结果、本轮下载结果、本轮 token 日志为准。

---

## 1. 测试目标

本轮测试目标：

1. 重新提交并测试以下模型，不复用旧结论：
   - `seedance1.5`
   - `seedance1.5-sr`
   - `seedance2`
   - `sd2.0fast`
   - `seedance2-sr`
   - `seedance2.0fast-sr`
   - `seedream4.5`
   - `seedream5.0lite`
2. 验证：
   - 请求接口是否可用
   - 业务状态查询是否可用
   - OpenAI 兼容查询是否可用
   - 下载接口是否可用
   - 返回结构是否正常
   - 实际扣费是否与定价对得上
3. 额外覆盖边界场景：
   - `seedance2` 的 1080p 档
   - `seedance2` 的视频输入档
   - 图片模型的非法尺寸
   - 真人资产的未实名拦截
   - 虚拟人像素材过小拦截

---

## 2. 测试基线

## 2.1 站点公共状态

`GET /api/status` 返回关键信息：

- `quota_display_type = CNY`
- `quota_per_unit = 500000`
- 换算公式：`金额(元) = quota / 500000`

## 2.2 当前价格表关键信息

`GET /api/pricing` 当前关键信息：

| 模型 | 当前价格配置 |
| --- | --- |
| `seedance1.5` | `model_ratio = 8` |
| `seedance1.5-sr` | `model_ratio = 10` |
| `sd2.0fast` | `model_ratio = 18.5` |
| `seedance2.0fast-sr` | `model_ratio = 23.15` |
| `seedance2-sr` | `model_ratio = 24.5` |
| `seedream4.5` | `model_price = 0.3` |
| `seedream5.0lite` | `model_price = 0.25` |
| `seedance2` | `billing_mode = output_tier_price` |

`seedance2` 当前分档：

| 条件 | `input_price` | 实际倍率 |
| --- | --- | --- |
| `720p` 无视频输入 | `46` | `23` |
| `720p` 有视频输入 | `31` | `15.5` |
| `1080p` 无视频输入 | `51` | `25.5` |
| `1080p` 有视频输入 | `34` | `17` |

## 2.3 当前对外可见模型

`GET /v1/models` 返回中，当前测试范围内可见：

- `seedance1.5`
- `seedance1.5-sr`
- `seedance2`
- `sd2.0fast`
- `seedance2-sr`
- `seedance2.0fast-sr`
- `seedream4.5`
- `seedream5.0lite`

---

## 3. 测试结论总表

| 模型 / 分支 | 提交 | 状态查询 | 下载 | 扣费核对 | 结论 |
| --- | --- | --- | --- | --- | --- |
| `seedance1.5` 720p | 成功 | 成功 | 成功 | 对得上 | 正常 |
| `seedance1.5-sr` 720p | 成功 | 成功 | 成功 | 对得上 | 正常 |
| `seedance2` 720p 无视频输入 | 成功 | 成功 | 成功 | 对得上 | 正常 |
| `seedance2` 1080p 无视频输入 | 成功 | 成功 | 成功 | 对得上 | 正常 |
| `seedance2` 720p 视频输入 | 成功 | 成功 | 成功 | 对得上 | 正常 |
| `sd2.0fast` 720p | 成功 | 成功 | 成功 | 对得上 | 正常 |
| `seedance2-sr` 720p | 成功 | 成功 | 成功 | 与价格页不一致 | 有问题 |
| `seedance2-sr` 1080p | 成功 | 成功 | 成功 | 与价格页不一致 | 有问题 |
| `seedance2.0fast-sr` 720p | 成功 | 成功 | 成功 | 对得上 | 有问题，输出清晰度不达标 |
| `seedream4.5` 2048x2048 | 成功 | 不适用 | 返回图片 URL | 对得上 | 正常 |
| `seedream4.5` 1024x1024 | 失败 | 不适用 | 不适用 | 不扣费 | 合理失败 |
| `seedream5.0lite` 2048x2048 | 成功 | 不适用 | 返回图片 URL | 对得上 | 正常 |
| `seedream5.0lite` 1024x1024 | 失败 | 不适用 | 不适用 | 不扣费 | 合理失败 |

---

## 4. 视频模型详细结果

## 4.1 `seedance1.5`

任务：

- `task_id = task_QYnSrLQyhZX83GaYNENzZEI67PTQdWIY`

请求结果：

- `POST /v1/video/generations` 成功
- OpenAI 返回 `queued`

状态结果：

- 业务查询：`SUCCESS`
- OpenAI 查询：`completed`
- `total_tokens = 108900`
- `resolution = 720p`
- `ratio = 16:9`

下载结果：

- `/content` 返回 `HTTP 200`
- `Content-Type = video/mp4`
- 文件大小：`7,473,124 bytes`
- 实际视频尺寸：`1280x720`
- 实际视频帧率：`24 fps`

计费核对：

- 预扣：`2000000 quota = ¥4.0000`
- 最终 `actual_quota = 871200`
- 实际金额：`871200 / 500000 = ¥1.7424`
- 退款：`1128800 quota = ¥2.2576`
- 核对公式：`108900 × 8 = 871200`

结论：

- `接口正常`
- `计费正常`

## 4.2 `seedance1.5-sr`

任务：

- `task_id = task_scvUXZ75JhSV9uMhyyMEhhvi4MPUsPBY`

状态结果：

- 业务查询：`SUCCESS`
- OpenAI 查询：`completed`
- `total_tokens = 108900`
- 业务记录 `resolution = 720p`

下载结果：

- `/content` 返回 `HTTP 200`
- 文件大小：`8,897,420 bytes`
- 实际视频尺寸：`1280x720`

计费核对：

- 预扣：`2500000 quota = ¥5.0000`
- 最终 `actual_quota = 1089000`
- 实际金额：`¥2.1780`
- 退款：`1411000 quota = ¥2.8220`
- 核对公式：`108900 × 10 = 1089000`

结论：

- `接口正常`
- `计费正常`

## 4.3 `seedance2` 720p 无视频输入

任务：

- `task_id = task_3WFTgFLZs0VydG4pRbQJXq4qRQ0xlCcg`

状态结果：

- 业务查询：`SUCCESS`
- OpenAI 查询：`completed`
- `total_tokens = 108900`
- 业务记录 `resolution = 720p`

下载结果：

- `/content` 返回 `HTTP 200`
- 文件大小：`2,365,942 bytes`
- 实际视频尺寸：`1280x720`

计费核对：

- 当前档位：`720p no video = 46`
- 换算倍率：`46 / 2 = 23`
- 预扣：`5750000 quota = ¥11.5000`
- 最终 `actual_quota = 2504700`
- 实际金额：`¥5.0094`
- 退款：`3245300 quota = ¥6.4906`
- 核对公式：`108900 × 23 = 2504700`

结论：

- `接口正常`
- `计费正常`

## 4.4 `seedance2` 1080p 无视频输入

任务：

- `task_id = task_a2c7YrdDp9LFYDZeJQ72uQrdU7fCVL4E`

状态结果：

- 业务查询：`SUCCESS`
- OpenAI 查询：`completed`
- `total_tokens = 245025`
- 业务记录 `resolution = 1080p`

下载结果：

- `/content` 返回 `HTTP 200`
- 文件大小：`4,056,113 bytes`
- 实际视频尺寸：`1920x1080`

计费核对：

- 当前档位：`1080p no video = 51`
- 换算倍率：`25.5`
- 预扣：`6375000 quota = ¥12.7500`
- 最终 `actual_quota = 6248137`
- 实际金额：`¥12.4963`
- 退款：`126863 quota = ¥0.2537`
- 核对公式：`245025 × 25.5 = 6248137.5`
- 日志按整数 quota 记为：`6248137`

结论：

- `接口正常`
- `计费正常`

## 4.5 `seedance2` 720p 视频输入

成功任务：

- `task_id = task_7cXrvbzeNGWbkDhk4ICZxscg82PMkHoI`

先测出的两个边界错误：

### 错误 1：少 `role`

结果：

- `HTTP 400`

错误原文：

```json
{
  "error": {
    "code": "InvalidParameter",
    "message": "The parameter `content` specified in the request is not valid: reference media mode requires video role to be reference_video.",
    "type": "BadRequest"
  }
}
```

### 错误 2：输入视频被隐私拦截

结果：

- `HTTP 400`

错误原文：

```json
{
  "error": {
    "code": "***.PrivacyInformation",
    "message": "The request failed because the input video may contain real person.",
    "type": "BadRequest"
  }
}
```

最终成功场景结果：

- 业务查询：`SUCCESS`
- OpenAI 查询：`completed`
- `total_tokens = 216900`
- 业务记录 `resolution = 720p`

下载结果：

- `/content` 返回 `HTTP 200`
- 文件大小：`2,059,794 bytes`
- 实际视频尺寸：`1280x720`

计费核对：

- 当前档位：`720p with video = 31`
- 换算倍率：`15.5`
- 预扣：`3875000 quota = ¥7.7500`
- 最终 `actual_quota = 3361950`
- 实际金额：`¥6.7239`
- 退款：`513050 quota = ¥1.0261`
- 核对公式：`216900 × 15.5 = 3361950`

结论：

- `接口正常`
- `计费正常`
- 但对输入格式和内容审核要求较严格

## 4.6 `sd2.0fast`

任务：

- `task_id = task_jK23tZdbjOKl5z5TrisFWhhxTTF1AFYM`

状态结果：

- 业务查询：`SUCCESS`
- OpenAI 查询：`completed`
- `total_tokens = 108900`

下载结果：

- `/content` 返回 `HTTP 200`
- 文件大小：`2,890,056 bytes`
- 实际视频尺寸：`1280x720`

计费核对：

- 预扣：`4625000 quota = ¥9.2500`
- 最终 `actual_quota = 2014650`
- 实际金额：`¥4.0293`
- 退款：`2610350 quota = ¥5.2207`
- 核对公式：`108900 × 18.5 = 2014650`

结论：

- `接口正常`
- `计费正常`

## 4.7 `seedance2-sr` 720p

任务：

- `task_id = task_9MyFZ9cQuSaw1UjgDovb2BsPiZtHquU1`

状态结果：

- 业务查询：`SUCCESS`
- OpenAI 查询：`completed`
- 业务记录 `resolution = 480p`
- `total_tokens = 50638`

下载结果：

- `/content` 返回 `HTTP 200`
- 文件大小：`998,462 bytes`
- 实际视频尺寸：`864x496`

计费核对：

- 预扣：`5750000 quota = ¥11.5000`
- 最终 `actual_quota = 1164674`
- 实际金额：`¥2.3293`
- 退款：`4585326 quota = ¥9.1707`
- 实际日志倍率：`23`
- 核对公式：`50638 × 23 = 1164674`

结论：

- `接口可用`
- `下载可用`
- `计费与价格页不一致`
- `实际输出分辨率低于用户请求`

## 4.8 `seedance2-sr` 1080p

任务：

- `task_id = task_0JGkH3QTx0O87huGepGdJuClWIWdviMI`

状态结果：

- 业务查询：`SUCCESS`
- OpenAI 查询：`completed`
- 业务记录 `resolution = 720p`
- `total_tokens = 108900`

下载结果：

- `/content` 返回 `HTTP 200`
- 文件大小：`1,941,720 bytes`
- 实际视频尺寸：`1280x720`

计费核对：

- 预扣：`6375000 quota = ¥12.7500`
- 最终 `actual_quota = 2776950`
- 实际金额：`¥5.5539`
- 退款：`3598050 quota = ¥7.1961`
- 实际日志倍率：`25.5`
- 核对公式：`108900 × 25.5 = 2776950`

结论：

- `接口可用`
- `下载可用`
- `计费与价格页不一致`
- `实际输出分辨率低于用户请求`

## 4.9 `seedance2.0fast-sr` 720p

任务：

- `task_id = task_ZOQvITphQWqiwz7Ice560EGuIVxvAOcf`

状态结果：

- 业务查询：`SUCCESS`
- OpenAI 查询：`completed`
- 业务记录 `resolution = 480p`
- `total_tokens = 50638`

下载结果：

- `/content` 返回 `HTTP 200`
- 文件大小：`1,138,252 bytes`
- 实际视频尺寸：`864x496`

计费核对：

- 预扣：`5787500 quota = ¥11.5750`
- 最终 `actual_quota = 1172269`
- 实际金额：`¥2.3445`
- 退款：`4615231 quota = ¥9.2305`
- 核对公式：`50638 × 23.15 = 1172269.7`
- 日志按整数 quota 记为：`1172269`

结论：

- `接口可用`
- `计费与价格页一致`
- `实际输出分辨率低于用户请求`

---

## 5. 图片模型详细结果

## 5.1 `seedream4.5`

成功请求：

- `size = 2048x2048`

结果：

- `POST /v1/images/generations` 成功
- 返回图片 URL
- `usage.total_tokens = 16384`

计费核对：

- 价格：`model_price = 0.3`
- 实际日志：`quota = 150000`
- 实际金额：`¥0.3000`

非法尺寸测试：

- `size = 1024x1024`
- 返回 `HTTP 400`

错误原文：

```json
{
  "error": {
    "message": "The parameter `size` specified in the request is not valid: image size must be at least 3686400 pixels.",
    "type": "upstream_error",
    "code": "InvalidParameter"
  }
}
```

结论：

- `接口正常`
- `定价正常`
- `1024x1024` 当前不可用

## 5.2 `seedream5.0lite`

成功请求：

- `size = 2048x2048`

结果：

- 请求成功
- 返回图片 URL
- `usage.total_tokens = 16384`

计费核对：

- 价格：`model_price = 0.25`
- 实际日志：`quota = 125000`
- 实际金额：`¥0.2500`

非法尺寸测试：

- `size = 1024x1024`
- 返回 `HTTP 400`

错误原文：

```json
{
  "error": {
    "message": "The parameter `size` specified in the request is not valid: image size must be at least 3686400 pixels.",
    "type": "upstream_error",
    "code": "InvalidParameter"
  }
}
```

结论：

- `接口正常`
- `定价正常`
- `1024x1024` 当前不可用

---

## 6. 资产接口测试结果

## 6.1 官方真人资产

已测接口：

- `GET /api/portrait_assets/official/config`
- `GET /api/portrait_assets/official/jobs`
- `POST /api/portrait_assets/official/jobs`
- `POST /api/portrait_assets/official/jobs/{id}/validation`
- `POST /api/portrait_assets/official/upload`
- `POST /api/portrait_assets/official/jobs/{id}/asset`
- `GET /api/portrait_assets/official/jobs/{id}/preview/{state}`

结果：

- 配置查询：成功
- 列表查询：成功
- 当前账号检测到已有进行中任务：`id = 16`
- 再次创建任务时返回：
  - `你已有进行中的官方真人资产任务，请完成后再创建`
- 对现有 `id = 16` 刷新校验链接：成功
- 上传素材：成功
- 未完成人脸实名前提交素材：被正确拦截
  - 返回：`请先完成真人认证`
- 预览接口：返回 `302` 跳转

结论：

- `接口可用`
- `流程约束清晰`
- `真人实名前无法闭环完成`，这属于预期限制，不算接口坏

## 6.2 虚拟人像资产

已测接口：

- `GET /api/portrait_assets/virtual/config`
- `GET /api/portrait_assets/virtual/group`
- `GET /api/portrait_assets/virtual/assets`
- `POST /api/portrait_assets/virtual/upload`
- `POST /api/portrait_assets/virtual/assets`
- `POST /api/portrait_assets/virtual/assets/{id}/sync`
- `GET /api/portrait_assets/virtual/assets/{id}/preview/{state}`

结果：

- 配置查询：成功
- 分组查询：成功
- 列表查询：成功
- 上传 `512x512` 图片：成功
- 创建资产：成功
- 同步一次即变成 `active`
- 预览：`302` 跳转到真实素材地址

异常场景：

- 上传小图 `200x80`
- 上传接口本身成功
- 但创建资产时返回：

```json
{
  "success": false,
  "message": "volc portrait CreateAsset failed: HTTP 400: ... Width must be between 300px and 6000px."
}
```

结论：

- `接口可用`
- `异常拦截正常`

---

## 7. 负向测试补充

## 7.1 视频请求缺少 `prompt`

请求：

```json
{
  "model": "seedance1.5",
  "duration": 5,
  "resolution": "720p",
  "ratio": "16:9"
}
```

结果：

- `HTTP 400`

错误原文：

```json
{
  "code": "invalid_request",
  "message": "prompt is required",
  "data": null
}
```

## 7.2 传不存在的模型

请求：

```json
{
  "model": "seedance-not-exist",
  "prompt": "test",
  "duration": 5,
  "resolution": "720p",
  "ratio": "16:9"
}
```

结果：

- `HTTP 503`

错误原文：

```json
{
  "error": {
    "code": "model_not_found",
    "message": "No available channel for model seedance-not-exist under group default (distributor)",
    "type": "new_api_error"
  }
}
```

## 7.3 虚拟预览错误状态

结果：

- `GET /api/portrait_assets/virtual/assets/{id}/preview/badstate`
- 返回 `HTTP 401 Unauthorized`

结论：

- 状态签名校验正常

---

## 8. 本轮净消耗汇总

按本轮重新测试窗口内的 token 日志汇总，净消耗约：

```text
¥42.9570
```

按模型聚合：

| 模型 | 本轮净消耗 |
| --- | --- |
| `seedance1.5` | `¥1.7424` |
| `seedance1.5-sr` | `¥2.1780` |
| `seedance2` | `¥24.2296` |
| `sd2.0fast` | `¥4.0293` |
| `seedance2-sr` | `¥7.8832` |
| `seedance2.0fast-sr` | `¥2.3445` |
| `seedream4.5` | `¥0.3000` |
| `seedream5.0lite` | `¥0.2500` |

说明：

- `seedance2` 的 `¥24.2296` 包含 3 条成功任务：
  - 720p 无视频输入
  - 1080p 无视频输入
  - 720p 视频输入
- `seedance2-sr` 的 `¥7.8832` 包含 2 条成功任务：
  - 720p
  - 1080p

---

## 9. 证据与产物

本轮下载到本地的样例视频位于：

- `C:\Users\A\Desktop\whdj\tmp_api_retest_downloads`

其中包含：

- `seedance1_5.mp4`
- `seedance1_5-sr.mp4`
- `seedance2.mp4`
- `seedance2-1080p.mp4`
- `seedance2-video-input-720p.mp4`
- `sd2_0fast.mp4`
- `seedance2-sr.mp4`
- `seedance2-sr-1080p.mp4`
- `seedance2_0fast-sr.mp4`

---

## 10. 最重要的结论

### 10.1 正常可用且计费对得上的

- `seedance1.5`
- `seedance1.5-sr`
- `seedance2`
- `sd2.0fast`
- `seedream4.5`
- `seedream5.0lite`

### 10.2 明确存在问题的

#### 问题 1：`seedance2-sr` 价格页与实际结算不一致

价格页显示：

- `model_ratio = 24.5`

但本轮实测：

- `720p` 实际按 `23x` 结算
- `1080p` 实际按 `25.5x` 结算

这说明：

- 当前价格页展示与实际结算逻辑不一致

#### 问题 2：SR 模型实际下载清晰度低于用户请求

本轮实测：

| 模型 | 请求 | 实际下载 |
| --- | --- | --- |
| `seedance2-sr` | `720p` | `864x496` |
| `seedance2-sr` | `1080p` | `1280x720` |
| `seedance2.0fast-sr` | `720p` | `864x496` |

这说明：

- “SR 别名”当前并不是简单地按用户请求输出目标清晰度
- 至少从最终下载文件看，当前结果分辨率低于用户期望

### 10.3 不是故障，但需要明确写进文档或前端提示的

- `seedance2` 视频输入必须写 `role=reference_video`
- 视频输入可能被上游按“可能包含真人”拦截
- 官方真人资产当前账号有进行中任务时，不能重复创建
- 未完成真人实名前，官方真人素材提交会被拦截

