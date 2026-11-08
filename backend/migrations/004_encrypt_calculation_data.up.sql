CREATE EXTENSION IF NOT EXISTS pgcrypto;

DELETE FROM calculations;

ALTER TABLE payroll_entries
    DROP COLUMN total_140_cents,
    DROP COLUMN gph_160_cents,
    ADD COLUMN total_140_ciphertext bytea NOT NULL,
    ADD COLUMN gph_160_ciphertext bytea NOT NULL,
    ADD CONSTRAINT payroll_entries_total_140_ciphertext_not_empty
        CHECK (octet_length(total_140_ciphertext) > 0),
    ADD CONSTRAINT payroll_entries_gph_160_ciphertext_not_empty
        CHECK (octet_length(gph_160_ciphertext) > 0);

ALTER TABLE monthly_results
    DROP COLUMN labor_payment_sum_cents,
    DROP COLUMN headcount,
    DROP COLUMN average_cents,
    ADD COLUMN labor_payment_sum_ciphertext bytea NOT NULL,
    ADD COLUMN headcount_ciphertext bytea NOT NULL,
    ADD COLUMN average_ciphertext bytea NOT NULL,
    ADD CONSTRAINT monthly_results_labor_payment_sum_ciphertext_not_empty
        CHECK (octet_length(labor_payment_sum_ciphertext) > 0),
    ADD CONSTRAINT monthly_results_headcount_ciphertext_not_empty
        CHECK (octet_length(headcount_ciphertext) > 0),
    ADD CONSTRAINT monthly_results_average_ciphertext_not_empty
        CHECK (octet_length(average_ciphertext) > 0);

ALTER TABLE result_snapshots
    DROP COLUMN monthly_average_sum_cents,
    DROP COLUMN quarter_average_cents,
    DROP COLUMN threshold_used_cents,
    DROP COLUMN delta_cents,
    DROP COLUMN delta_percent,
    DROP COLUMN verdict,
    DROP COLUMN recommendation_type,
    DROP COLUMN estimated_quarter_fund_change_cents,
    DROP COLUMN warnings,
    ADD COLUMN monthly_average_sum_ciphertext bytea NOT NULL,
    ADD COLUMN quarter_average_ciphertext bytea NOT NULL,
    ADD COLUMN threshold_used_ciphertext bytea,
    ADD COLUMN delta_ciphertext bytea,
    ADD COLUMN delta_percent_ciphertext bytea,
    ADD COLUMN verdict_ciphertext bytea NOT NULL,
    ADD COLUMN recommendation_type_ciphertext bytea NOT NULL,
    ADD COLUMN estimated_fund_change_ciphertext bytea NOT NULL,
    ADD COLUMN warnings_ciphertext bytea NOT NULL,
    ADD CONSTRAINT result_snapshots_monthly_average_sum_ciphertext_not_empty
        CHECK (octet_length(monthly_average_sum_ciphertext) > 0),
    ADD CONSTRAINT result_snapshots_quarter_average_ciphertext_not_empty
        CHECK (octet_length(quarter_average_ciphertext) > 0),
    ADD CONSTRAINT result_snapshots_threshold_used_ciphertext_not_empty
        CHECK (threshold_used_ciphertext IS NULL OR octet_length(threshold_used_ciphertext) > 0),
    ADD CONSTRAINT result_snapshots_delta_ciphertext_not_empty
        CHECK (delta_ciphertext IS NULL OR octet_length(delta_ciphertext) > 0),
    ADD CONSTRAINT result_snapshots_delta_percent_ciphertext_not_empty
        CHECK (delta_percent_ciphertext IS NULL OR octet_length(delta_percent_ciphertext) > 0),
    ADD CONSTRAINT result_snapshots_verdict_ciphertext_not_empty
        CHECK (octet_length(verdict_ciphertext) > 0),
    ADD CONSTRAINT result_snapshots_recommendation_type_ciphertext_not_empty
        CHECK (octet_length(recommendation_type_ciphertext) > 0),
    ADD CONSTRAINT result_snapshots_estimated_fund_change_ciphertext_not_empty
        CHECK (octet_length(estimated_fund_change_ciphertext) > 0),
    ADD CONSTRAINT result_snapshots_warnings_ciphertext_not_empty
        CHECK (octet_length(warnings_ciphertext) > 0);
