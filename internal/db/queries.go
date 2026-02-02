package db

import (
	"database/sql"
)

// Users

func (db *DB) CreateUser(email, passwordHash string) (*User, error) {
	var user User
	err := db.QueryRow(`
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, created_at, updated_at
	`, email, passwordHash).Scan(&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) GetUserByEmail(email string) (*User, error) {
	var user User
	err := db.QueryRow(`
		SELECT id, email, password_hash, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) GetUserByID(id string) (*User, error) {
	var user User
	err := db.QueryRow(`
		SELECT id, email, password_hash, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Subscriptions

func (db *DB) GetSubscription(userID string) (*Subscription, error) {
	var sub Subscription
	err := db.QueryRow(`
		SELECT user_id, status, plan, current_period_end, max_devices,
		       COALESCE(stripe_customer_id, ''), COALESCE(stripe_subscription_id, ''),
		       created_at, updated_at
		FROM subscriptions WHERE user_id = $1
	`, userID).Scan(
		&sub.UserID, &sub.Status, &sub.Plan, &sub.CurrentPeriodEnd, &sub.MaxDevices,
		&sub.StripeCustomerID, &sub.StripeSubID, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (db *DB) UpsertSubscription(sub *Subscription) error {
	_, err := db.Exec(`
		INSERT INTO subscriptions (user_id, status, plan, current_period_end, max_devices, stripe_customer_id, stripe_subscription_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id) DO UPDATE SET
			status = EXCLUDED.status,
			plan = EXCLUDED.plan,
			current_period_end = EXCLUDED.current_period_end,
			max_devices = EXCLUDED.max_devices,
			stripe_customer_id = EXCLUDED.stripe_customer_id,
			stripe_subscription_id = EXCLUDED.stripe_subscription_id,
			updated_at = NOW()
	`, sub.UserID, sub.Status, sub.Plan, sub.CurrentPeriodEnd, sub.MaxDevices, sub.StripeCustomerID, sub.StripeSubID)
	return err
}

// Entitlements

func (db *DB) GetUserTools(userID string) ([]string, error) {
	rows, err := db.Query(`
		SELECT tool_name FROM entitlements WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []string
	for rows.Next() {
		var tool string
		if err := rows.Scan(&tool); err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}
	return tools, nil
}

func (db *DB) UserHasTool(userID, toolName string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM entitlements WHERE user_id = $1 AND tool_name = $2)
	`, userID, toolName).Scan(&exists)
	return exists, err
}

func (db *DB) AddEntitlement(userID, toolName string) error {
	_, err := db.Exec(`
		INSERT INTO entitlements (user_id, tool_name)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, toolName)
	return err
}

// Activations

func (db *DB) CountActiveDevices(userID string) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(DISTINCT machine_fingerprint)
		FROM activations
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID).Scan(&count)
	return count, err
}

func (db *DB) CreateActivation(userID, toolName, fingerprint, nickname, platform, arch string) error {
	_, err := db.Exec(`
		INSERT INTO activations (user_id, tool_name, machine_fingerprint, nickname, platform, arch)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, machine_fingerprint) DO UPDATE SET
			last_seen = NOW(),
			nickname = EXCLUDED.nickname,
			platform = EXCLUDED.platform,
			arch = EXCLUDED.arch
	`, userID, toolName, fingerprint, nickname, platform, arch)
	return err
}

func (db *DB) GetActivations(userID string) ([]Activation, error) {
	rows, err := db.Query(`
		SELECT id, user_id, tool_name, machine_fingerprint, nickname, platform, arch,
		       activated_at, last_seen, revoked_at
		FROM activations
		WHERE user_id = $1 AND revoked_at IS NULL
		ORDER BY last_seen DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activations []Activation
	for rows.Next() {
		var a Activation
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.ToolName, &a.MachineFingerprint, &a.Nickname,
			&a.Platform, &a.Arch, &a.ActivatedAt, &a.LastSeen, &a.RevokedAt,
		); err != nil {
			return nil, err
		}
		activations = append(activations, a)
	}
	return activations, nil
}

func (db *DB) DeactivateDevice(userID, activationID string) error {
	result, err := db.Exec(`
		UPDATE activations
		SET revoked_at = NOW()
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL
	`, activationID, userID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (db *DB) ActivationExists(userID, fingerprint string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM activations
			WHERE user_id = $1 AND machine_fingerprint = $2 AND revoked_at IS NULL
		)
	`, userID, fingerprint).Scan(&exists)
	return exists, err
}

func (db *DB) UpdateLastSeen(userID, fingerprint string) error {
	_, err := db.Exec(`
		UPDATE activations
		SET last_seen = NOW()
		WHERE user_id = $1 AND machine_fingerprint = $2 AND revoked_at IS NULL
	`, userID, fingerprint)
	return err
}

// Tools Catalog

func (db *DB) GetAllTools() ([]Tool, error) {
	rows, err := db.Query(`
		SELECT name, display_name, description, price_monthly, price_yearly
		FROM tools ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []Tool
	for rows.Next() {
		var t Tool
		if err := rows.Scan(&t.Name, &t.DisplayName, &t.Description, &t.PriceMonthly, &t.PriceYearly); err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	return tools, nil
}

// Signing Keys

func (db *DB) GetPrivateKey() (string, error) {
	var key string
	err := db.QueryRow(`SELECT private_key FROM signing_keys WHERE id = 1`).Scan(&key)
	return key, err
}

func (db *DB) GetPublicKey() (string, error) {
	var key string
	err := db.QueryRow(`SELECT public_key FROM signing_keys WHERE id = 1`).Scan(&key)
	return key, err
}
