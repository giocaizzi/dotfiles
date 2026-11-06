package catalog

// Apply merges freshly fetched sources into c, keeping its manual keys.
func Apply(c Catalog, src Sources) (Catalog, Report) {
	_ = src
	return c, Report{} // TODO(unit 1)
}
