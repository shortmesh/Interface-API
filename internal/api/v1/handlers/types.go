package handlers

type TxError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *TxError) Error() string {
	return e.Err.Error()
}
