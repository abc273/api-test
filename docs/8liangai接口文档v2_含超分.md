1. # 8liangAI 接口文档

   ## 1. 文档说明

   - 站点地址：`https://8liangai.com`
   - 当前确认可用的 6 个模型别名：
     - 图片模型：`seedream4.5`、`seedream5.0lite`
     - 视频模型：`seedance2`、`sd2.0fast`、`seedance2-sr`、`sd2.0fast-sr`

   ## 2. 接入前必读

   ### 2.1 Base URL

   所有接口都基于以下地址调用：

   ```text
   https://8liangai.com
   ```

   ### 2.2 认证方式

   除公开预览地址外，本文大部分接口都需要 API Key 或登录态。

   推荐使用 API Key 调用时，在请求头中传：

   ```http
   Authorization: Bearer sk-你的_API_Key
   ```

   说明：

   - `Authorization` 必传。
   - `Bearer` 与密钥之间有一个空格。
   - 不要把 Key 放在 URL 查询参数中。

   ### 2.2.1 APIKey 怎么创建

   注意：创建 API 密钥时，分组需要设置为：默认分组。

   ### 2.2.2 统一鉴权失败返回示例

   如果未携带有效 API Key，或当前登录态无效，图片生成、视频生成、视频取流、资产列表等受保护接口通常会优先返回 `401 Unauthorized`。

   典型返回示例：

   ```json
   {
     "error": {
       "code": "",
       "message": "Invalid token (request id: xxxxx)",
       "type": "new_api_error"
     }
   }
   ```

   建议联调时优先先排查：

   1. `Authorization` 是否已正确传入
   2. `Bearer` 前缀是否带空格
   3. API Key 是否属于当前环境、当前用户、当前分组
   4. 是否把本应走登录态的接口误当成公开接口调用

   ### 2.3 Content-Type 规则

   | 场景      | Content-Type          | 说明                                             |
   | --------- | --------------------- | ------------------------------------------------ |
   | JSON 请求 | `application/json`    | 图片生成、视频生成、资产创建、状态同步等都用这个 |
   | 文件上传  | `multipart/form-data` | 真实资产/虚拟资产上传素材时使用                  |

   ### 2.4 常用的 4 个模型与接口对应关系

   | 模型别名          | 类型     | 调用接口                      | 推荐用途                                                 |
   | ----------------- | -------- | ----------------------------- | -------------------------------------------------------- |
   | `seedream4.5`     | 图片生成 | `POST /v1/images/generations` | 高质量图片生成                                           |
   | `seedream5.0lite` | 图片生成 | `POST /v1/images/generations` | 更轻量、更快的图片生成                                   |
   | `seedance2`       | 视频生成 | `POST /v1/video/generations`  | 标准视频生成，强烈建议显式传分辨率；联调时建议视为准必填 |
   | `sd2.0fast`       | 视频生成 | `POST /v1/video/generations`  | 更快的视频生成                                           |
   | `seedance2-sr`    | 视频生成 + 自动超分 | `POST /v1/video/generations`  | `seedance2` 专用超分版，适合最终拿到 `1080p`            |
   | `sd2.0fast-sr`    | 视频生成 + 自动超分 | `POST /v1/video/generations`  | `sd2.0fast` 专用超分版，适合最终拿到 `720p`              |

   ### 2.5 返回格式说明

   本项目不是所有接口都返回同一种 JSON 结构，请特别注意：

   1. 图片生成接口 `POST /v1/images/generations`
      直接返回图片结果对象，风格接近 OpenAI 图片接口。
   2. 视频提交接口 `POST /v1/video/generations`
      直接返回视频任务对象，包含 `task_id`。
   3. 视频任务查询接口 `GET /v1/video/generations/{task_id}`
      返回包装结构：`code`、`message`、`data`。
   4. 资产接口 `/api/portrait_assets/...`
      返回包装结构：`success`、`message`、`data`。

   接入时请不要把所有接口都按同一种返回体解析。

   ## 3. 调用顺序建议

   ### 3.1 图片生成

   1. 调用 `POST /v1/images/generations`
   2. 直接从返回里的 `url` 或 `b64_json` 取图

   说明：当前图片生成是同步返回结果，不需要额外轮询“图片状态”接口。

   ### 3.2 视频生成

   1. 调用 `POST /v1/video/generations` 提交任务
   2. 从返回中拿到 `task_id`
   3. 调用 `GET /v1/video/generations/{task_id}` 查询任务状态
   4. 任务成功后：
      - 可直接使用查询结果里的 `result_url`
      - 或调用 `GET /v1/videos/{task_id}/content` 取视频流

   #### 关于专用视频超分模型

   视频超分现在改为通过两个独立模型触发，不再通过 `metadata.enable_video_super_resolution` 开关触发：

   - `seedance2-sr`
   - `sd2.0fast-sr`

   对用户侧的实际表现是：

   1. 调用普通模型 `seedance2` / `sd2.0fast` 时，不会自动触发超分，也不会产生超分费用
   2. 调用专用模型 `seedance2-sr` / `sd2.0fast-sr` 时，系统会在原始视频生成完成后继续执行超分
   3. 如果您提交的是 `480p` 视频，最终返回的视频会提升到 `720p`
   4. 如果您提交的是 `720p` 视频，最终返回的视频会提升到 `1080p`
   5. 如果您提交的已经是 `1080p`，系统不会再继续超分
   6. 使用专用超分模型后，任务只有在超分处理完成后，才会对外显示为最终成功

   这意味着：

   - 普通模型下，`GET /v1/video/generations/{task_id}` 和 `GET /v1/videos/{task_id}/content` 返回的是原始生成视频
   - 专用超分模型下，`GET /v1/video/generations/{task_id}` 返回的最终 `result_url`，以及 `GET /v1/videos/{task_id}/content` 取到的，默认都是超分后的成品视频
   - 相比普通视频任务，专用超分模型的整体完成时间可能会略长，这是正常现象

   建议：

   - `seedance2`、`sd2.0fast`、`seedance2-sr`、`sd2.0fast-sr` 都显式传 `metadata.resolution`
   - 如果您希望最终拿到 `1080p`，最稳妥的传法是直接提交 `720p`，并使用 `seedance2-sr`
   - 如果您希望最终拿到 `720p`，可以直接提交 `480p`，并使用 `sd2.0fast-sr`
   - `sd2.0fast` 与 `sd2.0fast-sr` 仍然不要传 `1080p`

   ### 3.3 真实资产/虚拟资产使用

   当前只开放“查看真实资产、查看虚拟资产、拿到资产 ID、在视频生成中引用资产”这部分能力。

   建议顺序如下：

   1. 先查询真实资产或虚拟资产列表
   2. 找到可用资产对应的 `asset_id` / `volc_asset_id`
   3. 在视频生成接口中通过 `metadata.portrait_asset_id` 或 `metadata.asset_id` 引用

   ## 4. 图片生成接口

   ## 4.1 生成图片

   ### 接口地址

   ```text
   POST /v1/images/generations
   ```

   完整示例地址：

   ```text
   https://8liangai.com/v1/images/generations
   ```

   ### 适用模型

   - `seedream4.5`
   - `seedream5.0lite`

   ### 请求头

   | 请求头          | 是否必填 | 示例               | 说明        |
   | --------------- | -------- | ------------------ | ----------- |
   | `Authorization` | 是       | `Bearer sk-xxxx`   | API 认证    |
   | `Content-Type`  | 是       | `application/json` | JSON 请求体 |
   | `Accept`        | 否       | `application/json` | 建议带上    |

   ### 请求参数总表


   | 参数名               | 类型                     | 必填 | 传参位置  | 传法示例                         | 含义                           | 当前 4 模型建议                                              |
   | -------------------- | ------------------------ | ---- | --------- | -------------------------------- | ------------------------------ | ------------------------------------------------------------ |
   | `model`              | string                   | 是   | JSON body | `"seedream4.5"`                  | 模型别名                       | 必填                                                         |
   | `prompt`             | string                   | 是   | JSON body | `"写实商业人像，柔光，高清细节"` | 生成提示词                     | 必填                                                         |
   | `n`                  | uint                     | 否   | JSON body | `1`                              | 生成图片数量                   | 建议固定传 `1`                                               |
   | `size`               | string                   | 否   | JSON body | `"2048x2048"`                    | 输出尺寸                       | 建议显式传。`seedream5.0lite` 支持 `2K`/`3K`/`4K` 或明确像素值；`seedream4.5` 支持 `2K`/`4K` 或明确像素值。两者默认都是 `2048x2048` |
   | `quality`            | string                   | 否   | JSON body | `"hd"`                           | 生成质量档位                   | 如无明确要求可不传                                           |
   | `response_format`    | string                   | 否   | JSON body | `"url"` 或 `"b64_json"`          | 控制结果返回为 URL 还是 Base64 | 推荐 `url`                                                   |
   | `style`              | string / object / array  | 否   | JSON body | `"photorealistic"`               | 风格参数，原样透传             | 不作为标准参数                                               |
   | `user`               | string / object          | 否   | JSON body | `"customer-001"`                 | 业务侧用户标识                 | 建议传自己的用户 ID                                          |
   | `extra_fields`       | object / array / string  | 否   | JSON body | `{"foo":"bar"}`                  | 预留扩展字段                   | 无特殊需求不传                                               |
   | `background`         | string / object          | 否   | JSON body | `"transparent"`                  | 背景设置                       | 不作为标准参数，不建议传                                     |
   | `moderation`         | string / object          | 否   | JSON body | `"low"`                          | 安全审核相关参数               | 不作为标准参数，不建议传                                     |
   | `output_format`      | string / object          | 否   | JSON body | `"png"`                          | 输出格式                       | `seedream5.0lite` 支持 `png`/`jpeg`，默认 `jpeg`；`seedream4.5` 固定 `jpeg`，不要传这个字段 |
   | `output_compression` | number / string          | 否   | JSON body | `80`                             | 输出压缩率                     | 不作为标准参数，不建议传                                     |
   | `partial_images`     | number / object          | 否   | JSON body | `1`                              | 中间结果相关配置               | 不作为标准参数，不建议传                                     |
   | `watermark`          | boolean                  | 否   | JSON body | `false`                          | 是否加水印                     | 两个模型都支持；官方默认值为 `true`，如需无水印请显式传 `false` |
   | `watermark_enabled`  | boolean / object         | 否   | JSON body | `false`                          | 兼容部分上游的水印字段         | 当前这 2 个模型统一使用 `watermark`，不要再传这个字段        |
   | `user_id`            | string / number / object | 否   | JSON body | `"u_1001"`                       | 兼容上游的用户标识字段         | 当前这 2 个模型统一使用 `user` 即可                          |
   | `image`              | string / array / object  | 否   | JSON body | `"https://example.com/ref.png"`  | 图生图/多图生图输入            | 当前 `seedream4.5` 和 `seedream5.0lite` 已确认支持。可传单图或多图，最多 14 张参考图；支持 URL 或 Base64 |

   ### 当前这 2 个图片模型已确认支持的能力边界

   1. `seedream4.5` 和 `seedream5.0lite` 都支持：
      - 文生图
      - 单图生图
      - 多图生图
   2. 参考图 `image` 的已确认规则：
      - 最多 14 张
      - 支持 `jpeg`、`png`、`webp`、`bmp`、`tiff`、`gif`、`heic`、`heif`
      - 单张图大小不超过 30 MB
      - 单张图总像素不超过 `6000 x 6000 = 36000000`
   3. 如果要做“组图”而不是单图：
      - 请传 `sequential_image_generation: "auto"`
      - 如需限制最多返回多少张，再配 `sequential_image_generation_options.max_images`
      - 总规则是“输入参考图数量 + 最终生成图片数量 <= 15”
   4. 当前公开文档不把 `style`、`background`、`moderation`、`output_compression`、`partial_images` 当成标准交付参数承诺；这些字段虽可被网关接收，但不属于这 2 个模型的主推荐接入路径

   ### 关于 `size` 的建议

   1. 最稳妥的传法：直接传明确像素值，例如 `2048x2048`、`2304x1728`、`2848x1600`
   2. `seedream5.0lite` 也支持传分辨率档位：`2K`、`3K`、`4K`
   3. `seedream4.5` 支持传分辨率档位：`2K`、`4K`
   4. 如果传像素值，当前两模型都要同时满足：
      - 宽高比范围：`[1/16, 16]`
      - 总像素范围：`[3686400, 16777216]`

   ### 最常用的最小请求示例

   #### 示例 1：`seedream4.5`

   ```json
   {
     "model": "seedream4.5",
     "prompt": "一张写实风格的高端商业人像，柔和布光，皮肤纹理自然，细节丰富",
     "n": 1,
     "size": "2048x2048",
     "response_format": "url"
   }
   ```

   #### 示例 2：`seedream5.0lite`

   ```json
   {
     "model": "seedream5.0lite",
     "prompt": "干净的棚拍肖像，白色背景，真实肤感，商业摄影风格",
     "n": 1,
     "size": "2048x2048",
     "response_format": "url"
   }
   ```

   ### cURL 示例

   ```bash
   curl -X POST "https://8liangai.com/v1/images/generations" \
     -H "Authorization: Bearer sk-你的_API_Key" \
     -H "Content-Type: application/json" \
     -d '{
       "model": "seedream4.5",
       "prompt": "一张写实风格的高端商业人像，柔和布光，皮肤纹理自然，细节丰富",
       "n": 1,
       "size": "2048x2048",
       "response_format": "url"
     }'
   ```

   ### 返回字段说明

   | 字段                    | 类型   | 说明                                                |
   | ----------------------- | ------ | --------------------------------------------------- |
   | `created`               | int64  | 结果生成时间戳                                      |
   | `data`                  | array  | 图片结果数组                                        |
   | `data[].url`            | string | 图片 URL，`response_format=url` 时最常用            |
   | `data[].b64_json`       | string | 图片 Base64 内容，`response_format=b64_json` 时使用 |
   | `data[].revised_prompt` | string | 平台或上游可能改写后的提示词                        |
   | `metadata`              | object | 预留元数据，可能为空                                |

   ### 返回示例

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

   ### 接入注意事项

   1. 图片生成接口通常同步返回，不需要再查“图片任务状态”。
   2. 如果您要直接展示图片，优先取 `data[0].url`。
   3. 如果您要把图片存到自己的对象存储，再转发给前端，可取 `b64_json`。
   4. 如果您要做图生图，直接传 `image` 即可；当前 `seedream4.5` 和 `seedream5.0lite` 已确认支持该能力。

   ## 5. 视频生成接口

   ## 5.1 提交视频生成任务

   ### 接口地址

   ```text
   POST /v1/video/generations
   ```

   完整示例地址：

   ```text
   https://8liangai.com/v1/video/generations
   ```

   ### 适用模型

   - `seedance2`
   - `sd2.0fast`
   - `seedance2-sr`
   - `sd2.0fast-sr`

   ### 请求头

   | 请求头          | 是否必填 | 示例               | 说明        |
   | --------------- | -------- | ------------------ | ----------- |
   | `Authorization` | 是       | `Bearer sk-xxxx`   | API 认证    |
   | `Content-Type`  | 是       | `application/json` | JSON 请求体 |
   | `Accept`        | 否       | `application/json` | 建议带上    |

   ### 顶层请求参数

   | 参数名            | 类型     | 必填 | 传参位置  | 传法示例                        | 含义                             | 建议                                                         |
   | ----------------- | -------- | ---- | --------- | ------------------------------- | -------------------------------- | ------------------------------------------------------------ |
   | `model`           | string   | 是   | JSON body | `"seedance2"`                   | 视频模型别名                     | 必填                                                         |
   | `prompt`          | string   | 是   | JSON body | `"人物缓慢转头并向镜头挥手"`    | 视频生成提示词                   | 必填                                                         |
   | `mode`            | string   | 否   | JSON body | `"standard"`                    | 预留模式字段                     | 当前无特殊要求可不传                                         |
   | `image`           | string   | 否   | JSON body | `"https://example.com/1.png"`   | 单张图片输入                     | 当前视频模型更推荐用 `images`                                |
   | `images`          | string[] | 否   | JSON body | `["https://example.com/1.png"]` | 参考图数组，会作为参考图注入上游 | 强烈推荐使用这个字段传参考图。当前 `seedance2` / `sd2.0fast` 最多支持 9 张参考图 |
   | `size`            | string   | 否   | JSON body | `"1280x720"`                    | 预留尺寸字段                     | 当前更推荐传 `metadata.resolution` + `metadata.ratio`        |
   | `duration`        | int      | 否   | JSON body | `5`                             | 视频时长（秒）                   | 建议显式传 `4-15` 的正整数。当前 8liangAI 顶层字段最适合这么传 |
   | `seconds`         | string   | 否   | JSON body | `"5"`                           | 兼容字符串形式时长               | 有 `duration` 时不必再传                                     |
   | `input_reference` | string   | 否   | JSON body | `"asset://asset_xxx"`           | 预留参考输入字段                 | 当前不建议优先使用                                           |
   | `metadata`        | object   | 否   | JSON body | 见下方                          | 视频高级参数集合                 | 当前视频模型大量关键参数放这里；网关会把它整理成官方 `content[]` + 顶层参数结构发给上游 |

   ### `metadata` 参数总表

   当前 `seedance2` / `sd2.0fast` / `seedance2-sr` / `sd2.0fast-sr` 能接收并转上游的视频扩展参数如下：

   | 参数名                          | 类型         | 必填     | 传法示例                               | 含义                     | 建议                                                         |
   | ------------------------------- | ------------ | -------- | -------------------------------------- | ------------------------ | ------------------------------------------------------------ |
   | `resolution`                    | string       | 强烈建议 | `"720p"`、`"1080p"`                    | 输出分辨率               | 请显式传。`seedance2` 在不传分辨率时，实际联调中很容易直接触发价格档位匹配失败；`seedance2` 支持 `480p`/`720p`/`1080p`；`sd2.0fast` 只建议用 `480p`/`720p`，不支持 `1080p` |
   | `ratio`                         | string       | 强烈建议 | `"16:9"`、`"9:16"`、`"1:1"`、`"3:4"`   | 输出画幅比               | 建议显式传。当前支持 `16:9`、`4:3`、`1:1`、`3:4`、`9:16`、`21:9`、`adaptive` |
   | `watermark`                     | boolean      | 否       | `false`                                | 是否加水印               | 建议按业务需要明确传                                         |
   | `seed`                          | int          | 否       | `60740`                                | 随机种子，便于复现       | 需要可复现时传                                               |
   | `frames`                        | int          | 否       | `24`                                   | 帧相关控制参数           | 当前 `seedance2` / `sd2.0fast` 暂不支持，不要传              |
   | `camera_fixed`                  | boolean      | 否       | `true`                                 | 是否固定镜头             | 当前 `seedance2` / `sd2.0fast` 暂不支持，不要传              |
   | `callback_url`                  | string       | 否       | `"https://yourdomain.com/callback"`    | 上游完成后的回调地址     | 仅您有回调服务时传                                           |
   | `return_last_frame`             | boolean      | 否       | `true`                                 | 是否返回最后一帧         | 按需使用                                                     |
   | `service_tier`                  | string       | 否       | `"default"`                            | 服务等级                 | 当前 `seedance2` / `sd2.0fast` 不支持配置，直接不要传        |
   | `execution_expires_after`       | int          | 否       | `172800`                               | 任务过期秒数             | 一般不传                                                     |
   | `generate_audio`                | boolean      | 否       | `true`                                 | 是否生成音频             | 当前 `seedance2` / `sd2.0fast` 支持；官方默认值为 `true`，如需无声视频请显式传 `false` |
   | `draft`                         | boolean      | 否       | `false`                                | 是否草稿模式             | 当前 `seedance2` / `sd2.0fast` 不支持，不要传                |
   | `tools`                         | array        | 否       | `[{"type":"web_search"}]`              | 工具扩展                 | 当前四模型通常不需要                                         |
   | `content`                       | array        | 否       | 见后文                                 | 高级多模态内容数组       | 用来直通官方 `content[]` 结构；不建议手写，确需传音频/视频/首尾帧时再用 |
   | `duration`                      | int          | 否       | `5`                                    | 放在 metadata 里的时长   | 如需兼容官方结构可传；普通接入仍建议优先传顶层 `duration`    |
   | `portrait_asset_id`             | int / string | 否       | `12`                                   | 引用“真实资产任务 ID”    | 资产玩法强烈推荐                                             |
   | `asset_id`                      | string       | 否       | `"asset_xxx"` 或 `"asset://asset_xxx"` | 直接引用资产 ID          | 已有资产时可用                                               |

   ### 关于 `duration` 的优先级

   如果您既传了顶层 `duration`，又传了 `metadata.duration`，系统会优先使用顶层 `duration`。

   建议做法：

   - 普通：只传顶层 `duration`
   - 高级：如需跟上游参数结构保持一致，可在 `metadata` 中同步保留，但以顶层为准

   补充说明：

   - 火山方舟官方对 Seedance 2.0 系列支持 `duration=-1` 的“智能时长”模式
   - 但在当前 8liangAI 接入里，顶层 `duration` 最稳妥的用法仍是传 `4-15` 的正整数
   - 如果您确实要走上游的“智能时长”语义，请改传 `metadata.duration = -1`，不要只在顶层传 `-1`

   ### 关于参考图 `images`

   对当前 2 个视频模型，建议这样理解：

   - `images` 是“参考图数组”
   - 数组里的每一张图都会被作为 `reference_image` 传给上游
   - 如果您只传单个 `image` 而不传 `images`，不如直接传 `images: ["..."]` 稳妥
   - 当前 `seedance2` / `sd2.0fast` 最多支持 9 张参考图
   - 参考图支持 `jpeg`、`png`、`webp`、`bmp`、`tiff`、`gif`、`heic`、`heif`

   ### 关于 Seedance 2.0 系列的人脸素材限制

   官方文档明确说明：Seedance 2.0 系列不支持直接上传“含真人人脸”的参考图或参考视频。

   这对当前两个视频模型的接入意味着：

   1. 普通风景、物品、非真人脸参考图，可直接走 `images`
   2. 如果素材里有人脸，优先走平台的真实资产 / 虚拟资产能力，再通过 `metadata.portrait_asset_id` 或 `metadata.asset_id` 引用
   3. 这也是为什么本项目同时开放了真实资产、虚拟资产查询与引用能力

   ### 关于资产引用

   如果您已经通过资产接口拿到了可用资产，可以在视频生成接口里复用：

   - `metadata.portrait_asset_id`
     传真实资产任务 ID，系统会自动解析成资产引用。
   - `metadata.asset_id`
     直接传资产 ID 字符串，支持原始 `asset_xxx` 或 `asset://asset_xxx` 形式。

   前提：

   - 资产必须属于当前调用者
   - 资产必须已经是可用状态

   ### 最常用请求示例

   #### 示例 1：`seedance2` 文生视频

   ```json
   {
     "model": "seedance2",
     "prompt": "人物缓慢转头并向镜头挥手，写实风格，电影感光影，动作自然",
     "duration": 5,
     "metadata": {
       "resolution": "1080p",
       "ratio": "16:9",
       "watermark": false
     }
   }
   ```

   #### 示例 2：`sd2.0fast` 参考图视频

   ```json
   {
     "model": "sd2.0fast",
     "prompt": "基于参考图保持人物服装颜色和材质不变，人物缓慢向前走一步并轻微点头，写实风格",
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

   #### 示例 3：引用真实/虚拟资产

   ```json
   {
     "model": "seedance2",
     "prompt": "让该人物做一个自然抬头并微笑的动作，镜头稳定，写实风格",
     "duration": 5,
     "metadata": {
       "resolution": "1080p",
       "ratio": "9:16",
       "watermark": false,
       "portrait_asset_id": 12
     }
   }
   ```

   ### cURL 示例

   ```bash
   curl -X POST "https://8liangai.com/v1/video/generations" \
     -H "Authorization: Bearer sk-你的_API_Key" \
     -H "Content-Type: application/json" \
     -d '{
       "model": "seedance2",
       "prompt": "人物缓慢转头并向镜头挥手，写实风格，电影感光影，动作自然",
       "duration": 5,
       "metadata": {
         "resolution": "1080p",
         "ratio": "16:9",
         "watermark": false
       }
     }'
   ```

   ### 返回字段说明

   提交成功后，接口会直接返回一个视频任务对象：

   | 字段         | 类型   | 说明                             |
   | ------------ | ------ | -------------------------------- |
   | `id`         | string | 公开任务 ID                      |
   | `task_id`    | string | 与 `id` 一致，建议后续查询时使用 |
   | `object`     | string | 固定为 `video`                   |
   | `model`      | string | 您请求的模型别名                 |
   | `status`     | string | 初始通常是 `queued`              |
   | `progress`   | int    | 初始通常为 `0`                   |
   | `created_at` | int64  | 任务创建时间戳                   |

   ### 返回示例

   ```json
   {
     "id": "task_keM6sME4sTVzpWvYhgVQ2bFJVi04Bwwe",
     "task_id": "task_keM6sME4sTVzpWvYhgVQ2bFJVi04Bwwe",
     "object": "video",
     "model": "sd2.0fast",
     "status": "queued",
     "progress": 0,
     "created_at": 1779546653
   }
   ```

   ### 接入注意事项

   1. 视频生成是异步任务，不能像图片一样立即拿最终视频。
   2. `seedance2` 建议始终传 `metadata.resolution`，优先用 `720p` 或 `1080p`
   3. `sd2.0fast` 不要传 `1080p`，优先用 `720p`
   4. 如果您有参考图，优先用 `images` 数组，不要只传 `image`
   5. 如果参考素材里有真人人脸，优先改走真实资产 / 虚拟资产，不要直接把图片 URL 塞给 `images`
   6. 视频超分现在由 `seedance2-sr` 和 `sd2.0fast-sr` 两个专用模型触发，不再通过 `metadata.enable_video_super_resolution` 开关触发
   7. 如果您使用专用超分模型并提交 `480p`，最终通常会拿到 `720p`；如果提交 `720p`，最终通常会拿到 `1080p`
   8. 真实资产和虚拟资产接入成功后，视频接口可以直接复用资产，减少重复上传

   ## 5.2 查询视频任务状态（业务态接口）

   ### 接口地址

   ```text
   GET /v1/video/generations/{task_id}
   ```

   完整示例地址：

   ```text
   https://8liangai.com/v1/video/generations/task_xxx
   ```

   ### 用途

   这是最推荐使用的“视频任务状态查询接口”，适合：

   - 查询任务是否完成
   - 查询失败原因
   - 查询内部进度
   - 获取结果地址
   - 获取原始上游返回数据

   ### 路径参数

   | 参数名    | 类型   | 必填 | 示例                                    | 说明                        |
   | --------- | ------ | ---- | --------------------------------------- | --------------------------- |
   | `task_id` | string | 是   | `task_keM6sME4sTVzpWvYhgVQ2bFJVi04Bwwe` | 提交视频任务时返回的任务 ID |

   ### 请求头

   | 请求头          | 是否必填 | 示例             | 说明     |
   | --------------- | -------- | ---------------- | -------- |
   | `Authorization` | 是       | `Bearer sk-xxxx` | API 认证 |

   ### 返回结构

   该接口返回包装结构：

   | 字段      | 类型   | 说明                   |
   | --------- | ------ | ---------------------- |
   | `code`    | string | 成功时通常为 `success` |
   | `message` | string | 提示信息，可能为空     |
   | `data`    | object | 任务详情               |

   ### `data` 字段说明

   | 字段                             | 类型   | 说明                                                         |
   | -------------------------------- | ------ | ------------------------------------------------------------ |
   | `id`                             | int64  | 数据库记录 ID                                                |
   | `created_at`                     | int64  | 记录创建时间                                                 |
   | `updated_at`                     | int64  | 记录更新时间                                                 |
   | `task_id`                        | string | 公开任务 ID                                                  |
   | `platform`                       | string | 任务平台标识                                                 |
   | `user_id`                        | int    | 用户 ID                                                      |
   | `group`                          | string | 用户分组                                                     |
   | `channel_id`                     | int    | 渠道 ID                                                      |
   | `quota`                          | int    | 本次任务消耗额度                                             |
   | `action`                         | string | 动作类型，通常为 `generate`                                  |
   | `status`                         | string | 当前任务状态，当前四模型常见值为 `QUEUED`、`IN_PROGRESS`、`SUCCESS`、`FAILURE` |
   | `fail_reason`                    | string | 失败原因；成功时通常为空；部分实现里成功后也可能存放结果 URL，请以 `result_url` 和 `data.content.video_url` 为准 |
   | `result_url`                     | string | 最终视频地址，任务成功时最有用                               |
   | `submit_time`                    | int64  | 提交时间                                                     |
   | `start_time`                     | int64  | 开始处理时间                                                 |
   | `finish_time`                    | int64  | 处理完成时间                                                 |
   | `progress`                       | string | 进度字符串，如 `20%`、`30%`、`100%`                          |
   | `properties.input`               | string | 预留输入信息                                                 |
   | `properties.upstream_model_name` | string | 实际映射到上游的模型名                                       |
   | `properties.origin_model_name`   | string | 您调用时传入的模型别名                                       |
   | `data`                           | object | 上游原始任务数据，便于高级排障                               |

   ### 状态值建议解释

   | 状态值        | 建议业务解释               |
   | ------------- | -------------------------- |
   | `QUEUED`      | 已接单，排队中             |
   | `IN_PROGRESS` | 处理中                     |
   | `SUCCESS`     | 已成功，可取视频           |
   | `FAILURE`     | 失败，请读取 `fail_reason` |

   ### 返回示例

   ```json
   {
     "code": "success",
     "message": "",
     "data": {
       "id": 76,
       "created_at": 1779546653,
       "updated_at": 1779546796,
       "task_id": "task_keM6sME4sTVzpWvYhgVQ2bFJVi04Bwwe",
       "platform": "45",
       "user_id": 1,
       "group": "default",
       "channel_id": 7,
       "quota": 4625000,
       "action": "generate",
       "status": "SUCCESS",
       "fail_reason": "",
       "result_url": "https://example.com/result.mp4",
       "submit_time": 1779546653,
       "start_time": 1779546668,
       "finish_time": 1779546796,
       "progress": "100%",
       "properties": {
         "input": "",
         "upstream_model_name": "ep-20260430104948-vv5vf",
         "origin_model_name": "sd2.0fast"
       },
       "data": {
         "content": {
           "video_url": "https://example.com/result.mp4"
         },
         "duration": 5,
         "ratio": "3:4",
         "resolution": "720p",
         "seed": 60740,
         "status": "succeeded"
       }
     }
   }
   ```

   ## 5.3 查询视频任务状态（OpenAI 兼容接口）

   ### 接口地址

   ```text
   GET /v1/videos/{task_id}
   ```

   ### 适合谁用

   如果您的 SDK 或客户端已经按 OpenAI Video 风格封装，建议使用这个接口。

   ### 返回字段

   | 字段            | 类型   | 说明                                           |
   | --------------- | ------ | ---------------------------------------------- |
   | `id`            | string | 任务 ID                                        |
   | `task_id`       | string | 兼容字段，通常与 `id` 相同                     |
   | `object`        | string | 固定 `video`                                   |
   | `model`         | string | 模型别名                                       |
   | `status`        | string | `queued`、`in_progress`、`completed`、`failed` |
   | `progress`      | int    | 整数进度                                       |
   | `created_at`    | int64  | 创建时间                                       |
   | `completed_at`  | int64  | 完成时间                                       |
   | `metadata.url`  | string | 原始视频地址                                   |
   | `error.message` | string | 失败信息                                       |
   | `error.code`    | string | 失败码                                         |

   ### 返回示例

   ```json
   {
     "id": "task_xxx",
     "task_id": "task_xxx",
     "object": "video",
     "model": "seedance2",
     "status": "completed",
     "progress": 100,
     "created_at": 1779546653,
     "completed_at": 1779546796,
     "metadata": {
       "url": "https://example.com/result.mp4"
     }
   }
   ```

   ## 5.4 查看视频内容

   ### 接口地址

   ```text
   GET /v1/videos/{task_id}/content
   ```

   ### 用途

   直接拉取视频流内容，适合：

   - 前端 `<video src>` 直连
   - 服务端下载成文件
   - 统一走平台代理地址，而不是直接暴露上游原始地址

   ### 路径参数

   | 参数名    | 类型   | 必填 | 示例       | 说明        |
   | --------- | ------ | ---- | ---------- | ----------- |
   | `task_id` | string | 是   | `task_xxx` | 视频任务 ID |

   ### 调用前提

   - 任务必须已成功
   - 任务必须属于当前用户或当前 API Key 所属用户

   ### 请求头

   | 请求头          | 是否必填 | 示例             | 说明                                         |
   | --------------- | -------- | ---------------- | -------------------------------------------- |
   | `Authorization` | 是       | `Bearer sk-xxxx` | 鉴权必填。未携带有效凭证时会优先返回 `401`   |
   | `Accept`        | 否       | `video/mp4`      | 前端直连时可不传；服务端下载时也可按默认处理 |

   ### 返回说明

   - 成功时：直接返回视频二进制流，`Content-Type` 由上游决定，常见为 `video/mp4`
   - 失败时：返回 JSON 错误对象

   ### 常见错误

   | HTTP 状态码 | 场景             | 说明                                             |
   | ----------- | ---------------- | ------------------------------------------------ |
   | `401`       | 未鉴权或凭证无效 | 例如未传 API Key、Key 无效、当前任务不属于该凭证 |
   | `400`       | 任务未完成       | 例如任务还在排队或处理中                         |
   | `404`       | 任务不存在       | 任务 ID 错误或无权访问                           |
   | `502`       | 上游取流失败     | 上游视频地址失效或临时异常                       |

   ### 错误示例

   鉴权失败示例：

   ```json
   {
     "error": {
       "message": "Invalid token (request id: xxxxx)",
       "type": "new_api_error"
     }
   }
   ```

   任务未完成示例：

   ```json
   {
     "error": {
       "message": "Task is not completed yet, current status: IN_PROGRESS",
       "type": "invalid_request_error"
     }
   }
   ```

   ## 6. 图片查看方式说明

   当前四模型里的图片生成没有单独的“图片状态查询接口”或“图片内容代理接口”，查看图片有两种方式：

   1. 如果图片生成返回的是 `url`
      直接使用 `data[].url` 即可。
   2. 如果图片生成返回的是 `b64_json`
      由您的服务端或前端解码后再显示。

   如果您接的是“资产预览”而不是“图片生成结果”，请看第 7 节和第 8 节中的预览接口。

   ## 7. 真实资产接口

   说明：

   - 本节说的“真实资产”对应接口前缀：`/api/portrait_assets/official/...`
   - 当前仅开放“查看真实资产列表、查看真实资产 ID、查看预览”
   - 真实资产列表等查询接口属于受保护接口，需要有效 API Key 或有效登录态
   - 资产预览地址请直接使用列表接口返回的完整 URL，不要自行拼接
   - 这类接口返回统一包装：

   ```json
   {
     "success": true,
     "message": "",
     "data": {}
   }
   ```

   ## 7.1 查询真实资产列表

   ### 接口地址

   ```text
   GET /api/portrait_assets/official/jobs
   ```

   ### 请求头

   | 请求头          | 是否必填 | 示例               | 说明                          |
   | --------------- | -------- | ------------------ | ----------------------------- |
   | `Authorization` | 是       | `Bearer sk-xxxx`   | 需要有效 API Key 或有效登录态 |
   | `Accept`        | 否       | `application/json` | 建议带上                      |

   ### 分页参数

   | 参数名      | 类型 | 必填 | 默认值   | 说明                             |
   | ----------- | ---- | ---- | -------- | -------------------------------- |
   | `p`         | int  | 否   | `1`      | 页码，从 1 开始                  |
   | `page_size` | int  | 否   | 平台默认 | 每页数量，最大建议按平台限制使用 |

   ### 返回分页结构

   | 字段        | 类型  | 说明     |
   | ----------- | ----- | -------- |
   | `page`      | int   | 当前页   |
   | `page_size` | int   | 每页条数 |
   | `total`     | int   | 总条数   |
   | `items`     | array | 任务列表 |

   ### `items[]` 字段说明

   | 字段                   | 类型   | 说明                                                         |
   | ---------------------- | ------ | ------------------------------------------------------------ |
   | `id`                   | int    | 真实资产任务 ID。如需和平台沟通某条资产记录，通常报这个值    |
   | `user_id`              | int    | 用户 ID                                                      |
   | `name`                 | string | 任务/资产名称                                                |
   | `source`               | string | 固定为 `official`                                            |
   | `status`               | string | 当前业务状态                                                 |
   | `invite_url`           | string | 真人认证链接                                                 |
   | `validate_result_code` | string | 认证结果码                                                   |
   | `volc_group_id`        | string | 上游资产组 ID                                                |
   | `asset_id`             | string | 真实资产 ID。最需要关注的字段之一，后续视频生成引用资产时可用 |
   | `asset_status`         | string | 上游资产状态                                                 |
   | `asset_preview`        | string | 预览地址                                                     |
   | `asset_url`            | string | 上传时的原始素材 URL                                         |
   | `asset_type`           | string | `Image` / `Video` / `Audio`                                  |
   | `project_name`         | string | 项目名                                                       |
   | `error_message`        | string | 失败原因                                                     |
   | `created_time`         | int64  | 创建时间戳                                                   |
   | `updated_time`         | int64  | 更新时间戳                                                   |
   | `qr_expires_time`      | int64  | 认证链接过期时间                                             |
   | `ready_time`           | int64  | 资产可用时间                                                 |

   ### `status` 常见值

   | 状态值             | 说明                     |
   | ------------------ | ------------------------ |
   | `pending`          | 任务已创建，待处理       |
   | `validate_ready`   | 认证链接已生成           |
   | `validated`        | 真人认证已通过           |
   | `asset_processing` | 素材入库处理中           |
   | `pending_confirm`  | 预览已就绪，等待确认     |
   | `ready`            | 已可用，可被视频接口引用 |
   | `failed`           | 失败                     |
   | `expired`          | 认证链接已过期           |
   | `disabled`         | 已禁用                   |

   ### 建议重点关注字段

   如果当前只需要“查看真实资产并拿到资产 ID”，建议重点关注以下字段：

   | 字段            | 用途                            |
   | --------------- | ------------------------------- |
   | `id`            | 平台内部记录 ID，便于和平台沟通 |
   | `asset_id`      | 真实资产 ID，后续视频生成最常用 |
   | `name`          | 区分资产名称                    |
   | `status`        | 判断资产是否已可用              |
   | `asset_preview` | 查看资产预览                    |

   ## 7.2 查看真实资产预览

   ### 接口地址

   ```text
   GET /api/portrait_assets/official/jobs/{id}/preview/{state}
   ```

   ### 使用方式

   这个接口里的 `{state}` 不是让自己拼的，而是平台在列表接口返回的 `asset_preview` 完整 URL 中已经带好了。

   正确做法：

   1. 先调 `GET /api/portrait_assets/official/jobs`
   2. 直接读取返回里的 `asset_preview`
   3. 用这个完整 URL 做预览访问

   不要自己手工拼接 `{state}`，否则大概率鉴权失败。

   ## 8. 虚拟资产接口

   说明：

   - 本节说的“虚拟资产”对应接口前缀：`/api/portrait_assets/virtual/...`
   - 当前对仅开放“查看虚拟资产列表、查看虚拟资产 ID、查看预览”
   - 创建、上传、同步、分组等接口暂不在文档中披露
   - 虚拟资产列表等查询接口属于受保护接口，需要有效 API Key 或有效登录态

   ## 8.1 查询虚拟资产列表

   ### 接口地址

   ```text
   GET /api/portrait_assets/virtual/assets
   ```

   ### 请求头

   | 请求头          | 是否必填 | 示例               | 说明                          |
   | --------------- | -------- | ------------------ | ----------------------------- |
   | `Authorization` | 是       | `Bearer sk-xxxx`   | 需要有效 API Key 或有效登录态 |
   | `Accept`        | 否       | `application/json` | 建议带上                      |

   ### 分页参数

   | 参数名      | 类型 | 必填 | 默认值   | 说明            |
   | ----------- | ---- | ---- | -------- | --------------- |
   | `p`         | int  | 否   | `1`      | 页码，从 1 开始 |
   | `page_size` | int  | 否   | 平台默认 | 每页数量        |

   ### `items[]` 字段说明

   | 字段            | 类型   | 说明                                                         |
   | --------------- | ------ | ------------------------------------------------------------ |
   | `id`            | int    | 虚拟资产记录 ID。如需和平台沟通某条资产记录，通常报这个值    |
   | `user_id`       | int    | 用户 ID                                                      |
   | `group_id`      | int    | 分组 ID                                                      |
   | `name`          | string | 资产名称                                                     |
   | `asset_type`    | string | `Image` / `Video` / `Audio`                                  |
   | `source_url`    | string | 上传时的原始素材地址                                         |
   | `preview_url`   | string | 预览地址                                                     |
   | `project_name`  | string | 项目名                                                       |
   | `volc_group_id` | string | 上游组 ID                                                    |
   | `volc_asset_id` | string | 虚拟资产 ID。最需要关注的字段之一，后续视频生成引用资产时可用 |
   | `status`        | string | 当前业务状态                                                 |
   | `volc_status`   | string | 上游状态                                                     |
   | `error_message` | string | 错误原因                                                     |
   | `created_time`  | int64  | 创建时间                                                     |
   | `updated_time`  | int64  | 更新时间                                                     |
   | `ready_time`    | int64  | 可用时间                                                     |

   ### `status` 常见值

   | 状态值       | 说明   |
   | ------------ | ------ |
   | `processing` | 处理中 |
   | `active`     | 已可用 |
   | `failed`     | 失败   |

   ### 建议重点关注字段

   如果当前只需要“查看虚拟资产并拿到资产 ID”，建议重点关注以下字段：

   | 字段            | 用途                            |
   | --------------- | ------------------------------- |
   | `id`            | 平台内部记录 ID，便于和平台沟通 |
   | `volc_asset_id` | 虚拟资产 ID，后续视频生成最常用 |
   | `name`          | 区分资产名称                    |
   | `asset_type`    | 判断是图片、视频还是音频资产    |
   | `status`        | 判断资产是否已可用              |
   | `preview_url`   | 查看资产预览                    |

   ## 8.2 查看虚拟资产预览

   ### 接口地址

   ```text
   GET /api/portrait_assets/virtual/assets/{id}/preview/{state}
   ```

   ### 使用方式

   和真实资产预览一样，不要自己拼接 `{state}`。正确方式是：

   1. 先调用 `GET /api/portrait_assets/virtual/assets`
   2. 直接使用返回里的 `preview_url`

   ## 9. 资产与视频生成联动方式

   如果的核心业务是“先创建资产，再生成视频”，推荐按以下方式对接：

   ### 9.1 使用真实资产任务 ID

   在视频生成请求中这样传：

   ```json
   {
     "model": "seedance2",
     "prompt": "让该人物自然抬头并微笑",
     "duration": 5,
     "metadata": {
       "resolution": "1080p",
       "ratio": "9:16",
       "portrait_asset_id": 12
     }
   }
   ```

   说明：

   - `12` 是真实资产列表里的 `id`
   - 该真实资产必须是 `ready`

   ### 9.2 使用资产 ID

   在视频生成请求中这样传：

   ```json
   {
     "model": "sd2.0fast",
     "prompt": "让该资产人物轻轻挥手",
     "duration": 5,
     "metadata": {
       "resolution": "720p",
       "ratio": "16:9",
       "asset_id": "asset_xxx"
     }
   }
   ```

   也可以传：

   ```json
   {
     "metadata": {
       "asset_id": "asset://asset_xxx"
     }
   }
   ```

   ## 10. 常见错误与排查建议

   ### 10.1 视频提交成功，但一直拿不到最终视频

   排查顺序：

   1. 用 `GET /v1/video/generations/{task_id}` 看 `status`
   2. 如果 `status=FAILURE`，读 `fail_reason`
   3. 如果 `status=SUCCESS`，优先看 `result_url`
   4. 如果前端直连 `result_url` 有跨域或有效期问题，改用 `GET /v1/videos/{task_id}/content`
   5. 如果模型是 `seedance2-sr` 或 `sd2.0fast-sr`，任务在原始视频生成完成后还会继续进入超分阶段，因此整体耗时会比普通任务更长

   ### 10.2 `seedance2` 提交时报参数或价格档位错误

   建议优先检查：

   1. 是否显式传了 `metadata.resolution`
   2. 是否显式传了 `duration`
   3. 是否把参考图放在 `images` 数组，而不是只放 `image`
   4. 如果模型是 `sd2.0fast`，是否错误地传了 `1080p`
   5. 如果参考图里有人脸，是否应该改为走资产引用，而不是直接传 `images`

   ### 10.3 资产预览 URL 打不开

   先确认：

   1. 不是自己拼接的预览链接
   2. 是从列表接口返回里直接拿到的 `asset_preview` 或 `preview_url`
   3. 资产状态已经进入可预览或可用阶段

   ### 10.4 资产创建成功，但视频里引用时报无权限或未就绪

   常见原因：

   1. 资产不属于当前 API Key 对应用户
   2. 资产还不是 `ready` / `active`
   3. `portrait_asset_id` 传成了别的字段 ID
   4. `asset_id` 为空、格式不对，或引用了别人的资产

   ## 11. 接入建议总结

   1. 图片模型就用 `POST /v1/images/generations`，同步拿结果。
   2. 视频模型就用 `POST /v1/video/generations` 提交，再用 `GET /v1/video/generations/{task_id}` 轮询。
   3. 取视频成品时，优先用 `GET /v1/videos/{task_id}/content`，最稳定。
   4. `seedance2` 和 `sd2.0fast` 建议都显式传 `duration`、`metadata.resolution`、`metadata.ratio`、`metadata.watermark`。
   5. `sd2.0fast` 不支持 `1080p`，默认按 `720p` 思路接最稳。
   6. 图片模型当前已支持图生图；有参考图时，图片接口直接传 `image`，视频接口优先传 `images` 数组。
   7. 视频参考素材如果含真人人脸，优先走资产能力，不要直接把原图 / 原视频塞给 `images`。
   8. 做资产能力时，不要自己拼预览地址和 state 参数，直接用列表接口返回的完整预览 URL。
