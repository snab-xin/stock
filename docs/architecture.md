# 架构说明

## 目标

个人使用的 A 股日线系统，优先把数据和约定做对，而不是先上微服务。

第一版是 **一个 Go 命令行 + 一个 PostgreSQL**。采集、回测、实盘以后可以仍放在同一个模块里，等真有独立扩缩需求再拆进程。

## 数据流

```
Tushare Pro
        │
        ▼
  tushare.Source  →  ingest.MarketSource
        │
        ▼
  ingest.Service（记 ingest_run、按交易日灌数）
        │
        ▼
  repo（GORM 幂等 upsert）
        │
        ▼
  PostgreSQL
   ├─ 原始层: instrument / daily_bar / adj_factor / suspension / trade_calendar
   └─ 研究层: daily_bar_qfq（视图，可重建）
        │
        ├── 回测引擎（单票双均线，T+1 / 费用 / 涨跌停）
        └── 策略作业 → 目标仓位 → 执行/风控（未实现）
```

## 模块职责

| 包 | 职责 | 不要放什么 |
|---|---|---|
| `internal/config` | 读 `DATABASE_URL` / `.env` | 业务逻辑 |
| `internal/db` | GORM 连接、SQL 文件迁移 | 业务查询、AutoMigrate |
| `internal/domain` | 持久化实体（含 `id` / 时间戳） | 驱动细节 |
| `internal/repo` | GORM 映射与幂等 upsert | HTTP、数据源 SDK、手写长 SQL |
| `internal/ingest` | 采集编排、代码解析 | 表结构、HTTP |
| `internal/tushare` | Tushare HTTP 与单位换算 | 落库 |
| `internal/research` | 收益/回撤等只读统计 | 下单 |
| `internal/backtest` | 单票撮合：信号与成交分开 | 组合、实盘 |
| `internal/applog` | slog 初始化、GORM 适配 | 业务逻辑 |
| `internal/httpapi` | Gin 看板 API；路由分组、JSON、CORS、请求日志 | 下单 |
| `config.yaml` | 非密钥配置（含 tushare.rpm） | Token / 数据库密码 |
| `web/` | Vite React 数据看板（shadcn-admin） | 交易终端 |
| `cmd/ashare` | 子命令 | 复杂业务 |

## 为什么用 GORM（以及不用哪些功能）

引入 GORM 只为这三件事：结构体字段映射、少写 INSERT 样板、批量 `ON CONFLICT` upsert。

不用这些：

- **AutoMigrate**：表结构只来自 `migrations/*.sql`。字段中文备注用 PostgreSQL `COMMENT ON COLUMN`，不靠 GORM tag 建表。
- **Association / Preload**：行情是事实表，不做对象图
- **把研究 SQL 改成链式 API**：复权、截面、回测扫描继续写 SQL 或视图

列名默认用 GORM 的 snake_case（`TSCode` → `ts_code`）。对不上的才打 tag，例如 `AdjFactor.Factor` → `adj_factor`。

灌数走 `CreateInBatches`；以后如果全市场日线插入成为瓶颈，再对 `daily_bar` 单独上 `COPY`，不必换掉 GORM。

## 为什么是 PostgreSQL

A 股日线数据量不大，MySQL 也能存。选 PostgreSQL 是为了：

- 窗口函数、视图、部分索引写研究 SQL 更顺
- 以后分钟线可以上 TimescaleDB，仍是同一套 SQL
- `NUMERIC` 存价格，避免过早把精度问题藏进应用层

应用层目前用 `float64` 接价格，方便研究迭代；库内仍是 `NUMERIC`。若以后对账或净值计算出现误差，再改为 `shopspring/decimal`。

## 回测边界

日线回测按已落库的交易日推进，细节见 [backtest.md](backtest.md)：

1. 信号只用 T 日及更早的前复权收盘，禁止读到未来 K 线
2. T+1 未复权开盘撮合；涨跌停、停牌则不成交
3. 扣除佣金、印花税、滑点后再记账
4. 未覆盖的规则写进每次结果的 `limitations`，不假装已经模拟

信号计算和撮合必须分开，避免把「看见当天收盘再以当天收盘买入」写进策略。

## HTTP API

看板接口用 **Gin**，仍由 `net/http.Server` 托管以便优雅退出。Tushare 出站请求继续用标准库 HTTP 客户端，不经过 Gin。

选型理由：路由分组、JSON 绑定、Recovery/CORS 中间件后面都会用到；标准库 `ServeMux` 写到回测这一步已经开始散。不选 Fiber，避免和 `net/http` 生态分叉。

## 日志

统一用标准库 **slog**（`internal/applog`）：

- `config.yaml` 的 `log.level` / `log.format`，可用环境变量 `LOG_LEVEL` / `LOG_FORMAT` 覆盖
- Gin 记请求（health 走 debug）；GORM 默认只记慢查询和错误，`debug` 级别才打每条 SQL
- 采集进度、限流等待、任务起停走结构化字段；`show` / `backtest` 等人读表格仍用 `fmt`

## 配置与密钥

- 数据库连接走环境变量，不进 Git
- 数据源 Token 后续同样只放 `.env`
- 即使是个人项目，也不要把下单密钥写进仓库
