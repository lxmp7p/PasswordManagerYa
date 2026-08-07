package repository

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepository_Close(t *testing.T) {
	repo := &Repository{}

	require.NotPanics(t, func() {
		repo.Close()
	})
}
