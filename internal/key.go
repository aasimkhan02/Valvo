package internal


type RateLimitKey struct {
	Tenant   string
	Region   string
	Resource string
}
