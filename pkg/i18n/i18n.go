package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// simple in-memory catalogs; can be replaced by YAML/JSON loader later
var catalogs = map[string]map[string]string{
	"en": {
		"field_required":          "field %s is required",
		"min_len":                 "field %s must be at least %s characters",
		"email":                   "field %s must be a valid email",
		"strong_password":         "field %s must be a stronger password (>=12, upper/lower/digit/special, no spaces)",
		"malformed_json_at":       "malformed JSON at position %s",
		"invalid_type_for_field":  "invalid type for field %s",
		"invalid_value_for_field": "invalid value for field %s",
		"empty_body":              "request body is empty",
	},
	"vi": {
		"field_required":          "trường %s là bắt buộc",
		"min_len":                 "trường %s phải có ít nhất %s ký tự",
		"email":                   "trường %s phải là email hợp lệ",
		"strong_password":         "trường %s cần mật khẩu mạnh (>=12, có chữ hoa/thường/số/ký tự đặc biệt, không khoảng trắng)",
		"malformed_json_at":       "JSON không hợp lệ tại vị trí %s",
		"invalid_type_for_field":  "sai kiểu dữ liệu cho trường %s",
		"invalid_value_for_field": "giá trị không hợp lệ cho trường %s",
		"empty_body":              "nội dung yêu cầu trống",
	},
}

var defaultLocale = "en"
var mu sync.RWMutex

// RegisterCatalog allows adding/updating a locale at runtime.
func RegisterCatalog(locale string, templates map[string]string) {
	mu.Lock()
	catalogs[locale] = templates
	mu.Unlock()
}

// SetDefaultLocale sets fallback locale.
func SetDefaultLocale(locale string) {
	if locale != "" {
		mu.Lock()
		defaultLocale = locale
		mu.Unlock()
	}
}

// Clear removes all in-memory catalogs (use when you want only external YAML catalogs).
func Clear() { catalogs = map[string]map[string]string{} }

// RenderLocale formats a message by key using the specified locale.
func RenderLocale(locale, key string, a ...any) string {
	mu.RLock()
	def := defaultLocale
	mu.RUnlock()
	if locale == "" {
		locale = def
	}
	locale = primaryLanguage(locale)
	mu.RLock()
	if m, ok := catalogs[locale][key]; ok {
		mu.RUnlock()
		if strings.Contains(m, "%s") {
			return fmt.Sprintf(m, a...)
		}
		return m
	}
	if m, ok := catalogs[def][key]; ok {
		mu.RUnlock()
		if strings.Contains(m, "%s") {
			return fmt.Sprintf(m, a...)
		}
		return m
	}
	mu.RUnlock()
	// log key-miss (stdout for now; in production inject logger)
	fmt.Printf("[i18n] missing key=%q locale=%q\n", key, locale)
	return key
}

// Init loads locale catalogs from a directory (YAML files) and sets default locale.
// Each file name (without extension) is treated as the locale code, e.g., en.yaml → "en".
// YAML structure: flat key: value mapping.
func Init(dir string, defLocale string) error {
	if defLocale != "" {
		defaultLocale = defLocale
	}
	if dir == "" {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		locale := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
		p := filepath.Join(dir, name)
		if err := loadYAMLFile(locale, p); err != nil {
			// keep going on other files
			continue
		}
	}
	return nil
}

func loadYAMLFile(locale, path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var m map[string]string
	if err := yaml.Unmarshal(b, &m); err != nil {
		return err
	}
	// normalize keys to lower-case
	norm := make(map[string]string, len(m))
	for k, v := range m {
		norm[strings.ToLower(k)] = v
	}
	RegisterCatalog(locale, norm)
	return nil
}

// DefaultLocale returns the current default locale.
func DefaultLocale() string { mu.RLock(); defer mu.RUnlock(); return defaultLocale }

// ParseAcceptLanguage extracts primary language code.
func ParseAcceptLanguage(acceptLang string) string {
	if acceptLang == "" {
		return defaultLocale
	}
	parts := strings.Split(acceptLang, ",")
	if len(parts) == 0 {
		return defaultLocale
	}
	token := strings.TrimSpace(parts[0])
	if i := strings.Index(token, ";"); i >= 0 {
		token = token[:i]
	}
	if token == "" {
		return defaultLocale
	}
	return primaryLanguage(token)
}

// primaryLanguage reduces region variants vi-VN -> vi
func primaryLanguage(tag string) string {
	if i := strings.Index(tag, "-"); i >= 0 {
		return tag[:i]
	}
	return tag
}
