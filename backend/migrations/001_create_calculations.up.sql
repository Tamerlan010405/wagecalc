CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE organizations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,

    CONSTRAINT organizations_name_not_empty
        CHECK (btrim(name) <> '')
);

CREATE UNIQUE INDEX organizations_name_lower_unique
    ON organizations (lower(name));

CREATE TABLE calculations (
    id text PRIMARY KEY,
    organization_id uuid NOT NULL REFERENCES organizations(id),
    title text NOT NULL,
    period_year integer NOT NULL,
    period_quarter smallint NOT NULL,
    created_at timestamptz NOT NULL,

    CONSTRAINT calculations_id_not_empty
        CHECK (btrim(id) <> ''),
    CONSTRAINT calculations_title_not_empty
        CHECK (btrim(title) <> ''),
    CONSTRAINT calculations_period_year_valid
        CHECK (period_year >= 2021),
    CONSTRAINT calculations_period_quarter_valid
        CHECK (period_quarter BETWEEN 1 AND 4)
);

CREATE INDEX calculations_organization_id_idx
    ON calculations (organization_id);

CREATE INDEX calculations_period_idx
    ON calculations (period_year, period_quarter);

CREATE INDEX calculations_created_at_idx
    ON calculations (created_at DESC);

CREATE TABLE payroll_entries (
    calculation_id text NOT NULL REFERENCES calculations(id) ON DELETE CASCADE,
    employee_index integer NOT NULL,
    month_index smallint NOT NULL,
    total_140_cents bigint NOT NULL,
    gph_160_cents bigint NOT NULL,

    PRIMARY KEY (calculation_id, employee_index, month_index),

    CONSTRAINT payroll_entries_employee_index_positive
        CHECK (employee_index > 0),
    CONSTRAINT payroll_entries_month_index_valid
        CHECK (month_index BETWEEN 1 AND 3),
    CONSTRAINT payroll_entries_total_140_non_negative
        CHECK (total_140_cents >= 0),
    CONSTRAINT payroll_entries_gph_160_non_negative
        CHECK (gph_160_cents >= 0)
);

CREATE TABLE thresholds (
    calculation_id text PRIMARY KEY REFERENCES calculations(id) ON DELETE CASCADE,
    value_rf_cents bigint,
    value_region_cents bigint,
    region_code text,
    source text NOT NULL DEFAULT 'MANUAL',

    CONSTRAINT thresholds_value_rf_positive
        CHECK (value_rf_cents IS NULL OR value_rf_cents > 0),
    CONSTRAINT thresholds_value_region_positive
        CHECK (value_region_cents IS NULL OR value_region_cents > 0),
    CONSTRAINT thresholds_region_code_not_empty
        CHECK (region_code IS NULL OR btrim(region_code) <> ''),
    CONSTRAINT thresholds_source_valid
        CHECK (source IN ('MANUAL', 'ROSSTAT'))
);

CREATE TABLE monthly_results (
    calculation_id text NOT NULL REFERENCES calculations(id) ON DELETE CASCADE,
    month_index smallint NOT NULL,
    labor_payment_sum_cents bigint NOT NULL,
    headcount integer NOT NULL,
    average_cents bigint NOT NULL,
    included_in_quarter_average boolean NOT NULL,

    PRIMARY KEY (calculation_id, month_index),

    CONSTRAINT monthly_results_month_index_valid
        CHECK (month_index BETWEEN 1 AND 3),
    CONSTRAINT monthly_results_labor_payment_sum_non_negative
        CHECK (labor_payment_sum_cents >= 0),
    CONSTRAINT monthly_results_headcount_non_negative
        CHECK (headcount >= 0),
    CONSTRAINT monthly_results_average_non_negative
        CHECK (average_cents >= 0)
);

CREATE TABLE result_snapshots (
    calculation_id text PRIMARY KEY REFERENCES calculations(id) ON DELETE CASCADE,
    threshold_period_year integer,
    threshold_period_quarter smallint,
    months_with_payments smallint NOT NULL,
    monthly_average_sum_cents bigint NOT NULL,
    quarter_average_cents bigint NOT NULL,
    threshold_used_cents bigint,
    threshold_source text NOT NULL,
    delta_cents bigint,
    delta_percent numeric(12, 2),
    verdict text NOT NULL,
    recommendation_type text NOT NULL,
    estimated_quarter_fund_change_cents bigint NOT NULL DEFAULT 0,
    warnings jsonb NOT NULL DEFAULT '[]'::jsonb,
    engine_version text NOT NULL,
    checked_at timestamptz NOT NULL,

    CONSTRAINT result_snapshots_threshold_period_year_valid
        CHECK (threshold_period_year IS NULL OR threshold_period_year >= 2021),
    CONSTRAINT result_snapshots_threshold_period_quarter_valid
        CHECK (threshold_period_quarter IS NULL OR threshold_period_quarter BETWEEN 1 AND 4),
    CONSTRAINT result_snapshots_months_with_payments_valid
        CHECK (months_with_payments BETWEEN 0 AND 3),
    CONSTRAINT result_snapshots_monthly_average_sum_non_negative
        CHECK (monthly_average_sum_cents >= 0),
    CONSTRAINT result_snapshots_quarter_average_non_negative
        CHECK (quarter_average_cents >= 0),
    CONSTRAINT result_snapshots_threshold_used_positive
        CHECK (threshold_used_cents IS NULL OR threshold_used_cents > 0),
    CONSTRAINT result_snapshots_threshold_source_valid
        CHECK (threshold_source IN ('RF', 'REGION', 'NONE')),
    CONSTRAINT result_snapshots_verdict_valid
        CHECK (verdict IN ('PASS', 'FAIL', 'NO_THRESHOLD')),
    CONSTRAINT result_snapshots_recommendation_type_valid
        CHECK (recommendation_type IN ('CAN_DECREASE', 'NEED_INCREASE', 'NO_THRESHOLD')),
    CONSTRAINT result_snapshots_estimated_quarter_fund_change_non_negative
        CHECK (estimated_quarter_fund_change_cents >= 0),
    CONSTRAINT result_snapshots_warnings_is_array
        CHECK (jsonb_typeof(warnings) = 'array'),
    CONSTRAINT result_snapshots_engine_version_not_empty
        CHECK (btrim(engine_version) <> '')
);
