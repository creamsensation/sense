package sense

type Interceptor interface {
	OnJson(fn func(req RequestContext, json any) any) Interceptor
	OnBool(fn func(req RequestContext, value bool) bool) Interceptor
	OnXml(fn func(req RequestContext, xml any) any) Interceptor
	OnText(fn func(req RequestContext, text string) string) Interceptor
	OnError(fn func(req RequestContext, err error) error) Interceptor
}

type interceptor struct {
	onBool  func(req RequestContext, value bool) bool
	onJson  func(req RequestContext, json any) any
	onXml   func(req RequestContext, xml any) any
	onText  func(req RequestContext, text string) string
	onError func(req RequestContext, err error) error
}

func (i *interceptor) OnJson(fn func(req RequestContext, json any) any) Interceptor {
	i.onJson = fn
	return i
}

func (i *interceptor) OnBool(fn func(req RequestContext, value bool) bool) Interceptor {
	i.onBool = fn
	return i
}

func (i *interceptor) OnXml(fn func(req RequestContext, xml any) any) Interceptor {
	i.onXml = fn
	return i
}

func (i *interceptor) OnText(fn func(req RequestContext, text string) string) Interceptor {
	i.onText = fn
	return i
}

func (i *interceptor) OnError(fn func(req RequestContext, err error) error) Interceptor {
	i.onError = fn
	return i
}
