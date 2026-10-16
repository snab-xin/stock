-- 日线聚合为周线 / 月线（视图，不落实体表）。
-- trade_date = 该周期内最后一个交易日；open/close 取周期首/末日；high/low 取极值；量额求和。
-- PostgreSQL date_trunc('week') 以周一为一周起点（ISO）。

CREATE OR REPLACE VIEW daily_bar_week AS
SELECT
    b.ts_code,
    MAX(b.trade_date) AS trade_date,
    (ARRAY_AGG(b.open ORDER BY b.trade_date ASC))[1] AS open,
    MAX(b.high) AS high,
    MIN(b.low) AS low,
    (ARRAY_AGG(b.close ORDER BY b.trade_date DESC))[1] AS close,
    (ARRAY_AGG(b.pre_close ORDER BY b.trade_date ASC))[1] AS pre_close,
    SUM(b.volume) AS volume,
    SUM(b.amount) AS amount,
    NULL::NUMERIC AS turnover,
    NULL::NUMERIC AS high_limit,
    NULL::NUMERIC AS low_limit,
    NULL::BOOLEAN AS is_limit_up,
    NULL::BOOLEAN AS is_limit_down,
    BOOL_OR(b.is_paused) AS is_paused
FROM daily_bar b
GROUP BY b.ts_code, date_trunc('week', b.trade_date::timestamp);

CREATE OR REPLACE VIEW daily_bar_month AS
SELECT
    b.ts_code,
    MAX(b.trade_date) AS trade_date,
    (ARRAY_AGG(b.open ORDER BY b.trade_date ASC))[1] AS open,
    MAX(b.high) AS high,
    MIN(b.low) AS low,
    (ARRAY_AGG(b.close ORDER BY b.trade_date DESC))[1] AS close,
    (ARRAY_AGG(b.pre_close ORDER BY b.trade_date ASC))[1] AS pre_close,
    SUM(b.volume) AS volume,
    SUM(b.amount) AS amount,
    NULL::NUMERIC AS turnover,
    NULL::NUMERIC AS high_limit,
    NULL::NUMERIC AS low_limit,
    NULL::BOOLEAN AS is_limit_up,
    NULL::BOOLEAN AS is_limit_down,
    BOOL_OR(b.is_paused) AS is_paused
FROM daily_bar b
GROUP BY b.ts_code, date_trunc('month', b.trade_date::timestamp);

-- 前复权周 / 月：对已复权日线再聚合（不要先聚未复权再复权）。
CREATE OR REPLACE VIEW daily_bar_week_qfq AS
SELECT
    b.ts_code,
    MAX(b.trade_date) AS trade_date,
    (ARRAY_AGG(b.open ORDER BY b.trade_date ASC))[1] AS open,
    MAX(b.high) AS high,
    MIN(b.low) AS low,
    (ARRAY_AGG(b.close ORDER BY b.trade_date DESC))[1] AS close,
    (ARRAY_AGG(b.pre_close ORDER BY b.trade_date ASC))[1] AS pre_close,
    SUM(b.volume) AS volume,
    SUM(b.amount) AS amount,
    NULL::NUMERIC AS turnover,
    NULL::NUMERIC AS high_limit,
    NULL::NUMERIC AS low_limit,
    NULL::BOOLEAN AS is_limit_up,
    NULL::BOOLEAN AS is_limit_down,
    BOOL_OR(b.is_paused) AS is_paused
FROM daily_bar_qfq b
GROUP BY b.ts_code, date_trunc('week', b.trade_date::timestamp);

CREATE OR REPLACE VIEW daily_bar_month_qfq AS
SELECT
    b.ts_code,
    MAX(b.trade_date) AS trade_date,
    (ARRAY_AGG(b.open ORDER BY b.trade_date ASC))[1] AS open,
    MAX(b.high) AS high,
    MIN(b.low) AS low,
    (ARRAY_AGG(b.close ORDER BY b.trade_date DESC))[1] AS close,
    (ARRAY_AGG(b.pre_close ORDER BY b.trade_date ASC))[1] AS pre_close,
    SUM(b.volume) AS volume,
    SUM(b.amount) AS amount,
    NULL::NUMERIC AS turnover,
    NULL::NUMERIC AS high_limit,
    NULL::NUMERIC AS low_limit,
    NULL::BOOLEAN AS is_limit_up,
    NULL::BOOLEAN AS is_limit_down,
    BOOL_OR(b.is_paused) AS is_paused
FROM daily_bar_qfq b
GROUP BY b.ts_code, date_trunc('month', b.trade_date::timestamp);

COMMENT ON VIEW daily_bar_week IS '未复权周线，由 daily_bar 按自然周聚合';
COMMENT ON VIEW daily_bar_month IS '未复权月线，由 daily_bar 按自然月聚合';
COMMENT ON VIEW daily_bar_week_qfq IS '前复权周线，由 daily_bar_qfq 按自然周聚合';
COMMENT ON VIEW daily_bar_month_qfq IS '前复权月线，由 daily_bar_qfq 按自然月聚合';
