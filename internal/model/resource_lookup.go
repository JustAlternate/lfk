package model

import (
	"sort"
	"strings"
)

// FlattenedResourceTypes returns all resource types as a flat Item list with no filtering.
func FlattenedResourceTypes() []Item {
	return FlattenedResourceTypesFiltered(nil)
}

// FlattenedResourceTypesFiltered returns resource types as a flat list, optionally excluding
// CRD-dependent categories when the cluster doesn't have those CRDs installed.
// Core categories (Cluster, Workloads, Config, Networking, Storage, Access Control, Helm, API and CRDs) are
// always shown. Other categories are only shown if their API group name appears in availableGroups.
// Individual resource types marked with RequiresCRD are also filtered out unless their
// API group/resource appears in availableGroups. When availableGroups is nil, CRD-dependent
// entries are hidden (safe default before discovery completes).
func FlattenedResourceTypesFiltered(availableGroups map[string]bool) []Item {
	var items []Item
	// Add Cluster Dashboard and Monitoring as a dedicated "Dashboards" group.
	items = append(items, Item{
		Name:     "Cluster",
		Kind:     "__overview__",
		Extra:    "__overview__",
		Category: "Dashboards",
		Icon:     "◎",
	})
	items = append(items, Item{
		Name:     "Monitoring",
		Kind:     "__monitoring__",
		Extra:    "__monitoring__",
		Category: "Dashboards",
		Icon:     "⊙",
	})
	for _, cat := range TopLevelResourceTypes() {
		if !coreCategories[cat.Name] {
			// CRD-based category: only show if the API group is detected.
			if availableGroups == nil || !availableGroups[cat.Name] {
				continue
			}
		}
		for _, rt := range cat.Types {
			if rt.RequiresCRD && (availableGroups == nil || !availableGroups[rt.APIGroup+"/"+rt.Resource]) {
				continue
			}
			items = append(items, Item{
				Name:       rt.DisplayName,
				Kind:       rt.Kind,
				Extra:      rt.ResourceRef(),
				Category:   cat.Name,
				Icon:       rt.Icon,
				Deprecated: rt.Deprecated,
			})
		}
	}
	return items
}

// MergeWithCRDs returns the flattened resource type list with discovered CRDs appended
// as additional categories grouped by API group. CRDs that match a built-in resource
// type (same group + resource) are filtered out to avoid duplicates.
func MergeWithCRDs(discovered []ResourceTypeEntry) []Item {
	// Build the set of all API groups and specific resources present as discovered CRDs.
	availableGroups := make(map[string]bool, len(discovered)*2)
	for _, crd := range discovered {
		availableGroups[crd.APIGroup] = true
		availableGroups[crd.APIGroup+"/"+crd.Resource] = true
	}

	// Helm always shows (uses helm binary, not CRDs).
	// No special handling needed — Helm is a core category.

	items := FlattenedResourceTypesFiltered(availableGroups)
	if len(discovered) == 0 {
		return items
	}

	// Build a set of built-in resource identifiers (group/resource) to filter duplicates.
	builtIn := make(map[string]bool)
	for _, cat := range TopLevelResourceTypes() {
		for _, rt := range cat.Types {
			builtIn[rt.APIGroup+"/"+rt.Resource] = true
		}
	}

	// Build a map of discovered CRD versions so built-in entries can be updated
	// to match the version the cluster actually serves.
	discoveredVersion := make(map[string]string, len(discovered))
	for _, crd := range discovered {
		discoveredVersion[crd.APIGroup+"/"+crd.Resource] = crd.APIVersion
	}

	// Update built-in items whose API version differs from what the cluster serves.
	// This prevents stale hardcoded versions from causing "resource not found" errors.
	for i := range items {
		key := items[i].Extra
		if key == "" {
			continue
		}
		// Extra format is "group/version/resource" — extract group and resource.
		parts := strings.SplitN(key, "/", 3)
		if len(parts) != 3 {
			continue
		}
		groupResource := parts[0] + "/" + parts[2]
		if ver, ok := discoveredVersion[groupResource]; ok && ver != parts[1] {
			items[i].Extra = parts[0] + "/" + ver + "/" + parts[2]
		}
	}

	// Build builtInCategoryForGroup dynamically from TopLevelResourceTypes.
	// Maps API groups to their category name so discovered CRDs from the same group
	// get inserted alongside built-in entries.
	builtInCategoryForGroup := make(map[string]string)
	for _, cat := range TopLevelResourceTypes() {
		if coreCategories[cat.Name] {
			continue // Don't map core resource groups
		}
		for _, rt := range cat.Types {
			builtInCategoryForGroup[rt.APIGroup] = cat.Name
		}
	}

	// Group CRDs by API group, filtering out built-in duplicates.
	grouped := make(map[string][]ResourceTypeEntry)
	var groupOrder []string
	for _, crd := range discovered {
		key := crd.APIGroup + "/" + crd.Resource
		if builtIn[key] {
			continue
		}
		if _, seen := grouped[crd.APIGroup]; !seen {
			groupOrder = append(groupOrder, crd.APIGroup)
		}
		grouped[crd.APIGroup] = append(grouped[crd.APIGroup], crd)
	}

	// Separate groups into pinned (user-configured) and unpinned, preserving order.
	pinnedSet := make(map[string]bool, len(PinnedGroups))
	for _, g := range PinnedGroups {
		pinnedSet[g] = true
	}

	var pinnedOrder, unpinnedOrder []string
	for _, group := range groupOrder {
		if pinnedSet[group] {
			pinnedOrder = append(pinnedOrder, group)
		} else {
			unpinnedOrder = append(unpinnedOrder, group)
		}
	}

	// Sort pinnedOrder to match the user's configured order in PinnedGroups.
	pinnedOrderMap := make(map[string]int, len(PinnedGroups))
	for i, g := range PinnedGroups {
		pinnedOrderMap[g] = i
	}
	sort.SliceStable(pinnedOrder, func(i, j int) bool {
		return pinnedOrderMap[pinnedOrder[i]] < pinnedOrderMap[pinnedOrder[j]]
	})

	// Process groups: pinned first, then unpinned.
	orderedGroups := make([]string, 0, len(pinnedOrder)+len(unpinnedOrder))
	orderedGroups = append(orderedGroups, pinnedOrder...)
	orderedGroups = append(orderedGroups, unpinnedOrder...)

	// Build items for each discovered group (non-duplicate CRDs only).
	for _, group := range orderedGroups {
		categoryName, isBuiltInGroup := builtInCategoryForGroup[group]
		if !isBuiltInGroup {
			categoryName = group
		}

		crdItems := make([]Item, 0, len(grouped[group]))
		for _, rt := range grouped[group] {
			crdItems = append(crdItems, Item{
				Name:       rt.DisplayName,
				Kind:       rt.Kind,
				Extra:      rt.ResourceRef(),
				Category:   categoryName,
				Icon:       rt.Icon,
				Deprecated: rt.Deprecated,
			})
		}

		if isBuiltInGroup {
			// Merge extra discovered CRDs into their built-in category.
			insertIdx := -1
			for i, it := range items {
				if it.Category == categoryName {
					insertIdx = i
				}
			}
			if insertIdx >= 0 {
				tail := make([]Item, len(items[insertIdx+1:]))
				copy(tail, items[insertIdx+1:])
				items = append(items[:insertIdx+1], crdItems...)
				items = append(items, tail...)
				continue
			}
		}

		// Append non-built-in discovered groups at the end (sorted below).
		items = append(items, crdItems...)
	}

	// Sort all non-core CRD categories alphabetically by category name,
	// with pinned groups appearing first (in user-configured order).
	// Core categories retain their fixed position at the top.
	var coreItems, pinnedItems, crdItemsList []Item
	for _, it := range items {
		switch {
		case coreCategories[it.Category] || it.Category == "":
			coreItems = append(coreItems, it)
		case pinnedSet[it.Category]:
			pinnedItems = append(pinnedItems, it)
		default:
			crdItemsList = append(crdItemsList, it)
		}
	}

	// Sort pinned items by the user's configured pinned group order.
	sort.SliceStable(pinnedItems, func(i, j int) bool {
		return pinnedOrderMap[pinnedItems[i].Category] < pinnedOrderMap[pinnedItems[j].Category]
	})

	// Sort CRD items alphabetically by category name.
	sort.SliceStable(crdItemsList, func(i, j int) bool {
		return crdItemsList[i].Category < crdItemsList[j].Category
	})

	items = make([]Item, 0, len(coreItems)+len(pinnedItems)+len(crdItemsList))
	items = append(items, coreItems...)
	items = append(items, pinnedItems...)
	items = append(items, crdItemsList...)

	return items
}

// ResourceRef returns the "group/version/resource" reference string.
func (r ResourceTypeEntry) ResourceRef() string {
	return r.APIGroup + "/" + r.APIVersion + "/" + r.Resource
}

// FindResourceTypeByKind searches for a ResourceTypeEntry matching the given kind
// across all built-in types and the provided CRDs.
func FindResourceTypeByKind(kind string, crds []ResourceTypeEntry) (ResourceTypeEntry, bool) {
	// Build lookup of discovered CRDs by group/resource for version override and enrichment.
	discoveredByGR := make(map[string]*ResourceTypeEntry, len(crds))
	for i := range crds {
		key := crds[i].APIGroup + "/" + crds[i].Resource
		discoveredByGR[key] = &crds[i]
	}

	for _, cat := range TopLevelResourceTypes() {
		for _, rt := range cat.Types {
			if rt.Kind == kind {
				// Override version and enrich with PrinterColumns from discovered CRDs.
				grKey := rt.APIGroup + "/" + rt.Resource
				if crd, ok := discoveredByGR[grKey]; ok {
					rt.APIVersion = crd.APIVersion
					if len(crd.PrinterColumns) > 0 {
						rt.PrinterColumns = crd.PrinterColumns
					}
				}
				return rt, true
			}
		}
	}
	for _, crd := range crds {
		if crd.Kind == kind {
			return crd, true
		}
	}
	return ResourceTypeEntry{}, false
}

// FindResourceType looks up a ResourceTypeEntry by its ref string in built-in types.
func FindResourceType(ref string) (ResourceTypeEntry, bool) {
	return FindResourceTypeIn(ref, nil)
}

// FindResourceTypeIn looks up a ResourceTypeEntry by its ref string, searching both
// built-in types and the provided additional entries (e.g., discovered CRDs).
// The ref format is "group/version/resource". If a built-in entry matches by
// group and resource but has a different version (e.g., hardcoded v1beta1 vs
// cluster-served v1), the version from the ref is used.
func FindResourceTypeIn(ref string, additional []ResourceTypeEntry) (ResourceTypeEntry, bool) {
	// Parse the ref to extract version for potential override.
	refParts := strings.SplitN(ref, "/", 3)

	// Build a lookup of discovered CRDs by group/resource for enriching built-in types
	// with PrinterColumns from CRD discovery.
	discoveredByGR := make(map[string]*ResourceTypeEntry, len(additional))
	for i := range additional {
		key := additional[i].APIGroup + "/" + additional[i].Resource
		discoveredByGR[key] = &additional[i]
	}

	for _, cat := range TopLevelResourceTypes() {
		for _, rt := range cat.Types {
			matched := false
			if rt.ResourceRef() == ref {
				matched = true
			} else if len(refParts) == 3 && rt.APIGroup == refParts[0] && rt.Resource == refParts[2] {
				// Match by group/resource, override version from ref.
				rt.APIVersion = refParts[1]
				matched = true
			}
			if matched {
				// Enrich built-in type with PrinterColumns from discovered CRDs.
				grKey := rt.APIGroup + "/" + rt.Resource
				if crd, ok := discoveredByGR[grKey]; ok && len(crd.PrinterColumns) > 0 {
					rt.PrinterColumns = crd.PrinterColumns
				}
				return rt, true
			}
		}
	}
	for _, rt := range additional {
		if rt.ResourceRef() == ref {
			return rt, true
		}
	}
	return ResourceTypeEntry{}, false
}

// IsScaleableKind returns true if the given kind supports the scale operation.
func IsScaleableKind(kind string) bool {
	switch kind {
	case "Deployment", "StatefulSet", "ReplicaSet":
		return true
	default:
		return false
	}
}

// IsRestartableKind returns true if the given kind supports the restart operation.
func IsRestartableKind(kind string) bool {
	switch kind {
	case "Deployment", "StatefulSet", "DaemonSet":
		return true
	default:
		return false
	}
}

// IsForceDeleteableKind returns true if the given kind supports the force delete operation
// (kubectl delete --grace-period=0 --force, without removing finalizers).
func IsForceDeleteableKind(kind string) bool {
	switch kind {
	case "Pod", "Job":
		return true
	default:
		return false
	}
}
