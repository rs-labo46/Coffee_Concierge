package db

import (
	"coffee-spa/entity"
	"fmt"
)

// テーブル作成とindexの作成
func Migrate(d DB) error {
	if err := d.G.Exec(`CREATE EXTENSION IF NOT EXISTS pg_trgm;`).Error; err != nil {
		return fmt.Errorf("create extension pg_trgm: %w", err)
	}

	// テーブルを作成する
	if err := d.G.AutoMigrate(
		&entity.User{},
		&entity.EmailVerify{},
		&entity.PwReset{},
		&entity.RefreshToken{},
		&entity.Source{},
		&entity.Item{},
		&entity.AuditLog{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	if err := d.G.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_users_role'
			) THEN
				ALTER TABLE users
				ADD CONSTRAINT chk_users_role
				CHECK (role IN ('user', 'admin'));
			END IF;
		END$$;
	`).Error; err != nil {
		return fmt.Errorf("create chk_users_role: %w", err)
	}

	if err := d.G.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_items_kind'
			) THEN
				ALTER TABLE items
				ADD CONSTRAINT chk_items_kind
				CHECK (kind IN ('news', 'recipe', 'deal', 'shop'));
			END IF;
		END$$;
	`).Error; err != nil {
		return fmt.Errorf("create chk_items_kind: %w", err)
	}

	// itemsのindex
	if err := d.G.Exec(`
		CREATE INDEX IF NOT EXISTS idx_items_kind
		ON items (kind);
	`).Error; err != nil {
		return fmt.Errorf("create idx_items_kind: %w", err)
	}

	if err := d.G.Exec(`
		CREATE INDEX IF NOT EXISTS idx_items_source_id
		ON items (source_id);
	`).Error; err != nil {
		return fmt.Errorf("create idx_items_source_id: %w", err)
	}

	if err := d.G.Exec(`
		CREATE INDEX IF NOT EXISTS idx_items_published_at
		ON items (published_at DESC);
	`).Error; err != nil {
		return fmt.Errorf("create idx_items_published_at: %w", err)
	}

	if err := d.G.Exec(`
		CREATE INDEX IF NOT EXISTS idx_items_created_at
		ON items (created_at DESC);
	`).Error; err != nil {
		return fmt.Errorf("create idx_items_created_at: %w", err)
	}

	if err := d.G.Exec(`
		CREATE INDEX IF NOT EXISTS gin_items_title_trgm
		ON items
		USING gin (title gin_trgm_ops);
	`).Error; err != nil {
		return fmt.Errorf("create gin_items_title_trgm: %w", err)
	}

	if err := d.G.Exec(`
		CREATE INDEX IF NOT EXISTS gin_items_summary_trgm
		ON items
		USING gin (summary gin_trgm_ops);
	`).Error; err != nil {
		return fmt.Errorf("create gin_items_summary_trgm: %w", err)
	}

	if err := d.G.Exec(`
		CREATE INDEX IF NOT EXISTS gin_items_body_trgm
		ON items
		USING gin (body gin_trgm_ops);
	`).Error; err != nil {
		return fmt.Errorf("create gin_items_body_trgm: %w", err)
	}

	//refresh_tokensのindex
	if err := d.G.Exec(`
		CREATE INDEX IF NOT EXISTS idx_rt_user_id
		ON refresh_tokens (user_id);
	`).Error; err != nil {
		return fmt.Errorf("create idx_rt_user_id: %w", err)
	}

	if err := d.G.Exec(`
		CREATE INDEX IF NOT EXISTS idx_rt_family_id
		ON refresh_tokens (family_id);
	`).Error; err != nil {
		return fmt.Errorf("create idx_rt_family_id: %w", err)
	}

	//audit_logsのindex
	if err := d.G.Exec(`
		CREATE INDEX IF NOT EXISTS idx_audit_type
		ON audit_logs (type);
	`).Error; err != nil {
		return fmt.Errorf("create idx_audit_type: %w", err)
	}

	if err := d.G.Exec(`
		CREATE INDEX IF NOT EXISTS idx_audit_user_id
		ON audit_logs (user_id);
	`).Error; err != nil {
		return fmt.Errorf("create idx_audit_user_id: %w", err)
	}

	if err := d.G.Exec(`
		CREATE INDEX IF NOT EXISTS idx_audit_created_at
		ON audit_logs (created_at DESC);
	`).Error; err != nil {
		return fmt.Errorf("create idx_audit_created_at: %w", err)
	}

	return nil
}
