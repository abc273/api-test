# 服务器 Docker Compose 部署与升级

本文档适用于把当前仓库部署到 Linux 服务器，或在保留用户、令牌、渠道、系统配置等数据的前提下，将旧版 `new-api` 替换为当前代码。

## 适用场景

- 首次在 Ubuntu / Debian / CentOS 服务器部署 `new-api`
- 使用 Docker Compose 管理 `new-api + PostgreSQL + Redis`
- 需要保留现有用户、渠道、令牌和系统配置
- 需要从官方镜像切换到“以当前仓库代码为准”的部署方式

## 前置要求

| 项目 | 建议 |
| --- | --- |
| 操作系统 | Ubuntu 22.04+ / Debian 12+ / CentOS 7+ |
| Docker | 24+ |
| Docker Compose | `docker compose` 插件可用 |
| CPU / 内存 | 最低 2C4G，生产环境建议 4C8G 起 |
| 端口 | 放行 `3000`，如需域名访问再放行 `80/443` |

> [!WARNING]
> 当前仓库包含前端构建步骤。小内存服务器在 `docker build` 阶段容易卡住，建议提前准备 `4G-8G` swap。

## 目录建议

推荐将项目放在固定目录，例如：

```text
/www/wwwroot/new_api
```

常见保留目录如下：

- `docker-compose.yml`
- `Dockerfile`
- `.env`（如果你自己维护环境变量文件）
- `data/`
- `logs/`

## 一、首次部署当前仓库代码

### 1. 上传代码到服务器

将当前仓库完整上传到服务器，例如：

```bash
mkdir -p /www/wwwroot/new_api
```

然后使用 `scp`、`rsync`、Git 或面板上传文件到该目录。

### 2. 修改 `docker-compose.yml`

仓库根目录自带的 `docker-compose.yml` 默认使用官方镜像：

```yaml
image: calciumion/new-api:latest
```

如果你要部署“当前仓库代码”，请把 `new-api` 服务改成下面这种形式：

```yaml
services:
  new-api:
    build:
      context: .
      dockerfile: Dockerfile
    image: new-api-local:latest
```

同时至少确认以下配置已经改成生产值：

- `SQL_DSN`
- `REDIS_CONN_STRING`
- `POSTGRES_PASSWORD` 或 `MYSQL_ROOT_PASSWORD`
- `SESSION_SECRET`
- `CRYPTO_SECRET`（使用 Redis 或多节点时建议设置）

可以使用下面的命令生成随机密钥：

```bash
openssl rand -hex 32
```

### 3. 构建并启动

```bash
cd /www/wwwroot/new_api
docker compose build new-api
docker compose up -d
docker compose ps
```

### 4. 验证服务

```bash
curl http://127.0.0.1:3000/api/status
docker compose logs --tail=200 new-api
```

返回结果中包含 `success: true`，说明服务已经正常启动。

## 二、保留数据替换旧版本

这一节适合“服务器上已经有一个旧版 `new-api`，现在要切到当前仓库代码”的场景。

### 原则

只替换应用代码和 `new-api` 容器，不删除数据库卷，不删除 Redis，不执行 `docker compose down -v`。

以下内容通常需要保留：

- PostgreSQL 命名卷 `pg_data`
- MySQL 命名卷 `mysql_data`
- SQLite 使用的 `./data`
- 当前正在使用的 `SESSION_SECRET`
- 当前正在使用的 `CRYPTO_SECRET`

### 1. 先做完整备份

```bash
cd /www/wwwroot/new_api
ts=$(date +%Y%m%d_%H%M%S)
backup_dir=/www/backup/new_api/$ts
mkdir -p "$backup_dir"
tar czf "$backup_dir/app-files.tar.gz" docker-compose.yml Dockerfile data logs
```

如果你使用 PostgreSQL，再额外导出一份逻辑备份：

```bash
docker compose exec -T postgres pg_dump -U root new-api > "$backup_dir/postgres.sql"
```

如果你使用 MySQL：

```bash
docker compose exec -T mysql mysqldump -uroot -p'你的密码' new-api > "$backup_dir/mysql.sql"
```

如果你使用 SQLite，额外备份 `data/` 即可。

### 2. 上传新代码，但保留运行配置

替换代码时，保留旧环境里的以下文件和目录：

- `docker-compose.yml`
- `.env`
- `data/`
- `logs/`

如果你是整目录覆盖，先把这些文件单独拷走，再回填回来。

### 3. 仅重建 `new-api`

```bash
cd /www/wwwroot/new_api
docker compose build new-api
docker compose up -d --no-deps new-api
docker compose ps
```

`--no-deps` 的作用是只更新应用容器，不去重建数据库和 Redis。

### 4. 验证数据是否保留

重点检查：

- 管理员是否还能登录
- 用户列表是否还在
- 令牌是否还在
- 渠道和模型映射是否还在
- `设置` 页面里的系统配置是否保留

如果服务异常，先看日志：

```bash
docker compose logs --tail=200 new-api
```

## 三、小内存服务器建议

如果服务器在构建阶段卡死、SSH 无响应或 `docker build` 长时间不结束，通常是内存不够。

建议在构建前加 swap：

```bash
fallocate -l 8G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
echo '/swapfile none swap sw 0 0' >> /etc/fstab
free -h
```

如果机器本身只有 `2G-4G` 内存，这一步基本是必做项。

## 四、国内网络环境构建建议

如果构建卡在 `go mod download`，可以在 `Dockerfile` 的 Go 构建阶段加上：

```dockerfile
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn
```

加完后重新执行：

```bash
docker compose build --no-cache new-api
```

## 五、常见问题

### 1. 明明已经换了新代码，页面看起来还是老版本

先检查下面几项：

1. 浏览器是否用了旧缓存，先强制刷新一次
2. 系统配置里的 `theme.frontend` 是否为 `default`
3. 是否真的重新构建并重启了 `new-api` 容器

建议执行：

```bash
docker compose build new-api
docker compose up -d --no-deps new-api
```

### 2. 登录状态异常或重启后会话失效

确认 `SESSION_SECRET` 已设置，且升级前后保持一致。

### 3. 使用 Redis 后出现解密问题

确认 `CRYPTO_SECRET` 已设置，且升级前后保持一致。

### 4. 升级后数据丢了

最常见原因有两个：

1. 执行了 `docker compose down -v`
2. 覆盖代码时把 `data/`、数据库卷或旧配置一起删掉了

因此升级时只替换应用代码，不要删除卷。

### 5. 3000 端口被占用

如果旧版 `new-api` 还在运行，属于正常现象。直接执行：

```bash
docker compose up -d --no-deps new-api
```

Compose 会自动替换同名服务对应的旧容器。

## 六、推荐的升级顺序

```bash
cd /www/wwwroot/new_api
ts=$(date +%Y%m%d_%H%M%S)
backup_dir=/www/backup/new_api/$ts
mkdir -p "$backup_dir"
tar czf "$backup_dir/app-files.tar.gz" docker-compose.yml Dockerfile data logs
docker compose exec -T postgres pg_dump -U root new-api > "$backup_dir/postgres.sql"
docker compose build new-api
docker compose up -d --no-deps new-api
curl http://127.0.0.1:3000/api/status
```

这套流程的重点是：先备份，再重建应用，最后验证，不去动数据库卷。
