package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nickwells/english.mod/english"
	"github.com/nickwells/pusu.mod/pusu"
)

// namespaceMap records the allowed namespaces
type namespaceMap map[pusu.Namespace]bool

// namespaceRules records the constraints on the allowed Namespace
// values. Either the set of allowed namespaces can be non-empty or the slice
// of valid prefixes can be non-empty but not both (this is enforced at
// parameter setting time).
type namespaceRules struct {
	allowed  namespaceMap
	prefixes []string
}

// checkPrefixes returns a non-nil error if the set of prefixes is redundant
// - if one entry is itself a prefix of another.
func (rules namespaceRules) checkPrefixes() error {
	redundantCount := 0

	var errText string

	for i, pfx := range rules.prefixes {
	InnerLoop:
		for j, otherPfx := range rules.prefixes {
			if i == j {
				continue InnerLoop
			}
			if strings.HasPrefix(otherPfx, pfx) {
				redundantCount++
				errText = fmt.Sprintf(
					"%q has %q as a prefix of it",
					otherPfx, pfx)
			}
		}
	}

	if redundantCount > 0 {
		errText = "there are redundant prefixes: " + errText
		if redundantCount > 1 {
			errText += fmt.Sprintf(" and %d %s", redundantCount-1,
				english.Plural("other", redundantCount-1))
		}

		return errors.New(errText)
	}

	return nil
}

// isValid returns true if the given namespace is allowed. An empty namespace
// map allows any namespace.
func (rules namespaceRules) isValid(n pusu.Namespace) bool {
	if len(rules.allowed) > 0 {
		return rules.allowed[n]
	}

	if len(rules.prefixes) > 0 {
		for _, pfx := range rules.prefixes {
			if strings.HasPrefix(string(n), pfx) {
				return true
			}
		}

		return false
	}

	return true
}
