package cache

import "k8s.io/client-go/tools/cache"

var (
	storesMap = make(map[string]cache.Store)
)

// IngressStore returns ingress store
func IngressStore() cache.Store {
	return storesMap["ingress"]
}

// EndpointStore returns endpoint store
func EndpointStore() cache.Store {
	return storesMap["endpoint"]
}

// ServiceStore returns service store
func ServiceStore() cache.Store {
	return storesMap["service"]
}
