package main

import (
	"regexp"
	"strings"
)

// Merge a name and domain to create a key in the redis
func GetKey(domain string, name string) string {
	if name == "*" {
		return "&." + domain
	} else if name == "@" {
		return domain
	} else {
		return name + "." + domain
	}
}

// Check if an string contains wildcard syntax characters
func IsWildcard(s string) bool { return strings.IndexAny(s, "*?") != -1 }

// WildcardToRegexp Compile wildcard to a regexp pattern
func WildcardToRegexp(s string) string {
	pattern := ""
	for {
		i := strings.IndexAny(s, "?*")
		if i == -1 {
			pattern += regexp.QuoteMeta(s)
			break
		} else {
			pattern += regexp.QuoteMeta(s[:i])
			if s[i] == '?' {
				pattern += "."
			} else {
				pattern += ".*"
			}
			s = s[i+1:]
		}
	}
	return pattern
}

// CreateRegexp create a Regexp from a series of parts
func CreateRegexp(parts []string) (*regexp.Regexp, error) {
	if len(parts) == 0 {
		return nil, InvalidArgs{}
	} else if len(parts) == 1 {
		return regexp.Compile(parts[0])
	} else {
		return regexp.Compile("(" + strings.Join(parts, ")|(") + ")")
	}
}

// FindString return index of a value in a collection or -1 if value is not in the collection
func FindString(collection []string, value string) int {
	for i, item := range collection {
		if item == value {
			return i
		}
	}

	return -1
}

// FindStringIf return index of a value in a collection or -1 if value is not in the collection
func FindStringIf(collection []string, pred func(string) bool) int {
	for i, item := range collection {
		if pred(item) {
			return i
		}
	}

	return -1
}

// Contains return `true` if value is in the collection and `false` otherwise
func Contains(collection []string, value string) bool { return FindString(collection, value) != -1 }
