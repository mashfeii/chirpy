package stringshelpers

import (
	"slices"
	"strings"
)

func CleanString(initial string, stop []string) string {
	splitted := strings.Split(initial, " ")

	for s := range splitted {
		if slices.Contains(stop, strings.ToLower(splitted[s])) {
			splitted[s] = "****"
		}
	}

	return strings.Join(splitted, " ")
}
