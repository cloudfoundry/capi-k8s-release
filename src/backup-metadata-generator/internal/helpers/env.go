package helpers

import (
	"strings"
)

func EnvironToMap(env []string) map[string]string {
	mapEnv := map[string]string{}

	for _, e := range env {
		splits := strings.SplitN(e, "=", 2)
		mapEnv[splits[0]] = splits[1]
	}

	return mapEnv
}
