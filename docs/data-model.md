# 数据模型

所有原始行情表都按「自然键冲突则更新」设计，采集任务可重复跑。

每张业务表都有：

| 列 | 含义 |
|---|---|
| id | 代理主键 `BIGSERIAL`，Go 实体的 `Audit.ID` |
| created_at | 首次写入时间 |
| updated_at | 最后更新时间；GORM 写入时会带上，库内 `set_updated_at` 触发器也会维护 |

**不要用 `id` 做采集去重。** 去重和 upsert 一律走自然键（`UNIQUE`）。`id` 给关联、日志、后台展示用。

证券代码统一用 `ts_code`，例如 `000001.SZ`、`600000.SH`、`430047.BJ`。`exchange` 用 `SSE` / `SZSE` / `BSE`。

## instrument

证券主数据。自然键：`ts_code`。

| 列 | 含义 |
|---|---|
| ts_code | 带市场后缀的代码，唯一 |
| symbol | 纯数字代码 |
| name | 证券简称 |
| exchange | SSE / SZSE / BSE |
| asset_type | stock / etf / index |
| list_date / delist_date | 上市、退市日 |
| is_st | 是否 ST（随时间变化，后续可改为快照表） |
| status | L 上市 / D 退市 / P 暂停 |

## trade_calendar

交易所交易日历。自然键：`(exchange, cal_date)`。回测循环必须扫这张表里 `is_open = true` 的日期。

`pretrade_date` 是上一交易日，便于算收益率和 T+1。

上交所与深交所日历通常一致，但仍按 `exchange` 分开放，避免以后债市或北交所特例踩坑。

## daily_bar

**未复权**日线。自然键：`(ts_code, trade_date)`。这是唯一的价格事实表。

| 列 | 单位 / 说明 |
|---|---|
| open/high/low/close/pre_close | 元，未复权 |
| volume | **股**（不是手） |
| amount | **元** |
| turnover | 换手率，可空 |
| high_limit / low_limit | 涨跌停价 |
| is_limit_up / is_limit_down | 是否涨停、跌停 |
| is_paused | 当日是否停牌 |

入库时注意：

- Tushare `daily.vol` 一般是手，需要 `* 100` 再写入 `volume`
- 没有涨跌停价时不要假装算出 10%；ST、科创板、创业板、北交所限额不同
- 停牌日可以不写 K 线，或写 K 线同时 `is_paused = true`，但必须与 `suspension` 一致

## adj_factor

复权因子。自然键：`(ts_code, trade_date)`。

前复权（视图 `daily_bar_qfq`）：

```
qfq_price = raw_price * adj_factor / latest_adj_factor
```

后复权：

```
hfq_price = raw_price * adj_factor / first_adj_factor
```

分红送股后未复权价格会跳变，策略里的均线、动量应使用复权价；涨跌停判断必须用未复权价和停牌标记。

## 周线 / 月线视图

不单独落库，由日线聚合（migration `005_bar_week_month.sql`）：

| 视图 | 来源 |
|---|---|
| `daily_bar_week` / `daily_bar_month` | `daily_bar` |
| `daily_bar_week_qfq` / `daily_bar_month_qfq` | `daily_bar_qfq` |

规则：自然周（周一起点）/ 自然月；`trade_date` 为周期内最后交易日；open/close 取首末日；high/low 取极值；volume/amount 求和。涨跌停标记在周月线上为空；`is_paused` 为周期内是否曾停牌。

## suspension

按交易日记录停牌。自然键：`(ts_code, trade_date)`。回测撮合时：当日停牌则买卖都不成交。

## index_constituent

指数成分快照。自然键：`(index_code, ts_code, trade_date)`。做中证 500 / 沪深 300 增强时，用「当日可投资宇宙」，不要用今天的成分回测十年。

`trade_date` 表示该条成分在这一天有效（按日快照或按调仓日展开均可，但要在采集层统一）。

## cb_basic

可转债基础信息。自然键：`ts_code`。来源 Tushare `cb_basic`（约 2000 积分）。

字段备注写在迁移 `003_cb_basic_and_holders.sql` 的 `COMMENT ON COLUMN` 里。注意：`hold` 相关不在此表；条款类字段用 `TEXT`。

未收录已废弃的 `maturity_put_price`（文档要求改用 `maturity_call_price`）。

## cb_daily

可转债日线行情。自然键：`(ts_code, trade_date)`。来源 Tushare `cb_daily`（约 2000 积分）。

`volume` 单位为手，`amount` 单位为万元（与接口一致）。列表页用最新（或指定）交易日的 `close` / `change` / `pct_chg` 展示涨跌。

## cb_top_holder

可转债前十大持有人快照。自然键：`(ts_code, end_date, holder_rank)`。来源 Tushare `top10_cb_holders`（约 5000 积分）。

| 列 | 说明 |
|---|---|
| end_date | 报告期（年中/年报日） |
| hold_amount | **万张**（与接口一致，未换算） |
| hold_ratio | 持有比例（%） |

## ingest_run

采集任务流水，便于看「昨天日线有没有灌完」。

## daily_bar_qfq

前复权视图，依赖 `daily_bar` 与 `adj_factor` 在同一 `(ts_code, trade_date)` 上对齐。缺因子的行不会出现在视图里。

## 建议的灌数顺序

1. `trade_calendar`
2. `instrument`
3. `adj_factor`
4. `daily_bar`（含涨跌停字段）
5. `suspension`
6. `cb_basic`（可选；可转债研究）
7. `cb_daily`（可选；列表涨跌）
8. `cb_top_holder`（依赖 `cb_basic` 或显式代码）
9. `index_constituent`（做指数内选股时再灌）
