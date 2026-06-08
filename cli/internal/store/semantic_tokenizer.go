package store

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

const semanticMaxTokens = 256

type wordPieceTokenizer struct {
	vocab  map[string]int64
	clsID  int64
	sepID  int64
	padID  int64
	unkID  int64
	maxLen int
}

func loadWordPieceTokenizer(vocabPath string) (*wordPieceTokenizer, error) {
	f, err := os.Open(vocabPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	vocab := make(map[string]int64, 30522)
	scanner := bufio.NewScanner(f)
	var id int64
	for scanner.Scan() {
		token := strings.TrimSpace(scanner.Text())
		if token != "" {
			vocab[token] = id
		}
		id++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	t := &wordPieceTokenizer{vocab: vocab, maxLen: semanticMaxTokens}
	var ok bool
	if t.clsID, ok = vocab["[CLS]"]; !ok {
		return nil, fmt.Errorf("vocab missing [CLS]")
	}
	if t.sepID, ok = vocab["[SEP]"]; !ok {
		return nil, fmt.Errorf("vocab missing [SEP]")
	}
	if t.padID, ok = vocab["[PAD]"]; !ok {
		return nil, fmt.Errorf("vocab missing [PAD]")
	}
	if t.unkID, ok = vocab["[UNK]"]; !ok {
		return nil, fmt.Errorf("vocab missing [UNK]")
	}
	return t, nil
}

func (t *wordPieceTokenizer) Encode(text string) ([]int64, []int64, []int64, bool) {
	tokens := basicBertTokens(text)
	if len(tokens) == 0 {
		return nil, nil, nil, false
	}

	ids := make([]int64, 0, t.maxLen)
	ids = append(ids, t.clsID)
	for _, token := range tokens {
		pieces := t.wordPieces(token)
		if len(ids)+len(pieces)+1 > t.maxLen {
			break
		}
		ids = append(ids, pieces...)
	}
	ids = append(ids, t.sepID)
	mask := make([]int64, len(ids))
	tokenTypes := make([]int64, len(ids))
	for i := range mask {
		mask[i] = 1
	}
	return ids, mask, tokenTypes, true
}

func (t *wordPieceTokenizer) wordPieces(token string) []int64 {
	runes := []rune(token)
	if len(runes) == 0 {
		return nil
	}
	if len(runes) > 100 {
		return []int64{t.unkID}
	}
	var pieces []int64
	for start := 0; start < len(runes); {
		var found string
		for end := len(runes); end > start; end-- {
			candidate := string(runes[start:end])
			if start > 0 {
				candidate = "##" + candidate
			}
			if _, ok := t.vocab[candidate]; ok {
				found = candidate
				break
			}
		}
		if found == "" {
			return []int64{t.unkID}
		}
		pieces = append(pieces, t.vocab[found])
		if strings.HasPrefix(found, "##") {
			start += len([]rune(found[2:]))
		} else {
			start += len([]rune(found))
		}
	}
	return pieces
}

func basicBertTokens(text string) []string {
	var tokens []string
	var current strings.Builder
	flush := func() {
		if current.Len() == 0 {
			return
		}
		tokens = append(tokens, strings.ToLower(current.String()))
		current.Reset()
	}
	for _, r := range text {
		switch {
		case r == 0 || r == unicode.ReplacementChar || (unicode.IsControl(r) && !unicode.IsSpace(r)):
			continue
		case unicode.IsSpace(r):
			flush()
		case isCJK(r):
			flush()
			tokens = append(tokens, strings.ToLower(string(r)))
		case isBertPunctuation(r):
			flush()
			tokens = append(tokens, string(r))
		default:
			current.WriteRune(unicode.ToLower(r))
		}
	}
	flush()
	return tokens
}

func isBertPunctuation(r rune) bool {
	if (r >= 33 && r <= 47) || (r >= 58 && r <= 64) || (r >= 91 && r <= 96) || (r >= 123 && r <= 126) {
		return true
	}
	return unicode.IsPunct(r)
}

func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) ||
		(r >= 0x3400 && r <= 0x4DBF) ||
		(r >= 0x20000 && r <= 0x2A6DF) ||
		(r >= 0x2A700 && r <= 0x2B73F) ||
		(r >= 0x2B740 && r <= 0x2B81F) ||
		(r >= 0x2B820 && r <= 0x2CEAF) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0x2F800 && r <= 0x2FA1F)
}
