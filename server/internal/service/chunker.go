package service

import (
	"strings"
)

type TextSplitter struct {
	ChunkSize  int
	Overlap    int
	Separators []string
}

func (s *TextSplitter) SplitText(text string) []string {
	micro := s.recursiveSplit(text, s.Separators)
	return s.mergeWithOverlap(micro)
}

func (s *TextSplitter) recursiveSplit(text string, seps []string) []string {
	if len(text) <= s.ChunkSize {
		return []string{text}
	}
	if len(seps) == 0 {
		return []string{text[:s.ChunkSize]}
	}

	sep := seps[0]
	parts := strings.SplitAfter(text, sep)

	var result []string
	for _, part := range parts {
		if part == "" {
			continue
		}
		if len(part) > s.ChunkSize && len(seps) > 1 {
			subParts := s.recursiveSplit(part, seps[1:])
			result = append(result, subParts...)
		} else {
			result = append(result, part)
		}
	}
	return result
}

func (s *TextSplitter) mergeWithOverlap(parts []string) []string {
	var chunks []string
	var current []string
	currentSize := 0

	for _, part := range parts {
		partSize := len(part)

		if partSize > s.ChunkSize {
			if len(current) > 0 {
				chunks = append(chunks, strings.Join(current, " "))
				current = nil
				currentSize = 0
			}
			chunks = append(chunks, part)
			continue
		}

		if currentSize+partSize > s.ChunkSize && len(current) > 0 {
			chunks = append(chunks, strings.Join(current, " "))

			overlapStart := len(current) - s.Overlap
			if overlapStart < 0 {
				overlapStart = 0
			}
			current = append([]string{}, current[overlapStart:]...)
			currentSize = 0
			for _, p := range current {
				currentSize += len(p)
			}
		}

		current = append(current, part)
		currentSize += partSize
	}

	if len(current) > 0 {
		chunk := strings.Join(current, " ")
		chunk = strings.Join(strings.Fields(chunk), " ")
		chunks = append(chunks, chunk)
	}
	return chunks
}
