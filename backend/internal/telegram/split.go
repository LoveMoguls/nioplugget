package telegram

// telegramMessageLimit is Telegram's hard cap on message text length.
const telegramMessageLimit = 4096

// splitMessage splits text into chunks of at most max runes, preferring to
// break at a newline, then a space, in the last half of the chunk.
func splitMessage(text string, max int) []string {
	runes := []rune(text)
	if len(runes) <= max {
		return []string{text}
	}
	var parts []string
	for len(runes) > max {
		cut := 0
		for i := max; i > max/2; i-- {
			if runes[i-1] == '\n' {
				cut = i
				break
			}
		}
		if cut == 0 {
			for i := max; i > max/2; i-- {
				if runes[i-1] == ' ' {
					cut = i
					break
				}
			}
		}
		if cut == 0 {
			cut = max
		}
		parts = append(parts, string(runes[:cut]))
		runes = runes[cut:]
	}
	if len(runes) > 0 {
		parts = append(parts, string(runes))
	}
	return parts
}
