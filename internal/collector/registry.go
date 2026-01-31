package collector

import (
	"fmt"
	"sync"
)

// Registry manages available collectors
type Registry struct {
	collectors map[string]Collector
	mu         sync.RWMutex
}

// NewRegistry creates a new collector registry
func NewRegistry() *Registry {
	return &Registry{
		collectors: make(map[string]Collector),
	}
}

// Register registers a new collector
func (r *Registry) Register(collector Collector) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	collectorType := collector.Type()
	if _, exists := r.collectors[collectorType]; exists {
		return fmt.Errorf("collector for type %s already registered", collectorType)
	}

	r.collectors[collectorType] = collector
	return nil
}

// Get retrieves a collector by router type
func (r *Registry) Get(routerType string) (Collector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	collector, exists := r.collectors[routerType]
	if !exists {
		return nil, fmt.Errorf("no collector found for router type: %s", routerType)
	}

	return collector, nil
}

// List returns all registered collector types
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.collectors))
	for t := range r.collectors {
		types = append(types, t)
	}

	return types
}

// Has checks if a collector exists for the given type
func (r *Registry) Has(routerType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.collectors[routerType]
	return exists
}
