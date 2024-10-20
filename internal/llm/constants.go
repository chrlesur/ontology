// internal/llm/constants.go

package llm

import "time"


const (
    MaxRetries        = 5
    InitialRetryDelay = 1 * time.Second
    MaxRetryDelay     = 32 * time.Second
)

var ModelContextLimits = map[string]int{
    "GPT-4o":                 8192,
    "GPT-4o mini":            4096,
    "o1-preview":             16384,
    "o1-mini":                2048,
    "claude-3-5-sonnet-20240620": 200000,
    "claude-3-opus-20240229":     200000,
    "claude-3-haiku-20240307":    200000,
    "llama3.2:3B":            4096,
    "llama3.1:8B":            4096,
    "mistral-nemo:12B":       8192,
    "mixtral:7B":             32768,
    "mistral:7B":             8192,
    "mistral-small:22B":      16384,
}