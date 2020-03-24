package cache

import "k8s.io/client-go/tools/cache"

var (
	storesMap = make(map[string]cache.Store)
)

// GetStore returns store by kind
// kind: ingress,endpoint,service,pod
func GetStore(kind string) cache.Store {
	return storesMap[kind]
}
