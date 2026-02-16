package vault

import "runtime"

// New returns the best available vault for the current platform.
// On macOS, it uses the system Keychain. On other platforms, it falls
// back to an AES-256-GCM encrypted file at ~/.config/palm/vault.enc.
func New() Vault {
	if runtime.GOOS == "darwin" {
		return NewKeychain()
	}
	return NewFileVault()
}
