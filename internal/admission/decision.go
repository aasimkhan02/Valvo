package admission

type Decision int

const (
	ALLOW Decision = iota
	SOFT_DENY
	HARD_DENY
)

type RateLimitResult struct {
	Decision        Decision
	RemainingTokens int64
	RetryAfterMs    int64
}
