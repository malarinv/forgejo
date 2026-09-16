// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package convert

import (
	auth_model "forgejo.org/models/auth"
	auth_module "forgejo.org/modules/auth"
	api "forgejo.org/modules/structs"
	"forgejo.org/services/auth/source/oauth2"
)

// ToAuthenticationSource converts an auth_model.Source to api.AuthenticationSource
func ToAuthenticationSource(s *auth_model.Source) *api.AuthenticationSource {
	res := &api.AuthenticationSource{
		ID:            s.ID,
		Name:          s.Name,
		Type:          s.TypeName(),
		IsActive:      s.IsActive,
		IsSyncEnabled: s.IsSyncEnabled,
		CreatedAt:     s.CreatedUnix.AsTime(),
		UpdatedAt:     s.UpdatedUnix.AsTime(),
	}

	if s.IsOAuth2() {
		if oauth2Source, ok := s.Cfg.(*oauth2.Source); ok {
			groupTeamMap, _ := auth_module.UnmarshalGroupTeamMapping(oauth2Source.GroupTeamMap)
			if groupTeamMap == nil {
				groupTeamMap = make(map[string]map[string][]string)
			}
			res.OAuth2Config = &api.OAuth2AuthenticationSource{
				Provider:                      oauth2Source.Provider,
				ClientID:                      oauth2Source.ClientID,
				OpenIDConnectAutoDiscoveryURL: oauth2Source.OpenIDConnectAutoDiscoveryURL,
				IconURL:                       oauth2Source.IconURL,
				Scopes:                        oauth2Source.Scopes,
				AttributeSSHPublicKey:         oauth2Source.AttributeSSHPublicKey,
				RequiredClaimName:             oauth2Source.RequiredClaimName,
				RequiredClaimValue:            oauth2Source.RequiredClaimValue,
				GroupClaimName:                oauth2Source.GroupClaimName,
				AdminGroup:                    oauth2Source.AdminGroup,
				RestrictedGroup:               oauth2Source.RestrictedGroup,
				GroupTeamMap:                  groupTeamMap,
				GroupTeamMapRemoval:           oauth2Source.GroupTeamMapRemoval,
				AllowUsernameChange:           oauth2Source.AllowUsernameChange,
				QuotaGroupClaimName:           oauth2Source.QuotaGroupClaimName,
				QuotaGroupMap:                 oauth2Source.QuotaGroupMap,
				QuotaGroupMapRemoval:          oauth2Source.QuotaGroupMapRemoval,
				SkipLocalTwoFA:                oauth2Source.SkipLocalTwoFA,
			}
		}
	}

	return res
}
