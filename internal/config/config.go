package config

const (
	// Application configuration
	AppID      = "seekr"
	AppName    = "SeekR"
	AppVersion = "0.1.0"
	DBFileExt  = "skdb"
)

const (
	// Embedding configuration settings
	EmbeddingDimension = 768
	MaxChunkChars      = 500
	ChunkOverlapping   = 100
	// ollama pull hf.co/nomic-ai/nomic-embed-text-v2-moe-gguf
	DefaultEmbeddingModel = "hf.co/nomic-ai/nomic-embed-text-v2-moe-gguf"
	MaxContent            = 10_000
)
