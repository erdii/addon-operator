package apihelpers

import (
	"fmt"
	"strings"
)

func SplitAddonMetadataVersionName(name string) (string, string, error) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf(`could not parse AddonMetadataVersion name into addon.Name + version. example: reference-addon.v0.0.1`)
	}

	return parts[0], parts[1], nil
}
