package catalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
)

// Parse decodes models.json.
func Parse(data []byte) (Catalog, error) {
	var f File
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return Catalog{}, fmt.Errorf("decode model catalog: %w", err)
	}
	return f.Catalog, nil
}

// Marshal encodes c deterministically: sorted map keys, 2-space indent,
// trailing newline.
func Marshal(c Catalog) ([]byte, error) {
	data, err := json.MarshalIndent(File{Catalog: c}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode model catalog: %w", err)
	}
	return append(data, '\n'), nil
}

// Validate checks the manual keys: known schema and tier names, every tiered
// id present in Models, no id in two tiers.
func (c Catalog) Validate() error {
	var errs []error
	if c.Schema != Schema {
		errs = append(errs, fmt.Errorf("schema %d, want %d", c.Schema, Schema))
	}
	seen := map[string]string{}
	for _, tier := range sortedKeys(c.Tiers) {
		if !slices.Contains(Tiers(), tier) {
			errs = append(errs, fmt.Errorf("tiers.%s: unknown tier (want one of %v)", tier, Tiers()))
		}
		for _, id := range c.Tiers[tier] {
			if prev, ok := seen[id]; ok {
				errs = append(errs, fmt.Errorf("tiers.%s: %s is already in tiers.%s", tier, id, prev))
				continue
			}
			seen[id] = tier
			if _, ok := c.Models[id]; !ok {
				errs = append(errs, fmt.Errorf("tiers.%s: %s is not in models (run models sync, or fix the id)", tier, id))
			}
		}
	}
	return errors.Join(errs...)
}

// TierOf returns the tier of id, or "" when it has none.
func (c Catalog) TierOf(id string) string {
	for tier, ids := range c.Tiers {
		if slices.Contains(ids, id) {
			return tier
		}
	}
	return ""
}

// Untiered lists, sorted, the ids in Models that have no tier and are not ignored.
func (c Catalog) Untiered() []string {
	var out []string
	for _, id := range sortedKeys(c.Models) {
		if c.TierOf(id) == "" && !slices.Contains(c.Ignore, id) {
			out = append(out, id)
		}
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
