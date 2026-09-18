package repository

type StoreRepository interface {
	Set(key string, value string) error
	Get(key string) (string, error)
	Update(key string, value string) error
	Delete(key string) error
	Exists(key string) bool
	Keys() []string
	Clear() error
}
