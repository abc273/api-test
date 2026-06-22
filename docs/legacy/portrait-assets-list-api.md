# 画像资产查询接口文档

## 结论

当前代码库中**没有**一个“同时返回用户全部真人资产 + 虚拟资产”的统一接口。

现有可用于查询用户已上传资产的接口是两套分页列表接口：

1. 官方真人资产列表：`GET /api/portrait_assets/official/jobs`
2. 虚拟资产列表：`GET /api/portrait_assets/virtual/assets`

如果业务侧希望前端一次性拿到“全部资产”，当前需要分别调用这两个接口后在前端聚合。

## 通用说明

- 认证方式：`middleware.UserAuth()`，需要用户登录态
- 返回包装格式：

```json
{
  "success": true,
  "message": "",
  "data": {}
}
```

- 分页参数：
  - `p`: 页码，从 `1` 开始
  - `page_size`: 每页数量，最大 `100`
- 兼容参数：
  - 页码兼容 `p`
  - 每页数量兼容 `ps`、`size`

分页返回结构：

```json
{
  "page": 1,
  "page_size": 10,
  "total": 2,
  "items": []
}
```

---

## 1. 官方真人资产列表

### 接口地址

`GET /api/portrait_assets/official/jobs`

### 代码位置

- 路由定义：`router/api-router.go`
- 控制器：`controller/portrait_asset.go` 中 `ListOfficialPortraitAssetJobs`
- 查询模型：`model/portrait_asset.go` 中 `GetUserOfficialPortraitAssetJobs`

### 用途

查询当前用户创建过的官方真人资产任务列表。  
该列表中包含资产上传、认证、入库、确认等全过程状态，以及资产相关字段。

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `p` | int | 否 | 页码，默认 `1` |
| `page_size` | int | 否 | 每页数量，默认系统分页大小，最大 `100` |

### 返回字段

`data.items[]` 中每项字段如下：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int | 任务 ID |
| `user_id` | int | 用户 ID |
| `name` | string | 任务/资产名称 |
| `source` | string | 固定为 `official` |
| `status` | string | 任务状态 |
| `invite_url` | string | 实人认证链接，部分状态下有值 |
| `validate_result_code` | string | 实人认证结果码 |
| `volc_group_id` | string | 火山资产组 ID |
| `asset_id` | string | 资产 ID |
| `asset_status` | string | 火山侧资产状态 |
| `asset_preview` | string | 资产预览地址 |
| `asset_url` | string | 用户提交的原始素材 URL |
| `asset_type` | string | 资产类型，通常为 `Image` / `Video` / `Audio` |
| `project_name` | string | 项目名 |
| `error_message` | string | 错误信息 |
| `created_time` | int64 | 创建时间，Unix 时间戳（秒） |
| `updated_time` | int64 | 更新时间，Unix 时间戳（秒） |
| `qr_expires_time` | int64 | 认证链接过期时间，Unix 时间戳（秒） |
| `ready_time` | int64 | 资产就绪时间，Unix 时间戳（秒） |

### `status` 枚举

| 值 | 说明 |
| --- | --- |
| `pending` | 任务已创建，待处理 |
| `validate_ready` | 实人认证链接已生成 |
| `validated` | 实人认证已通过 |
| `asset_processing` | 资产入库处理中 |
| `pending_confirm` | 资产已入库，待用户确认 |
| `ready` | 资产已确认可用 |
| `failed` | 失败 |
| `expired` | 认证链接已过期 |
| `disabled` | 已禁用 |

### 返回示例

```json
{
  "success": true,
  "message": "",
  "data": {
    "page": 1,
    "page_size": 10,
    "total": 1,
    "items": [
      {
        "id": 12,
        "user_id": 1001,
        "name": "主播真人资产",
        "source": "official",
        "status": "ready",
        "validate_result_code": "10000",
        "volc_group_id": "group_xxx",
        "asset_id": "asset_xxx",
        "asset_status": "Active",
        "asset_preview": "https://example.com/api/portrait_assets/official/jobs/12/preview/xxxx",
        "asset_url": "https://example.com/uploads/20260523/demo.jpg",
        "asset_type": "Image",
        "project_name": "portrait-project",
        "error_message": "",
        "created_time": 1747980000,
        "updated_time": 1747980300,
        "qr_expires_time": 1747980120,
        "ready_time": 1747980300
      }
    ]
  }
}
```

---

## 2. 虚拟资产列表

### 接口地址

`GET /api/portrait_assets/virtual/assets`

### 代码位置

- 路由定义：`router/api-router.go`
- 控制器：`controller/virtual_portrait_asset.go` 中 `ListUserVirtualPortraitAssets`
- 查询模型：`model/virtual_portrait_asset.go` 中 `GetUserVirtualPortraitAssets`

### 用途

查询当前用户创建过的虚拟资产列表。  
该列表直接对应用户的虚拟素材记录，可用于“查看我上传过哪些虚拟资产”。

### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `p` | int | 否 | 页码，默认 `1` |
| `page_size` | int | 否 | 每页数量，默认系统分页大小，最大 `100` |

### 返回字段

`data.items[]` 中每项字段如下：

| 字段名 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int | 资产记录 ID |
| `user_id` | int | 用户 ID |
| `group_id` | int | 资产组记录 ID |
| `name` | string | 资产名称 |
| `asset_type` | string | 资产类型，`Image` / `Video` / `Audio` |
| `source_url` | string | 用户提交的原始素材 URL |
| `preview_url` | string | 预览地址 |
| `project_name` | string | 项目名 |
| `volc_group_id` | string | 火山资产组 ID |
| `volc_asset_id` | string | 火山资产 ID |
| `status` | string | 资产状态 |
| `volc_status` | string | 火山侧状态 |
| `error_message` | string | 错误信息 |
| `created_time` | int64 | 创建时间，Unix 时间戳（秒） |
| `updated_time` | int64 | 更新时间，Unix 时间戳（秒） |
| `ready_time` | int64 | 就绪时间，Unix 时间戳（秒） |

### `status` 枚举

| 值 | 说明 |
| --- | --- |
| `processing` | 处理中 |
| `active` | 已可用 |
| `failed` | 失败 |

### 返回示例

```json
{
  "success": true,
  "message": "",
  "data": {
    "page": 1,
    "page_size": 10,
    "total": 1,
    "items": [
      {
        "id": 25,
        "user_id": 1001,
        "group_id": 8,
        "name": "数字人声音素材",
        "asset_type": "Audio",
        "source_url": "https://example.com/uploads/20260523/demo.m4a",
        "preview_url": "https://example.com/api/portrait_assets/virtual/assets/25/preview/xxxx",
        "project_name": "portrait-project",
        "volc_group_id": "group_xxx",
        "volc_asset_id": "asset_xxx",
        "status": "active",
        "volc_status": "Active",
        "error_message": "",
        "created_time": 1747981000,
        "updated_time": 1747981200,
        "ready_time": 1747981200
      }
    ]
  }
}
```

---

## 是否满足“查询用户上传的所有资产”

### 当前能力

部分满足。

- 查官方真人资产：有
- 查虚拟资产：有
- 一次性查询“真人资产 + 虚拟资产全部列表”：没有

### 当前推荐调用方式

前端分别调用：

1. `GET /api/portrait_assets/official/jobs`
2. `GET /api/portrait_assets/virtual/assets`

然后按业务需要在前端合并展示。

### 如果要做成统一接口

建议新增例如：

`GET /api/portrait_assets/all`

返回结构可按以下两种方式设计：

1. 分组返回

```json
{
  "official": [],
  "virtual": []
}
```

2. 扁平返回并带类型字段

```json
{
  "items": [
    {
      "asset_kind": "official",
      "id": 12
    },
    {
      "asset_kind": "virtual",
      "id": 25
    }
  ]
}
```

