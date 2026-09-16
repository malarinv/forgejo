// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package admin

import (
	"net/http"
	"net/url"

	"forgejo.org/models/auth"
	"forgejo.org/models/db"
	"forgejo.org/modules/json"
	api "forgejo.org/modules/structs"
	"forgejo.org/modules/web"
	"forgejo.org/services/auth/source/oauth2"
	"forgejo.org/services/context"
	"forgejo.org/services/convert"
)

// ListAuthSources lists all authentication sources
func ListAuthSources(ctx *context.APIContext) {
	// swagger:operation GET /admin/auth-sources admin adminListAuthSources
	// ---
	// summary: List all authentication sources
	// produces:
	// - application/json
	// responses:
	//   "200":
	//     "$ref": "#/responses/AuthenticationSourceList"
	//   "403":
	//     "$ref": "#/responses/forbidden"

	sources, err := db.Find[auth.Source](ctx, auth.FindSourcesOptions{})
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "db.Find", err)
		return
	}

	apiSources := make([]*api.AuthenticationSource, len(sources))
	for i, s := range sources {
		apiSources[i] = convert.ToAuthenticationSource(s)
	}

	ctx.JSON(http.StatusOK, apiSources)
}

// GetAuthSource gets an authentication source by ID
func GetAuthSource(ctx *context.APIContext) {
	// swagger:operation GET /admin/auth-sources/{id} admin adminGetAuthSource
	// ---
	// summary: Get an authentication source by ID
	// produces:
	// - application/json
	// parameters:
	// - name: id
	//   in: path
	//   description: ID of authentication source to get
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/AuthenticationSource"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"

	id := ctx.ParamsInt64(":id")
	source, err := auth.GetSourceByID(ctx, id)
	if err != nil {
		if auth.IsErrSourceNotExist(err) {
			ctx.NotFound()
		} else {
			ctx.Error(http.StatusInternalServerError, "auth.GetSourceByID", err)
		}
		return
	}

	ctx.JSON(http.StatusOK, convert.ToAuthenticationSource(source))
}

// EditOAuth2AuthSource updates an OAuth2 authentication source
func EditOAuth2AuthSource(ctx *context.APIContext) {
	// swagger:operation PATCH /admin/auth-sources/{id} admin adminEditAuthSource
	// ---
	// summary: Update an authentication source
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: id
	//   in: path
	//   description: ID of authentication source to update
	//   type: integer
	//   format: int64
	//   required: true
	// - name: body
	//   in: body
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/EditOAuth2AuthSourceOption"
	// responses:
	//   "200":
	//     "$ref": "#/responses/AuthenticationSource"
	//   "400":
	//     "$ref": "#/responses/error"
	//   "403":
	//     "$ref": "#/responses/forbidden"
	//   "404":
	//     "$ref": "#/responses/notFound"
	//   "409":
	//     "$ref": "#/responses/conflict"

	id := ctx.ParamsInt64(":id")
	source, err := auth.GetSourceByID(ctx, id)
	if err != nil {
		if auth.IsErrSourceNotExist(err) {
			ctx.NotFound()
		} else {
			ctx.Error(http.StatusInternalServerError, "auth.GetSourceByID", err)
		}
		return
	}

	if !source.IsOAuth2() {
		ctx.Error(http.StatusBadRequest, "", "Authentication source is not of OAuth2 type")
		return
	}

	form := web.GetForm(ctx).(*api.EditOAuth2AuthSourceOption)
	oauth2Config := source.Cfg.(*oauth2.Source)

	if form.Name != nil {
		source.Name = *form.Name
	}
	if form.IsActive != nil {
		source.IsActive = *form.IsActive
	}
	if form.ClientID != nil {
		oauth2Config.ClientID = *form.ClientID
	}
	if form.ClientSecret != nil {
		oauth2Config.ClientSecret = *form.ClientSecret
	}
	if form.OpenIDConnectAutoDiscoveryURL != nil {
		discoveryURL, err := url.Parse(*form.OpenIDConnectAutoDiscoveryURL)
		if err != nil || (discoveryURL.Scheme != "http" && discoveryURL.Scheme != "https") {
			ctx.Error(http.StatusBadRequest, "", "invalid OpenIDConnectAutoDiscoveryURL: must start with http:// or https://")
			return
		}
		oauth2Config.OpenIDConnectAutoDiscoveryURL = *form.OpenIDConnectAutoDiscoveryURL
	}
	if form.IconURL != nil {
		oauth2Config.IconURL = *form.IconURL
	}
	if form.Scopes != nil {
		oauth2Config.Scopes = *form.Scopes
	}
	if form.AttributeSSHPublicKey != nil {
		oauth2Config.AttributeSSHPublicKey = *form.AttributeSSHPublicKey
	}
	if form.RequiredClaimName != nil {
		oauth2Config.RequiredClaimName = *form.RequiredClaimName
	}
	if form.RequiredClaimValue != nil {
		oauth2Config.RequiredClaimValue = *form.RequiredClaimValue
	}
	if form.GroupClaimName != nil {
		oauth2Config.GroupClaimName = *form.GroupClaimName
	}
	if form.AdminGroup != nil {
		oauth2Config.AdminGroup = *form.AdminGroup
	}
	if form.RestrictedGroup != nil {
		oauth2Config.RestrictedGroup = *form.RestrictedGroup
	}
	if form.GroupTeamMapRemoval != nil {
		oauth2Config.GroupTeamMapRemoval = *form.GroupTeamMapRemoval
	}
	if form.AllowUsernameChange != nil {
		oauth2Config.AllowUsernameChange = *form.AllowUsernameChange
	}
	if form.QuotaGroupClaimName != nil {
		oauth2Config.QuotaGroupClaimName = *form.QuotaGroupClaimName
	}
	if form.QuotaGroupMap != nil {
		oauth2Config.QuotaGroupMap = *form.QuotaGroupMap
	}
	if form.QuotaGroupMapRemoval != nil {
		oauth2Config.QuotaGroupMapRemoval = *form.QuotaGroupMapRemoval
	}
	if form.SkipLocalTwoFA != nil {
		oauth2Config.SkipLocalTwoFA = *form.SkipLocalTwoFA
	}
	if form.GroupTeamMap != nil {
		rawBytes, err := json.Marshal(form.GroupTeamMap)
		if err != nil {
			ctx.Error(http.StatusBadRequest, "json.Marshal", err)
			return
		}
		oauth2Config.GroupTeamMap = string(rawBytes)
	}

	if err := auth.UpdateSource(ctx, source); err != nil {
		if auth.IsErrSourceAlreadyExist(err) {
			ctx.Error(http.StatusConflict, "", err)
		} else if oauth2.IsErrOpenIDConnectInitialize(err) {
			ctx.Error(http.StatusBadRequest, "", err)
		} else {
			ctx.Error(http.StatusInternalServerError, "auth.UpdateSource", err)
		}
		return
	}

	ctx.JSON(http.StatusOK, convert.ToAuthenticationSource(source))
}
