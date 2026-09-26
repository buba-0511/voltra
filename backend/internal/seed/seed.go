package seed

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"energy-platform/internal/auth"
	"energy-platform/internal/config"
	sqlcgen "energy-platform/internal/db/sqlc"
)

const (
	DemoUserEmail    = "admin@energy-platform.local"
	DemoUserPassword = "demo1234"
)

// Run seeds the default tenant and demo user (idempotent via upsert), then
// loads readings.csv/events.csv into that tenant if it has no readings yet.
func Run(ctx context.Context, pool *pgxpool.Pool, dataDir string) error {
	q := sqlcgen.New(pool)

	tenant, err := q.CreateTenant(ctx, sqlcgen.CreateTenantParams{
		Slug: config.DefaultTenantSlug,
		Name: "Default",
	})
	if err != nil {
		return fmt.Errorf("create tenant: %w", err)
	}

	hash, err := auth.HashPassword(DemoUserPassword)
	if err != nil {
		return fmt.Errorf("hash demo password: %w", err)
	}
	if _, err := q.CreateUser(ctx, sqlcgen.CreateUserParams{
		TenantID:     tenant.ID,
		Email:        DemoUserEmail,
		PasswordHash: hash,
		Name:         "Demo Admin",
	}); err != nil {
		return fmt.Errorf("create demo user: %w", err)
	}

	count, err := q.CountReadings(ctx, tenant.ID)
	if err != nil {
		return fmt.Errorf("count readings: %w", err)
	}
	if count > 0 {
		return nil
	}

	readings, meterIDs, err := parseReadings(dataDir)
	if err != nil {
		return fmt.Errorf("parse readings.csv: %w", err)
	}
	events, err := parseEvents(dataDir)
	if err != nil {
		return fmt.Errorf("parse events.csv: %w", err)
	}

	sort.Strings(meterIDs)
	for i, id := range meterIDs {
		_, err := q.UpsertMeter(ctx, sqlcgen.UpsertMeterParams{
			TenantID: tenant.ID,
			MeterID:  id,
			Name:     fmt.Sprintf("Meter %s", id),
			Location: fmt.Sprintf("Building %d - Floor %d", (i/4)+1, (i%4)+1),
		})
		if err != nil {
			return fmt.Errorf("upsert meter %s: %w", id, err)
		}
	}

	for _, r := range readings {
		if err := q.InsertReading(ctx, sqlcgen.InsertReadingParams{
			TenantID:       tenant.ID,
			MeterID:        r.MeterID,
			Timestamp:      toTimestamptz(r.Timestamp),
			ConsumptionKwh: r.ConsumptionKWh,
			VoltageV:       r.VoltageV,
			CurrentA:       r.CurrentA,
			PowerFactor:    r.PowerFactor,
			Status:         r.Status,
		}); err != nil {
			return fmt.Errorf("insert reading: %w", err)
		}
	}

	for _, e := range events {
		if err := q.InsertEvent(ctx, sqlcgen.InsertEventParams{
			TenantID:    tenant.ID,
			MeterID:     e.MeterID,
			Timestamp:   toTimestamptz(e.Timestamp),
			Type:        e.Type,
			Description: e.Description,
		}); err != nil {
			return fmt.Errorf("insert event: %w", err)
		}
	}

	return nil
}

func toTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
