package types

import (
	"fmt"
	"strings"
)

func CanonicalRef(imageRef string) string {
	parts := strings.Split(imageRef, "/")

	// get the registry domain if there is more than two parts or if the first part contains a .
	registry := "kapsule.io"
	if len(parts) > 2 || strings.Contains(parts[0], ".") {
		registry = parts[0]
		parts = parts[1:]
	}

	// get the workspace
	workspace := "library"
	if len(parts) > 1 {
		workspace = parts[0]
		parts = parts[1:]
	}

	// set the tag
	tagParts := strings.Split(parts[0], ":")
	tag := "latest"
	if len(tagParts) > 1 && tagParts[1] != "" {
		tag = tagParts[1]
	}

	// get the image name
	image := tagParts[0]

	return fmt.Sprintf("%s/%s/%s:%s", registry, workspace, image, tag)
}
