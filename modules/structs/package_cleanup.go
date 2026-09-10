// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package structs

import (
	"time"
)

// PackageCleanupRule represents a package cleanup rule
type PackageCleanupRule struct {
	ID            int64  `json:"id"`
	Enabled       bool   `json:"enabled"`
	Type          string `json:"type"`
	KeepCount     int    `json:"keep_count"`
	KeepPattern   string `json:"keep_pattern"`
	RemoveDays    int    `json:"remove_days"`
	RemovePattern string `json:"remove_pattern"`
	MatchFullName bool   `json:"match_full_name"`
	// swagger:strfmt date-time
	CreatedAt time.Time `json:"created_at"`
	// swagger:strfmt date-time
	UpdatedAt time.Time `json:"updated_at"`
}

// CreatePackageCleanupRuleOption options when creating a package cleanup rule
type CreatePackageCleanupRuleOption struct {
	Enabled       bool   `json:"enabled"`
	Type          string `json:"type" binding:"Required"`
	KeepCount     int    `json:"keep_count"`
	KeepPattern   string `json:"keep_pattern"`
	RemoveDays    int    `json:"remove_days"`
	RemovePattern string `json:"remove_pattern"`
	MatchFullName bool   `json:"match_full_name"`
}

// EditPackageCleanupRuleOption options when updating a package cleanup rule
type EditPackageCleanupRuleOption struct {
	Enabled       *bool   `json:"enabled"`
	KeepCount     *int    `json:"keep_count"`
	KeepPattern   *string `json:"keep_pattern"`
	RemoveDays    *int    `json:"remove_days"`
	RemovePattern *string `json:"remove_pattern"`
	MatchFullName *bool   `json:"match_full_name"`
}
