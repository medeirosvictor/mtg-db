package app

import "strings"

// generateSlug creates a filesystem-safe slug from a deck title.
func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	var cleaned strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleaned.WriteRune(r)
		}
	}
	slug = cleaned.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if len(slug) > 60 {
		slug = slug[:60]
	}
	return slug
}

// normalizeStatus extracts a clean status label.
func normalizeStatus(raw string) string {
	lower := strings.ToLower(raw)
	switch {
	case strings.Contains(lower, "owned"):
		return "Owned"
	case strings.Contains(lower, "planned"):
		return "Planned"
	case strings.Contains(lower, "disassembled"):
		return "Disassembled"
	default:
		return raw
	}
}
