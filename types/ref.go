package types

import (
	"fmt"
	"strings"
)

func CanonicalRef(imageRef string) string {
	parts := strings.Split(imageRef, "/")
	tagParts := strings.Split(parts[len(parts)-1], ":")

	// set the tag
	tag := "latest"
	if len(tagParts) > 1 && tagParts[1] != "" {
		tag = tagParts[1]
	}

	// get the registry domain
	registry := "kapsule.io"
	if len(parts) > 2 {
		registry = parts[0]
	}

	// get the image name
	image := tagParts[0]

	// get the workspace
	workspace := "library"
	if len(parts) > 1 {
		workspace = parts[len(parts)-2]
	}

	return fmt.Sprintf("%s/%s/%s:%s", registry, workspace, image, tag)
}
