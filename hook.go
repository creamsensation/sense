package sense

type Hook interface {
	OnError(fn func(req RequestContext, err error)) Hook
}

type hook struct {
	onError func(req RequestContext, err error)
}

func (h *hook) OnError(fn func(req RequestContext, err error)) Hook {
	h.onError = fn
	return h
}
