package cache

import "k8s.io/client-go/tools/cache"

var (
	storesMap map[string]cache.Store = make(map[string]cache.Store)
)
