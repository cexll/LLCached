package LLCached

type Cache interface {
	Get(key string) ([]byte, error)
}
