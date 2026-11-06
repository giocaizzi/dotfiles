package catalog

import (
	"cmp"
	"slices"
	"strings"
)

// Tiers are computed by sync, never edited by hand; Catalog.Pin overrides a
// model and Catalog.Ignore drops it. The rule, per model of Catalog.Models:
//
//  1. A model with no API price, or with neither an arena rank nor an Epoch ECI,
//     has no tier.
//  2. The blended price p = (3*In + Out) / 4 $ per million tokens
//     (BlendInWeight, BlendOutWeight) picks the band: below BandVersatile is
//     lightweight, below BandPowerful versatile, below BandExtra powerful,
//     else extra. A price exactly on a limit belongs to the higher band.
//  3. A model whose ECI is in the top quartile of the catalog (ECI > 0 and at
//     least the k-th highest ECI, k = max(1, n/4), n the models with an ECI)
//     moves up one band, never into extra.
//  4. A pin wins over the rule.
//
// Each tier lists its models by release date descending, then canonical id.
const (
	BlendInWeight  = 3.0
	BlendOutWeight = 1.0

	BandVersatile = 2.5
	BandPowerful  = 7.0
	BandExtra     = 15.0
)

// Untiered reasons.
const (
	ReasonNoPrice   = "no API price"
	ReasonNoRanking = "no ranking data"
)

// BlendedPrice is the input-weighted mean of the input and output price, $/Mtok.
func BlendedPrice(p Price) float64 {
	return (BlendInWeight*p.In + BlendOutWeight*p.Out) / (BlendInWeight + BlendOutWeight)
}

// priceTier is the tier of a blended price.
func priceTier(p float64) string {
	switch {
	case p < BandVersatile:
		return TierLightweight
	case p < BandPowerful:
		return TierVersatile
	case p < BandExtra:
		return TierPowerful
	}
	return TierExtra
}

// ECICut is the ECI a model must reach to be promoted: the k-th highest ECI of
// models, k = max(1, n/4). It is 0 when no model has an ECI.
func ECICut(models map[string]Model) float64 {
	var ecis []float64
	for _, m := range models {
		if m.External.ECI > 0 {
			ecis = append(ecis, m.External.ECI)
		}
	}
	if len(ecis) == 0 {
		return 0
	}
	slices.SortFunc(ecis, func(a, b float64) int { return cmp.Compare(b, a) })
	return ecis[max(1, len(ecis)/4)-1]
}

// UntieredReason says why the rule gives m no tier, or "" when it has one.
func UntieredReason(m Model) string {
	if m.API == nil {
		return ReasonNoPrice
	}
	if m.External.ECI == 0 && len(m.Arena) == 0 {
		return ReasonNoRanking
	}
	return ""
}

// ComputeTiers applies the tier rule and the pins to models.
func ComputeTiers(models map[string]Model, pin map[string]string) map[string][]string {
	cut := ECICut(models)
	tiers := map[string][]string{}
	for _, id := range sortedKeys(models) {
		m := models[id]
		tier, pinned := pin[id]
		if !pinned {
			if UntieredReason(m) != "" {
				continue
			}
			tier = priceTier(BlendedPrice(*m.API))
			// Promotion stops below extra: powerful and extra stay.
			if cut > 0 && m.External.ECI >= cut && (tier == TierLightweight || tier == TierVersatile) {
				tier = Tiers()[slices.Index(Tiers(), tier)+1]
			}
		}
		tiers[tier] = append(tiers[tier], id)
	}
	for _, ids := range tiers {
		slices.SortFunc(ids, func(a, b string) int {
			return cmp.Or(strings.Compare(models[b].ReleaseDate, models[a].ReleaseDate), strings.Compare(a, b))
		})
	}
	return tiers
}
