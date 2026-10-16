-- 指数：基本信息 / 日线 / 申万行业分类 / 申万日线
-- Tushare index_basic(doc_id=94)、index_daily(95)、index_classify(181)、sw_daily(327)

CREATE TABLE IF NOT EXISTS index_basic (
    id           BIGSERIAL PRIMARY KEY,
    ts_code      VARCHAR(24)  NOT NULL,
    name         VARCHAR(64)  NOT NULL DEFAULT '',
    fullname     TEXT         NOT NULL DEFAULT '',
    market       VARCHAR(16)  NOT NULL DEFAULT '',
    publisher    VARCHAR(64)  NOT NULL DEFAULT '',
    index_type   VARCHAR(64)  NOT NULL DEFAULT '',
    category     VARCHAR(64)  NOT NULL DEFAULT '',
    base_date    DATE,
    base_point   NUMERIC(18, 6),
    list_date    DATE,
    weight_rule  TEXT         NOT NULL DEFAULT '',
    description  TEXT         NOT NULL DEFAULT '',
    exp_date     DATE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_index_basic_ts_code UNIQUE (ts_code)
);

CREATE INDEX IF NOT EXISTS idx_index_basic_market ON index_basic (market);
CREATE INDEX IF NOT EXISTS idx_index_basic_category ON index_basic (category);

COMMENT ON TABLE index_basic IS '指数基本信息，来源 Tushare index_basic';
COMMENT ON COLUMN index_basic.id IS '代理主键，不去重';
COMMENT ON COLUMN index_basic.ts_code IS 'TS指数代码，自然键';
COMMENT ON COLUMN index_basic.name IS '简称';
COMMENT ON COLUMN index_basic.fullname IS '指数全称';
COMMENT ON COLUMN index_basic.market IS '市场 MSCI/CSI/SSE/SZSE/CICC/SW/OTH';
COMMENT ON COLUMN index_basic.publisher IS '发布方';
COMMENT ON COLUMN index_basic.index_type IS '指数风格';
COMMENT ON COLUMN index_basic.category IS '指数类别';
COMMENT ON COLUMN index_basic.base_date IS '基期';
COMMENT ON COLUMN index_basic.base_point IS '基点';
COMMENT ON COLUMN index_basic.list_date IS '发布日期';
COMMENT ON COLUMN index_basic.weight_rule IS '加权方式';
COMMENT ON COLUMN index_basic.description IS '描述，对应 Tushare desc';
COMMENT ON COLUMN index_basic.exp_date IS '终止日期';
COMMENT ON COLUMN index_basic.created_at IS '首次写入时间';
COMMENT ON COLUMN index_basic.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_index_basic_updated_at ON index_basic;
CREATE TRIGGER trg_index_basic_updated_at
    BEFORE UPDATE ON index_basic
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS index_daily (
    id          BIGSERIAL PRIMARY KEY,
    ts_code     VARCHAR(24)  NOT NULL,
    trade_date  DATE         NOT NULL,
    close       NUMERIC(18, 6),
    open        NUMERIC(18, 6),
    high        NUMERIC(18, 6),
    low         NUMERIC(18, 6),
    pre_close   NUMERIC(18, 6),
    change      NUMERIC(18, 6),
    pct_chg     NUMERIC(18, 6),
    vol         NUMERIC(24, 4),
    amount      NUMERIC(24, 4),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_index_daily UNIQUE (ts_code, trade_date)
);

CREATE INDEX IF NOT EXISTS idx_index_daily_trade_date ON index_daily (trade_date);
CREATE INDEX IF NOT EXISTS idx_index_daily_ts_date ON index_daily (ts_code, trade_date DESC);

COMMENT ON TABLE index_daily IS '指数日线行情，来源 Tushare index_daily（不含申万行业）';
COMMENT ON COLUMN index_daily.id IS '代理主键，不去重';
COMMENT ON COLUMN index_daily.ts_code IS 'TS指数代码';
COMMENT ON COLUMN index_daily.trade_date IS '交易日';
COMMENT ON COLUMN index_daily.close IS '收盘点位';
COMMENT ON COLUMN index_daily.open IS '开盘点位';
COMMENT ON COLUMN index_daily.high IS '最高点位';
COMMENT ON COLUMN index_daily.low IS '最低点位';
COMMENT ON COLUMN index_daily.pre_close IS '昨日收盘点';
COMMENT ON COLUMN index_daily.change IS '涨跌点';
COMMENT ON COLUMN index_daily.pct_chg IS '涨跌幅（%）';
COMMENT ON COLUMN index_daily.vol IS '成交量（手，与 Tushare 一致）';
COMMENT ON COLUMN index_daily.amount IS '成交额（千元，与 Tushare 一致）';
COMMENT ON COLUMN index_daily.created_at IS '首次写入时间';
COMMENT ON COLUMN index_daily.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_index_daily_updated_at ON index_daily;
CREATE TRIGGER trg_index_daily_updated_at
    BEFORE UPDATE ON index_daily
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS index_classify (
    id             BIGSERIAL PRIMARY KEY,
    index_code     VARCHAR(24)  NOT NULL,
    industry_name  VARCHAR(64)  NOT NULL DEFAULT '',
    parent_code    VARCHAR(24)  NOT NULL DEFAULT '',
    level          VARCHAR(8)   NOT NULL DEFAULT '',
    industry_code  VARCHAR(24)  NOT NULL DEFAULT '',
    is_pub         VARCHAR(8)   NOT NULL DEFAULT '',
    src            VARCHAR(16)  NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_index_classify UNIQUE (index_code, src)
);

CREATE INDEX IF NOT EXISTS idx_index_classify_level ON index_classify (src, level);
CREATE INDEX IF NOT EXISTS idx_index_classify_parent ON index_classify (parent_code);

COMMENT ON TABLE index_classify IS '申万行业分类，来源 Tushare index_classify';
COMMENT ON COLUMN index_classify.id IS '代理主键，不去重';
COMMENT ON COLUMN index_classify.index_code IS '指数代码';
COMMENT ON COLUMN index_classify.industry_name IS '行业名称';
COMMENT ON COLUMN index_classify.parent_code IS '父级代码（一级为0）';
COMMENT ON COLUMN index_classify.level IS '行业层级 L1/L2/L3';
COMMENT ON COLUMN index_classify.industry_code IS '行业代码';
COMMENT ON COLUMN index_classify.is_pub IS '是否发布了指数';
COMMENT ON COLUMN index_classify.src IS '分类来源 SW2014 / SW2021';
COMMENT ON COLUMN index_classify.created_at IS '首次写入时间';
COMMENT ON COLUMN index_classify.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_index_classify_updated_at ON index_classify;
CREATE TRIGGER trg_index_classify_updated_at
    BEFORE UPDATE ON index_classify
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS sw_daily (
    id          BIGSERIAL PRIMARY KEY,
    ts_code     VARCHAR(24)  NOT NULL,
    trade_date  DATE         NOT NULL,
    name        VARCHAR(64)  NOT NULL DEFAULT '',
    open        NUMERIC(18, 6),
    low         NUMERIC(18, 6),
    high        NUMERIC(18, 6),
    close       NUMERIC(18, 6),
    change      NUMERIC(18, 6),
    pct_change  NUMERIC(18, 6),
    vol         NUMERIC(24, 4),
    amount      NUMERIC(24, 4),
    pe          NUMERIC(18, 6),
    pb          NUMERIC(18, 6),
    float_mv    NUMERIC(24, 6),
    total_mv    NUMERIC(24, 6),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_sw_daily UNIQUE (ts_code, trade_date)
);

CREATE INDEX IF NOT EXISTS idx_sw_daily_trade_date ON sw_daily (trade_date);
CREATE INDEX IF NOT EXISTS idx_sw_daily_ts_date ON sw_daily (ts_code, trade_date DESC);

COMMENT ON TABLE sw_daily IS '申万行业日线行情，来源 Tushare sw_daily（默认申万2021版）';
COMMENT ON COLUMN sw_daily.id IS '代理主键，不去重';
COMMENT ON COLUMN sw_daily.ts_code IS '指数代码';
COMMENT ON COLUMN sw_daily.trade_date IS '交易日期';
COMMENT ON COLUMN sw_daily.name IS '指数名称';
COMMENT ON COLUMN sw_daily.open IS '开盘点位';
COMMENT ON COLUMN sw_daily.low IS '最低点位';
COMMENT ON COLUMN sw_daily.high IS '最高点位';
COMMENT ON COLUMN sw_daily.close IS '收盘点位';
COMMENT ON COLUMN sw_daily.change IS '涨跌点位';
COMMENT ON COLUMN sw_daily.pct_change IS '涨跌幅';
COMMENT ON COLUMN sw_daily.vol IS '成交量（万股）';
COMMENT ON COLUMN sw_daily.amount IS '成交额（万元）';
COMMENT ON COLUMN sw_daily.pe IS '市盈率';
COMMENT ON COLUMN sw_daily.pb IS '市净率';
COMMENT ON COLUMN sw_daily.float_mv IS '流通市值（万元）';
COMMENT ON COLUMN sw_daily.total_mv IS '总市值（万元）';
COMMENT ON COLUMN sw_daily.created_at IS '首次写入时间';
COMMENT ON COLUMN sw_daily.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_sw_daily_updated_at ON sw_daily;
CREATE TRIGGER trg_sw_daily_updated_at
    BEFORE UPDATE ON sw_daily
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
