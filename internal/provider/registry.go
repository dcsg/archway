package provider

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Registry struct {
	mu        sync.RWMutex
	providers map[string]LanguageProvider
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]LanguageProvider),
	}
}

func (r *Registry) Register(language string, provider LanguageProvider) {
	if provider == nil {
		return
	}
	language = strings.ToLower(strings.TrimSpace(language))
	if language == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[language] = provider
}

func (r *Registry) Get(language string) (LanguageProvider, error) {
	language = strings.ToLower(strings.TrimSpace(language))
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[language]
	if !ok {
		return nil, fmt.Errorf("provider %q not registered", language)
	}
	return p, nil
}

func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	langs := make([]string, 0, len(r.providers))
	for lang := range r.providers {
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	return langs
}

var defaultRegistry = NewRegistry()

func Register(language string, provider LanguageProvider) {
	defaultRegistry.Register(language, provider)
}

func Get(language string) (LanguageProvider, error) {
	return defaultRegistry.Get(language)
}

func List() []string {
	return defaultRegistry.List()
}
