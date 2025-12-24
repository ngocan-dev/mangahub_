package env

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Parse populates the fields of a struct based on environment variables.
// It honors the following struct tags:
//   - env: overrides the environment variable name (defaults to uppercase field name)
//   - envDefault: default value when env var is missing
//   - envRequired: when set to "true" or "required", the env var must be present
//
// Only exported struct fields are processed. Supported kinds are string, bool,
// all signed/unsigned integers, and float types.
func Parse(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("env: expected a non-nil pointer to struct")
	}
	elem := rv.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("env: expected a pointer to struct")
	}
	rt := elem.Type()

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}

		tagName := field.Tag.Get("env")
		if tagName == "" {
			tagName = strings.ToUpper(field.Name)
		}

		requiredTag := strings.ToLower(field.Tag.Get("envRequired"))
		required := requiredTag == "true" || requiredTag == "required"

		value, ok := os.LookupEnv(tagName)
		if !ok {
			if def := field.Tag.Get("envDefault"); def != "" {
				value = def
			}
		}

		if value == "" && required {
			return fmt.Errorf("env: required environment variable %q is missing", tagName)
		}

		if value == "" {
			continue
		}

		if err := setFieldValue(elem.Field(i), value); err != nil {
			return fmt.Errorf("env: parse %q: %w", tagName, err)
		}
	}

	return nil
}

func setFieldValue(v reflect.Value, raw string) error {
	if !v.CanSet() {
		return errors.New("field cannot be set")
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(raw)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		v.SetBool(parsed)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(raw, 10, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		parsed, err := strconv.ParseUint(raw, 10, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(raw, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetFloat(parsed)
	default:
		return fmt.Errorf("unsupported kind %s", v.Kind())
	}
	return nil
}
