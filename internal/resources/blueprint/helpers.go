// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package blueprint

import (
	"context"
	"net/http"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

const (
	relationshipApps           = "apps"
	relationshipConfigurations = "configurations"
	relationshipPackages       = "packages"
	relationshipOrgDevices     = "orgDevices"
	relationshipUsers          = "users"
	relationshipUserGroups     = "userGroups"
)

// buildRelationshipData converts IDs into relationship linkage data.
func buildRelationshipData(resourceType string, ids []string) []client.Data {
	data := make([]client.Data, len(ids))
	for i, id := range ids {
		data[i] = client.Data{
			Type: resourceType,
			ID:   id,
		}
	}
	return data
}

// diffIDs compares current and desired IDs to determine adds and removals.
func diffIDs(current, desired []string) ([]string, []string) {
	currentMap := make(map[string]bool, len(current))
	for _, id := range current {
		currentMap[id] = true
	}

	desiredMap := make(map[string]bool, len(desired))
	for _, id := range desired {
		desiredMap[id] = true
	}

	var toAdd []string
	for _, id := range desired {
		if !currentMap[id] {
			toAdd = append(toAdd, id)
		}
	}

	var toRemove []string
	for _, id := range current {
		if !desiredMap[id] {
			toRemove = append(toRemove, id)
		}
	}

	return toAdd, toRemove
}

// readBlueprintRelationshipIDs fetches IDs for a relationship and converts them to a set.
func (r *BlueprintResource) readBlueprintRelationshipIDs(ctx context.Context, blueprintID, relationship string) ([]string, error) {
	return r.client.GetBlueprintRelationshipIDs(ctx, blueprintID, relationship)
}

// updateBlueprintRelationship applies add/remove operations for a relationship.
func (r *BlueprintResource) updateBlueprintRelationship(ctx context.Context, blueprintID, relationship, resourceType string, toAdd, toRemove []string) error {
	if len(toAdd) > 0 {
		if err := r.client.UpdateBlueprintRelationship(ctx, blueprintID, relationship, resourceType, http.MethodPost, toAdd); err != nil {
			return err
		}
	}

	if len(toRemove) > 0 {
		if err := r.client.UpdateBlueprintRelationship(ctx, blueprintID, relationship, resourceType, http.MethodDelete, toRemove); err != nil {
			return err
		}
	}

	return nil
}
