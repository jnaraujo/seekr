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
	EmbeddingDimension    = 1536
	MaxChunkChars         = 1000
	ChunkOverlapping      = 200
	DefaultEmbeddingModel = "qwen2:1.5b-instruct"
	MaxContentChars       = 10_000_000
)
