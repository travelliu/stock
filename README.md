# Stock — A股日线价差分析

基于 Go + Vue 的 A 股日线数据管理与价差分析平台，包含 Web 界面、REST API 服务端（`stockd`）和命令行客户端（`stockctl`）。

## 功能

- **持仓管理**：添加/删除关注股票，Web 界面实时显示最新行情
- **日线数据同步**：通过 Tushare API 拉取 OHLCV 数据，支持定时增量更新
- **价差分析**：6 种价差（高-开、开-低、高-低、开-收、高-收、低-收）多时间窗口统计
- **交易参考**：基于历史价差均值生成当日最高/最低/收盘参考价预测表
- **预测记录**：保存每日预测并与实际值对比，持续追踪预测准确度
- **用户管理**：多用户支持，管理员可管理用户和数据同步任务

## 技术栈

- **后端**：Go 1.25，Gin，GORM，MySQL / SQLite
- **前端**：Vue 3，Element Plus，Vite（打包后嵌入 Go binary）
- **CLI**：Cobra（`stockctl`）
- **部署**：Docker + Docker Compose

## 快速开始

### 本地开发

```bash
# 1. 复制环境变量
cp .env.example .env

# 2. 启动 MySQL
make dev-up

# 3. 构建前端（嵌入 Go binary）
make web-build

# 4. 运行服务端
go run ./cmd/stockd

# 访问 http://localhost:8443
```

### Docker 部署（从源码构建）

```bash
cp .env.example .env
# 编辑 .env，填入必要的密钥（见下方配置说明）

make selfhost-build
# 构建并启动后访问 http://localhost:8443
```

停止：

```bash
make selfhost-stop
```

## 配置

服务端配置文件默认路径为 `/etc/stockd/config.yaml`，可通过环境变量 `STOCKD_CONFIG` 覆盖路径。

```yaml
server:
  listen: ":8443"
  session_secret: "至少32字节的随机字符串"

database:
  driver: mysql          # sqlite | mysql
  dsn: "user:pass@tcp(localhost:3306)/stockd?parseTime=true"

tushare:
  default_token: ""      # Tushare token，数据同步必填

scheduler:
  enabled: true
  daily_fetch_cron: "0 22 * * 1-5"   # 每个交易日 22:00 同步当日数据
```

所有配置项均可通过 `STOCKD_` 前缀的环境变量覆盖，例如：

```bash
STOCKD_DATABASE_DSN="user:pass@tcp(mysql:3306)/stockd?parseTime=true"
STOCKD_SERVER_SESSION_SECRET="your-secret"
STOCKD_TUSHARE_DEFAULT_TOKEN="your-token"
```

## CLI 使用

`stockctl` 通过 REST API 与 `stockd` 通信，需先配置服务地址和 API Token。

```bash
# 搜索股票
stockctl stock search 茅台

# 查看价差分析（传入开盘价生成参考价预测）
stockctl stock analysis 600519.SH --actual-open 1800.00

# 触发行情同步
stockctl stock fetch
```

## Makefile

```
make build          # 构建 stockd 和 stockctl
make web-build      # 构建前端（输出到 web/dist）
make test           # 运行 Go 测试
make lint           # go vet + staticcheck

make dev-up         # 启动本地 MySQL（开发用）
make dev-down       # 停止本地 MySQL

make selfhost-build # 构建 web + Docker 镜像并启动
make selfhost-stop  # 停止生产栈
make selfhost-logs  # 查看 stockd 日志
```

## 项目结构

```
stock/
├── cmd/
│   ├── stockd/        # 服务端入口
│   └── stockctl/      # CLI 入口
├── pkg/
│   ├── cli/           # CLI 命令与渲染
│   ├── models/        # 数据模型
│   ├── stockd/        # HTTP 路由、服务、配置、数据库
│   └── tushare/       # Tushare API 客户端
├── web/               # Vue 3 前端（构建后嵌入 binary）
├── docker/
├── embed.go           # web/dist 静态文件嵌入
├── Dockerfile
├── docker-compose.yml        # 本地 dev（MySQL only）
├── docker-compose.self.yml   # 生产构建 override
└── config.example.yaml
```

## 价差说明

所有价差均取绝对值存储，均值计算方式为 `(算术平均 + 中位数) / 2`。

| 价差  | 计算        | 用途                       |
|-------|-------------|----------------------------|
| 高-开 | 最高 − 开盘 | 高抛目标：开盘 + 此值附近卖 |
| 开-低 | 开盘 − 最低 | 低吸目标：开盘 − 此值附近买 |
| 高-低 | 最高 − 最低 | 全日振幅，衡量做T空间       |
| 开-收 | 开盘 − 收盘 | 日内偏离幅度               |
| 高-收 | 最高 − 收盘 | 高点到收盘回落空间         |
| 低-收 | 收盘 − 最低 | 低点到收盘反弹空间         |

## License

MIT
