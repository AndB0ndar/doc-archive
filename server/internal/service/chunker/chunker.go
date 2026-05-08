package chunker

import (
	"fmt"
	"strings"

	"github.com/AndB0ndar/doc-archive/internal/domain"
)

type chunker struct {
	ChunkSize        int
	Overlap          int
	Separators       []string
	SplitBySentences bool
}

type Config struct {
	ChunkSize        int
	Overlap          int
	SplitBySentences bool
}

func New(
	size, overlap int, separators []string, splitby bool,
) domain.ChunkerService {
	return &chunker{
		ChunkSize:        size,
		Overlap:          overlap,
		Separators:       separators,
		SplitBySentences: splitby,
	}
}

func (c *chunker) Split(text string) ([]string, error) {
	if c.ChunkSize <= 0 {
		return nil, fmt.Errorf("chunk size must be positive, got %d", c.ChunkSize)
	}
	if c.Overlap < 0 {
		return nil, fmt.Errorf("overlap must be non-negative, got %d", c.Overlap)
	}
	if c.Overlap >= c.ChunkSize {
		return nil, fmt.Errorf("overlap (%d) must be less than chunk size (%d)", c.Overlap, c.ChunkSize)
	}
	if c.SplitBySentences {
		return c.mergeWithOverlap(c.recursiveSplit(text, c.Separators)), nil
	}
	return c.splitByFixedSize(text), nil
}

func (c *chunker) splitByFixedSize(text string) []string {
	chunks := make([]string, 0)
	textRunes := []rune(text)
	total := len(textRunes)

	if total <= c.ChunkSize {
		return []string{text}
	}

	start := 0
	for start < total {
		end := start + c.ChunkSize
		if end > total {
			end = total
		}
		chunk := string(textRunes[start:end])
		chunks = append(chunks, chunk)

		start += c.ChunkSize - c.Overlap
		if start < 0 {
			start = 0
		}
		if start >= total {
			break
		}
	}
	return chunks
}

func (c *chunker) recursiveSplit(text string, seps []string) []string {
	if len(text) <= c.ChunkSize {
		return []string{text}
	}
	if len(seps) == 0 {
		return []string{text[:c.ChunkSize]}
	}

	sep := seps[0]
	parts := strings.SplitAfter(text, sep)

	var result []string
	for _, part := range parts {
		if part == "" {
			continue
		}
		if len(part) > c.ChunkSize && len(seps) > 1 {
			subParts := c.recursiveSplit(part, seps[1:])
			result = append(result, subParts...)
		} else {
			result = append(result, part)
		}
	}
	return result
}

func (c *chunker) mergeWithOverlap(parts []string) []string {
	var chunks []string
	var current []string
	currentSize := 0

	for _, part := range parts {
		partSize := len(part)

		if partSize > c.ChunkSize {
			if len(current) > 0 {
				chunks = append(chunks, strings.Join(current, " "))
				current = nil
				currentSize = 0
			}
			chunks = append(chunks, part)
			continue
		}

		if currentSize+partSize > c.ChunkSize && len(current) > 0 {
			chunks = append(chunks, strings.Join(current, " "))

			overlapStart := len(current) - c.Overlap
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
