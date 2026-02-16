package vault

import (
	"fmt"
	"os/exec"
	"strings"
)

const serviceName = "tamr-vault"

// KeychainVault stores API keys in the macOS Keychain.
type KeychainVault struct{}

// NewKeychain creates a new macOS Keychain vault.
func NewKeychain() *KeychainVault {
	return &KeychainVault{}
}

// Set stores a key-value pair in the Keychain.
func (k *KeychainVault) Set(key, value string) error {
	// Delete existing entry first (ignore error if not found)
	_ = k.Delete(key)

	cmd := exec.Command("security", "add-generic-password",
		"-s", serviceName,
		"-a", key,
		"-w", value,
		"-U", // update if exists
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("keychain set: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// Get retrieves a value from the Keychain.
func (k *KeychainVault) Get(key string) (string, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-s", serviceName,
		"-a", key,
		"-w", // output password only
	)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return strings.TrimSpace(string(out)), nil
}

// Delete removes a key from the Keychain.
func (k *KeychainVault) Delete(key string) error {
	cmd := exec.Command("security", "delete-generic-password",
		"-s", serviceName,
		"-a", key,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("keychain delete: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// List returns all key names stored in the vault.
func (k *KeychainVault) List() ([]string, error) {
	cmd := exec.Command("security", "dump-keychain")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("keychain list: %w", err)
	}

	var keys []string
	lines := strings.Split(string(out), "\n")
	inTamrEntry := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, fmt.Sprintf(`"svce"<blob>="%s"`, serviceName)) {
			inTamrEntry = true
			continue
		}
		if inTamrEntry && strings.Contains(line, `"acct"<blob>="`) {
			// Extract account name (the key)
			start := strings.Index(line, `"acct"<blob>="`) + len(`"acct"<blob>="`)
			end := strings.LastIndex(line, `"`)
			if start > 0 && end > start {
				keys = append(keys, line[start:end])
			}
			inTamrEntry = false
		}
	}
	return keys, nil
}

// Mask returns a masked version of a value for display.
func Mask(value string) string {
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + "..." + value[len(value)-4:]
}
