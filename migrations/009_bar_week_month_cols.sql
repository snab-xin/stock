-- 已应用旧版 005 的库：给周/月视图补上 change/pct_chg/ah_vol/ah_amount，与 005 当前定义一致。
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
    SUM(b.ah_vol) AS ah_vol,
    SUM(b.ah_amount) AS ah_amount,
    NULL::NUMERIC AS change,
    NULL::NUMERIC AS pct_chg,
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
    SUM(b.ah_vol) AS ah_vol,
    SUM(b.ah_amount) AS ah_amount,
    NULL::NUMERIC AS change,
    NULL::NUMERIC AS pct_chg,
    NULL::NUMERIC AS turnover,
    NULL::NUMERIC AS high_limit,
    NULL::NUMERIC AS low_limit,
    NULL::BOOLEAN AS is_limit_up,
    NULL::BOOLEAN AS is_limit_down,
    BOOL_OR(b.is_paused) AS is_paused
FROM daily_bar b
GROUP BY b.ts_code, date_trunc('month', b.trade_date::timestamp);

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
    SUM(b.ah_vol) AS ah_vol,
    SUM(b.ah_amount) AS ah_amount,
    NULL::NUMERIC AS change,
    NULL::NUMERIC AS pct_chg,
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
    SUM(b.ah_vol) AS ah_vol,
    SUM(b.ah_amount) AS ah_amount,
    NULL::NUMERIC AS change,
    NULL::NUMERIC AS pct_chg,
    NULL::NUMERIC AS turnover,
    NULL::NUMERIC AS high_limit,
    NULL::NUMERIC AS low_limit,
    NULL::BOOLEAN AS is_limit_up,
    NULL::BOOLEAN AS is_limit_down,
    BOOL_OR(b.is_paused) AS is_paused
FROM daily_bar_qfq b
GROUP BY b.ts_code, date_trunc('month', b.trade_date::timestamp);
