-- ETF：基础信息 / 日线 / 复权因子 / 份额规模 / 沪深持仓组合
-- Tushare etf_basic(385)、fund_daily(127)、fund_adj(199)、etf_share_size(408)、etf_sh_cons(471)、etf_sz_cons(472)

CREATE TABLE IF NOT EXISTS etf_basic (
    id           BIGSERIAL PRIMARY KEY,
    ts_code      VARCHAR(16)  NOT NULL,
    csname       VARCHAR(64)  NOT NULL DEFAULT '',
    extname      VARCHAR(64)  NOT NULL DEFAULT '',
    cname        TEXT         NOT NULL DEFAULT '',
    index_code   VARCHAR(24)  NOT NULL DEFAULT '',
    index_name   TEXT         NOT NULL DEFAULT '',
    setup_date   DATE,
    list_date    DATE,
    list_status  VARCHAR(8)   NOT NULL DEFAULT '',
    exchange     VARCHAR(8)   NOT NULL DEFAULT '',
    mgr_name     VARCHAR(64)  NOT NULL DEFAULT '',
    custod_name  TEXT         NOT NULL DEFAULT '',
    mgt_fee      NUMERIC(12, 6),
    etf_type     VARCHAR(16)  NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_etf_basic_ts_code UNIQUE (ts_code)
);

CREATE INDEX IF NOT EXISTS idx_etf_basic_index_code ON etf_basic (index_code);
CREATE INDEX IF NOT EXISTS idx_etf_basic_list_status ON etf_basic (list_status);
CREATE INDEX IF NOT EXISTS idx_etf_basic_mgr ON etf_basic (mgr_name);

COMMENT ON TABLE etf_basic IS 'ETF基础信息，来源 Tushare etf_basic';
COMMENT ON COLUMN etf_basic.id IS '代理主键，不去重';
COMMENT ON COLUMN etf_basic.ts_code IS '基金交易代码，自然键';
COMMENT ON COLUMN etf_basic.csname IS 'ETF中文简称';
COMMENT ON COLUMN etf_basic.extname IS 'ETF扩位简称（交易所简称）';
COMMENT ON COLUMN etf_basic.cname IS '基金中文全称';
COMMENT ON COLUMN etf_basic.index_code IS 'ETF基准指数代码';
COMMENT ON COLUMN etf_basic.index_name IS 'ETF基准指数中文全称';
COMMENT ON COLUMN etf_basic.setup_date IS '设立日期';
COMMENT ON COLUMN etf_basic.list_date IS '上市日期';
COMMENT ON COLUMN etf_basic.list_status IS '存续状态 L上市 D退市 P待上市';
COMMENT ON COLUMN etf_basic.exchange IS '交易所 上交所SH 深交所SZ';
COMMENT ON COLUMN etf_basic.mgr_name IS '基金管理人简称';
COMMENT ON COLUMN etf_basic.custod_name IS '基金托管人名称';
COMMENT ON COLUMN etf_basic.mgt_fee IS '基金管理人收取的费用';
COMMENT ON COLUMN etf_basic.etf_type IS '基金投资通道类型（境内、QDII）';
COMMENT ON COLUMN etf_basic.created_at IS '首次写入时间';
COMMENT ON COLUMN etf_basic.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_etf_basic_updated_at ON etf_basic;
CREATE TRIGGER trg_etf_basic_updated_at
    BEFORE UPDATE ON etf_basic
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS etf_daily (
    id          BIGSERIAL PRIMARY KEY,
    ts_code     VARCHAR(16)  NOT NULL,
    trade_date  DATE         NOT NULL,
    open        NUMERIC(18, 6),
    high        NUMERIC(18, 6),
    low         NUMERIC(18, 6),
    close       NUMERIC(18, 6),
    pre_close   NUMERIC(18, 6),
    change      NUMERIC(18, 6),
    pct_chg     NUMERIC(18, 6),
    vol         NUMERIC(24, 4),
    amount      NUMERIC(24, 4),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_etf_daily UNIQUE (ts_code, trade_date)
);

CREATE INDEX IF NOT EXISTS idx_etf_daily_trade_date ON etf_daily (trade_date);
CREATE INDEX IF NOT EXISTS idx_etf_daily_ts_date ON etf_daily (ts_code, trade_date DESC);

COMMENT ON TABLE etf_daily IS 'ETF日线行情，来源 Tushare fund_daily';
COMMENT ON COLUMN etf_daily.id IS '代理主键，不去重';
COMMENT ON COLUMN etf_daily.ts_code IS 'TS代码';
COMMENT ON COLUMN etf_daily.trade_date IS '交易日期';
COMMENT ON COLUMN etf_daily.open IS '开盘价（元）';
COMMENT ON COLUMN etf_daily.high IS '最高价（元）';
COMMENT ON COLUMN etf_daily.low IS '最低价（元）';
COMMENT ON COLUMN etf_daily.close IS '收盘价（元）';
COMMENT ON COLUMN etf_daily.pre_close IS '昨收盘价（元）';
COMMENT ON COLUMN etf_daily.change IS '涨跌额（元）';
COMMENT ON COLUMN etf_daily.pct_chg IS '涨跌幅（%）';
COMMENT ON COLUMN etf_daily.vol IS '成交量（手，与 Tushare 一致）';
COMMENT ON COLUMN etf_daily.amount IS '成交额（千元，与 Tushare 一致）';
COMMENT ON COLUMN etf_daily.created_at IS '首次写入时间';
COMMENT ON COLUMN etf_daily.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_etf_daily_updated_at ON etf_daily;
CREATE TRIGGER trg_etf_daily_updated_at
    BEFORE UPDATE ON etf_daily
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS etf_adj (
    id          BIGSERIAL PRIMARY KEY,
    ts_code     VARCHAR(16)  NOT NULL,
    trade_date  DATE         NOT NULL,
    adj_factor  NUMERIC(18, 6) NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_etf_adj UNIQUE (ts_code, trade_date)
);

CREATE INDEX IF NOT EXISTS idx_etf_adj_trade_date ON etf_adj (trade_date);

COMMENT ON TABLE etf_adj IS 'ETF复权因子，来源 Tushare fund_adj';
COMMENT ON COLUMN etf_adj.id IS '代理主键，不去重';
COMMENT ON COLUMN etf_adj.ts_code IS 'ts基金代码';
COMMENT ON COLUMN etf_adj.trade_date IS '交易日期';
COMMENT ON COLUMN etf_adj.adj_factor IS '复权因子';
COMMENT ON COLUMN etf_adj.created_at IS '首次写入时间';
COMMENT ON COLUMN etf_adj.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_etf_adj_updated_at ON etf_adj;
CREATE TRIGGER trg_etf_adj_updated_at
    BEFORE UPDATE ON etf_adj
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS etf_share_size (
    id           BIGSERIAL PRIMARY KEY,
    trade_date   DATE         NOT NULL,
    ts_code      VARCHAR(16)  NOT NULL,
    etf_name     VARCHAR(64)  NOT NULL DEFAULT '',
    total_share  NUMERIC(24, 6),
    total_size   NUMERIC(24, 6),
    nav          NUMERIC(18, 6),
    close        NUMERIC(18, 6),
    exchange     VARCHAR(8)   NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_etf_share_size UNIQUE (ts_code, trade_date)
);

CREATE INDEX IF NOT EXISTS idx_etf_share_size_trade_date ON etf_share_size (trade_date);
CREATE INDEX IF NOT EXISTS idx_etf_share_size_exchange ON etf_share_size (exchange, trade_date);

COMMENT ON TABLE etf_share_size IS 'ETF份额规模，来源 Tushare etf_share_size';
COMMENT ON COLUMN etf_share_size.id IS '代理主键，不去重';
COMMENT ON COLUMN etf_share_size.trade_date IS '交易日期';
COMMENT ON COLUMN etf_share_size.ts_code IS 'ETF代码';
COMMENT ON COLUMN etf_share_size.etf_name IS '基金名称';
COMMENT ON COLUMN etf_share_size.total_share IS '总份额（万份）';
COMMENT ON COLUMN etf_share_size.total_size IS '总规模（万元）';
COMMENT ON COLUMN etf_share_size.nav IS '基金份额净值（元）';
COMMENT ON COLUMN etf_share_size.close IS '收盘价（元）';
COMMENT ON COLUMN etf_share_size.exchange IS '交易所 SSE/SZSE/BSE';
COMMENT ON COLUMN etf_share_size.created_at IS '首次写入时间';
COMMENT ON COLUMN etf_share_size.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_etf_share_size_updated_at ON etf_share_size;
CREATE TRIGGER trg_etf_share_size_updated_at
    BEFORE UPDATE ON etf_share_size
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS etf_sh_cons (
    id          BIGSERIAL PRIMARY KEY,
    trade_date  DATE         NOT NULL,
    ts_code     VARCHAR(16)  NOT NULL,
    con_code    VARCHAR(16)  NOT NULL,
    con_name    VARCHAR(64)  NOT NULL DEFAULT '',
    qty         BIGINT,
    sub_flag    VARCHAR(16)  NOT NULL DEFAULT '',
    cpr         NUMERIC(18, 6),
    rdr         NUMERIC(18, 6),
    sca         NUMERIC(24, 6),
    exchange    VARCHAR(8)   NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_etf_sh_cons UNIQUE (ts_code, trade_date, con_code)
);

CREATE INDEX IF NOT EXISTS idx_etf_sh_cons_trade_date ON etf_sh_cons (trade_date);
CREATE INDEX IF NOT EXISTS idx_etf_sh_cons_con ON etf_sh_cons (con_code, trade_date);

COMMENT ON TABLE etf_sh_cons IS 'ETF每日持仓组合（沪市），来源 Tushare etf_sh_cons';
COMMENT ON COLUMN etf_sh_cons.id IS '代理主键，不去重';
COMMENT ON COLUMN etf_sh_cons.trade_date IS '交易日期';
COMMENT ON COLUMN etf_sh_cons.ts_code IS 'ETF代码';
COMMENT ON COLUMN etf_sh_cons.con_code IS '成分代码';
COMMENT ON COLUMN etf_sh_cons.con_name IS '成分名称';
COMMENT ON COLUMN etf_sh_cons.qty IS '股票数量（股）';
COMMENT ON COLUMN etf_sh_cons.sub_flag IS '现金替代标志：允许/必须';
COMMENT ON COLUMN etf_sh_cons.cpr IS '申购现金替代溢价比率（%）';
COMMENT ON COLUMN etf_sh_cons.rdr IS '赎回现金替代折价比率（%）';
COMMENT ON COLUMN etf_sh_cons.sca IS '替代金额（人民币元）';
COMMENT ON COLUMN etf_sh_cons.exchange IS '交易所代码 HK/SH/SZ/OTH';
COMMENT ON COLUMN etf_sh_cons.created_at IS '首次写入时间';
COMMENT ON COLUMN etf_sh_cons.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_etf_sh_cons_updated_at ON etf_sh_cons;
CREATE TRIGGER trg_etf_sh_cons_updated_at
    BEFORE UPDATE ON etf_sh_cons
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS etf_sz_cons (
    id          BIGSERIAL PRIMARY KEY,
    trade_date  DATE         NOT NULL,
    ts_code     VARCHAR(16)  NOT NULL,
    con_code    VARCHAR(16)  NOT NULL,
    con_name    VARCHAR(64)  NOT NULL DEFAULT '',
    qty         BIGINT,
    sub_flag    VARCHAR(16)  NOT NULL DEFAULT '',
    cpr         NUMERIC(18, 6),
    rdr         NUMERIC(18, 6),
    sub_cc      NUMERIC(24, 6),
    red_cc      NUMERIC(24, 6),
    exchange    VARCHAR(8)   NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_etf_sz_cons UNIQUE (ts_code, trade_date, con_code)
);

CREATE INDEX IF NOT EXISTS idx_etf_sz_cons_trade_date ON etf_sz_cons (trade_date);
CREATE INDEX IF NOT EXISTS idx_etf_sz_cons_con ON etf_sz_cons (con_code, trade_date);

COMMENT ON TABLE etf_sz_cons IS 'ETF每日持仓组合（深市），来源 Tushare etf_sz_cons';
COMMENT ON COLUMN etf_sz_cons.id IS '代理主键，不去重';
COMMENT ON COLUMN etf_sz_cons.trade_date IS '交易日期';
COMMENT ON COLUMN etf_sz_cons.ts_code IS 'ETF代码';
COMMENT ON COLUMN etf_sz_cons.con_code IS '成分代码';
COMMENT ON COLUMN etf_sz_cons.con_name IS '成分名称';
COMMENT ON COLUMN etf_sz_cons.qty IS '股票数量（股）';
COMMENT ON COLUMN etf_sz_cons.sub_flag IS '现金替代标志';
COMMENT ON COLUMN etf_sz_cons.cpr IS '申购现金替代保证金率（%）';
COMMENT ON COLUMN etf_sz_cons.rdr IS '赎回现金替代保证金率（%）';
COMMENT ON COLUMN etf_sz_cons.sub_cc IS '申购替代金额（人民币元）';
COMMENT ON COLUMN etf_sz_cons.red_cc IS '赎回替代金额（人民币元）';
COMMENT ON COLUMN etf_sz_cons.exchange IS '交易所代码 HK/SH/SZ/OTH';
COMMENT ON COLUMN etf_sz_cons.created_at IS '首次写入时间';
COMMENT ON COLUMN etf_sz_cons.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_etf_sz_cons_updated_at ON etf_sz_cons;
CREATE TRIGGER trg_etf_sz_cons_updated_at
    BEFORE UPDATE ON etf_sz_cons
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
