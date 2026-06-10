# 8liangai Vidu 四模型接口测试报告

测试时间：2026-06-09  
测试域名：`https://8liangai.com`

本报告仅记录本次在线实测结果，与接口文档分离。

## 1. 测试范围

已测试模型：
- `vidu-q1`
- `vidu-q2`
- `vidu-q3-turbo`
- `viduq-3`

已测试接口：
- `GET /v1/models`
- `POST /v1/images/generations`
- `POST /v1/video/generations`
- `GET /v1/video/generations/{task_id}`
- `GET /v1/videos/{task_id}`
- `GET /v1/videos/{task_id}/content`

## 2. 模型可见性

`GET /v1/models` 可见以下 4 个模型：
- `vidu-q1`
- `vidu-q2`
- `vidu-q3-turbo`
- `viduq-3`

结论：模型列表已正确展示。

## 3. 生图模型测试

### `vidu-q1`

请求接口：
- `POST /v1/images/generations`

结果：
- HTTP 状态码：`404`
- 返回：

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

结论：
- 模型已展示在模型列表中。
- 官方能力包含图片生成，但当前站点的该图片调用未打通。

### `vidu-q2`

请求接口：
- `POST /v1/images/generations`

结果：
- HTTP 状态码：`404`
- 返回：

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

结论：
- 模型已展示在模型列表中。
- 官方能力包含图片生成，但当前站点的该图片调用未打通。

## 4. 视频模型测试

### `vidu-q1`

#### 1080p

- 提交结果：成功
- `task_id`：`task_ejoGkhb7TSxKn8rJdaX8VgKa7dITzF7s`
- 轮询结果：`SUCCESS`
- OpenAI 兼容结果：成功
- 输出文件：`tmp\vidu_q1q2_video_20260609\vidu-q1_1080p.mp4`
- 实际分辨率：`1920x1080`
- 帧率：`24fps`
- 时长：`5.208333s`
- 上游模型：`viduq1`

结论：
- `vidu-q1` 视频链路正常。
- `1080p / 5s` 已实测通过。

### `vidu-q2`

#### 720p

- 提交结果：成功
- `task_id`：`task_us9bgPkYWeiJv3EUZS0E030MextBtBLZ`
- 轮询结果：`SUCCESS`
- OpenAI 兼容结果：成功
- 输出文件：`tmp\vidu_q1q2_video_20260609\vidu-q2_720p.mp4`
- 实际分辨率：`1280x720`
- 帧率：`24fps`
- 时长：`5.041667s`
- 上游模型：`viduq2`

#### 1080p

- 提交结果：成功
- `task_id`：`task_WEnxWXdwxzNSISYJh2Osnnrz5ekGuDFe`
- 轮询结果：`SUCCESS`
- OpenAI 兼容结果：成功
- 输出文件：`tmp\vidu_q1q2_video_20260609\vidu-q2_1080p.mp4`
- 实际分辨率：`1920x1080`
- 帧率：`24fps`
- 时长：`5.041667s`
- 上游模型：`viduq2`

结论：
- `vidu-q2` 视频链路正常。
- `720p`、`1080p` 均已实测通过。

### `vidu-q3-turbo`

#### 720p

- 提交结果：成功
- `task_id`：`task_14tQTiRlZWA28r0EhwyRjpOIMm1KOTj0`
- 轮询结果：`SUCCESS`
- OpenAI 兼容结果：成功
- 输出文件：`tmp\vidu_correct_matrix_20260609\vidu-q3-turbo_720p.mp4`
- 实际分辨率：`1280x720`
- 帧率：`24fps`
- 时长：`5.041667s`
- 上游模型：`viduq3-turbo`

#### 1080p

- 提交结果：成功
- `task_id`：`task_O7Hu1nI47gJ3DviLo85vEtvKZpkgdSpZ`
- 轮询结果：`SUCCESS`
- OpenAI 兼容结果：成功
- 输出文件：`tmp\vidu_correct_matrix_20260609\vidu-q3-turbo_1080p.mp4`
- 实际分辨率：`1920x1080`
- 帧率：`24fps`
- 时长：`5.041667s`
- 上游模型：`viduq3-turbo`

结论：
- `vidu-q3-turbo` 文生视频链路正常。
- `720p`、`1080p` 两档都已实测通过。

### `viduq-3`

#### 720p

- 提交结果：成功
- `task_id`：`task_rGPoTBQYwheCVzOwXDjLDEEveLqfGLxP`
- 轮询结果：`SUCCESS`
- OpenAI 兼容结果：成功
- 输出文件：`tmp\vidu_correct_matrix_20260609\viduq-3_720p.mp4`
- 实际分辨率：`1280x720`
- 帧率：`24fps`
- 时长：`5.041667s`
- 上游模型：`viduq3-pro`

#### 1080p

- 提交结果：成功
- `task_id`：`task_igvDJ9LfAJfOtTcuFBGfuZQ3owJl6Alo`
- 轮询结果：`SUCCESS`
- OpenAI 兼容结果：成功
- 输出文件：`tmp\vidu_correct_matrix_20260609\viduq-3_1080p.mp4`
- 实际分辨率：`1920x1080`
- 帧率：`24fps`
- 时长：`5.041667s`
- 上游模型：`viduq3-pro`

结论：
- `viduq-3` 文生视频链路正常。
- `720p`、`1080p` 两档都已实测通过。

## 5. 总结

通过：
- `vidu-q1` 视频生成可用，`1080p / 5s` 已通过。
- `vidu-q2` 视频生成可用，`720p`、`1080p` 均通过。
- `vidu-q3-turbo` 视频生成可用，`720p`、`1080p` 均通过。
- `viduq-3` 视频生成可用，`720p`、`1080p` 均通过。

未通过：
- `vidu-q1` 生图当前返回 `404 bad_response_status_code`
- `vidu-q2` 生图当前返回 `404 bad_response_status_code`

建议：
- `vidu-q1`、`vidu-q2` 需要继续检查线上图片通道映射、上游路由或模型配置。
- `vidu-q1` 可按 `1080p / 5s` 作为视频对外推荐规格。
- `vidu-q2`、`vidu-q3-turbo`、`viduq-3` 可按 `720p` / `1080p` 作为视频对外推荐规格。
