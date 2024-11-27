package bindings

import (
	"fmt"

	"github.com/dop251/goja"

	typescriptengine "github.com/coder/guts/typescript-engine"
)

type Bindings struct {
	vm *goja.Runtime
}

func New() (*Bindings, error) {
	vm := goja.New()
	_, err := vm.RunString(string(typescriptengine.JSScript))
	if err != nil {
		return nil, fmt.Errorf("failed to run script: %v", err)
	}

	return &Bindings{
		vm: vm,
	}, nil
}

func (b *Bindings) f(name string) (goja.Callable, error) {
	f, ok := goja.AssertFunction(b.vm.Get("guts").ToObject(b.vm).Get(name))
	if !ok {
		return nil, fmt.Errorf("%q is not a function", name)
	}
	return f, nil
}
