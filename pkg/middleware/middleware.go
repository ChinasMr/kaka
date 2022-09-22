package middleware

import "context"

type Handler func(ctx context.Context, req interface{}) (interface{}, error)

type Middleware func(handler Handler) Handler
