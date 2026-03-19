package content

import "unicode/utf8"

func DetectLanguage(text string) string {
	total := 0
	cjk := 0
	for _, r := range text {
		if r == '\n' || r == '\r' || r == '\t' || r == ' ' {
			continue
		}
		total++
		if (r >= 0x4E00 && r <= 0x9FFF) || (r >= 0x3400 && r <= 0x4DBF) {
			cjk++
		}
	}
	if total == 0 {
		return "unknown"
	}
	if cjk*100/total >= 20 {
		return "zh"
	}
	if utf8.ValidString(text) {
		return "en"
	}
	return "unknown"
}
