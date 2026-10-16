-- 股票主数据、日历、日线及相关表（Tushare stock_basic/trade_cal/daily/stk_limit/daily_basic/adj_factor/suspend_d）。
-- 表结构只来自本目录 SQL；GORM 不做 AutoMigrate。

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION set_updated_at() IS '更新行时自动刷新 updated_at';

-- ---------------------------------------------------------------------------
-- instrument ← stock_basic（doc_id=25）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS instrument (
    id             BIGSERIAL PRIMARY KEY,
    ts_code        VARCHAR(16)  NOT NULL,
    symbol         VARCHAR(16)  NOT NULL DEFAULT '',
    name           VARCHAR(64)  NOT NULL DEFAULT '',
    area           VARCHAR(32)  NOT NULL DEFAULT '',
    industry       VARCHAR(64)  NOT NULL DEFAULT '',
    fullname       TEXT         NOT NULL DEFAULT '',
    enname         TEXT         NOT NULL DEFAULT '',
    cnspell        VARCHAR(64)  NOT NULL DEFAULT '',
    market         VARCHAR(32)  NOT NULL DEFAULT '',
    exchange       VARCHAR(8)   NOT NULL DEFAULT '',
    curr_type      VARCHAR(16)  NOT NULL DEFAULT '',
    list_status    VARCHAR(8)   NOT NULL DEFAULT '',
    list_date      DATE,
    delist_date    DATE,
    is_hs          VARCHAR(8)   NOT NULL DEFAULT '',
    act_name       TEXT         NOT NULL DEFAULT '',
    act_ent_type   VARCHAR(64)  NOT NULL DEFAULT '',
    asset_type     VARCHAR(16)  NOT NULL DEFAULT 'stock',
    is_st          BOOLEAN      NOT NULL DEFAULT false,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_instrument_ts_code UNIQUE (ts_code)
);

CREATE INDEX IF NOT EXISTS idx_instrument_symbol ON instrument (symbol);
CREATE INDEX IF NOT EXISTS idx_instrument_asset_status ON instrument (asset_type, list_status);
CREATE INDEX IF NOT EXISTS idx_instrument_name ON instrument (name);

COMMENT ON TABLE instrument IS '证券主数据；股票来源 Tushare stock_basic，ETF/指数同步时也会写入';
COMMENT ON COLUMN instrument.id IS '代理主键，不去重';
COMMENT ON COLUMN instrument.ts_code IS 'TS代码，如 000001.SZ；自然键';
COMMENT ON COLUMN instrument.symbol IS '股票代码（不含后缀）';
COMMENT ON COLUMN instrument.name IS '股票名称';
COMMENT ON COLUMN instrument.area IS '地域';
COMMENT ON COLUMN instrument.industry IS '所属行业';
COMMENT ON COLUMN instrument.fullname IS '股票全称';
COMMENT ON COLUMN instrument.enname IS '英文全称';
COMMENT ON COLUMN instrument.cnspell IS '拼音缩写';
COMMENT ON COLUMN instrument.market IS '市场类型（主板/创业板/科创板/CDR/北交所）';
COMMENT ON COLUMN instrument.exchange IS '交易所代码 SSE/SZSE/BSE';
COMMENT ON COLUMN instrument.curr_type IS '交易货币';
COMMENT ON COLUMN instrument.list_status IS '上市状态 L上市 D退市 G过会未交易 P暂停上市 UN未上市';
COMMENT ON COLUMN instrument.list_date IS '上市日期';
COMMENT ON COLUMN instrument.delist_date IS '退市日期';
COMMENT ON COLUMN instrument.is_hs IS '是否沪深港通标的 N否 H沪股通 S深股通';
COMMENT ON COLUMN instrument.act_name IS '实控人名称';
COMMENT ON COLUMN instrument.act_ent_type IS '实控人企业性质';
COMMENT ON COLUMN instrument.asset_type IS '资产类型 stock/etf/index，非 Tushare 字段';
COMMENT ON COLUMN instrument.is_st IS '名称是否含 ST，由应用根据 name 派生';
COMMENT ON COLUMN instrument.created_at IS '首次写入时间';
COMMENT ON COLUMN instrument.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_instrument_updated_at ON instrument;
CREATE TRIGGER trg_instrument_updated_at
    BEFORE UPDATE ON instrument
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- trade_calendar ← trade_cal（doc_id=26）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS trade_calendar (
    id             BIGSERIAL PRIMARY KEY,
    exchange       VARCHAR(16)  NOT NULL,
    cal_date       DATE         NOT NULL,
    is_open        BOOLEAN      NOT NULL DEFAULT false,
    pretrade_date  DATE,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_trade_calendar UNIQUE (exchange, cal_date)
);

CREATE INDEX IF NOT EXISTS idx_trade_calendar_open ON trade_calendar (exchange, is_open, cal_date);

COMMENT ON TABLE trade_calendar IS '交易日历，来源 Tushare trade_cal';
COMMENT ON COLUMN trade_calendar.id IS '代理主键，不去重';
COMMENT ON COLUMN trade_calendar.exchange IS '交易所 SSE上交所 SZSE深交所；BSE 由应用从 SSE 复制';
COMMENT ON COLUMN trade_calendar.cal_date IS '日历日期';
COMMENT ON COLUMN trade_calendar.is_open IS '是否交易：false休市 true交易';
COMMENT ON COLUMN trade_calendar.pretrade_date IS '上一个交易日';
COMMENT ON COLUMN trade_calendar.created_at IS '首次写入时间';
COMMENT ON COLUMN trade_calendar.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_trade_calendar_updated_at ON trade_calendar;
CREATE TRIGGER trg_trade_calendar_updated_at
    BEFORE UPDATE ON trade_calendar
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- daily_bar ← daily（doc_id=27）+ stk_limit 涨跌停 + daily_basic.turnover_rate + 停牌标记
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS daily_bar (
    id             BIGSERIAL PRIMARY KEY,
    ts_code        VARCHAR(16)  NOT NULL,
    trade_date     DATE         NOT NULL,
    open           NUMERIC(18, 6) NOT NULL DEFAULT 0,
    high           NUMERIC(18, 6) NOT NULL DEFAULT 0,
    low            NUMERIC(18, 6) NOT NULL DEFAULT 0,
    close          NUMERIC(18, 6) NOT NULL DEFAULT 0,
    pre_close      NUMERIC(18, 6),
    change         NUMERIC(18, 6),
    pct_chg        NUMERIC(18, 6),
    volume         BIGINT       NOT NULL DEFAULT 0,
    amount         NUMERIC(24, 4) NOT NULL DEFAULT 0,
    ah_vol         BIGINT       NOT NULL DEFAULT 0,
    ah_amount      NUMERIC(24, 4) NOT NULL DEFAULT 0,
    turnover       NUMERIC(18, 6),
    high_limit     NUMERIC(18, 6),
    low_limit      NUMERIC(18, 6),
    is_limit_up    BOOLEAN,
    is_limit_down  BOOLEAN,
    is_paused      BOOLEAN      NOT NULL DEFAULT false,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_daily_bar UNIQUE (ts_code, trade_date)
);

CREATE INDEX IF NOT EXISTS idx_daily_bar_trade_date ON daily_bar (trade_date);
CREATE INDEX IF NOT EXISTS idx_daily_bar_ts_date ON daily_bar (ts_code, trade_date DESC);

COMMENT ON TABLE daily_bar IS 'A股未复权日线，来源 Tushare daily；涨跌停来自 stk_limit，换手率来自 daily_basic';
COMMENT ON COLUMN daily_bar.id IS '代理主键，不去重';
COMMENT ON COLUMN daily_bar.ts_code IS '股票代码';
COMMENT ON COLUMN daily_bar.trade_date IS '交易日期';
COMMENT ON COLUMN daily_bar.open IS '开盘价（元，未复权）';
COMMENT ON COLUMN daily_bar.high IS '最高价（元，未复权）';
COMMENT ON COLUMN daily_bar.low IS '最低价（元，未复权）';
COMMENT ON COLUMN daily_bar.close IS '收盘价（元，未复权）';
COMMENT ON COLUMN daily_bar.pre_close IS '昨收价（除权价，元）';
COMMENT ON COLUMN daily_bar.change IS '涨跌额（元）';
COMMENT ON COLUMN daily_bar.pct_chg IS '涨跌幅（%），基于除权后昨收';
COMMENT ON COLUMN daily_bar.volume IS '成交量，单位：股（Tushare daily.vol 为手，入库时 ×100）';
COMMENT ON COLUMN daily_bar.amount IS '成交额，单位：元（Tushare daily.amount 为千元，入库时 ×1000）';
COMMENT ON COLUMN daily_bar.ah_vol IS '盘后成交量，单位：股（Tushare ah_vol 为手，入库时 ×100）';
COMMENT ON COLUMN daily_bar.ah_amount IS '盘后成交额，单位：元（Tushare ah_amount 为千元，入库时 ×1000）';
COMMENT ON COLUMN daily_bar.turnover IS '换手率（%），来源 daily_basic.turnover_rate';
COMMENT ON COLUMN daily_bar.high_limit IS '涨停价（元），来源 stk_limit.up_limit';
COMMENT ON COLUMN daily_bar.low_limit IS '跌停价（元），来源 stk_limit.down_limit';
COMMENT ON COLUMN daily_bar.is_limit_up IS '收盘是否涨停（接近涨停价）';
COMMENT ON COLUMN daily_bar.is_limit_down IS '收盘是否跌停（接近跌停价）';
COMMENT ON COLUMN daily_bar.is_paused IS '当日是否停牌';
COMMENT ON COLUMN daily_bar.created_at IS '首次写入时间';
COMMENT ON COLUMN daily_bar.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_daily_bar_updated_at ON daily_bar;
CREATE TRIGGER trg_daily_bar_updated_at
    BEFORE UPDATE ON daily_bar
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- stk_limit ← stk_limit（doc_id=183）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS stk_limit (
    id          BIGSERIAL PRIMARY KEY,
    trade_date  DATE         NOT NULL,
    ts_code     VARCHAR(16)  NOT NULL,
    pre_close   NUMERIC(18, 6),
    up_limit    NUMERIC(18, 6),
    down_limit  NUMERIC(18, 6),
    asset_type  VARCHAR(16)  NOT NULL DEFAULT '',
    exchange    VARCHAR(8)   NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_stk_limit UNIQUE (ts_code, trade_date)
);

CREATE INDEX IF NOT EXISTS idx_stk_limit_trade_date ON stk_limit (trade_date);

COMMENT ON TABLE stk_limit IS '每日涨跌停价格，来源 Tushare stk_limit';
COMMENT ON COLUMN stk_limit.id IS '代理主键，不去重';
COMMENT ON COLUMN stk_limit.trade_date IS '交易日期';
COMMENT ON COLUMN stk_limit.ts_code IS 'TS股票代码';
COMMENT ON COLUMN stk_limit.pre_close IS '昨日收盘价';
COMMENT ON COLUMN stk_limit.up_limit IS '涨停价';
COMMENT ON COLUMN stk_limit.down_limit IS '跌停价';
COMMENT ON COLUMN stk_limit.asset_type IS '类型，如 STK';
COMMENT ON COLUMN stk_limit.exchange IS '交易所 SSE/SZSE/BSE';
COMMENT ON COLUMN stk_limit.created_at IS '首次写入时间';
COMMENT ON COLUMN stk_limit.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_stk_limit_updated_at ON stk_limit;
CREATE TRIGGER trg_stk_limit_updated_at
    BEFORE UPDATE ON stk_limit
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- daily_basic ← daily_basic（doc_id=32）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS daily_basic (
    id               BIGSERIAL PRIMARY KEY,
    ts_code          VARCHAR(16)  NOT NULL,
    trade_date       DATE         NOT NULL,
    close            NUMERIC(18, 6),
    turnover_rate    NUMERIC(18, 6),
    turnover_rate_f  NUMERIC(18, 6),
    volume_ratio     NUMERIC(18, 6),
    pe               NUMERIC(18, 6),
    pe_ttm           NUMERIC(18, 6),
    pb               NUMERIC(18, 6),
    ps               NUMERIC(18, 6),
    ps_ttm           NUMERIC(18, 6),
    dv_ratio         NUMERIC(18, 6),
    dv_ttm           NUMERIC(18, 6),
    total_share      NUMERIC(24, 6),
    float_share      NUMERIC(24, 6),
    free_share       NUMERIC(24, 6),
    total_mv         NUMERIC(24, 6),
    circ_mv          NUMERIC(24, 6),
    limit_status     INT,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_daily_basic UNIQUE (ts_code, trade_date)
);

CREATE INDEX IF NOT EXISTS idx_daily_basic_trade_date ON daily_basic (trade_date);

COMMENT ON TABLE daily_basic IS '每日指标，来源 Tushare daily_basic';
COMMENT ON COLUMN daily_basic.id IS '代理主键，不去重';
COMMENT ON COLUMN daily_basic.ts_code IS 'TS股票代码';
COMMENT ON COLUMN daily_basic.trade_date IS '交易日期';
COMMENT ON COLUMN daily_basic.close IS '当日收盘价';
COMMENT ON COLUMN daily_basic.turnover_rate IS '换手率（%），成交量/无限售流通股数';
COMMENT ON COLUMN daily_basic.turnover_rate_f IS '换手率（%），自由流通股';
COMMENT ON COLUMN daily_basic.volume_ratio IS '量比';
COMMENT ON COLUMN daily_basic.pe IS '市盈率（总市值/净利润，亏损为空）';
COMMENT ON COLUMN daily_basic.pe_ttm IS '市盈率 TTM（亏损为空）';
COMMENT ON COLUMN daily_basic.pb IS '市净率（总市值/(净资产-其他权益工具)）';
COMMENT ON COLUMN daily_basic.ps IS '市销率（总市值/最新年报营业收入）';
COMMENT ON COLUMN daily_basic.ps_ttm IS '市销率 TTM';
COMMENT ON COLUMN daily_basic.dv_ratio IS '股息率（%）';
COMMENT ON COLUMN daily_basic.dv_ttm IS '股息率 TTM（%）';
COMMENT ON COLUMN daily_basic.total_share IS '总股本（万股）';
COMMENT ON COLUMN daily_basic.float_share IS '流通股本（万股）';
COMMENT ON COLUMN daily_basic.free_share IS '自由流通股本（万股）';
COMMENT ON COLUMN daily_basic.total_mv IS '总市值（万元）';
COMMENT ON COLUMN daily_basic.circ_mv IS '流通市值（万元）';
COMMENT ON COLUMN daily_basic.limit_status IS '收盘涨跌状态：0平盘 1上涨 2涨停 3一字涨停 4下跌 5跌停 6一字跌停';
COMMENT ON COLUMN daily_basic.created_at IS '首次写入时间';
COMMENT ON COLUMN daily_basic.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_daily_basic_updated_at ON daily_basic;
CREATE TRIGGER trg_daily_basic_updated_at
    BEFORE UPDATE ON daily_basic
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- adj_factor ← adj_factor（doc_id=28）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS adj_factor (
    id          BIGSERIAL PRIMARY KEY,
    ts_code     VARCHAR(16)  NOT NULL,
    trade_date  DATE         NOT NULL,
    adj_factor  NUMERIC(18, 6) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_adj_factor UNIQUE (ts_code, trade_date)
);

CREATE INDEX IF NOT EXISTS idx_adj_factor_trade_date ON adj_factor (trade_date);

COMMENT ON TABLE adj_factor IS '复权因子，来源 Tushare adj_factor';
COMMENT ON COLUMN adj_factor.id IS '代理主键，不去重';
COMMENT ON COLUMN adj_factor.ts_code IS '股票代码';
COMMENT ON COLUMN adj_factor.trade_date IS '交易日期';
COMMENT ON COLUMN adj_factor.adj_factor IS '复权因子';
COMMENT ON COLUMN adj_factor.created_at IS '首次写入时间';
COMMENT ON COLUMN adj_factor.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_adj_factor_updated_at ON adj_factor;
CREATE TRIGGER trg_adj_factor_updated_at
    BEFORE UPDATE ON adj_factor
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- suspension ← suspend_d（doc_id=214）
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS suspension (
    id              BIGSERIAL PRIMARY KEY,
    ts_code         VARCHAR(16)  NOT NULL,
    trade_date      DATE         NOT NULL,
    suspend_timing  VARCHAR(64)  NOT NULL DEFAULT '',
    suspend_type    VARCHAR(8)   NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_suspension UNIQUE (ts_code, trade_date)
);

CREATE INDEX IF NOT EXISTS idx_suspension_trade_date ON suspension (trade_date);

COMMENT ON TABLE suspension IS '每日停复牌信息，来源 Tushare suspend_d';
COMMENT ON COLUMN suspension.id IS '代理主键，不去重';
COMMENT ON COLUMN suspension.ts_code IS 'TS代码';
COMMENT ON COLUMN suspension.trade_date IS '停复牌日期';
COMMENT ON COLUMN suspension.suspend_timing IS '日内停牌时间段（全日停牌为空）';
COMMENT ON COLUMN suspension.suspend_type IS '停复牌类型 S停牌 R复牌';
COMMENT ON COLUMN suspension.created_at IS '首次写入时间';
COMMENT ON COLUMN suspension.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_suspension_updated_at ON suspension;
CREATE TRIGGER trg_suspension_updated_at
    BEFORE UPDATE ON suspension
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- ingest_run 采集流水
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ingest_run (
    id             BIGSERIAL PRIMARY KEY,
    job_name       VARCHAR(64)  NOT NULL DEFAULT '',
    started_at     TIMESTAMPTZ  NOT NULL,
    finished_at    TIMESTAMPTZ,
    status         VARCHAR(16)  NOT NULL DEFAULT '',
    message        TEXT         NOT NULL DEFAULT '',
    rows_affected  BIGINT       NOT NULL DEFAULT 0,
    params         JSONB        NOT NULL DEFAULT '{}'::jsonb,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ingest_run_started_at ON ingest_run (started_at DESC);

COMMENT ON TABLE ingest_run IS '采集任务流水';
COMMENT ON COLUMN ingest_run.id IS '代理主键';
COMMENT ON COLUMN ingest_run.job_name IS '任务名，如 instruments/daily/index_basic';
COMMENT ON COLUMN ingest_run.started_at IS '开始时间';
COMMENT ON COLUMN ingest_run.finished_at IS '结束时间';
COMMENT ON COLUMN ingest_run.status IS 'running/success/failed';
COMMENT ON COLUMN ingest_run.message IS '失败原因或备注';
COMMENT ON COLUMN ingest_run.rows_affected IS '写入行数';
COMMENT ON COLUMN ingest_run.params IS '任务参数 JSON';
COMMENT ON COLUMN ingest_run.created_at IS '首次写入时间';
COMMENT ON COLUMN ingest_run.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_ingest_run_updated_at ON ingest_run;
CREATE TRIGGER trg_ingest_run_updated_at
    BEFORE UPDATE ON ingest_run
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- daily_bar_qfq 前复权视图
-- ---------------------------------------------------------------------------
CREATE OR REPLACE VIEW daily_bar_qfq AS
SELECT
    b.ts_code,
    b.trade_date,
    b.open * a.adj_factor / l.adj_factor AS open,
    b.high * a.adj_factor / l.adj_factor AS high,
    b.low * a.adj_factor / l.adj_factor AS low,
    b.close * a.adj_factor / l.adj_factor AS close,
    b.pre_close * a.adj_factor / l.adj_factor AS pre_close,
    b.change,
    b.pct_chg,
    b.volume,
    b.amount,
    b.ah_vol,
    b.ah_amount,
    b.turnover,
    b.high_limit,
    b.low_limit,
    b.is_limit_up,
    b.is_limit_down,
    b.is_paused
FROM daily_bar b
JOIN adj_factor a
    ON a.ts_code = b.ts_code AND a.trade_date = b.trade_date
JOIN (
    SELECT ts_code, adj_factor
    FROM (
        SELECT ts_code, adj_factor,
               ROW_NUMBER() OVER (PARTITION BY ts_code ORDER BY trade_date DESC) AS rn
        FROM adj_factor
    ) x
    WHERE rn = 1
) l ON l.ts_code = b.ts_code
WHERE l.adj_factor <> 0;

COMMENT ON VIEW daily_bar_qfq IS '前复权日线：raw * adj_factor / 最新 adj_factor；涨跌停价不复权';
