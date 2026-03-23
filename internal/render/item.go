package render

import "strings"

func normalizeRenderableItemID(raw string) (string, bool) {
	itemID := strings.TrimSpace(raw)
	if itemID == "" {
		return "", false
	}
	return itemID, true
}
