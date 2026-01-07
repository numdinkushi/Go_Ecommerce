package payment

import (
	"errors"
	"sync"
)

// ProviderRegistry manages payment provider registration and retrieval
type ProviderRegistry struct {
	providers map[string]PaymentProvider
	mu        sync.RWMutex
}

var globalRegistry = &ProviderRegistry{
	providers: make(map[string]PaymentProvider),
}

// RegisterProvider registers a payment provider
func RegisterProvider(provider PaymentProvider) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.providers[provider.GetProviderName()] = provider
}

// GetProvider retrieves a payment provider by name
func GetProvider(name string) (PaymentProvider, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	provider, exists := globalRegistry.providers[name]
	if !exists {
		return nil, errors.New("payment provider not found: " + name)
	}
	return provider, nil
}

// ListProviders returns a list of all registered provider names
func ListProviders() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	names := make([]string, 0, len(globalRegistry.providers))
	for name := range globalRegistry.providers {
		names = append(names, name)
	}
	return names
}

// IsProviderRegistered checks if a provider is registered
func IsProviderRegistered(name string) bool {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	_, exists := globalRegistry.providers[name]
	return exists
}
