package client

import (
	"context"
)

type Client interface {
	Exec(context.Context, bool)
	Done() <-chan struct{}
	Err() error
}
