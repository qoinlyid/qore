package qore

// HttpValidatorErr defines http validator error wrap.
type HttpValidatorErr struct {
	ErrMandatory error
	ErrFormat    error

	// Internal used for auto put object to the pool.
	recycle func(*HttpValidatorErr)
}

// Error implements error interface.
func (e *HttpValidatorErr) Error() string {
	if e.ErrMandatory != nil {
		return e.ErrMandatory.Error()
	}
	if e.ErrFormat != nil {
		return e.ErrFormat.Error()
	}
	return ""
}

// Close releases instance to the pool.
func (e *HttpValidatorErr) Close() {
	if e.recycle != nil {
		e.recycle(e)
	}
}
