// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package structs

import (
	"time"
)

// AuthenticationSource represents an authentication source
type AuthenticationSource struct {
	ID            int64                       `json:"id"`
	Name          string                      `json:"name"`
	Type          string                      `json:"type"`
	IsActive      bool                        `json:"is_active"`
	IsSyncEnabled bool                        `json:"is_sync_enabled"`
	OAuth2Config  *OAuth2AuthenticationSource `json:"oauth2_config,omitempty"`
	// swagger:strfmt date-time
	CreatedAt time.Time `json:"created_at"`
	// swagger:strfmt date-time
	UpdatedAt time.Time `json:"updated_at"`
}

// OAuth2AuthenticationSource represents OAuth2/OIDC specific configuration
type OAuth2AuthenticationSource struct {
	Provider                      string                         `json:"provider"`
	ClientID                      string                         `json:"client_id"`
	OpenIDConnectAutoDiscoveryURL string                         `json:"open_id_connect_auto_discovery_url"`
	IconURL                       string                         `json:"icon_url"`
	Scopes                        []string                       `json:"scopes"`
	AttributeSSHPublicKey         string                         `json:"attribute_ssh_public_key"`
	RequiredClaimName             string                         `json:"required_claim_name"`
	RequiredClaimValue            string                         `json:"required_claim_value"`
	GroupClaimName                string                         `json:"group_claim_name"`
	AdminGroup                    string                         `json:"admin_group"`
	RestrictedGroup               string                         `json:"restricted_group"`
	GroupTeamMap                  map[string]map[string][]string `json:"group_team_map"`
	GroupTeamMapRemoval           bool                           `json:"group_team_map_removal"`
	AllowUsernameChange           bool                           `json:"allow_username_change"`
	QuotaGroupClaimName           string                         `json:"quota_group_claim_name"`
	QuotaGroupMap                 string                         `json:"quota_group_map"`
	QuotaGroupMapRemoval          bool                           `json:"quota_group_map_removal"`
	SkipLocalTwoFA                bool                           `json:"skip_local_two_fa"`
}

// EditOAuth2AuthSourceOption options when updating an OAuth2 authentication source
type EditOAuth2AuthSourceOption struct {
	Name                          *string                        `json:"name"`
	IsActive                      *bool                          `json:"is_active"`
	ClientID                      *string                        `json:"client_id"`
	ClientSecret                  *string                        `json:"client_secret"`
	OpenIDConnectAutoDiscoveryURL *string                        `json:"open_id_connect_auto_discovery_url"`
	IconURL                       *string                        `json:"icon_url"`
	Scopes                        *[]string                      `json:"scopes"`
	AttributeSSHPublicKey         *string                        `json:"attribute_ssh_public_key"`
	RequiredClaimName             *string                        `json:"required_claim_name"`
	RequiredClaimValue            *string                        `json:"required_claim_value"`
	GroupClaimName                *string                        `json:"group_claim_name"`
	AdminGroup                    *string                        `json:"admin_group"`
	RestrictedGroup               *string                        `json:"restricted_group"`
	GroupTeamMap                  map[string]map[string][]string `json:"group_team_map"`
	GroupTeamMapRemoval           *bool                          `json:"group_team_map_removal"`
	AllowUsernameChange           *bool                          `json:"allow_username_change"`
	QuotaGroupClaimName           *string                        `json:"quota_group_claim_name"`
	QuotaGroupMap                 *string                        `json:"quota_group_map"`
	QuotaGroupMapRemoval          *bool                          `json:"quota_group_map_removal"`
	SkipLocalTwoFA                *bool                          `json:"skip_local_two_fa"`
}
