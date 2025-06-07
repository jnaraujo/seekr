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
	EmbeddingDimension    = 768
	MaxChunkChars         = 1000
	ChunkOverlapping      = 200
	DefaultEmbeddingModel = "hf.co/nomic-ai/nomic-embed-text-v2-moe-gguf"
	MaxContentChars       = 10_000_000
)
