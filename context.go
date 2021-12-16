package stopctx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ErrStopped returns from ctx.Err() when stopctx.Context stopped.
var ErrStopped = errors.New("context stopped")

// Context represents a stop context.
//
// Stop context is used as a base context (instead of TODO or Background)
// and allows to determine which instance of stop context stopped.
type Context struct {
	stopCh  chan struct{}
	stopErr error
}

// New creates a new stop context instance with current nano timestamp as ID.
func New() (*Context, context.CancelFunc) {
	return NewWithID(time.Now().UnixNano())
}

// NewWithID creates a new stop context instance with provided ID.
func NewWithID(id interface{}) (*Context, context.CancelFunc) {
	ctx := &Context{
		stopCh:  make(chan struct{}),
		stopErr: fmt.Errorf("%w [id = %v]", ErrStopped, id),
	}

	var once sync.Once
	stopFn := func() {
		once.Do(func() {
			close(ctx.stopCh)
		})
	}

	return ctx, stopFn
}

// Deadline implements context.Context interface.
func (*Context) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done implements context.Context interface.
func (c *Context) Done() <-chan struct{} {
	return c.stopCh
}

// Err implements context.Context interface.
func (c *Context) Err() error {
	select {
	case <-c.stopCh:
		return c.stopErr
	default:
		return nil
	}
}

// Value implements context.Context interface.
func (*Context) Value(key interface{}) interface{} {
	return nil
}

// IsMyErr returns true if err contains text or wraps error raised by this stop context.
func (c *Context) IsMyErr(err error) bool {
	return errors.Is(err, c.stopErr) || strings.Contains(err.Error(), c.stopErr.Error())
}
