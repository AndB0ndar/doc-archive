package chunker

func Chunk(text string, chunkSize, overlap int) []string {
	if len(text) == 0 {
		return nil
	}
	runes := []rune(text)
	totalRunes := len(runes)
	if totalRunes <= chunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0
	for start < totalRunes {
		end := start + chunkSize
		if end > totalRunes {
			end = totalRunes
		}
		chunkRunes := runes[start:end]
		chunks = append(chunks, string(chunkRunes))
		start += chunkSize - overlap
		if start < 0 {
			start = 0
		}
	}
	return chunks
}
