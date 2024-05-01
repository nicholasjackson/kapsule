package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanonicalRefWithOnlyImage(t *testing.T) {
	imageRef := "test"
	expected := "kapsule/library/test:latest"

	actual := CanonicalRef(imageRef)
	require.Equal(t, expected, actual)
}

func TestCanonicalRefWithOnlyImageAndTag(t *testing.T) {
	imageRef := "test:v1"
	expected := "kapsule/library/test:v1"

	actual := CanonicalRef(imageRef)
	require.Equal(t, expected, actual)
}

func TestCanonicalRefWithNoRegistry(t *testing.T) {
	imageRef := "nicholasjackson/test:v1"
	expected := "kapsule/nicholasjackson/test:v1"

	actual := CanonicalRef(imageRef)
	require.Equal(t, expected, actual)
}

func TestCanonicalRefValid(t *testing.T) {
	imageRef := "docker.io/nicholasjackson/test:v1"
	expected := "docker.io/nicholasjackson/test:v1"

	actual := CanonicalRef(imageRef)
	require.Equal(t, expected, actual)
}
