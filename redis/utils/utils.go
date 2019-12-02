package utils

// GenStateKey returns state key by address and task type
func GenStateKey(addr string, task string) string {
	return addr + "+" + task
}

// GenCacheKey returns cache key by address and task type
func GenCacheKey(addr string, task string) string {
	return addr + ":cache+" + task
}
