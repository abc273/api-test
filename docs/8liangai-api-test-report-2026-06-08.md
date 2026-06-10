# 8liangai.com 接口完整回归测试报告

- 测试日期：2026-06-08（Asia/Shanghai）
- 测试站点：`https://8liangai.com`
- 测试认证：`Authorization: Bearer sk-***`
- 测试方式：真实提交任务、轮询业务状态接口、轮询 OpenAI 视频查询接口、下载视频内容、使用 `ffprobe` 校验实际宽高、读取 `/api/log/token` 核对真实扣费。
- 原始结果：`tmp/8liangai_full_retest_20260608_094650/results.json`、`tmp/8liangai_full_retest_20260608_102358/results.json`、`tmp/8liangai_full_retest_20260608_094548/results.json`

## 1. 总结论

| 类别 | 覆盖数 | 通过 | 失败 | 实际扣费 |
| --- | ---: | ---: | ---: | ---: |
| 视频生成 / 查询 / 下载 | 17 | 13 | 4 | 72.502640 CNY |
| 图片生成 | 4 | 4 | 0 | 0.550000 CNY |
| 真人 / 虚拟人像资产接口 | 7 | 7 | 0 | 0 CNY |

生产环境仍存在 2 类 SR 问题：

- `seedance2-sr`：实际计费仍按 `seedance2` base output-tier 档位。`720p` 实扣倍率 `23.0`，不是价格页 `24.5`；`1080p` 实扣倍率 `25.5`，也不是价格页 `24.5`。输出分别只有 `864x496` / `1280x720`，未达到对外承诺的 `720p` / `1080p`。
- `seedance2.0fast-sr`：计费倍率 `23.15` 与价格页一致，但输出分别只有 `864x496` / `1280x720`，未达到对外承诺的 `720p` / `1080p`。

本地代码已针对上述根因修复并通过编译和单测；但本轮线上复测仍失败，说明生产环境尚未部署本地修复，或部署后仍有运行配置未生效。

## 2. 已修复的代码问题

| 问题 | 修复点 | 本地验证 |
| --- | --- | --- |
| SR 失败 / 无输出时返回 base 低清视频并标记成功 | SR 后处理失败时改为任务失败，不再把原视频当最终成功结果 | `service` SR 失败用例通过 |
| `seedance2-sr` 显式价格被 base `seedance2` output-tier 抢走 | 价格模式解析先检查显式模型价格 / 倍率，再回退 base 模型计费模式 | `relay/helper` 价格测试通过 |
| `seedance1.5-sr` 未纳入 SR 专用别名处理 | 增加 `seedance1.5` / `seedance1.5-sr` alias、能力回退、默认分辨率映射 | `service`、`model`、`doubao adaptor` 测试通过 |
| 创建任务时未持久化专用 SR alias 的 SR 请求标记 | `RelayTask` 创建任务后写入 `VideoSuperResolutionRequested` | 编译通过，相关 SR 轮询测试通过 |
| 豆包 payload 对普通空 `RelayInfo` 存在空指针风险 | 访问 `ChannelMeta` 前加 guard，资产校验支持无资产普通请求 | `doubao adaptor` 全包测试通过 |

本地验证命令：

```powershell
go test ./relay/channel/task/doubao -count=1
go test ./service -run "TestResolveVideoSuperResolution|TestReadVideoSuperResolution|TestInfer|TestIsVideoSuperResolution|TestNormalizeVideoSuperResolution|TestShouldAutoEnableVideoSuperResolution|TestMaybeStartVideoSuperResolution|TestPollVideoSuperResolution" -count=1
go test ./model -run "TestBuildAbilityLookupCandidates|TestGetRandomSatisfiedChannel" -count=1
go test ./relay/helper -run "TestModelPriceHelper" -count=1
go build ./...
```

## 3. 视频接口测试结果

提交 / 查询 / 下载列格式为：`提交 HTTP / 业务状态 / OpenAI 状态 / 下载 HTTP`。

| 用例 | 模型 | 请求分辨率 | 视频输入 | 提交 / 查询 / 下载 | 实际宽高 | 期望宽高 | 实扣倍率 | 实扣金额 | 结论 |
| --- | --- | --- | --- | --- | --- | --- | ---: | ---: | --- |
| `seedance15_480p_16x9` | `seedance1.5` | `480p` | `False` | `200 / SUCCESS / completed / 200` | `864x496` | `864x496` | `8.0` | `0.810208` | 通过 |
| `seedance15_720p_16x9` | `seedance1.5` | `720p` | `False` | `200 / SUCCESS / completed / 200` | `1280x720` | `1280x720` | `8.0` | `1.7424` | 通过 |
| `seedance15_1080p_16x9` | `seedance1.5` | `1080p` | `False` | `200 / SUCCESS / completed / 200` | `1920x1080` | `1920x1080` | `8.0` | `3.9204` | 通过 |
| `seedance15sr_720p_16x9` | `seedance1.5-sr` | `720p` | `False` | `200 / SUCCESS / completed / 200` | `1280x720` | `1280x720` | `10.0` | `2.178` | 通过 |
| `seedance15sr_1080p_16x9` | `seedance1.5-sr` | `1080p` | `False` | `200 / SUCCESS / completed / 200` | `1920x1080` | `1920x1080` | `10.0` | `4.9005` | 通过 |
| `seedance2_480p_no_video_16x9` | `seedance2` | `480p` | `False` | `200 / SUCCESS / completed / 200` | `864x496` | `864x496` | `23.0` | `2.329348` | 通过 |
| `seedance2_720p_no_video_16x9` | `seedance2` | `720p` | `False` | `200 / SUCCESS / completed / 200` | `1280x720` | `1280x720` | `23.0` | `5.0094` | 通过 |
| `seedance2_1080p_no_video_16x9` | `seedance2` | `1080p` | `False` | `200 / SUCCESS / completed / 200` | `1920x1080` | `1920x1080` | `25.5` | `12.496274` | 通过 |
| `seedance2sr_720p_16x9` | `seedance2-sr` | `720p` | `False` | `200 / SUCCESS / completed / 200` | `864x496` | `1280x720` | `23.0` | `2.329348` | 失败 |
| `seedance2sr_1080p_16x9` | `seedance2-sr` | `1080p` | `False` | `200 / SUCCESS / completed / 200` | `1280x720` | `1920x1080` | `25.5` | `5.5539` | 失败 |
| `sd20fast_480p_16x9` | `sd2.0fast` | `480p` | `False` | `200 / SUCCESS / completed / 200` | `864x496` | `864x496` | `18.5` | `1.873606` | 通过 |
| `sd20fast_720p_16x9` | `sd2.0fast` | `720p` | `False` | `200 / SUCCESS / completed / 200` | `1280x720` | `1280x720` | `18.5` | `4.0293` | 通过 |
| `seedance20fastsr_720p_16x9` | `seedance2.0fast-sr` | `720p` | `False` | `200 / SUCCESS / completed / 200` | `864x496` | `1280x720` | `23.15` | `2.344538` | 失败 |
| `seedance20fastsr_1080p_16x9` | `seedance2.0fast-sr` | `1080p` | `False` | `200 / SUCCESS / completed / 200` | `1280x720` | `1920x1080` | `23.15` | `5.04207` | 失败 |
| `seedance2_480p_with_video_16x9` | `seedance2` | `480p` | `True` | `200 / SUCCESS / completed / 200` | `864x496` | `864x496` | `15.5` | `3.126598` | 通过 |
| `seedance2_720p_with_video_16x9` | `seedance2` | `720p` | `True` | `200 / SUCCESS / completed / 200` | `1280x720` | `1280x720` | `15.5` | `6.7239` | 通过 |
| `seedance2_1080p_with_video_16x9` | `seedance2` | `1080p` | `True` | `200 / SUCCESS / completed / 200` | `1920x1080` | `1920x1080` | `17.0` | `8.09285` | 通过 |

说明：

- 期望宽高按字节 Seedance 常用 16:9 尺寸核对：`480p=864x496`，`720p=1280x720`，`1080p=1920x1080`。
- `seedance2` 视频输入用例必须在 `metadata.content[]` 中传 `role: "reference_video"`；缺失该字段时上游返回 400，且不扣费。

## 4. 图片接口测试结果

| 用例 | 模型 | size | HTTP | 实扣金额 | 结论 |
| --- | --- | --- | ---: | ---: | --- |
| `seedream45_valid_2048` | `seedream4.5` | `2048x2048` | 200 | `0.3` | 通过 |
| `seedream45_invalid_1024` | `seedream4.5` | `1024x1024` | 400 | `0.0` | 通过 |
| `seedream50lite_valid_2048` | `seedream5.0lite` | `2048x2048` | 200 | `0.25` | 通过 |
| `seedream50lite_invalid_1024` | `seedream5.0lite` | `1024x1024` | 400 | `0.0` | 通过 |

## 5. 资产接口测试结果

| 用例 | 方法 | 路径 | 预期业务成功 | HTTP | 结论 |
| --- | --- | --- | --- | ---: | --- |
| `official_config` | `GET` | `/api/portrait_assets/official/config` | `True` | 200 | 通过 |
| `official_jobs` | `GET` | `/api/portrait_assets/official/jobs?p=1&page_size=5` | `True` | 200 | 通过 |
| `official_create_missing_name` | `POST` | `/api/portrait_assets/official/jobs` | `False` | 200 | 通过 |
| `virtual_config` | `GET` | `/api/portrait_assets/virtual/config` | `True` | 200 | 通过 |
| `virtual_group` | `GET` | `/api/portrait_assets/virtual/group` | `True` | 200 | 通过 |
| `virtual_assets` | `GET` | `/api/portrait_assets/virtual/assets?p=1&page_size=5` | `True` | 200 | 通过 |
| `virtual_create_invalid` | `POST` | `/api/portrait_assets/virtual/assets` | `False` | 200 | 通过 |

## 6. 失败详情与处理建议

### `seedance2sr_720p_16x9`

- 模型：`seedance2-sr`
- 请求分辨率：`720p`
- 期望实际宽高：`1280x720`
- 本次实际宽高：`864x496`
- 实扣倍率：`23.0`
- 实扣金额：`2.329348 CNY`
- 判断：生产仍走 base `seedance2` 计费 / 输出路径，未使用本地修复后的显式 SR alias 计费优先级和 SR 失败阻断逻辑。

### `seedance2sr_1080p_16x9`

- 模型：`seedance2-sr`
- 请求分辨率：`1080p`
- 期望实际宽高：`1920x1080`
- 本次实际宽高：`1280x720`
- 实扣倍率：`25.5`
- 实扣金额：`5.5539 CNY`
- 判断：生产仍走 base `seedance2` 计费 / 输出路径，未使用本地修复后的显式 SR alias 计费优先级和 SR 失败阻断逻辑。

### `seedance20fastsr_720p_16x9`

- 模型：`seedance2.0fast-sr`
- 请求分辨率：`720p`
- 期望实际宽高：`1280x720`
- 本次实际宽高：`864x496`
- 实扣倍率：`23.15`
- 实扣金额：`2.344538 CNY`
- 判断：生产能按 `23.15` 计费，但 SR 后处理没有产出目标清晰度，仍返回 base 分辨率视频。

### `seedance20fastsr_1080p_16x9`

- 模型：`seedance2.0fast-sr`
- 请求分辨率：`1080p`
- 期望实际宽高：`1920x1080`
- 本次实际宽高：`1280x720`
- 实扣倍率：`23.15`
- 实扣金额：`5.04207 CNY`
- 判断：生产能按 `23.15` 计费，但 SR 后处理没有产出目标清晰度，仍返回 base 分辨率视频。

建议上线流程：

1. 部署本地修复后的后端代码。
2. 确认生产模型价格配置仍包含 `seedance2-sr=24.5`、`seedance2.0fast-sr=23.15`、`seedance1.5-sr=10`。
3. 部署后优先重跑 4 条失败 SR 用例：`seedance2-sr 720p/1080p`、`seedance2.0fast-sr 720p/1080p`。
4. 通过后再跑完整主矩阵，防止未来新增模型再次出现 alias、价格、输出分辨率颗粒度不一致。
