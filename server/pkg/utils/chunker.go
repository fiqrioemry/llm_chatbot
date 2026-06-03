package utils

import "strings"

// ChunkText splits text into overlapping word-based chunks using a sliding window.
// chunkSize = number of words per chunk, overlap = number of words shared with next chunk.
func ChunkText(text string, chunkSize, overlap int) []string {
	if text == "" {
		return nil
	}
	if chunkSize <= 0 {
		chunkSize = 512
	}
	if overlap < 0 || overlap >= chunkSize {
		overlap = 0
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	step := chunkSize - overlap
	if step <= 0 {
		step = 1
	}

	var chunks []string
	for start := 0; start < len(words); start += step {
		end := start + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunk := strings.Join(words[start:end], " ")
		chunks = append(chunks, chunk)
		if end == len(words) {
			break
		}
	}

	return chunks
}
