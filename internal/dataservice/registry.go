package dataservice

import (
	"fmt"
	"sync"
)

// DataServiceRegistry manages multiple data sources
type DataServiceRegistry struct {
	mu         sync.RWMutex
	services   map[string]DataService
	defaultSvc DataService
}

var (
	registry *DataServiceRegistry
	once     sync.Once
)

func GetRegistry() *DataServiceRegistry {
	once.Do(func() {
		registry = &DataServiceRegistry{
			services: make(map[string]DataService),
		}
	})
	return registry
}

// Register adds a new data source
func (r *DataServiceRegistry) Register(name string, svc DataService) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[name] = svc
	// First registered becomes default if not set
	if r.defaultSvc == nil {
		r.defaultSvc = svc
	}
}

// SetDefault sets the default data source
func (r *DataServiceRegistry) SetDefault(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	svc, ok := r.services[name]
	if !ok {
		return fmt.Errorf("data service not found: %s", name)
	}
	r.defaultSvc = svc
	return nil
}

// GetDefault returns the default data service
func (r *DataServiceRegistry) GetDefault() DataService {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.defaultSvc
}

// CompositeDataService can try multiple sources or specific sources based on query
// This is an example of how to extend capabilities
type CompositeDataService struct {
	Primary   DataService
	Secondary DataService
}

// Implement DataService interface for CompositeDataService... (omitted for brevity, can be added if requested)
