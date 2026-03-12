package middleware

type AuthMethod string

const (
	AuthMethodSession     AuthMethod = "session"
	AuthMethodMatrixToken AuthMethod = "matrix_token"
)
