-- 可转债基础信息（Tushare cb_basic，doc_id=185）与十大持有人（top10_cb_holders，doc_id=459）。
-- 表结构只来自本目录 SQL；GORM 不做 AutoMigrate。

CREATE TABLE IF NOT EXISTS cb_basic (
    id                    BIGSERIAL PRIMARY KEY,
    ts_code               VARCHAR(16)  NOT NULL,
    bond_full_name        TEXT         NOT NULL DEFAULT '',
    bond_short_name       VARCHAR(64)  NOT NULL DEFAULT '',
    cb_code               VARCHAR(16)  NOT NULL DEFAULT '',
    cb_type               VARCHAR(8)   NOT NULL DEFAULT '',
    stk_code              VARCHAR(16)  NOT NULL DEFAULT '',
    stk_short_name        VARCHAR(64)  NOT NULL DEFAULT '',
    maturity              NUMERIC(12, 4),
    par                   NUMERIC(18, 6),
    issue_price           NUMERIC(18, 6),
    issue_size            NUMERIC(24, 4),
    remain_size           NUMERIC(24, 4),
    value_date            DATE,
    maturity_date         DATE,
    rate_type             VARCHAR(32)  NOT NULL DEFAULT '',
    coupon_rate           NUMERIC(12, 6),
    add_rate              NUMERIC(12, 6),
    pay_per_year          INT,
    list_date             DATE,
    delist_date           DATE,
    exchange              VARCHAR(8)   NOT NULL DEFAULT '',
    conv_start_date       DATE,
    conv_end_date         DATE,
    conv_stop_date        DATE,
    first_conv_price      NUMERIC(18, 6),
    conv_price            NUMERIC(18, 6),
    rate_clause           TEXT         NOT NULL DEFAULT '',
    put_clause            TEXT         NOT NULL DEFAULT '',
    maturity_call_price   NUMERIC(18, 6),
    call_clause           TEXT         NOT NULL DEFAULT '',
    reset_clause          TEXT         NOT NULL DEFAULT '',
    conv_clause           TEXT         NOT NULL DEFAULT '',
    guarantor             TEXT         NOT NULL DEFAULT '',
    guarantee_type        TEXT         NOT NULL DEFAULT '',
    issue_rating          VARCHAR(32)  NOT NULL DEFAULT '',
    newest_rating         VARCHAR(32)  NOT NULL DEFAULT '',
    rating_comp           TEXT         NOT NULL DEFAULT '',
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_cb_basic_ts_code UNIQUE (ts_code)
);

CREATE INDEX IF NOT EXISTS idx_cb_basic_stk_code ON cb_basic (stk_code);
CREATE INDEX IF NOT EXISTS idx_cb_basic_list_date ON cb_basic (list_date);

COMMENT ON TABLE cb_basic IS '可转债基础信息，来源 Tushare cb_basic';
COMMENT ON COLUMN cb_basic.id IS '代理主键，不去重';
COMMENT ON COLUMN cb_basic.ts_code IS '转债代码，如 110059.SH；自然键';
COMMENT ON COLUMN cb_basic.bond_full_name IS '转债全称';
COMMENT ON COLUMN cb_basic.bond_short_name IS '转债简称';
COMMENT ON COLUMN cb_basic.cb_code IS '转股申报代码';
COMMENT ON COLUMN cb_basic.cb_type IS '转债类型：CB 可转债 / EB 可交换债';
COMMENT ON COLUMN cb_basic.stk_code IS '正股 ts_code';
COMMENT ON COLUMN cb_basic.stk_short_name IS '正股简称';
COMMENT ON COLUMN cb_basic.maturity IS '发行期限（年）';
COMMENT ON COLUMN cb_basic.par IS '面值（元）';
COMMENT ON COLUMN cb_basic.issue_price IS '发行价格（元）';
COMMENT ON COLUMN cb_basic.issue_size IS '发行总额（元）';
COMMENT ON COLUMN cb_basic.remain_size IS '债券余额（元）';
COMMENT ON COLUMN cb_basic.value_date IS '起息日期';
COMMENT ON COLUMN cb_basic.maturity_date IS '到期日期';
COMMENT ON COLUMN cb_basic.rate_type IS '利率类型';
COMMENT ON COLUMN cb_basic.coupon_rate IS '票面利率（%）';
COMMENT ON COLUMN cb_basic.add_rate IS '补偿利率（%）';
COMMENT ON COLUMN cb_basic.pay_per_year IS '年付息次数';
COMMENT ON COLUMN cb_basic.list_date IS '上市日期';
COMMENT ON COLUMN cb_basic.delist_date IS '摘牌日';
COMMENT ON COLUMN cb_basic.exchange IS '上市交易所原始字段（如 SSE/SZSE 或 SH/SZ）';
COMMENT ON COLUMN cb_basic.conv_start_date IS '转股起始日';
COMMENT ON COLUMN cb_basic.conv_end_date IS '转股截止日';
COMMENT ON COLUMN cb_basic.conv_stop_date IS '停止转股日（提前到期）';
COMMENT ON COLUMN cb_basic.first_conv_price IS '初始转股价（元）';
COMMENT ON COLUMN cb_basic.conv_price IS '最新转股价（元）';
COMMENT ON COLUMN cb_basic.rate_clause IS '利率说明';
COMMENT ON COLUMN cb_basic.put_clause IS '回售条款';
COMMENT ON COLUMN cb_basic.maturity_call_price IS '到期赎回价格含税（元）；对应 Tushare maturity_call_price';
COMMENT ON COLUMN cb_basic.call_clause IS '赎回条款';
COMMENT ON COLUMN cb_basic.reset_clause IS '特别向下修正条款';
COMMENT ON COLUMN cb_basic.conv_clause IS '转股条款';
COMMENT ON COLUMN cb_basic.guarantor IS '担保人';
COMMENT ON COLUMN cb_basic.guarantee_type IS '担保方式';
COMMENT ON COLUMN cb_basic.issue_rating IS '发行信用等级';
COMMENT ON COLUMN cb_basic.newest_rating IS '最新信用等级';
COMMENT ON COLUMN cb_basic.rating_comp IS '最新评级机构';
COMMENT ON COLUMN cb_basic.created_at IS '首次写入时间';
COMMENT ON COLUMN cb_basic.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_cb_basic_updated_at ON cb_basic;
CREATE TRIGGER trg_cb_basic_updated_at
    BEFORE UPDATE ON cb_basic
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS cb_top_holder (
    id           BIGSERIAL PRIMARY KEY,
    ts_code      VARCHAR(16)  NOT NULL,
    end_date     DATE         NOT NULL,
    holder_rank  INT          NOT NULL,
    holder_name  TEXT         NOT NULL DEFAULT '',
    hold_amount  NUMERIC(24, 6),
    hold_ratio   NUMERIC(12, 6),
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT uk_cb_top_holder UNIQUE (ts_code, end_date, holder_rank)
);

CREATE INDEX IF NOT EXISTS idx_cb_top_holder_end_date
    ON cb_top_holder (end_date);

CREATE INDEX IF NOT EXISTS idx_cb_top_holder_name
    ON cb_top_holder (holder_name);

COMMENT ON TABLE cb_top_holder IS '可转债前十大持有人，来源 Tushare top10_cb_holders；按报告期快照';
COMMENT ON COLUMN cb_top_holder.id IS '代理主键，不去重';
COMMENT ON COLUMN cb_top_holder.ts_code IS '转债代码';
COMMENT ON COLUMN cb_top_holder.end_date IS '报告期（年中/年报日）';
COMMENT ON COLUMN cb_top_holder.holder_rank IS '持有排名 1～10；与 ts_code+end_date 组成自然键';
COMMENT ON COLUMN cb_top_holder.holder_name IS '持有人名称';
COMMENT ON COLUMN cb_top_holder.hold_amount IS '持有数量，单位：万张（与 Tushare 一致，未换算）';
COMMENT ON COLUMN cb_top_holder.hold_ratio IS '持有比例（%）';
COMMENT ON COLUMN cb_top_holder.created_at IS '首次写入时间';
COMMENT ON COLUMN cb_top_holder.updated_at IS '最后更新时间';

DROP TRIGGER IF EXISTS trg_cb_top_holder_updated_at ON cb_top_holder;
CREATE TRIGGER trg_cb_top_holder_updated_at
    BEFORE UPDATE ON cb_top_holder
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
