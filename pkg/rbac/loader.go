package rbac

import (
    "fmt"
    "os"
    "strings"
    "gopkg.in/yaml.v3"
)

// File format:
// roles:
//   admin:
//     - "*"
//   user:
//     - "user:read"
//   viewer:
//     - "user:read"

type filePolicy struct {
    Roles map[string][]string `yaml:"roles"`
}

// LoadFromYAML replaces the global policy with rules from a YAML file.
func LoadFromYAML(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("read policy file: %w", err)
    }
    var fp filePolicy
    if err := yaml.Unmarshal(data, &fp); err != nil {
        return fmt.Errorf("parse policy yaml: %w", err)
    }
    if len(fp.Roles) == 0 {
        return fmt.Errorf("empty roles in policy file")
    }
    // Validate roles and permissions non-empty and normalized
    for role, perms := range fp.Roles {
        if strings.TrimSpace(role) == "" {
            return fmt.Errorf("invalid role name (empty)")
        }
        for _, p := range perms {
            if strings.TrimSpace(p) == "" {
                return fmt.Errorf("role %q has empty permission entry", role)
            }
        }
    }
    Replace(fp.Roles)
    return nil
}

