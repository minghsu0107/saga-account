package config

import "github.com/minghsu0107/saga-account/pkg"

type HTTPContextKey string

var (
	// JWTAuthHeader is the auth header containing customer ID
	JWTAuthHeader = "Authorization"
	// InvalidationTopic is the cache invalidation topic
	InvalidationTopic = pkg.Join("invalidate_cache:", "account")
	// CustomerKey is the key name for retrieving jwt-decoded customer id in a http request context
	CustomerKey HTTPContextKey = "customer_key"
)
