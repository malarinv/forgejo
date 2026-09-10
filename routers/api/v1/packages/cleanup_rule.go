// Copyright 2026 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package packages

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"

	packages_model "forgejo.org/models/packages"
	api "forgejo.org/modules/structs"
	"forgejo.org/modules/web"
	"forgejo.org/services/context"
	"forgejo.org/services/convert"
)

// ListCleanupRules gets all package cleanup rules of an owner
func ListCleanupRules(ctx *context.APIContext) {
	// swagger:operation GET /packages/{owner}/rules package listCleanupRules
	// ---
	// summary: Gets all package cleanup rules of an owner
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the packages
	//   type: string
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/PackageCleanupRuleList"
	//   "404":
	//     "$ref": "#/responses/notFound"

	pcrs, err := packages_model.GetCleanupRulesByOwner(ctx, ctx.Package.Owner.ID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "GetCleanupRulesByOwner", err)
		return
	}

	apiRules := make([]*api.PackageCleanupRule, 0, len(pcrs))
	for _, pcr := range pcrs {
		apiRules = append(apiRules, convert.ToPackageCleanupRule(pcr))
	}

	ctx.JSON(http.StatusOK, apiRules)
}

// CreateCleanupRule creates a new cleanup rule for an owner
func CreateCleanupRule(ctx *context.APIContext) {
	// swagger:operation POST /packages/{owner}/rules package createCleanupRule
	// ---
	// summary: Creates a package cleanup rule for an owner
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the packages
	//   type: string
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/CreatePackageCleanupRuleOption"
	// responses:
	//   "201":
	//     "$ref": "#/responses/PackageCleanupRule"
	//   "400":
	//     "$ref": "#/responses/error"
	//   "409":
	//     "$ref": "#/responses/error"

	form := web.GetForm(ctx).(*api.CreatePackageCleanupRuleOption)
	pType := packages_model.Type(form.Type)

	// Validate type exists
	validType := false
	for _, t := range packages_model.TypeList {
		if t == pType {
			validType = true
			break
		}
	}
	if !validType {
		ctx.Error(http.StatusBadRequest, "InvalidPackageType", fmt.Sprintf("invalid package type: %s", form.Type))
		return
	}

	// Validate regex patterns if provided
	if form.KeepPattern != "" {
		if _, err := regexp.Compile(fmt.Sprintf(`(?i)\A%s\z`, form.KeepPattern)); err != nil {
			ctx.Error(http.StatusBadRequest, "InvalidKeepPattern", err)
			return
		}
	}
	if form.RemovePattern != "" {
		if _, err := regexp.Compile(fmt.Sprintf(`(?i)\A%s\z`, form.RemovePattern)); err != nil {
			ctx.Error(http.StatusBadRequest, "InvalidRemovePattern", err)
			return
		}
	}

	// Check if rule already exists for this package type
	has, err := packages_model.HasOwnerCleanupRuleForPackageType(ctx, ctx.Package.Owner.ID, pType)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "HasOwnerCleanupRuleForPackageType", err)
		return
	}
	if has {
		ctx.Error(http.StatusConflict, "CleanupRuleExists", fmt.Sprintf("cleanup rule for type %s already exists", form.Type))
		return
	}

	pcr := &packages_model.PackageCleanupRule{
		Enabled:       form.Enabled,
		OwnerID:       ctx.Package.Owner.ID,
		Type:          pType,
		KeepCount:     form.KeepCount,
		KeepPattern:   form.KeepPattern,
		RemoveDays:    form.RemoveDays,
		RemovePattern: form.RemovePattern,
		MatchFullName: form.MatchFullName,
	}

	pcr, err = packages_model.InsertCleanupRule(ctx, pcr)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "InsertCleanupRule", err)
		return
	}

	ctx.JSON(http.StatusCreated, convert.ToPackageCleanupRule(pcr))
}

// GetCleanupRule gets a single cleanup rule
func GetCleanupRule(ctx *context.APIContext) {
	// swagger:operation GET /packages/{owner}/rules/{id} package getCleanupRule
	// ---
	// summary: Gets a package cleanup rule
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the packages
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: ID of the cleanup rule
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/PackageCleanupRule"
	//   "404":
	//     "$ref": "#/responses/notFound"

	pcr := getCleanupRuleFromContext(ctx)
	if pcr == nil {
		return
	}

	ctx.JSON(http.StatusOK, convert.ToPackageCleanupRule(pcr))
}

// EditCleanupRule updates an existing cleanup rule
func EditCleanupRule(ctx *context.APIContext) {
	// swagger:operation PATCH /packages/{owner}/rules/{id} package editCleanupRule
	// ---
	// summary: Updates a package cleanup rule
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the packages
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: ID of the cleanup rule
	//   type: integer
	//   format: int64
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/EditPackageCleanupRuleOption"
	// responses:
	//   "200":
	//     "$ref": "#/responses/PackageCleanupRule"
	//   "400":
	//     "$ref": "#/responses/error"
	//   "404":
	//     "$ref": "#/responses/notFound"

	pcr := getCleanupRuleFromContext(ctx)
	if pcr == nil {
		return
	}

	form := web.GetForm(ctx).(*api.EditPackageCleanupRuleOption)

	if form.Enabled != nil {
		pcr.Enabled = *form.Enabled
	}
	if form.KeepCount != nil {
		pcr.KeepCount = *form.KeepCount
	}
	if form.KeepPattern != nil {
		if *form.KeepPattern != "" {
			if _, err := regexp.Compile(fmt.Sprintf(`(?i)\A%s\z`, *form.KeepPattern)); err != nil {
				ctx.Error(http.StatusBadRequest, "InvalidKeepPattern", err)
				return
			}
		}
		pcr.KeepPattern = *form.KeepPattern
	}
	if form.RemoveDays != nil {
		pcr.RemoveDays = *form.RemoveDays
	}
	if form.RemovePattern != nil {
		if *form.RemovePattern != "" {
			if _, err := regexp.Compile(fmt.Sprintf(`(?i)\A%s\z`, *form.RemovePattern)); err != nil {
				ctx.Error(http.StatusBadRequest, "InvalidRemovePattern", err)
				return
			}
		}
		pcr.RemovePattern = *form.RemovePattern
	}
	if form.MatchFullName != nil {
		pcr.MatchFullName = *form.MatchFullName
	}

	if err := packages_model.UpdateCleanupRule(ctx, pcr); err != nil {
		ctx.Error(http.StatusInternalServerError, "UpdateCleanupRule", err)
		return
	}

	ctx.JSON(http.StatusOK, convert.ToPackageCleanupRule(pcr))
}

// DeleteCleanupRule deletes an existing cleanup rule
func DeleteCleanupRule(ctx *context.APIContext) {
	// swagger:operation DELETE /packages/{owner}/rules/{id} package deleteCleanupRule
	// ---
	// summary: Deletes a package cleanup rule
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the packages
	//   type: string
	//   required: true
	// - name: id
	//   in: path
	//   description: ID of the cleanup rule
	//   type: integer
	//   format: int64
	//   required: true
	// responses:
	//   "204":
	//     "$ref": "#/responses/empty"
	//   "404":
	//     "$ref": "#/responses/notFound"

	pcr := getCleanupRuleFromContext(ctx)
	if pcr == nil {
		return
	}

	if err := packages_model.DeleteCleanupRuleByID(ctx, pcr.ID); err != nil {
		ctx.Error(http.StatusInternalServerError, "DeleteCleanupRuleByID", err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func getCleanupRuleFromContext(ctx *context.APIContext) *packages_model.PackageCleanupRule {
	id := ctx.ParamsInt64(":id")
	pcr, err := packages_model.GetCleanupRuleByID(ctx, id)
	if err != nil {
		if errors.Is(err, packages_model.ErrPackageCleanupRuleNotExist) {
			ctx.NotFound()
		} else {
			ctx.Error(http.StatusInternalServerError, "GetCleanupRuleByID", err)
		}
		return nil
	}
	if pcr.OwnerID != ctx.Package.Owner.ID {
		ctx.NotFound()
		return nil
	}
	return pcr
}
