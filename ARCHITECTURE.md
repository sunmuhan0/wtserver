# War Thunder Assistant - 架构与部署文档

## 系统架构

```
用户浏览器/小程序
       │
       ▼
   nginx (:443)
   ├── /api/*  ──► Go 后端 wtserver (:8080)
   └── /*      ──► 静态前端 /var/www/html/
                         │
              Go wtserver │
              ┌───────────┤
              │  V1 端点   │  ThunderSkill / 自有数据（无需 Token）
              │  V3 端点   │──────► Python 代理 ss_proxy (:8082)
              └───────────┘              │
                                         ▼
                                  Chrome (DrissionPage)
                                  非 headless + Xvfb
                                         │
                                         ▼
                                  statshark.net API
                                  (Cloudflare 保护)
```

### 组件说明

| 组件 | 端口 | 进程管理 | 作用 |
|------|------|----------|------|
| nginx | 443/80 | systemd | HTTPS 终端、静态文件、反向代理 |
| wtserver (Go) | 8080 | pm2 | REST API 后端 |
| ss_proxy (Python) | 8082 | systemd | Chrome 浏览器代理 |
| Xvfb | :99 (X11) | systemd | 虚拟显示器，供 Chrome 非 headless 运行 |

### 为什么需要浏览器代理

statshark.net 使用 Cloudflare 保护，包含 Turnstile 验证和 TLS 指纹检测：
- Go 的 `net/http` 客户端 TLS 指纹与 Chrome 不同，即使携带有效的 `cf_clearance` cookie 和 `turnstile_token`，仍会被 Cloudflare 返回 406
- 非 headless Chrome 通过 Xvfb 虚拟显示器运行，Turnstile 会自动求解
- 所有 statshark API 请求必须通过浏览器内部的 XHR/fetch 发出，才能通过 Cloudflare 验证

## 项目结构

```
wtserver/                        # Go 后端
├── cmd/server/main.go           # 入口、路由注册
├── config/config.go             # 环境变量配置
├── internal/
│   ├── handler/handler.go       # HTTP 处理器
│   └── service/
│       ├── browser.go           # 浏览器代理客户端（连接 ss_proxy）
│       ├── statshark.go         # statshark API 调用（通过浏览器代理）
│       ├── token_client.go      # 旧 token 管理（保留兼容）
│       └── turnstile.go         # capsolver 备用方案
├── scripts/
│   └── ss_proxy.py              # Python 浏览器代理服务
└── token.json                   # 运行时 token 存储

wthtml/                          # 前端 (uni-app Vue3)
├── src/
│   ├── api/
│   │   ├── index.ts             # 请求基础配置 (BASE_URL)
│   │   └── player.ts            # 玩家相关 API + 类型定义
│   └── pages/
│       └── player/player.vue    # 玩家详情页
└── dist/build/h5/               # 构建产物 → 部署到 /var/www/html/
```

## API 端点

### V1（旧版，部分不依赖 statshark）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/player-ts/:nickname | ThunderSkill 玩家数据 |
| GET | /api/v1/player-search/:nickname | 搜索玩家（自有多源） |
| GET | /api/v1/player-ss/:nickname | statshark 玩家档案（旧） |
| GET | /api/v1/player-search-ss/:nickname | statshark 搜索（旧） |
| GET | /api/v1/squadron/:name | 中队信息 |
| GET | /api/v1/globalstats | 全局统计 |
| GET | /api/v1/vehicle/:name | 载具信息 |
| GET | /api/v1/vehicles | 载具列表 |
| GET | /api/v1/vehicle-filters | 载具筛选项 |
| GET | /api/v1/news | 新闻列表 |
| GET | /api/v1/news/detail | 新闻详情 |
| POST | /api/v1/token | 手动设置 token |
| GET | /api/v1/token/status | token 状态 |

### V3（新版，全部通过浏览器代理）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v3/player-detail/:nickname | 玩家详情（pvp 数据） |
| GET | /api/v3/player-ss/:nickname | 玩家 statshark 档案 |
| GET | /api/v3/player-search-ss/:nickname | 搜索 statshark 玩家 |
| GET | /api/v3/player-leaderboard-ss/:nickname | 天梯历史 |

V3 端点**不需要前端传 token**，Go 后端自动通过浏览器代理获取和使用 token。

## 部署方式

### 系统服务

```bash
# Xvfb 虚拟显示器
systemctl status xvfb.service
systemctl restart xvfb.service

# Python 浏览器代理（依赖 Xvfb）
systemctl status ss-proxy.service
systemctl restart ss-proxy.service
```

### Go 后端

```bash
# 编译
cd /root/project/wtserver
CGO_ENABLED=0 go build -o wtserver ./cmd/server

# 部署
pm2 restart wtserver

# 查看日志
pm2 logs wtserver
```

### 前端

```bash
cd /root/project/wthtml
npx uni build
cp -r dist/build/h5/* /var/www/html/
```

### nginx

```bash
systemctl status nginx
systemctl reload nginx
```

### 完整重启流程

```bash
systemctl restart xvfb.service
sleep 2
systemctl restart ss-proxy.service
sleep 35              # 等待 Turnstile 求解
pm2 restart wtserver
systemctl reload nginx
```

## 关键依赖

- **Go 1.22.2** + Gin 框架
- **Python 3** + DrissionPage 4.1.1.4（Chrome 自动化）
- **Google Chrome 150** (`/usr/bin/google-chrome-stable`)
- **Xvfb**（虚拟显示器 :99）
- **nginx 1.24** + Let's Encrypt SSL
- **pm2**（Go 进程管理）
- 域名: `wt.0x53.cn`

## 数据流示例：查询玩家 Dark#598

1. 前端 `GET /api/v3/player-detail/Dark%23598`
2. nginx → Go wtserver `:8080`
3. Go 调用 `ensureBrowser()` → 检查 `ss_proxy` 是否就绪
4. Go `GET http://127.0.0.1:8082/api/stat/GetIdByName?Name=Dark%23598&...`
5. Python 代理通过 Chrome XHR 发出请求（携带 Turnstile Token + cf_clearance）
6. statshark 返回 `{"224501637":"Dark#598"}`
7. Go 再通过代理 `POST /api/stat/MakeStatRequestById/224501637?update=true`
8. Chrome XHR 返回完整玩家数据
9. Go 解析并返回 JSON 给前端

## 故障排查

```bash
# 检查所有服务状态
systemctl is-active xvfb.service ss-proxy.service
pm2 list | grep wtserver

# 检查浏览器代理
curl http://127.0.0.1:8082/health
# 返回 {"ready":true,"token_len":730,"cf_clearance":"yes"} 表示正常

# 手动刷新代理 session
curl http://127.0.0.1:8082/refresh

# 查看代理日志
journalctl -u ss-proxy.service -f

# 查看 Go 后端日志
pm2 logs wtserver --lines 50
```
