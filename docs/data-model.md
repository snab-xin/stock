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

## suspension

按交易日记录停牌。自然键：`(ts_code, trade_date)`。回测撮合时：当日停牌则买卖都不成交。

## index_constituent

指数成分快照。自然键：`(index_code, ts_code, trade_date)`。做中证 500 / 沪深 300 增强时，用「当日可投资宇宙」，不要用今天的成分回测十年。

`trade_date` 表示该条成分在这一天有效（按日快照或按调仓日展开均可，但要在采集层统一）。

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
6. `index_constituent`（做指数内选股时再灌）
