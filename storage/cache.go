package storage

type Cache interface {
	Get(key string) (string, error)
	Set(key string, value interface{})
	Del(key string) error
	GetCeq(key string) (int64, error)
}

var cache Cache = newRedis()

func Get(key string) (string, error) {
	return cache.Get(key)
}

func Set(key string, val interface{}) {
	cache.Set(key, val)
}
func Del(key string) error {
	return cache.Del(key)
}
func GetCeq(key string) (int64, error) {
	return cache.GetCeq(key)
}
