package license

import "time"

// LicensePayload is what gets signed and sent to client
type LicensePayload struct {
	UserID             string    `json:"user_id"`
	Entitlements       []string  `json:"entitlements"`
	MachineFingerprint string    `json:"machine_fingerprint"`
	IssuedAt           time.Time `json:"issued_at"`
	ExpiresAt          time.Time `json:"expires_at"`
	Plan               string    `json:"plan"`
}

// SignedLicense is the complete license with signature
type SignedLicense struct {
	Payload   LicensePayload `json:"payload"`
	Signature string         `json:"signature"` // base64 encoded ed25519 signature
}

// ActivationRequest from toolbox client
type ActivationRequest struct {
	MachineFingerprint string `json:"machine_fingerprint"`
	Nickname           string `json:"nickname"`
	Platform           string `json:"platform"`
	Arch               string `json:"arch"`
}

// ActivationResponse sent to toolbox client
type ActivationResponse struct {
	License     *SignedLicense `json:"license"`
	DownloadURL string         `json:"download_url,omitempty"`
}

// RenewRequest from toolbox client
type RenewRequest struct {
	MachineFingerprint string `json:"machine_fingerprint"`
}

// RenewResponse sent to toolbox client
type RenewResponse struct {
	Denied  bool           `json:"denied"`
	Reason  string         `json:"reason,omitempty"`
	License *SignedLicense `json:"license,omitempty"`
}
