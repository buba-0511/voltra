CREATE TABLE IF NOT EXISTS tenants (
    id          SERIAL PRIMARY KEY,
    slug        TEXT UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
    id              SERIAL PRIMARY KEY,
    tenant_id       INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email           TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    name            TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, email)
);

CREATE TABLE IF NOT EXISTS meters (
    id          SERIAL PRIMARY KEY,
    tenant_id   INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    meter_id    TEXT NOT NULL,
    name        TEXT NOT NULL,
    location    TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'OK',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, meter_id)
);

CREATE TABLE IF NOT EXISTS readings (
    id                  BIGSERIAL PRIMARY KEY,
    tenant_id           INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    meter_id            TEXT NOT NULL,
    timestamp           TIMESTAMPTZ NOT NULL,
    consumption_kwh     DOUBLE PRECISION NOT NULL,
    voltage_v           DOUBLE PRECISION NOT NULL,
    current_a           DOUBLE PRECISION NOT NULL,
    power_factor        DOUBLE PRECISION NOT NULL,
    status              TEXT NOT NULL DEFAULT 'OK',
    UNIQUE (tenant_id, meter_id, timestamp),
    FOREIGN KEY (tenant_id, meter_id) REFERENCES meters (tenant_id, meter_id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_readings_meter_ts ON readings (tenant_id, meter_id, timestamp);

CREATE TABLE IF NOT EXISTS events (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    meter_id        TEXT NOT NULL,
    timestamp       TIMESTAMPTZ NOT NULL,
    type            TEXT NOT NULL,
    description     TEXT NOT NULL,
    FOREIGN KEY (tenant_id, meter_id) REFERENCES meters (tenant_id, meter_id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_events_meter_ts ON events (tenant_id, meter_id, timestamp);

CREATE TABLE IF NOT EXISTS analysis_runs (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at     TIMESTAMPTZ,
    status          TEXT NOT NULL DEFAULT 'RUNNING',
    stage           TEXT NOT NULL DEFAULT 'READINGS',
    meters_analyzed INT NOT NULL DEFAULT 0,
    anomalies_found INT NOT NULL DEFAULT 0,
    high_priority   INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS anomalies (
    id                  BIGSERIAL PRIMARY KEY,
    tenant_id           INT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    analysis_run_id     BIGINT NOT NULL REFERENCES analysis_runs(id) ON DELETE CASCADE,
    meter_id            TEXT NOT NULL,
    detected_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    window_start        TIMESTAMPTZ NOT NULL,
    window_end          TIMESTAMPTZ NOT NULL,
    type                TEXT NOT NULL,       -- REAL_ANOMALY | EXPLAINABLE_ANOMALY | FALSE_POSITIVE | DATA_QUALITY
    severity            TEXT NOT NULL,       -- HIGH | MEDIUM | LOW
    confidence          DOUBLE PRECISION NOT NULL,
    reason              TEXT NOT NULL,
    recommended_action  TEXT NOT NULL,
    status              TEXT NOT NULL DEFAULT 'OPEN',
    baseline_kwh        DOUBLE PRECISION NOT NULL,
    actual_kwh          DOUBLE PRECISION NOT NULL,
    variation_pct       DOUBLE PRECISION NOT NULL,
    evidence            JSONB NOT NULL DEFAULT '{}'::jsonb,
    FOREIGN KEY (tenant_id, meter_id) REFERENCES meters (tenant_id, meter_id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_anomalies_meter ON anomalies (tenant_id, meter_id);
CREATE INDEX IF NOT EXISTS idx_anomalies_run ON anomalies (analysis_run_id);
