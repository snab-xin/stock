-- A 股日线研究库：原始行情只追加/幂等写入，复权与研究视图可重建。
-- schema_migrations 由迁移程序创建，不在本文件里建。
-- 约定：id 为代理主键；业务唯一性用 UNIQUE 自然键，供 ON CONFLICT upsert。

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 证券主数据（股票 / ETF / 指数）
CREATE TABLE IF NOT EXISTS instrument (
    id            BIGSERIAL PRIMARY KEY,
    ts_code       VARCHAR(16) NOT NULL,             -- 000001.SZ
    symbol        VARCHAR(8)  NOT NULL,             -- 000001
    name          VARCHAR(64) NOT NULL,
    exchange      VARCHAR(8)  NOT NULL,             -- SSE / SZSE / BSE
    asset_type    VARCHAR(16) NOT NULL DEFAULT 'stock', -- stock / etf / index
    list_date     DATE,
    delist_date   DATE,
    is_st         BOOLEAN     NOT NULL DEFAULT FALSE,
    status        VARCHAR(8)  NOT NULL DEFAULT 'L', -- L 上市 / D 退市 / P 暂停
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uk_instrument_ts_code UNIQUE (ts_code)
);

CREATE INDEX IF NOT EXISTS idx_instrument_exchange_status
    ON instrument (exchange, status);

CREATE INDEX IF NOT EXISTS idx_instrument_asset_type
    ON instrument (asset_type);

DROP TRIGGER IF EXISTS trg_instrument_updated_at ON instrument;
CREATE TRIGGER trg_instrument_updated_at
    BEFORE UPDATE ON instrument
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 交易所交易日历（回测必须用这个，不要用自然日）
CREATE TABLE IF NOT EXISTS trade_calendar (
    id            BIGSERIAL PRIMARY KEY,
    exchange      VARCHAR(8) NOT NULL,
    cal_date      DATE       NOT NULL,
    is_open       BOOLEAN    NOT NULL,
    pretrade_date DATE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uk_trade_calendar_exchange_date UNIQUE (exchange, cal_date)
);

CREATE INDEX IF NOT EXISTS idx_trade_calendar_open
    ON trade_calendar (exchange, cal_date)
    WHERE is_open;

DROP TRIGGER IF EXISTS trg_trade_calendar_updated_at ON trade_calendar;
CREATE TRIGGER trg_trade_calendar_updated_at
    BEFORE UPDATE ON trade_calendar
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 未复权日线（原始行情）
-- volume 单位：股；amount 单位：元
CREATE TABLE IF NOT EXISTS daily_bar (
    id            BIGSERIAL PRIMARY KEY,
    ts_code       VARCHAR(16) NOT NULL,
    trade_date    DATE        NOT NULL,
    open          NUMERIC(16,4) NOT NULL,
    high          NUMERIC(16,4) NOT NULL,
    low           NUMERIC(16,4) NOT NULL,
    close         NUMERIC(16,4) NOT NULL,
    pre_close     NUMERIC(16,4),
    volume        BIGINT        NOT NULL DEFAULT 0,
    amount        NUMERIC(20,2) NOT NULL DEFAULT 0,
    turnover      NUMERIC(12,6),
    high_limit    NUMERIC(16,4),
    low_limit     NUMERIC(16,4),
    is_limit_up   BOOLEAN,
    is_limit_down BOOLEAN,
    is_paused     BOOLEAN       NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uk_daily_bar_ts_code_date UNIQUE (ts_code, trade_date)
);

CREATE INDEX IF NOT EXISTS idx_daily_bar_trade_date
    ON daily_bar (trade_date);

DROP TRIGGER IF EXISTS trg_daily_bar_updated_at ON daily_bar;
CREATE TRIGGER trg_daily_bar_updated_at
    BEFORE UPDATE ON daily_bar
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 复权因子。前复权价 = 未复权价 * adj_factor / 最新 adj_factor
CREATE TABLE IF NOT EXISTS adj_factor (
    id            BIGSERIAL PRIMARY KEY,
    ts_code       VARCHAR(16) NOT NULL,
    trade_date    DATE        NOT NULL,
    adj_factor    NUMERIC(20,8) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uk_adj_factor_ts_code_date UNIQUE (ts_code, trade_date)
);

DROP TRIGGER IF EXISTS trg_adj_factor_updated_at ON adj_factor;
CREATE TRIGGER trg_adj_factor_updated_at
    BEFORE UPDATE ON adj_factor
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 停牌（按交易日）
CREATE TABLE IF NOT EXISTS suspension (
    id            BIGSERIAL PRIMARY KEY,
    ts_code       VARCHAR(16) NOT NULL,
    trade_date    DATE        NOT NULL,
    suspend_type  VARCHAR(32),
    reason        TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uk_suspension_ts_code_date UNIQUE (ts_code, trade_date)
);

DROP TRIGGER IF EXISTS trg_suspension_updated_at ON suspension;
CREATE TRIGGER trg_suspension_updated_at
    BEFORE UPDATE ON suspension
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 指数成分快照，后续组合回测用
CREATE TABLE IF NOT EXISTS index_constituent (
    id            BIGSERIAL PRIMARY KEY,
    index_code    VARCHAR(16) NOT NULL,
    ts_code       VARCHAR(16) NOT NULL,
    trade_date    DATE        NOT NULL,
    weight        NUMERIC(12,6),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uk_index_constituent UNIQUE (index_code, ts_code, trade_date)
);

CREATE INDEX IF NOT EXISTS idx_index_constituent_date
    ON index_constituent (index_code, trade_date);

DROP TRIGGER IF EXISTS trg_index_constituent_updated_at ON index_constituent;
CREATE TRIGGER trg_index_constituent_updated_at
    BEFORE UPDATE ON index_constituent
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 采集任务记录
CREATE TABLE IF NOT EXISTS ingest_run (
    id            BIGSERIAL PRIMARY KEY,
    job_name      TEXT        NOT NULL,
    started_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at   TIMESTAMPTZ,
    status        TEXT        NOT NULL, -- running / success / failed
    message       TEXT,
    rows_affected BIGINT      NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ingest_run_job_started
    ON ingest_run (job_name, started_at DESC);

DROP TRIGGER IF EXISTS trg_ingest_run_updated_at ON ingest_run;
CREATE TRIGGER trg_ingest_run_updated_at
    BEFORE UPDATE ON ingest_run
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 前复权日线视图：用每只股票最新复权因子做归一
CREATE OR REPLACE VIEW daily_bar_qfq AS
SELECT
    b.id,
    b.ts_code,
    b.trade_date,
    b.open  * f.adj_factor / lf.latest_factor AS open,
    b.high  * f.adj_factor / lf.latest_factor AS high,
    b.low   * f.adj_factor / lf.latest_factor AS low,
    b.close * f.adj_factor / lf.latest_factor AS close,
    b.pre_close * f.adj_factor / lf.latest_factor AS pre_close,
    b.volume,
    b.amount,
    b.turnover,
    b.high_limit,
    b.low_limit,
    b.is_limit_up,
    b.is_limit_down,
    b.is_paused,
    b.created_at,
    b.updated_at
FROM daily_bar b
JOIN adj_factor f
  ON f.ts_code = b.ts_code AND f.trade_date = b.trade_date
JOIN (
    SELECT ts_code, adj_factor AS latest_factor
    FROM (
        SELECT ts_code, adj_factor,
               ROW_NUMBER() OVER (PARTITION BY ts_code ORDER BY trade_date DESC) AS rn
        FROM adj_factor
    ) t
    WHERE rn = 1
) lf ON lf.ts_code = b.ts_code;
