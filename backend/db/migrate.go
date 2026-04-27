package db

import (
	"coffee-spa/entity"
	"fmt"
)

func Migrate(d DB) error {
	if err := d.G.Exec(`CREATE EXTENSION IF NOT EXISTS pg_trgm;`).Error; err != nil {
		return fmt.Errorf("create extension pg_trgm: %w", err)
	}

	if err := d.G.AutoMigrate(
		&entity.User{},
		&entity.EmailVerify{},
		&entity.PwReset{},
		&entity.Rt{},
		&entity.AuditLog{},
		&entity.Source{},
		&entity.Item{},
		&entity.Bean{},
		&entity.Recipe{},
		&entity.Session{},
		&entity.Turn{},
		&entity.Pref{},
		&entity.Suggestion{},
		&entity.SavedSuggestion{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	constraints := []struct {
		name string
		sql  string
	}{
		{
			name: "chk_users_role",
			sql: `
				ALTER TABLE users
				ADD CONSTRAINT chk_users_role
				CHECK (role IN ('user', 'admin'));
			`,
		},
		{
			name: "chk_items_kind",
			sql: `
				ALTER TABLE items
				ADD CONSTRAINT chk_items_kind
				CHECK (kind IN ('news', 'recipe', 'deal', 'shop'));
			`,
		},
		{
			name: "chk_beans_roast",
			sql: `
				ALTER TABLE beans
				ADD CONSTRAINT chk_beans_roast
				CHECK (roast IN ('light', 'medium', 'dark'));
			`,
		},
		{
			name: "chk_beans_flavor",
			sql: `
				ALTER TABLE beans
				ADD CONSTRAINT chk_beans_flavor
				CHECK (flavor BETWEEN 1 AND 5);
			`,
		},
		{
			name: "chk_beans_acidity",
			sql: `
				ALTER TABLE beans
				ADD CONSTRAINT chk_beans_acidity
				CHECK (acidity BETWEEN 1 AND 5);
			`,
		},
		{
			name: "chk_beans_bitterness",
			sql: `
				ALTER TABLE beans
				ADD CONSTRAINT chk_beans_bitterness
				CHECK (bitterness BETWEEN 1 AND 5);
			`,
		},
		{
			name: "chk_beans_body",
			sql: `
				ALTER TABLE beans
				ADD CONSTRAINT chk_beans_body
				CHECK (body BETWEEN 1 AND 5);
			`,
		},
		{
			name: "chk_beans_aroma",
			sql: `
				ALTER TABLE beans
				ADD CONSTRAINT chk_beans_aroma
				CHECK (aroma BETWEEN 1 AND 5);
			`,
		},
		{
			name: "chk_recipes_method",
			sql: `
				ALTER TABLE recipes
				ADD CONSTRAINT chk_recipes_method
				CHECK (method IN ('drip', 'espresso', 'milk', 'iced'));
			`,
		},
		{
			name: "chk_recipes_temp_pref",
			sql: `
				ALTER TABLE recipes
				ADD CONSTRAINT chk_recipes_temp_pref
				CHECK (temp_pref IN ('hot', 'ice'));
			`,
		},
		{
			name: "chk_sessions_status",
			sql: `
				ALTER TABLE sessions
				ADD CONSTRAINT chk_sessions_status
				CHECK (status IN ('active', 'closed'));
			`,
		},
		{
			name: "chk_turns_role",
			sql: `
				ALTER TABLE turns
				ADD CONSTRAINT chk_turns_role
				CHECK (role IN ('user', 'assistant', 'system'));
			`,
		},
		{
			name: "chk_turns_kind",
			sql: `
				ALTER TABLE turns
				ADD CONSTRAINT chk_turns_kind
				CHECK (kind IN ('message', 'followup', 'notice'));
			`,
		},
		{
			name: "chk_prefs_flavor",
			sql: `
				ALTER TABLE prefs
				ADD CONSTRAINT chk_prefs_flavor
				CHECK (flavor BETWEEN 1 AND 5);
			`,
		},
		{
			name: "chk_prefs_acidity",
			sql: `
				ALTER TABLE prefs
				ADD CONSTRAINT chk_prefs_acidity
				CHECK (acidity BETWEEN 1 AND 5);
			`,
		},
		{
			name: "chk_prefs_bitterness",
			sql: `
				ALTER TABLE prefs
				ADD CONSTRAINT chk_prefs_bitterness
				CHECK (bitterness BETWEEN 1 AND 5);
			`,
		},
		{
			name: "chk_prefs_body",
			sql: `
				ALTER TABLE prefs
				ADD CONSTRAINT chk_prefs_body
				CHECK (body BETWEEN 1 AND 5);
			`,
		},
		{
			name: "chk_prefs_aroma",
			sql: `
				ALTER TABLE prefs
				ADD CONSTRAINT chk_prefs_aroma
				CHECK (aroma BETWEEN 1 AND 5);
			`,
		},
		{
			name: "chk_prefs_mood",
			sql: `
				ALTER TABLE prefs
				ADD CONSTRAINT chk_prefs_mood
				CHECK (mood IN ('morning', 'work', 'relax', 'night'));
			`,
		},
		{
			name: "chk_prefs_method",
			sql: `
				ALTER TABLE prefs
				ADD CONSTRAINT chk_prefs_method
				CHECK (method IN ('drip', 'espresso', 'milk', 'iced'));
			`,
		},
		{
			name: "chk_prefs_scene",
			sql: `
				ALTER TABLE prefs
				ADD CONSTRAINT chk_prefs_scene
				CHECK (scene IN ('work', 'break', 'after_meal', 'relax'));
			`,
		},
		{
			name: "chk_prefs_temp_pref",
			sql: `
				ALTER TABLE prefs
				ADD CONSTRAINT chk_prefs_temp_pref
				CHECK (temp_pref IN ('hot', 'ice'));
			`,
		},
	}

	for _, c := range constraints {
		q := fmt.Sprintf(`
			DO $$
			BEGIN
				IF NOT EXISTS (
					SELECT 1
					FROM pg_constraint
					WHERE conname = '%s'
				) THEN
					%s
				END IF;
			END$$;
		`, c.name, c.sql)

		if err := d.G.Exec(q).Error; err != nil {
			return fmt.Errorf("create constraint %s: %w", c.name, err)
		}
	}

	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_family_id ON refresh_tokens (family_id);`,

		`CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id_created_at ON audit_logs (user_id, created_at DESC);`,

		`CREATE INDEX IF NOT EXISTS idx_items_kind_published_at ON items (kind, published_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_items_source_id ON items (source_id);`,

		`CREATE INDEX IF NOT EXISTS idx_beans_active ON beans (active);`,
		`CREATE INDEX IF NOT EXISTS idx_beans_roast ON beans (roast);`,

		`CREATE INDEX IF NOT EXISTS idx_recipes_bean_id ON recipes (bean_id);`,
		`CREATE INDEX IF NOT EXISTS idx_recipes_method ON recipes (method);`,

		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id_created_at ON sessions (user_id, created_at DESC);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_session_key_hash_unique ON sessions (session_key_hash);`,

		`CREATE INDEX IF NOT EXISTS idx_turns_session_id_created_at ON turns (session_id, created_at);`,

		`CREATE UNIQUE INDEX IF NOT EXISTS idx_prefs_session_id_unique ON prefs (session_id);`,

		`CREATE INDEX IF NOT EXISTS idx_suggestions_session_id_rank ON suggestions (session_id, rank);`,

		`CREATE INDEX IF NOT EXISTS idx_saved_suggestions_user_id_created_at ON saved_suggestions (user_id, created_at DESC);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_saved_suggestions_user_suggestion_unique ON saved_suggestions (user_id, suggestion_id);`,
		`CREATE INDEX IF NOT EXISTS gin_items_title_trgm ON items USING gin (title gin_trgm_ops);`,
		`CREATE INDEX IF NOT EXISTS gin_items_summary_trgm ON items USING gin (summary gin_trgm_ops);`,
	}

	for _, q := range indexes {
		if err := d.G.Exec(q).Error; err != nil {
			return fmt.Errorf("create index: %w", err)
		}
	}

	return nil
}
