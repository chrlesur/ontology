package tokenizer

import (
	"github.com/pkoukk/tiktoken-go"
)

// CountTokens returns the number of tokens in the given text
func CountTokens(text string) (int, error) {
	encoding, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return 0, err
	}
	tokens := encoding.Encode(text, nil, nil)
	return len(tokens), nil
}
