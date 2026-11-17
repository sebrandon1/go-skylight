package cmd

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/sebrandon1/go-skylight/lib"
)

// isCrossFrame reports whether the export came from a different frame than the
// import target, so source category and recipe IDs mean nothing there.
func isCrossFrame(data ExportData) bool {
	return data.FrameID != "" && data.FrameID != frameID
}

// forEachCategoryRef calls fn with a pointer to every assignee or category ID
// in the resources an import would create. Rows import skips (routine chores
// and finished one-off chores) are left out.
func forEachCategoryRef(data *ExportData, want map[string]bool, fn func(*string)) {
	if want[exportResourceChores] {
		for i := range data.Chores {
			c := &data.Chores[i]
			if c.Routine || (len(c.RecurrenceSet) == 0 && isDone(*c)) {
				continue
			}
			fn(&c.AssigneeID)
		}
	}
	if want[exportResourceRewards] {
		for i := range data.Rewards {
			fn(&data.Rewards[i].CategoryID)
		}
	}
	if want[exportResourceRoutines] {
		for i := range data.Routines {
			fn(&data.Routines[i].AssigneeID)
		}
	}
	if want[exportResourceCalendar] {
		for i := range data.CalendarEvents {
			fn(&data.CalendarEvents[i].CategoryID)
		}
	}
	if shouldImportBounties(want) {
		for i := range data.Bounties {
			fn(&data.Bounties[i].Chore.AssigneeID)
			fn(&data.Bounties[i].Reward.CategoryID)
		}
	}
}

// remapCategories returns data with every category reference replaced through
// mapping (source ID to target ID). The caller's slices are not modified.
func remapCategories(data ExportData, want map[string]bool, mapping map[string]string) ExportData {
	data.Chores = slices.Clone(data.Chores)
	data.Rewards = slices.Clone(data.Rewards)
	data.Routines = slices.Clone(data.Routines)
	data.CalendarEvents = slices.Clone(data.CalendarEvents)
	data.Bounties = slices.Clone(data.Bounties)
	forEachCategoryRef(&data, want, func(id *string) {
		if n, ok := mapping[*id]; ok {
			*id = n
		}
	})
	return data
}

// remapSittingRecipes points sittings at the recipes created in this run. A
// sitting whose recipe was not imported loses the link, because the source ID
// is meaningless on the target frame; it keeps its summary.
func remapSittingRecipes(sittings []lib.MealSitting, recipeIDs *sync.Map) []lib.MealSitting {
	out := slices.Clone(sittings)
	for i := range out {
		if out[i].RecipeID == "" {
			continue
		}
		if n, ok := recipeIDs.Load(out[i].RecipeID); ok {
			out[i].RecipeID, _ = n.(string)
		} else {
			fmt.Fprintf(os.Stderr, "Meal sitting %q: recipe %s was not imported; creating it without a recipe\n", out[i].Summary, out[i].RecipeID)
			out[i].RecipeID = ""
		}
	}
	return out
}

// remapPlan is how the export's categories line up with the target frame's.
type remapPlan struct {
	mapping    map[string]string // source ID to target ID, for matched labels
	matched    []lib.Category    // exported categories with a target match
	missing    []lib.Category    // exported categories with no target match
	unresolved []string          // referenced IDs absent from the export's categories
	needed     []string          // labels of missing categories that items reference
}

// planCategoryRemap matches the export's categories to the target frame's by
// exact label. It only reads.
func planCategoryRemap(ctx context.Context, client *lib.Client, data ExportData, want map[string]bool) (*remapPlan, error) {
	target, err := client.ListCategories(ctx, frameID)
	if err != nil {
		return nil, fmt.Errorf("listing target categories: %w", err)
	}
	byLabel := make(map[string]string, len(target))
	for _, c := range target {
		if _, dup := byLabel[c.Name]; !dup { // the first of duplicate labels wins
			byLabel[c.Name] = c.ID
		}
	}

	p := &remapPlan{mapping: map[string]string{}}
	known := map[string]string{} // source ID to label
	for _, c := range data.Categories {
		known[c.ID] = c.Name
		if id, ok := byLabel[c.Name]; ok {
			p.mapping[c.ID] = id
			p.matched = append(p.matched, c)
		} else {
			p.missing = append(p.missing, c)
		}
	}
	unresolved, needed := map[string]bool{}, map[string]bool{}
	forEachCategoryRef(&data, want, func(id *string) {
		if *id == "" {
			return
		}
		if name, ok := known[*id]; !ok {
			unresolved[*id] = true
		} else if _, ok := p.mapping[*id]; !ok {
			needed[fmt.Sprintf("%q", name)] = true
		}
	})
	p.unresolved = slices.Sorted(maps.Keys(unresolved))
	p.needed = slices.Sorted(maps.Keys(needed))
	return p, nil
}

// check reports why the import cannot go ahead: items that reference a
// category the export does not describe, or one that has no match on the
// target and may not be created.
func (p *remapPlan) check(want map[string]bool) error {
	var problems []string
	if len(p.unresolved) > 0 {
		problems = append(problems, "category IDs not in the export's categories (export with categories included): "+strings.Join(p.unresolved, ", "))
	}
	if !want[exportResourceCategories] && len(p.needed) > 0 {
		problems = append(problems, "no category with these labels on the target frame (add them, or include categories in --resources to create them): "+strings.Join(p.needed, ", "))
	}
	if len(problems) == 0 {
		return nil
	}
	return errors.New("cannot remap categories to frame " + frameID + ": " + strings.Join(problems, "; "))
}

// print shows the mapping and anything that does not match.
func (p *remapPlan) print() {
	fmt.Printf("Category mapping (export frame -> frame %s):\n", frameID)
	for _, c := range p.matched {
		fmt.Printf("  %q: %s -> %s\n", c.Name, c.ID, p.mapping[c.ID])
	}
	for _, c := range p.missing {
		fmt.Printf("  %q: %s -> no match\n", c.Name, c.ID)
	}
	for _, id := range p.unresolved {
		fmt.Printf("  %s: not in the export's categories\n", id)
	}
}

// apply creates the missing categories when categories are being imported and
// returns the full source-to-target ID mapping. Linked-profile status is not
// copied: the API may not allow creating linked profiles.
func (p *remapPlan) apply(ctx context.Context, client *lib.Client, want map[string]bool) (map[string]string, error) {
	mapping := maps.Clone(p.mapping)
	if !want[exportResourceCategories] {
		return mapping, nil
	}
	for _, c := range p.missing {
		created, err := client.CreateCategory(ctx, frameID, lib.CategoryData{Name: c.Name, Color: c.Color})
		if err != nil {
			return nil, fmt.Errorf("creating category %q: %w", c.Name, err)
		}
		mapping[c.ID] = created.ID
	}
	return mapping, nil
}

// remapForTargetFrame is a no-op within one frame. Across frames it creates any
// categories the export has and the target lacks, if categories are being
// imported, and returns data with its category references pointing at the
// target frame. It fails before writing anything when a referenced category
// cannot be matched or created.
func remapForTargetFrame(ctx context.Context, client *lib.Client, data ExportData, want map[string]bool) (ExportData, error) {
	if !isCrossFrame(data) {
		return data, nil
	}
	plan, err := planCategoryRemap(ctx, client, data, want)
	if err != nil {
		return data, err
	}
	if err := plan.check(want); err != nil {
		return data, err
	}
	mapping, err := plan.apply(ctx, client, want)
	if err != nil {
		return data, err
	}
	return remapCategories(data, want, mapping), nil
}
