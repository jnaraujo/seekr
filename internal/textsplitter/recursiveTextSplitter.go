package textsplitter

// // This function/section is adapted from https://github.com/henomis/lingoose/blob/780ad85cb0a4a227ce20bca0e5ddf2ba402fdedf/textsplitter/recursiveTextSplitter.go

import (
	"strings"
)

var (
	defaultSeparators                 = []string{"\n\n", "\n", " ", ""}
	defaultLengthFunction LenFunction = func(s string) int { return len(s) }
)

type RecursiveCharacterTextSplitter struct {
	TextSplitter
	separators []string
}

func NewRecursiveCharacterTextSplitter(chunkSize int, chunkOverlap int) *RecursiveCharacterTextSplitter {
	return &RecursiveCharacterTextSplitter{
		TextSplitter: TextSplitter{
			chunkSize:      chunkSize,
			chunkOverlap:   chunkOverlap,
			lengthFunction: defaultLengthFunction,
		},
		separators: defaultSeparators,
	}
}

func (r *RecursiveCharacterTextSplitter) WithSeparators(separators []string) *RecursiveCharacterTextSplitter {
	r.separators = separators
	return r
}

func (r *RecursiveCharacterTextSplitter) WithLengthFunction(
	lengthFunction LenFunction,
) *RecursiveCharacterTextSplitter {
	r.lengthFunction = lengthFunction
	return r
}

func (r *RecursiveCharacterTextSplitter) SplitText(text string) []string {
	// Split incoming text and return chunks.
	finalChunks := []string{}
	// Get appropriate separator to use
	separator := r.separators[len(r.separators)-1]
	newSeparators := []string{}
	for i, s := range r.separators {
		if s == "" {
			separator = s
			break
		}

		if strings.Contains(text, s) {
			separator = s
			newSeparators = r.separators[i+1:]
			break
		}
	}
	// Now that we have the separator, split the text
	splits := strings.Split(text, separator)
	// Now go merging things, recursively splitting longer texts.
	goodSplits := []string{}
	for _, s := range splits {
		if r.lengthFunction(s) < r.chunkSize {
			goodSplits = append(goodSplits, s)
		} else {
			if len(goodSplits) > 0 {
				mergedText := r.mergeSplits(goodSplits, separator)
				finalChunks = append(finalChunks, mergedText...)
				goodSplits = []string{}
			}
			if len(newSeparators) == 0 {
				finalChunks = append(finalChunks, s)
			} else {
				otherInfo := r.SplitText(s)
				finalChunks = append(finalChunks, otherInfo...)
			}
		}
	}
	if len(goodSplits) > 0 {
		mergedText := r.mergeSplits(goodSplits, separator)
		finalChunks = append(finalChunks, mergedText...)
	}
	return finalChunks
}
