package db

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Subscription struct {
	UserID           string    `json:"user_id"`
	Status           string    `json:"status"` // active, canceled, past_due
	Plan             string    `json:"plan"`   // monthly, yearly
	CurrentPeriodEnd time.Time `json:"current_period_end"`
	MaxDevices       int       `json:"max_devices"`
	StripeCustomerID string    `json:"-"`
	StripeSubID      string    `json:"-"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Tool struct {
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	Description  string `json:"description"`
	PriceMonthly int    `json:"price_monthly"` // cents
	PriceYearly  int    `json:"price_yearly"`  // cents
}

type Entitlement struct {
	UserID     string    `json:"user_id"`
	ToolName   string    `json:"tool_name"`
	AcquiredAt time.Time `json:"acquired_at"`
}

type Activation struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	ToolName           string     `json:"tool_name"`
	MachineFingerprint string     `json:"machine_fingerprint"`
	Nickname           string     `json:"nickname"`
	Platform           string     `json:"platform"`
	Arch               string     `json:"arch"`
	ActivatedAt        time.Time  `json:"activated_at"`
	LastSeen           time.Time  `json:"last_seen"`
	RevokedAt          *time.Time `json:"revoked_at,omitempty"`
}
