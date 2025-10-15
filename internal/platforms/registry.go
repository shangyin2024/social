package platforms

import (
	"fmt"
	"social/internal/types"
)

// Registry manages platform implementations
type Registry struct {
	platforms map[string]types.Platform
}

// NewRegistry creates a new platform registry
func NewRegistry() *Registry {
	registry := &Registry{
		platforms: make(map[string]types.Platform),
	}

	// Register all platforms
	registry.Register(NewXPlatform())
	registry.Register(NewYouTubePlatform())
	registry.Register(NewFacebookPlatform())
	registry.Register(NewTikTokPlatform())
	registry.Register(NewInstagramPlatform())

	return registry
}

// Register registers a platform implementation
func (r *Registry) Register(platform types.Platform) {
	r.platforms[platform.GetName()] = platform
}

// GetPlatform returns a platform implementation by name
func (r *Registry) GetPlatform(name string) (types.Platform, error) {
	platform, exists := r.platforms[name]
	if !exists {
		return nil, fmt.Errorf("platform %s not supported", name)
	}
	return platform, nil
}

// GetSupportedPlatforms returns a list of supported platform names
func (r *Registry) GetSupportedPlatforms() []string {
	var platforms []string
	for name := range r.platforms {
		platforms = append(platforms, name)
	}
	return platforms
}
