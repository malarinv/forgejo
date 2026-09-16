// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"testing"

	auth_model "forgejo.org/models/auth"
	api "forgejo.org/modules/structs"
	"forgejo.org/tests"

	"github.com/stretchr/testify/assert"
)

func TestAPIAdminAuthSources(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	adminToken := getUserToken(t, "user1", auth_model.AccessTokenScopeWriteAdmin)
	nonAdminToken := getUserToken(t, "user2", auth_model.AccessTokenScopeReadUser)

	// Create an OAuth2 source for testing
	source := addAuthSource(t, map[string]string{
		"type":            fmt.Sprintf("%d", auth_model.OAuth2),
		"name":            "oauth2-test-source",
		"is_active":       "on",
		"oauth2_provider": "gitlab",
		"oauth2_key":      "test-client-id",
		"oauth2_secret":   "super-secret-value",
	})

	t.Run("NonAdminForbidden", func(t *testing.T) {
		req := NewRequest(t, "GET", "/api/v1/admin/auth-sources").AddTokenAuth(nonAdminToken)
		MakeRequest(t, req, http.StatusForbidden)

		req = NewRequestf(t, "GET", "/api/v1/admin/auth-sources/%d", source.ID).AddTokenAuth(nonAdminToken)
		MakeRequest(t, req, http.StatusForbidden)
	})

	t.Run("ListAuthSources", func(t *testing.T) {
		req := NewRequest(t, "GET", "/api/v1/admin/auth-sources").AddTokenAuth(adminToken)
		resp := MakeRequest(t, req, http.StatusOK)

		var sources []*api.AuthenticationSource
		DecodeJSON(t, resp, &sources)
		assert.NotEmpty(t, sources)

		found := false
		for _, s := range sources {
			if s.ID == source.ID {
				found = true
				assert.Equal(t, "oauth2-test-source", s.Name)
				assert.Equal(t, "OAuth2", s.Type)
				assert.True(t, s.IsActive)
				assert.NotNil(t, s.OAuth2Config)
				assert.Equal(t, "test-client-id", s.OAuth2Config.ClientID)
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("GetAuthSource", func(t *testing.T) {
		req := NewRequestf(t, "GET", "/api/v1/admin/auth-sources/%d", source.ID).AddTokenAuth(adminToken)
		resp := MakeRequest(t, req, http.StatusOK)

		var res api.AuthenticationSource
		DecodeJSON(t, resp, &res)
		assert.Equal(t, source.ID, res.ID)
		assert.Equal(t, "oauth2-test-source", res.Name)
		assert.NotNil(t, res.OAuth2Config)
		assert.Equal(t, "test-client-id", res.OAuth2Config.ClientID)
		// Ensure secret is not leaked in the response
		assert.NotContains(t, resp.Body.String(), "super-secret-value")
	})

	t.Run("GetNotFound", func(t *testing.T) {
		req := NewRequest(t, "GET", "/api/v1/admin/auth-sources/99999").AddTokenAuth(adminToken)
		MakeRequest(t, req, http.StatusNotFound)
	})

	t.Run("EditOAuth2AuthSource", func(t *testing.T) {
		adminGroup := "core-team"
		groupRemoval := true
		groupTeamMap := map[string]map[string][]string{
			"core-team": {
				"whiteblossom": {"Owners", "platform-team", "security-team", "app-team"},
			},
			"platform-team": {
				"whiteblossom": {"platform-team", "app-team"},
			},
		}

		editOption := api.EditOAuth2AuthSourceOption{
			AdminGroup:          &adminGroup,
			GroupTeamMapRemoval: &groupRemoval,
			GroupTeamMap:        groupTeamMap,
		}

		req := NewRequestWithJSON(t, "PATCH", fmt.Sprintf("/api/v1/admin/auth-sources/%d", source.ID), &editOption).
			AddTokenAuth(adminToken)
		resp := MakeRequest(t, req, http.StatusOK)

		var res api.AuthenticationSource
		DecodeJSON(t, resp, &res)
		assert.Equal(t, source.ID, res.ID)
		assert.NotNil(t, res.OAuth2Config)
		assert.Equal(t, "core-team", res.OAuth2Config.AdminGroup)
		assert.True(t, res.OAuth2Config.GroupTeamMapRemoval)
		assert.Equal(t, groupTeamMap, res.OAuth2Config.GroupTeamMap)

		// Verify subsequent GET returns the persisted changes
		reqGet := NewRequestf(t, "GET", "/api/v1/admin/auth-sources/%d", source.ID).AddTokenAuth(adminToken)
		respGet := MakeRequest(t, reqGet, http.StatusOK)

		var resGet api.AuthenticationSource
		DecodeJSON(t, respGet, &resGet)
		assert.Equal(t, "core-team", resGet.OAuth2Config.AdminGroup)
		assert.True(t, resGet.OAuth2Config.GroupTeamMapRemoval)
		assert.Equal(t, groupTeamMap, resGet.OAuth2Config.GroupTeamMap)
	})
}
