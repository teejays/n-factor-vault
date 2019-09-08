package util

import "fmt"

// ValidateUniqueStrings validates that strings in the array are unique
func ValidateUniqueStrings(strs []string) []error {
	var errs []error
	var cache = make(map[string]bool)
	var notUnique = make(map[string]bool)
	for _, s := range strs {
		if seen := cache[s]; seen {
			notUnique[s] = true
		}
		cache[s] = true
	}

	for _, s := range notUnique {
		errs = append(errs, fmt.Errorf("'%s' occurs more than once", s))
	}

	return errs

}
