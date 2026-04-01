package checker

import (
	"fmt"
	"sync"
)

type FactoryFunc func(config map[string]interface{}) (Checker, error)

var (
	registryMu sync.RWMutex
	registry   = make(map[string]FactoryFunc)
)

func Register(name string, factory FactoryFunc) {
	registryMu.Lock()
	defer registryMu.Unlock()

	if name == "" {
		panic("checker: register called with empty name")
	}

	if factory == nil {
		panic("checker: register called with nil factory")
	}

	registry[name] = factory
}

func Create(name string, config map[string]interface{}) (Checker, error) {
	registryMu.RLock()
	factory, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("checker %q is not registered", name)
	}

	instance, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("create checker %q: %w", name, err)
	}

	return instance, nil
}
