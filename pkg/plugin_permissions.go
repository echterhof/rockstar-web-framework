package pkg

import (
	"fmt"
	"sync"
)

// Standard permission names
const (
	PermissionDatabase   = "database"
	PermissionCache      = "cache"
	PermissionConfig     = "config"
	PermissionRouter     = "router"
	PermissionFileSystem = "filesystem"
	PermissionNetwork    = "network"
	PermissionExec       = "exec"
)

// permissionCheckerImpl is the default implementation of PermissionChecker
type permissionCheckerImpl struct {
	mu          sync.RWMutex
	permissions map[string]PluginPermissions // pluginName -> permissions
	logger      Logger
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker(logger Logger) PermissionChecker {
	return &permissionCheckerImpl{
		permissions: make(map[string]PluginPermissions),
		logger:      logger,
	}
}

// CheckPermission verifies if a plugin has a specific permission
func (p *permissionCheckerImpl) CheckPermission(pluginName string, permission string) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if permission == "" {
		return fmt.Errorf("permission cannot be empty")
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	perms, exists := p.permissions[pluginName]
	if !exists {
		// Log security violation
		if p.logger != nil {
			p.logger.Warn(fmt.Sprintf("Security violation: Plugin %s has no permissions assigned", pluginName))
		}
		return fmt.Errorf("plugin %s has no permissions assigned", pluginName)
	}

	// Check standard permissions
	hasPermission := false
	switch permission {
	case PermissionDatabase:
		hasPermission = perms.AllowDatabase
	case PermissionCache:
		hasPermission = perms.AllowCache
	case PermissionConfig:
		hasPermission = perms.AllowConfig
	case PermissionRouter:
		hasPermission = perms.AllowRouter
	case PermissionFileSystem:
		hasPermission = perms.AllowFileSystem
	case PermissionNetwork:
		hasPermission = perms.AllowNetwork
	case PermissionExec:
		hasPermission = perms.AllowExec
	default:
		// Check custom permissions
		if perms.CustomPermissions != nil {
			hasPermission = perms.CustomPermissions[permission]
		}
	}

	if !hasPermission {
		// Log security violation
		if p.logger != nil {
			p.logger.Warn(fmt.Sprintf("Security violation: Plugin %s denied access to %s", pluginName, permission))
		}
		return fmt.Errorf("plugin %s does not have permission for %s", pluginName, permission)
	}

	return nil
}

// GrantPermission grants a specific permission to a plugin
func (p *permissionCheckerImpl) GrantPermission(pluginName string, permission string) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if permission == "" {
		return fmt.Errorf("permission cannot be empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	perms, exists := p.permissions[pluginName]
	if !exists {
		perms = PluginPermissions{
			CustomPermissions: make(map[string]bool),
		}
	}

	// Grant standard permissions
	switch permission {
	case PermissionDatabase:
		perms.AllowDatabase = true
	case PermissionCache:
		perms.AllowCache = true
	case PermissionConfig:
		perms.AllowConfig = true
	case PermissionRouter:
		perms.AllowRouter = true
	case PermissionFileSystem:
		perms.AllowFileSystem = true
	case PermissionNetwork:
		perms.AllowNetwork = true
	case PermissionExec:
		perms.AllowExec = true
	default:
		// Grant custom permission
		if perms.CustomPermissions == nil {
			perms.CustomPermissions = make(map[string]bool)
		}
		perms.CustomPermissions[permission] = true
	}

	p.permissions[pluginName] = perms

	if p.logger != nil {
		p.logger.Info(fmt.Sprintf("Granted permission %s to plugin %s", permission, pluginName))
	}

	return nil
}

// RevokePermission revokes a specific permission from a plugin
func (p *permissionCheckerImpl) RevokePermission(pluginName string, permission string) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if permission == "" {
		return fmt.Errorf("permission cannot be empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	perms, exists := p.permissions[pluginName]
	if !exists {
		return fmt.Errorf("plugin %s has no permissions assigned", pluginName)
	}

	// Revoke standard permissions
	switch permission {
	case PermissionDatabase:
		perms.AllowDatabase = false
	case PermissionCache:
		perms.AllowCache = false
	case PermissionConfig:
		perms.AllowConfig = false
	case PermissionRouter:
		perms.AllowRouter = false
	case PermissionFileSystem:
		perms.AllowFileSystem = false
	case PermissionNetwork:
		perms.AllowNetwork = false
	case PermissionExec:
		perms.AllowExec = false
	default:
		// Revoke custom permission
		if perms.CustomPermissions != nil {
			delete(perms.CustomPermissions, permission)
		}
	}

	p.permissions[pluginName] = perms

	if p.logger != nil {
		p.logger.Info(fmt.Sprintf("Revoked permission %s from plugin %s", permission, pluginName))
	}

	return nil
}

// GetPermissions retrieves all permissions for a plugin
func (p *permissionCheckerImpl) GetPermissions(pluginName string) PluginPermissions {
	if pluginName == "" {
		return PluginPermissions{}
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	perms, exists := p.permissions[pluginName]
	if !exists {
		return PluginPermissions{
			CustomPermissions: make(map[string]bool),
		}
	}

	// Return a copy to prevent external modification
	result := PluginPermissions{
		AllowDatabase:   perms.AllowDatabase,
		AllowCache:      perms.AllowCache,
		AllowConfig:     perms.AllowConfig,
		AllowRouter:     perms.AllowRouter,
		AllowFileSystem: perms.AllowFileSystem,
		AllowNetwork:    perms.AllowNetwork,
		AllowExec:       perms.AllowExec,
	}

	if perms.CustomPermissions != nil {
		result.CustomPermissions = make(map[string]bool, len(perms.CustomPermissions))
		for k, v := range perms.CustomPermissions {
			result.CustomPermissions[k] = v
		}
	} else {
		result.CustomPermissions = make(map[string]bool)
	}

	return result
}

// SetPermissions sets all permissions for a plugin at once
func (p *permissionCheckerImpl) SetPermissions(pluginName string, permissions PluginPermissions) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Create a copy to prevent external modification
	perms := PluginPermissions{
		AllowDatabase:   permissions.AllowDatabase,
		AllowCache:      permissions.AllowCache,
		AllowConfig:     permissions.AllowConfig,
		AllowRouter:     permissions.AllowRouter,
		AllowFileSystem: permissions.AllowFileSystem,
		AllowNetwork:    permissions.AllowNetwork,
		AllowExec:       permissions.AllowExec,
	}

	if permissions.CustomPermissions != nil {
		perms.CustomPermissions = make(map[string]bool, len(permissions.CustomPermissions))
		for k, v := range permissions.CustomPermissions {
			perms.CustomPermissions[k] = v
		}
	} else {
		perms.CustomPermissions = make(map[string]bool)
	}

	p.permissions[pluginName] = perms

	if p.logger != nil {
		p.logger.Info(fmt.Sprintf("Set permissions for plugin %s", pluginName))
	}

	return nil
}

// RemovePlugin removes all permissions for a plugin
func (p *permissionCheckerImpl) RemovePlugin(pluginName string) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.permissions, pluginName)

	if p.logger != nil {
		p.logger.Info(fmt.Sprintf("Removed all permissions for plugin %s", pluginName))
	}

	return nil
}
