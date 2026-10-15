# ashare

个人 A 股量化系统：Go 负责采集、存储和后续回测/执行，PostgreSQL 存行情与主数据。

当前已打通 **Tushare → 日线库 → 命令行/看板质检 → 单票最小回测**。组合、模拟盘和实盘还没有做。

## 能做什么

- Docker 启动 PostgreSQL 16（端口 `55432`）
- `ashare migrate` 建立研究库
- 从 Tushare 灌入交易日历、证券列表、未复权日线、复权因子、涨跌停、停牌
- `config.yaml` 配置 Tushare 每分钟请求上限（默认 200）
- Web「采集任务」页可直接同步行情；也可用 `ashare ingest`
- `ashare show` 看 K 线，`ashare stats` 看前复权收益和最大回撤
- GORM 只做映射和 upsert；表结构以 SQL 迁移为准

## 目录

```
ashare/
  cmd/ashare/           ping / migrate / ingest / show / stats / backtest / web
  web/                  Vite + React 数据看板
  internal/httpapi/     Gin JSON API（采集 / 回测 / 质检）
  internal/applog/      slog 日志（Gin / GORM / 采集）
  internal/tushare/     Tushare 适配（手→股、千元→元）
  internal/ingest/      采集编排
  internal/research/    简单绩效
  internal/backtest/    单票双均线回测
  internal/repo/        GORM 仓储
  migrations/           SQL 迁移
  docs/
```

## 环境

- Go 1.25+
- PostgreSQL（可用 `docker compose up -d`）
- [Tushare Pro](https://tushare.pro) token

## 快速开始

```powershell
copy .env.example .env
# 编辑 .env：DATABASE_URL、TUSHARE_TOKEN
docker compose up -d
go run ./cmd/ashare migrate
go run ./cmd/ashare ping
```

先用一只股票打通（省积分）：

```powershell
go run ./cmd/ashare ingest calendar -start 20260101 -end 20261013
go run ./cmd/ashare ingest instruments
go run ./cmd/ashare ingest daily -start 20260101 -end 20261013 -code 000001.SZ
go run ./cmd/ashare show 000001.SZ -qfq
go run ./cmd/ashare stats 000001.SZ
go run ./cmd/ashare backtest 000001.SZ -start 20260101 -end 20261013
```

数据看板（两个终端）：

```powershell
go run ./cmd/ashare web
```

```powershell
cd web
npm install
npm run dev
```

打开 http://127.0.0.1:5173

一次做完（仍建议带 `-code`，全市场按交易日拉截面，积分消耗大）：

```powershell
go run ./cmd/ashare ingest sync -start 20260101 -end 20261013 -code 000001.SZ
go run ./cmd/ashare status
```

## 设计约定

1. **日线原始行情不复权。** 研究用视图 `daily_bar_qfq`。
2. **回测用交易所日历，不用自然日。**
3. **成交量单位是股，成交额单位是元。** Tushare 的 `vol`（手）入库乘 100，`amount`（千元）乘 1000。
4. **涨跌停、停牌单独落地。** `stk_limit` / `daily_basic` 若积分不够会跳过并打日志。
5. **不要用 `id` 做采集去重。** 去重走自然键。

字段说明见 [docs/data-model.md](docs/data-model.md)，模块边界见 [docs/architecture.md](docs/architecture.md)。

回测约定见 [docs/backtest.md](docs/backtest.md)。

## 下一步

1. 组合与风控（指数成分、仓位上限）
2. 模拟盘
3. 实盘（默认关闭，需明确要求）
