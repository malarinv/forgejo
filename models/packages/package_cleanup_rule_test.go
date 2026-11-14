// Copyright 2026 The Forgejo Authors.
// SPDX-License-Identifier: GPL-3.0-or-later

package packages

import (
	"testing"

	"forgejo.org/models/unittest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPackageCleanupRuleCRUD is a fork protecting test for the
// package-cleanup-rules-api feature: the REST CRUD handlers in
// routers/api/v1/packages/cleanup_rule.go are thin wrappers over these model
// functions, so the model round-trip is the layer the API contract rests on.
// The functions are upstream (v15 and v17 both have them), so the test is
// valid on either base.
func TestPackageCleanupRuleCRUD(t *testing.T) {
	require.NoError(t, unittest.PrepareTestDatabase())

	// Create
	pcr := &PackageCleanupRule{
		Enabled:       true,
		OwnerID:       1,
		Type:          "container",
		KeepCount:     5,
		KeepPattern:   "latest|development",
		RemoveDays:    7,
		RemovePattern: ".*",
	}
	created, err := InsertCleanupRule(t.Context(), pcr)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "container", string(created.Type))

	// Read
	got, err := GetCleanupRuleByID(t.Context(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), got.OwnerID)
	assert.Equal(t, 5, got.KeepCount)
	assert.Equal(t, "latest|development", got.KeepPattern)

	// CompiledPattern: the keep pattern compiles and matches the tag it must
	// retain (case-insensitively) but not an unrelated tag.
	require.NoError(t, got.CompiledPattern())
	require.NotNil(t, got.KeepPatternMatcher)
	assert.True(t, got.KeepPatternMatcher.MatchString("latest"))
	assert.True(t, got.KeepPatternMatcher.MatchString("Development"))
	assert.False(t, got.KeepPatternMatcher.MatchString("v1.2.3"))

	// Update
	got.KeepCount = 10
	got.Enabled = false
	require.NoError(t, UpdateCleanupRule(t.Context(), got))

	// List by owner
	rules, err := GetCleanupRulesByOwner(t.Context(), 1)
	require.NoError(t, err)
	var found *PackageCleanupRule
	for _, r := range rules {
		if r.ID == created.ID {
			found = r
		}
	}
	require.NotNil(t, found, "updated rule should be listed for owner 1")
	assert.False(t, found.Enabled)
	assert.Equal(t, 10, found.KeepCount)

	// Delete
	require.NoError(t, DeleteCleanupRuleByID(t.Context(), created.ID))
	_, err = GetCleanupRuleByID(t.Context(), created.ID)
	assert.Error(t, err, "deleted rule should no longer be found")
}
