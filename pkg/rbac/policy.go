package rbac

import (
    "strings"
    "sync"
)

// Policy defines an in-memory RBAC ruleset with thread-safe access.
// Rules are role -> set of permission patterns. Patterns can include:
//   - exact permission match (e.g., "admin:read")
//   - prefix wildcard (e.g., "admin:*")
//   - global wildcard "*" (all permissions)
type Policy struct {
    mu    sync.RWMutex
    rules map[string]map[string]struct{}
}

// NewPolicy creates a new Policy from role -> permissions list.
func NewPolicy(rules map[string][]string) *Policy {
    p := &Policy{rules: make(map[string]map[string]struct{}, len(rules))}
    for role, perms := range rules {
        set := make(map[string]struct{}, len(perms))
        for _, perm := range perms {
            set[strings.TrimSpace(perm)] = struct{}{}
        }
        p.rules[role] = set
    }
    return p
}

// DefaultRules returns a sensible default RBAC mapping.
func DefaultRules() map[string][]string {
    return map[string][]string{
        "admin":  {"*"},           // full access
        "user":   {"user:read"},   // basic read
        "viewer": {"user:read"},   // readonly
    }
}

// Replace sets the entire ruleset at once.
func (p *Policy) Replace(rules map[string][]string) {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.rules = make(map[string]map[string]struct{}, len(rules))
    for role, perms := range rules {
        set := make(map[string]struct{}, len(perms))
        for _, perm := range perms {
            set[strings.TrimSpace(perm)] = struct{}{}
        }
        p.rules[role] = set
    }
}

// SetRolePermissions sets/overwrites permissions for a role.
func (p *Policy) SetRolePermissions(role string, perms ...string) {
    p.mu.Lock()
    defer p.mu.Unlock()
    set := make(map[string]struct{}, len(perms))
    for _, perm := range perms {
        set[strings.TrimSpace(perm)] = struct{}{}
    }
    p.rules[role] = set
}

// HasPermission checks if a role grants a permission using wildcard matching.
func (p *Policy) HasPermission(role string, permission string) bool {
    p.mu.RLock()
    set, ok := p.rules[role]
    p.mu.RUnlock()
    if !ok {
        return false
    }
    // exact match
    if _, ok := set[permission]; ok {
        return true
    }
    // wildcard patterns
    for pattern := range set {
        if wildcardMatch(pattern, permission) {
            return true
        }
    }
    return false
}

// wildcardMatch supports "*" and suffix ":*" patterns.
func wildcardMatch(pattern, permission string) bool {
    if pattern == "*" {
        return true
    }
    if strings.HasSuffix(pattern, ":*") {
        prefix := strings.TrimSuffix(pattern, ":*")
        return strings.HasPrefix(permission, prefix+":") || permission == prefix
    }
    return false
}

// Global default policy instance
var defaultPolicy = NewPolicy(DefaultRules())

// Replace replaces the global policy ruleset.
func Replace(rules map[string][]string) { defaultPolicy.Replace(rules) }

// AddRolePermissions overwrites role permissions in the global policy.
func AddRolePermissions(role string, perms ...string) { defaultPolicy.SetRolePermissions(role, perms...) }

// HasPermission checks permission against the global policy.
func HasPermission(role string, permission string) bool { return defaultPolicy.HasPermission(role, permission) }

// Roles returns the list of role names in the current global policy.
func Roles() []string {
    defaultPolicy.mu.RLock()
    defer defaultPolicy.mu.RUnlock()
    out := make([]string, 0, len(defaultPolicy.rules))
    for role := range defaultPolicy.rules {
        out = append(out, role)
    }
    return out
}

// RoleExists returns true if the role is present in the current policy.
func RoleExists(role string) bool {
    defaultPolicy.mu.RLock()
    defer defaultPolicy.mu.RUnlock()
    _, ok := defaultPolicy.rules[role]
    return ok
}

