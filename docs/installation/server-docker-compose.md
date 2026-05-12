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
- `temp_images/`

> [!TIP]
> 如果服务器上的 `Dockerfile` 做过针对部署环境的定制，例如加过 `GOPROXY`、`GOSUMDB`、额外系统依赖或镜像源优化，升级代码时不要直接覆盖掉它，除非这些改动已经合并回当前仓库。

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
docker compose --progress plain build new-api
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
tar czf "$backup_dir/app-files.tar.gz" docker-compose.yml Dockerfile .env data logs temp_images
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
- `Dockerfile`（如果服务器上做过部署优化）
- `.env`
- `data/`
- `logs/`
- `temp_images/`

如果你是整目录覆盖，先把这些文件单独拷走，再回填回来。

### 3. 仅重建 `new-api`

```bash
cd /www/wwwroot/new_api
docker compose --progress plain build new-api
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

### 5. 推荐的同步方式

如果你使用 `rsync` 同步本地代码，建议明确排除运行数据和本地缓存，不要整仓裸传。推荐命令类似这样：

```bash
rsync -az --delete \
  --exclude .git/ \
  --exclude .gocache/ \
  --exclude .gomodcache/ \
  --exclude .run/ \
  --exclude .tools/ \
  --exclude data/ \
  --exclude logs/ \
  --exclude temp_images/ \
  --exclude .env \
  --exclude docker-compose.yml \
  --exclude Dockerfile \
  --exclude web/default/.codex-local/ \
  --exclude web/default/node_modules/ \
  --exclude web/classic/node_modules/ \
  /本地/new-api/ root@你的服务器:/www/wwwroot/new_api/
```

这样做有两个好处：

1. 不会把线上配置、数据库文件、日志和缓存误覆盖
2. 不会把本地开发缓存一并同步到服务器，避免构建上下文变得非常大

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

## 四、构建上下文瘦身

如果服务器目录里混进了本地缓存、运行文件或调试产物，`docker build` 可能看起来像“卡住”，本质上是在搬运一个过大的构建上下文。

推荐至少确认 `.dockerignore` 包含这些内容：

```text
.git
.github
*.md
docs
.gocache
.gomodcache
.run
.tools
web/default/node_modules
web/classic/node_modules
web/default/.codex-local
data
logs
temp_images
one-api.db
new-api
.env
```

构建前建议先看一眼项目目录体积：

```bash
cd /www/wwwroot/new_api
du -sh .
```

如果目录体积已经达到数 GB，先清理本地缓存和误同步上去的运行目录，再开始构建。

## 五、国内网络环境构建建议

如果构建卡在 `go mod download`，可以在 `Dockerfile` 的 Go 构建阶段加上：

```dockerfile
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn
```

加完后重新执行：

```bash
docker compose build --no-cache new-api
```

## 六、后台构建推荐方式

如果你通过 SSH 远程发布，推荐把构建放到后台并写入日志文件，不要一直挂在交互式终端里等。

推荐命令：

```bash
cd /www/wwwroot/new_api
ts=$(date +%Y%m%d_%H%M%S)
log=/tmp/new_api_deploy_$ts.log
status=/tmp/new_api_deploy_$ts.status

nohup bash -lc '
  cd /www/wwwroot/new_api &&
  docker compose --progress plain build new-api &&
  docker compose up -d --no-deps new-api
  rc=$?
  echo $rc > "$1"
  exit $rc
' _ "$status" > "$log" 2>&1 < /dev/null &

echo "log=$log"
echo "status=$status"
```

查看进度：

```bash
tail -n 200 "$log"
```

查看是否完成：

```bash
cat "$status"
```

返回 `0` 表示构建和替换成功。

> [!TIP]
> 当前仓库构建时间可能明显长于普通 Go 项目，因为它还会构建两个前端主题，其中 `web/classic` 的 `vite build` 很可能需要几十分钟。这种情况不一定是卡死，先结合 CPU、内存和日志一起判断。

## 七、域名与 443 推荐方式

生产环境建议保留应用监听 `3000`，再由 Nginx 负责 `80/443` 和 HTTPS。

### 1. DNS

先把域名解析到服务器公网 IP，例如：

- `example.com -> 116.62.175.161`
- `www.example.com -> 116.62.175.161`

### 2. Nginx 反向代理

推荐配置结构：

```nginx
server {
    listen 80 default_server;
    server_name _;
    return 404;
}

server {
    listen 80;
    server_name example.com www.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name example.com www.example.com;

    ssl_certificate /etc/nginx/ssl/example.com.pem;
    ssl_certificate_key /etc/nginx/ssl/example.com.key;

    client_max_body_size 200M;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 3600;
        proxy_send_timeout 3600;
        proxy_buffering off;
    }
}
```

检查并重载：

```bash
nginx -t
systemctl reload nginx
```

### 3. 不要用端口转发替代 HTTPS

不要用 `iptables`、安全软件或面板端口转发把：

- `443 -> 3000`
- 或 `3000 -> 443`

这种“改口”规则加在公网入口上。

这样做的常见后果是：

- 外部访问 `443` 时拿到明文 HTTP
- TLS 握手失败，浏览器报 `ERR_SSL_PROTOCOL_ERROR`
- 域名证书明明存在，但访问的其实不是 Nginx 的 HTTPS 入口

HTTPS 应该由 Nginx 终止，应用只监听内部 `3000`。

## 八、升级后必须联动检查的配置

如果服务从 IP 切到域名，或部署地址有变化，升级后要检查这些配置是否仍然指向旧地址：

- `ServerAddress`
- `portrait_asset.callback_base_url`
- `Passkey` 相关 origin / rp id

建议直接访问：

```bash
curl https://你的域名/api/status
```

重点确认返回值里的：

- `server_address`
- `passkey_origins`
- `passkey_rp_id`

如果这里还是旧 IP，说明运行配置没有更新干净。真人资产回调、支付回跳、OAuth、Passkey 都可能继续走旧地址。

## 九、常见问题

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

### 6. 构建看起来像卡住了

先不要立刻中断，先检查这几项：

```bash
ps -eo pid,ppid,%cpu,%mem,etime,cmd --sort=-%cpu | head -n 20
tail -n 200 /tmp/new_api_deploy_xxx.log
free -h
```

如果你看到类似：

- `node /build/node_modules/.bin/vite build`
- CPU 仍在持续占用
- 日志仍在推进

通常说明前端还在构建，不是死锁。

### 7. 域名 443 打开后是默认 Nginx 页面

通常有两个原因：

1. `server_name` 没匹配到你的域名
2. 默认站点先吃掉了请求，而你的业务站点没正确启用

先检查：

```bash
nginx -T | grep -nE 'server_name|listen 443|proxy_pass'
```

### 8. 域名 443 出现 SSL 协议错误

优先排查是否存在错误的 NAT / 端口转发规则：

```bash
iptables -t nat -S
```

如果看到类似：

```text
-A PREROUTING -p tcp --dport 443 -j REDIRECT --to-ports 3000
```

就说明 HTTPS 被错误地改口了。

## 十、推荐的升级顺序

```bash
cd /www/wwwroot/new_api
ts=$(date +%Y%m%d_%H%M%S)
backup_dir=/www/backup/new_api/$ts
mkdir -p "$backup_dir"
tar czf "$backup_dir/app-files.tar.gz" docker-compose.yml Dockerfile .env data logs temp_images
docker compose exec -T postgres pg_dump -U root new-api > "$backup_dir/postgres.sql"
rsync -az --delete \
  --exclude .git/ \
  --exclude .gocache/ \
  --exclude .gomodcache/ \
  --exclude .run/ \
  --exclude .tools/ \
  --exclude data/ \
  --exclude logs/ \
  --exclude temp_images/ \
  --exclude .env \
  --exclude docker-compose.yml \
  --exclude Dockerfile \
  --exclude web/default/.codex-local/ \
  --exclude web/default/node_modules/ \
  --exclude web/classic/node_modules/ \
  /本地/new-api/ root@你的服务器:/www/wwwroot/new_api/
docker compose --progress plain build new-api
docker compose up -d --no-deps new-api
curl http://127.0.0.1:3000/api/status
curl https://你的域名/api/status
```

这套流程的重点是：

1. 先备份
2. 再用排除规则同步代码
3. 只重建 `new-api`
4. 最后同时验证 `127.0.0.1:3000` 和域名 `443`

整个过程中不去动数据库卷，也不依赖公网 `443 -> 3000` 端口转发。
