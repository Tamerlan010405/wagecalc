DELETE FROM calculations;

ALTER TABLE payroll_entries
    DROP COLUMN total_140_ciphertext,
    DROP COLUMN gph_160_ciphertext,
    ADD COLUMN total_140_cents bigint NOT NULL,
    ADD COLUMN gph_160_cents bigint NOT NULL;

ALTER TABLE monthly_results
    DROP COLUMN labor_payment_sum_ciphertext,
    DROP COLUMN headcount_ciphertext,
    DROP COLUMN average_ciphertext,
    ADD COLUMN labor_payment_sum_cents bigint NOT NULL,
    ADD COLUMN headcount integer NOT NULL,
    ADD COLUMN average_cents bigint NOT NULL;

ALTER TABLE result_snapshots
    DROP COLUMN monthly_average_sum_ciphertext,
    DROP COLUMN quarter_average_ciphertext,
    DROP COLUMN threshold_used_ciphertext,
    DROP COLUMN delta_ciphertext,
    DROP COLUMN delta_percent_ciphertext,
    DROP COLUMN verdict_ciphertext,
    DROP COLUMN recommendation_type_ciphertext,
    DROP COLUMN estimated_fund_change_ciphertext,
    DROP COLUMN warnings_ciphertext,
    ADD COLUMN monthly_average_sum_cents bigint NOT NULL,
    ADD COLUMN quarter_average_cents bigint NOT NULL,
    ADD COLUMN threshold_used_cents bigint,
    ADD COLUMN delta_cents bigint,
    ADD COLUMN delta_percent numeric(12, 2),
    ADD COLUMN verdict text NOT NULL,
    ADD COLUMN recommendation_type text NOT NULL,
    ADD COLUMN estimated_quarter_fund_change_cents bigint NOT NULL DEFAULT 0,
    ADD COLUMN warnings jsonb NOT NULL DEFAULT '[]'::jsonb;
