package errx

func NotImplementedError() Exception {
	return NewException("not implemented")
}

func TODO(reason string) Exception {
	return NewExceptionWithCause(NotImplementedError(), reason)
}

func IllegalStateError() Exception {
	return NewException("ni shi zen me zuo dao de?")
}

func UnsupportedOperationException() Exception {
	return NewException("unsupported operation")
}
