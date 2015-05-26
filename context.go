package task

import "sync"

// Context is an interface for helping you share data between tasks.
type Context interface {
	// Get gets the value associated to the given key, or nil if there is none.
	Get(key string) interface{}

	// GetOk gets the value, and a boolean indicating the existence
	// of such key. When the key exists (value, true) will be returned, otherwise
	// (nil, false) will be returned.
	GetOk(key string) (interface{}, bool)

	// Set sets the given value in the given key.
	Set(key string, value interface{})

	// Delete deletes the given key.
	Delete(key string)
}

type context struct {
	sync.RWMutex
	keys map[string]interface{}
}

// NewContext creates a new Context.
func NewContext() Context {
	return &context{}
}

func (ctx *context) Get(key string) interface{} {
	ctx.RLock()
	value := ctx.keys[key]
	ctx.RUnlock()
	return value
}

func (ctx *context) GetOk(key string) (interface{}, bool) {
	ctx.RLock()
	value, ok := ctx.keys[key]
	ctx.RUnlock()
	return value, ok
}

func (ctx *context) Set(key string, value interface{}) {
	ctx.Lock()
	if ctx.keys == nil {
		ctx.keys = make(map[string]interface{})
	}
	ctx.keys[key] = value
	ctx.Unlock()
}

func (ctx *context) Delete(key string) {
	ctx.Lock()
	delete(ctx.keys, key)
	ctx.Unlock()
}
