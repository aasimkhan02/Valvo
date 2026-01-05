package admission

type Decision type

const (
	ALLOW Decision = iota
	SOFT_DENY
	HARD_DENY
)

type result struct {
	Decision Decision
	Remaining int64
	RetryAfterNanos int64
}