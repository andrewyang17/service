package web

type Middleware func(Handler) Handler

func wrapMiddleware(mv []Middleware, handler Handler) Handler {
	for i := len(mv) - 1; i >= 0; i-- {
		h := mv[i]
		if h != nil {
			handler = h(handler)
		}
	}
	
	return handler
}