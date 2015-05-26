package task

import (
	"testing"
	"time"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext()
	if ctx == nil {
		t.Fatalf("Not expecting nil")
	}
}

func TestContextGet(t *testing.T) {
	ctx := NewContext()
	if val := ctx.Get("foo"); val != nil {
		t.Fatalf("Not expecting ctx")
	}
}

func TestContextGetOk(t *testing.T) {
	ctx := NewContext()
	if val, ok := ctx.GetOk("foo"); ok || val != nil {
		t.Fatalf("Not expecting ctx")
	}
}

func TestContextDeleteGet(t *testing.T) {
	ctx := NewContext()
	ctx.Delete("foo")
	if val := ctx.Get("foo"); val != nil {
		t.Fatalf("Not expecting ctx")
	}
}

func TestContextDeleteGetOk(t *testing.T) {
	ctx := NewContext()
	ctx.Delete("foo")
	if val, ok := ctx.GetOk("foo"); ok || val != nil {
		t.Fatalf("Not expecting ctx")
	}
}

func TestContextSetGet(t *testing.T) {
	ctx := NewContext()
	ctx.Set("foo", 123)
	if val := ctx.Get("foo"); val != 123 {
		t.Fatalf("Expecting ctx")
	}
}

func TestContextSetGetOk(t *testing.T) {
	ctx := NewContext()
	ctx.Set("foo", 123)
	if val, ok := ctx.GetOk("foo"); !ok || val != 123 {
		t.Fatalf("Expecting ctx")
	}
}

func TestContextSetNilGetOk(t *testing.T) {
	ctx := NewContext()
	ctx.Set("foo", nil)
	if val, ok := ctx.GetOk("foo"); !ok || val != nil {
		t.Fatalf("Expecting ctx")
	}
}

func TestContextReplaceGetOk(t *testing.T) {
	ctx := NewContext()
	ctx.Set("foo", 123)
	ctx.Set("foo", "bar")
	if val, ok := ctx.GetOk("foo"); !ok || val != "bar" {
		t.Fatalf("Expecting ctx")
	}
}

func TestContextSetDeleteGetOk(t *testing.T) {
	ctx := NewContext()
	ctx.Set("foo", 123)
	ctx.Delete("foo")
	if val, ok := ctx.GetOk("foo"); ok || val != nil {
		t.Fatalf("Not expecting ctx")
	}
}

func TestContextSetNilDeleteGetOk(t *testing.T) {
	ctx := NewContext()
	ctx.Set("foo", nil)
	ctx.Delete("foo")
	if val, ok := ctx.GetOk("foo"); ok || val != nil {
		t.Fatalf("Not expecting ctx")
	}
}

func TestContextSetMultipleGetOk(t *testing.T) {
	ctx := NewContext()
	ctx.Set("foo", 123)
	ctx.Set("bar", true)
	ctx.Set("dead", "beef")
	if val, ok := ctx.GetOk("foo"); !ok || val != 123 {
		t.Fatalf("Expecting ctx")
	}
	if val, ok := ctx.GetOk("bar"); !ok || val != true {
		t.Fatalf("Expecting ctx")
	}
	if val, ok := ctx.GetOk("dead"); !ok || val != "beef" {
		t.Fatalf("Expecting ctx")
	}
}

func TestContextSetComplexGetOk(t *testing.T) {
	var bar = struct {
		str  string
		f    float32
		time time.Time
	}{"bar", 3.14, time.Now()}
	ctx := NewContext()
	ctx.Set("foo", &bar)
	val, ok := ctx.GetOk("foo")
	if !ok || val != &bar {
		t.Fatalf("Expecting ctx")
	}
}
