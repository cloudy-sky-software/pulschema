// Copyright 2022, Cloudy Sky Software.

package pkg

import (
	"fmt"
	"regexp"
	"strings"
)

var numbersRegexp = regexp.MustCompile("[0-9]+[_]*[a-zA-Z]+")

// ToSdkName converts a property or attribute name to the lowerCamelCase convention that
// is used in Pulumi schema's properties.
func ToSdkName(s string) string {
	if r := rune(s[0]); r >= 'A' && r <= 'Z' {
		s = strings.ToLower(string(r)) + s[1:]
	}
	return s
}

func snakeCaseToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts[1:] {
		parts[i+1] = ToPascalCase(p)
	}

	return strings.Join(parts, "")
}

func startsWithNumber(s string) bool {
	return numbersRegexp.Match([]byte(s))
}

// ToPascalCase converts a string to PascalCase.
func ToPascalCase(s string) string {
	return toCamelInitCase(s, true)
}

// moduleToPascalCase converts a module name to PascalCase.
func moduleToPascalCase(mod string) string {
	parts := strings.Split(mod, "/")

	for i, p := range parts {
		parts[i] = ToPascalCase(p)
	}

	return strings.Join(parts, "")
}

func toCamelInitCase(s string, initCase bool) string {
	if s == strings.ToUpper(s) {
		// lowercase the UPPER_SNAKE_CASE
		s = strings.ToLower(s)
	}

	s = strings.Trim(s, " ")
	n := ""
	capNext := initCase
	for _, v := range s {
		if v >= 'A' && v <= 'Z' {
			n += string(v)
		}
		if v >= '0' && v <= '9' {
			n += string(v)
		}
		if v >= 'a' && v <= 'z' {
			if capNext {
				n += strings.ToUpper(string(v))
			} else {
				n += string(v)
			}
		}
		if v == '_' || v == ' ' || v == '-' || v == '.' {
			capNext = true
		} else {
			capNext = false
		}
	}
	return n
}

func addAPINameOverride(sdkName, apiPropName string, overrides map[string]string) {
	if v, ok := overrides[sdkName]; ok && apiPropName != v {
		panic(fmt.Errorf("mapping for sdkName %s already exists and has a value %s but a new mapping with value %s was request", sdkName, v, apiPropName))
	}

	overrides[sdkName] = apiPropName
}
