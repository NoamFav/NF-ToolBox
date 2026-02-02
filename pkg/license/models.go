package license

import "time"

// LicensePayload is the signed data - MUST match server exactly
type LicensePayload struct {
	UserID             string    `json:"user_id"`
	Entitlements       []string  `json:"entitlements"`
	MachineFingerprint string    `json:"machine_fingerprint"`
	IssuedAt           time.Time `json:"issued_at"`
	ExpiresAt          time.Time `json:"expires_at"`
	Plan               string    `json:"plan"`
}

// SignedLicense is stored locally and verified on each run
type SignedLicense struct {
	Payload   LicensePayload `json:"payload"`
	Signature string         `json:"signature"`
}

// HasAccess checks if license includes a specific tool
func (l *SignedLicense) HasAccess(toolName string) bool {
	for _, t := range l.Payload.Entitlements {
		if t == toolName {
			return true
		}
	}
	return false
}

// IsExpired checks if license has expired (including grace period)
func (l *SignedLicense) IsExpired() bool {
	return time.Now().After(l.Payload.ExpiresAt)
}

// IsInGracePeriod checks if within 72-hour grace period
func (l *SignedLicense) IsInGracePeriod() bool {
	if !l.IsExpired() {
		return false
	}
	grace := l.Payload.ExpiresAt.Add(72 * time.Hour)
	return time.Now().Before(grace)
}

// DaysUntilExpiry returns days until license expires
func (l *SignedLicense) DaysUntilExpiry() int {
	duration := time.Until(l.Payload.ExpiresAt)
	return int(duration.Hours() / 24)
}

// NeedsRenewal checks if license is in renewal window
func (l *SignedLicense) NeedsRenewal() bool {
	days := l.DaysUntilExpiry()
	if l.Payload.Plan == "yearly" {
		return days <= 30
	}
	return days <= 5 // monthly
}
