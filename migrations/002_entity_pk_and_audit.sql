-- 给已执行旧版 001 的库补代理主键与审计字段。
-- 新库 001 已包含这些结构，本脚本全程幂等。

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ---------- 通用：加列 ----------
ALTER TABLE instrument ADD COLUMN IF NOT EXISTS id BIGINT;
ALTER TABLE instrument ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE instrument ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE trade_calendar ADD COLUMN IF NOT EXISTS id BIGINT;
ALTER TABLE trade_calendar ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE trade_calendar ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE daily_bar ADD COLUMN IF NOT EXISTS id BIGINT;
ALTER TABLE daily_bar ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE daily_bar ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE adj_factor ADD COLUMN IF NOT EXISTS id BIGINT;
ALTER TABLE adj_factor ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE adj_factor ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE suspension ADD COLUMN IF NOT EXISTS id BIGINT;
ALTER TABLE suspension ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE suspension ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE index_constituent ADD COLUMN IF NOT EXISTS id BIGINT;
ALTER TABLE index_constituent ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE index_constituent ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE ingest_run ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();
ALTER TABLE ingest_run ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- ---------- 回填序列主键 ----------
DO $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN
        SELECT *
        FROM (VALUES
            ('instrument', 'instrument_id_seq'),
            ('trade_calendar', 'trade_calendar_id_seq'),
            ('daily_bar', 'daily_bar_id_seq'),
            ('adj_factor', 'adj_factor_id_seq'),
            ('suspension', 'suspension_id_seq'),
            ('index_constituent', 'index_constituent_id_seq')
        ) AS t(tbl, seq)
    LOOP
        EXECUTE format('CREATE SEQUENCE IF NOT EXISTS %I', rec.seq);
        EXECUTE format(
            'ALTER TABLE %I ALTER COLUMN id SET DEFAULT nextval(%L)',
            rec.tbl, rec.seq
        );
        EXECUTE format(
            'UPDATE %I SET id = nextval(%L) WHERE id IS NULL',
            rec.tbl, rec.seq
        );
        EXECUTE format('ALTER TABLE %I ALTER COLUMN id SET NOT NULL', rec.tbl);
        EXECUTE format('ALTER SEQUENCE %I OWNED BY %I.id', rec.seq, rec.tbl);
        EXECUTE format(
            'SELECT setval(%L, COALESCE((SELECT MAX(id) FROM %I), 1), (SELECT COUNT(*) > 0 FROM %I))',
            rec.seq, rec.tbl, rec.tbl
        );
    END LOOP;
END $$;

-- ---------- 主键切到 id，自然键改为 UNIQUE ----------
DO $$
BEGIN
    -- instrument
    IF EXISTS (
        SELECT 1 FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        WHERE t.relname = 'instrument' AND c.contype = 'p' AND c.conname = 'instrument_pkey'
          AND pg_get_constraintdef(c.oid) ILIKE '%ts_code%'
    ) THEN
        ALTER TABLE instrument DROP CONSTRAINT instrument_pkey;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        WHERE t.relname = 'instrument' AND c.contype = 'p'
    ) THEN
        ALTER TABLE instrument ADD CONSTRAINT instrument_pkey PRIMARY KEY (id);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uk_instrument_ts_code'
    ) THEN
        ALTER TABLE instrument ADD CONSTRAINT uk_instrument_ts_code UNIQUE (ts_code);
    END IF;

    -- trade_calendar
    IF EXISTS (
        SELECT 1 FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        WHERE t.relname = 'trade_calendar' AND c.contype = 'p'
          AND pg_get_constraintdef(c.oid) ILIKE '%cal_date%'
    ) THEN
        ALTER TABLE trade_calendar DROP CONSTRAINT trade_calendar_pkey;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        WHERE t.relname = 'trade_calendar' AND c.contype = 'p'
    ) THEN
        ALTER TABLE trade_calendar ADD CONSTRAINT trade_calendar_pkey PRIMARY KEY (id);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uk_trade_calendar_exchange_date'
    ) THEN
        ALTER TABLE trade_calendar ADD CONSTRAINT uk_trade_calendar_exchange_date UNIQUE (exchange, cal_date);
    END IF;

    -- daily_bar
    IF EXISTS (
        SELECT 1 FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        WHERE t.relname = 'daily_bar' AND c.contype = 'p'
          AND pg_get_constraintdef(c.oid) ILIKE '%trade_date%'
    ) THEN
        ALTER TABLE daily_bar DROP CONSTRAINT daily_bar_pkey;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        WHERE t.relname = 'daily_bar' AND c.contype = 'p'
    ) THEN
        ALTER TABLE daily_bar ADD CONSTRAINT daily_bar_pkey PRIMARY KEY (id);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uk_daily_bar_ts_code_date'
    ) THEN
        ALTER TABLE daily_bar ADD CONSTRAINT uk_daily_bar_ts_code_date UNIQUE (ts_code, trade_date);
    END IF;

    -- adj_factor
    IF EXISTS (
        SELECT 1 FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        WHERE t.relname = 'adj_factor' AND c.contype = 'p'
          AND pg_get_constraintdef(c.oid) ILIKE '%trade_date%'
    ) THEN
        ALTER TABLE adj_factor DROP CONSTRAINT adj_factor_pkey;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        WHERE t.relname = 'adj_factor' AND c.contype = 'p'
    ) THEN
        ALTER TABLE adj_factor ADD CONSTRAINT adj_factor_pkey PRIMARY KEY (id);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uk_adj_factor_ts_code_date'
    ) THEN
        ALTER TABLE adj_factor ADD CONSTRAINT uk_adj_factor_ts_code_date UNIQUE (ts_code, trade_date);
    END IF;

    -- suspension
    IF EXISTS (
        SELECT 1 FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        WHERE t.relname = 'suspension' AND c.contype = 'p'
          AND pg_get_constraintdef(c.oid) ILIKE '%trade_date%'
    ) THEN
        ALTER TABLE suspension DROP CONSTRAINT suspension_pkey;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        WHERE t.relname = 'suspension' AND c.contype = 'p'
    ) THEN
        ALTER TABLE suspension ADD CONSTRAINT suspension_pkey PRIMARY KEY (id);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uk_suspension_ts_code_date'
    ) THEN
        ALTER TABLE suspension ADD CONSTRAINT uk_suspension_ts_code_date UNIQUE (ts_code, trade_date);
    END IF;

    -- index_constituent
    IF EXISTS (
        SELECT 1 FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        WHERE t.relname = 'index_constituent' AND c.contype = 'p'
          AND pg_get_constraintdef(c.oid) ILIKE '%trade_date%'
    ) THEN
        ALTER TABLE index_constituent DROP CONSTRAINT index_constituent_pkey;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        WHERE t.relname = 'index_constituent' AND c.contype = 'p'
    ) THEN
        ALTER TABLE index_constituent ADD CONSTRAINT index_constituent_pkey PRIMARY KEY (id);
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uk_index_constituent'
    ) THEN
        ALTER TABLE index_constituent ADD CONSTRAINT uk_index_constituent UNIQUE (index_code, ts_code, trade_date);
    END IF;
END $$;

DROP TRIGGER IF EXISTS trg_instrument_updated_at ON instrument;
CREATE TRIGGER trg_instrument_updated_at
    BEFORE UPDATE ON instrument
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_trade_calendar_updated_at ON trade_calendar;
CREATE TRIGGER trg_trade_calendar_updated_at
    BEFORE UPDATE ON trade_calendar
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_daily_bar_updated_at ON daily_bar;
CREATE TRIGGER trg_daily_bar_updated_at
    BEFORE UPDATE ON daily_bar
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_adj_factor_updated_at ON adj_factor;
CREATE TRIGGER trg_adj_factor_updated_at
    BEFORE UPDATE ON adj_factor
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_suspension_updated_at ON suspension;
CREATE TRIGGER trg_suspension_updated_at
    BEFORE UPDATE ON suspension
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_index_constituent_updated_at ON index_constituent;
CREATE TRIGGER trg_index_constituent_updated_at
    BEFORE UPDATE ON index_constituent
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_ingest_run_updated_at ON ingest_run;
CREATE TRIGGER trg_ingest_run_updated_at
    BEFORE UPDATE ON ingest_run
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

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
