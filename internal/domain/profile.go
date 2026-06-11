// SPDX-License-Identifier: AGPL-3.0-only

package domain

import (
	"fmt"
	"strings"
)

type BusinessProfile struct {
	Sector         string
	Segment        string
	Label          string
	Description    string
	DiscoveryTerms []string
	IncludeTerms   []string
	ExcludeTerms   []string
	MinScore       int
}

func (p BusinessProfile) Key() string {
	return ProfileKey(p.Sector, p.Segment)
}

func ProfileKey(sector, segment string) string {
	return normalizeProfilePart(sector) + "/" + normalizeProfilePart(segment)
}

func BusinessProfiles() []BusinessProfile {
	profiles := []BusinessProfile{
		{
			Sector:         "commerce",
			Segment:        "wholesalers",
			Label:          "Commerce / Wholesalers",
			Description:    "Suppliers, distributors, bulk sellers, trade stores, and wholesale businesses.",
			DiscoveryTerms: []string{"wholesale", "wholesaler", "wholesalers", "supplier", "distributor", "bulk", "trade"},
			IncludeTerms:   []string{"wholesale", "wholesaler", "wholesalers", "supplier", "distributor", "bulk order", "trade price", "trade account", "b2b", "dealer"},
			ExcludeTerms:   blockedBusinessTerms(),
			MinScore:       35,
		},
		{
			Sector:         "commerce",
			Segment:        "ecommerce",
			Label:          "Commerce / Ecommerce",
			Description:    "Online stores with product, cart, checkout, or storefront signals.",
			DiscoveryTerms: []string{"shop", "store", "cart", "checkout", "product"},
			IncludeTerms:   []string{"shop", "store", "cart", "checkout", "product", "buy now", "add to cart", "woocommerce", "shopify"},
			ExcludeTerms:   blockedBusinessTerms(),
			MinScore:       35,
		},
		{
			Sector:         "commerce",
			Segment:        "retailers",
			Label:          "Commerce / Retailers",
			Description:    "Retail stores and consumer-facing merchants with purchasable goods.",
			DiscoveryTerms: []string{"retail", "retailer", "retailers", "shop", "store", "products"},
			IncludeTerms:   []string{"retail", "retailer", "retailers", "shop", "store", "products", "cart", "checkout", "delivery", "opening hours"},
			ExcludeTerms:   blockedBusinessTerms(),
			MinScore:       35,
		},
		{
			Sector:         "services",
			Segment:        "salons",
			Label:          "Services / Salons",
			Description:    "Hair, beauty, grooming, and salon service businesses.",
			DiscoveryTerms: []string{"salon", "salons", "hairdresser", "beauty", "barber", "spa"},
			IncludeTerms:   []string{"salon", "salons", "hairdresser", "beauty", "barber", "spa", "manicure", "pedicure", "stylist", "booking"},
			ExcludeTerms:   blockedBusinessTerms(),
			MinScore:       30,
		},
	}
	return profiles
}

func FindBusinessProfile(sector, segment string) (BusinessProfile, error) {
	key := ProfileKey(sector, segment)
	for _, profile := range BusinessProfiles() {
		if profile.Key() == key {
			return profile, nil
		}
	}
	return BusinessProfile{}, fmt.Errorf("unknown business profile %q", key)
}

func CustomBusinessProfile(query string) (BusinessProfile, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return BusinessProfile{}, fmt.Errorf("query is required")
	}
	terms := ExpandIntentTerms(query)
	if len(terms) == 0 {
		return BusinessProfile{}, fmt.Errorf("query must contain searchable words")
	}
	sector, segment := InferIntentProfile(query)
	if segment == "" {
		segment = strings.Join(splitProfileTerms(query), "-")
	}
	return BusinessProfile{
		Sector:         sector,
		Segment:        segment,
		Label:          "Custom / " + query,
		Description:    "Businesses matching the operator search terms.",
		DiscoveryTerms: terms,
		IncludeTerms:   terms,
		ExcludeTerms:   blockedBusinessTerms(),
		MinScore:       25,
	}, nil
}

func InferIntentProfile(query string) (string, string) {
	terms := splitProfileTerms(query)
	for _, term := range terms {
		switch term {
		case "shop", "store", "retail", "retailer", "boutique", "mall":
			return "commerce", "retailers"
		case "wholesale", "wholesaler", "supplier", "distributor", "bulk":
			return "commerce", "wholesalers"
		case "ecommerce", "checkout", "cart", "online":
			return "commerce", "ecommerce"
		case "salon", "hair", "hairdresser", "beauty", "barber", "spa":
			return "services", "salons"
		}
	}
	return "custom", strings.Join(terms, "-")
}

func ExpandIntentTerms(query string) []string {
	terms := splitProfileTerms(query)
	seen := map[string]bool{}
	var out []string
	add := func(values ...string) {
		for _, value := range values {
			if value == "" || seen[value] {
				continue
			}
			seen[value] = true
			out = append(out, value)
		}
	}
	for _, term := range terms {
		add(term)
		switch term {
		case "shop", "store", "retail", "retailer", "boutique":
			add("store", "shop", "retail", "products", "cart", "checkout", "delivery")
		case "wholesale", "wholesaler", "supplier", "distributor", "bulk":
			add("wholesale", "wholesaler", "supplier", "distributor", "bulk", "trade account", "b2b")
		case "ecommerce", "online", "checkout", "cart":
			add("ecommerce", "online store", "checkout", "cart", "add to cart", "product")
		case "salon", "hair", "hairdresser", "beauty", "barber", "spa":
			add("salon", "hairdresser", "beauty", "barber", "spa", "booking", "stylist")
		}
	}
	return out
}

func normalizeProfilePart(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return value
}

func splitProfileTerms(query string) []string {
	fields := strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	})
	seen := map[string]bool{}
	var out []string
	for _, field := range fields {
		if len(field) < 3 || seen[field] {
			continue
		}
		seen[field] = true
		out = append(out, field)
	}
	return out
}

func blockedBusinessTerms() []string {
	return []string{"casino", "gambling", "porn", "adult", "betting"}
}
