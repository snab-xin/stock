-- 可转债日线行情（Tushare cb_daily，doc_id=187）

CREATE TABLE IF NOT EXISTS cb_daily (
    id              BIGSERIAL PRIMARY KEY,
    ts_code         VARCHAR(16)  NOT NULL,
    trade_date      DATE         NOT NULL,
    pre_close       NUMERIC(18, 6),
    open            NUMERIC(18, 6),
    high            NUMERIC(18, 6),
    low             NUMERIC(18, 6),
    close           NUMERIC(18, 6),
    change          NUMERIC(18, 6),
    pct_chg         NUMERIC(18, 6),
    volume          NUMERIC(24, 4),
    amount          NUMERIC(24, 4),
    bond_value      NUMERIC(18, 6),
    bond_over_rate  NUMERIC(18, 6),
    cb_value        NUMERIC(18, 6),
    cb_over_rate    NUMERIC(18, 6),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_cb_daily UNIQUE (ts_code, trade_date)
);

CREATE INDEX IF NOT EXISTS idx_cb_daily_trade_date ON cb_daily (trade_date);
CREATE INDEX IF NOT EXISTS idx_cb_daily_ts_date ON cb_daily (ts_code, trade_date DESC);

COMMENT ON TABLE cb_daily IS '可转债日线行情，来源 Tushare cb_daily';
COMMENT ON COLUMN cb_daily.id IS '代理主键，不去重';
COMMENT ON COLUMN cb_daily.ts_code IS '转债代码';
COMMENT ON COLUMN cb_daily.trade_date IS '交易日期';
COMMENT ON COLUMN cb_daily.pre_close IS '昨收盘价（元）';
COMMENT ON COLUMN cb_daily.open IS '开盘价（元）';
COMMENT ON COLUMN cb_daily.high IS '最高价（元）';
COMMENT ON COLUMN cb_daily.low IS '最低价（元）';
COMMENT ON COLUMN cb_daily.close IS '收盘价（元）';
COMMENT ON COLUMN cb_daily.change IS '涨跌额（元）';
COMMENT ON COLUMN cb_daily.pct_chg IS '涨跌幅（%）';
COMMENT ON COLUMN cb_daily.volume IS '成交量，单位：手（与 Tushare 一致）';
COMMENT ON COLUMN cb_daily.amount IS '成交金额，单位：万元（与 Tushare 一致）';
COMMENT ON COLUMN cb_daily.bond_value IS '纯债价值';
COMMENT ON COLUMN cb_daily.bond_over_rate IS '纯债溢价率（%）';
COMMENT ON COLUMN cb_daily.cb_value IS '转股价值';
COMMENT ON COLUMN cb_daily.cb_over_rate IS '转股溢价率（%）';
COMMENT ON COLUMN cb_daily.created_at IS '首次写入时间';
COMMENT ON COLUMN cb_daily.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_cb_daily_updated_at ON cb_daily;
CREATE TRIGGER trg_cb_daily_updated_at
    BEFORE UPDATE ON cb_daily
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
