package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
)

// Sign creates a signed license from payload using private key
func Sign(payload LicensePayload, privateKeyB64 string) (*SignedLicense, error) {
	// Decode private key
	privKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return nil, fmt.Errorf("decode private key: %w", err)
	}
	privateKey := ed25519.PrivateKey(privKeyBytes)

	// Create canonical JSON (deterministic)
	canonical, err := canonicalJSON(payload)
	if err != nil {
		return nil, fmt.Errorf("canonical json: %w", err)
	}

	// Sign
	signature := ed25519.Sign(privateKey, canonical)

	return &SignedLicense{
		Payload:   payload,
		Signature: base64.StdEncoding.EncodeToString(signature),
	}, nil
}

// Verify checks if a signed license is valid using public key
func Verify(license SignedLicense, publicKeyB64 string) error {
	// Decode public key
	pubKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return fmt.Errorf("decode public key: %w", err)
	}
	publicKey := ed25519.PublicKey(pubKeyBytes)

	// Canonical JSON
	canonical, err := canonicalJSON(license.Payload)
	if err != nil {
		return fmt.Errorf("canonical json: %w", err)
	}

	// Decode signature
	sig, err := base64.StdEncoding.DecodeString(license.Signature)
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}

	// Verify
	if !ed25519.Verify(publicKey, canonical, sig) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// canonicalJSON creates deterministic JSON output
// Keys are sorted alphabetically for consistent hashing
func canonicalJSON(v interface{}) ([]byte, error) {
	// Marshal to get map
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// Unmarshal to map
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	// Sort and remarshal
	return sortedMarshal(m)
}

func sortedMarshal(m map[string]interface{}) ([]byte, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build sorted map
	sorted := make([]byte, 0, 256)
	sorted = append(sorted, '{')

	for i, k := range keys {
		if i > 0 {
			sorted = append(sorted, ',')
		}

		// Add key
		keyBytes, _ := json.Marshal(k)
		sorted = append(sorted, keyBytes...)
		sorted = append(sorted, ':')

		// Add value
		valBytes, err := json.Marshal(m[k])
		if err != nil {
			return nil, err
		}
		sorted = append(sorted, valBytes...)
	}

	sorted = append(sorted, '}')
	return sorted, nil
}
