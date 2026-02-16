package vault

// Vault provides secure storage for API keys.
type Vault interface {
	Set(key, value string) error
	Get(key string) (string, error)
	Delete(key string) error
	List() ([]string, error)
}
