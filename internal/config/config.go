package config

const (
	// Application configuration
	AppID      = "seekr"
	AppName    = "SeekR"
	AppVersion = 0.1
	DBFileExt  = "skdb"
)

const (
	// Embedding configuration settings
	EmbeddingDimension    = 768
	MaxChunkChars         = 500
	ChunkOverlapping      = 100
	DefaultEmbeddingModel = "nomic-embed-text"
	MaxContent            = 10_000
)
