package domain

type EmbedResult struct {
	Embeddings []float32
}

type RerankResult struct {
	Scores []float32
}

type Answer struct {
	Answer     string
	Confidence float64
	Start      int
	End        int
}

type EmbedderClient interface {
	Embed(text string) (*EmbedResult, error)
}

type RerankerClient interface {
	Rerank(query string, passages []string) (*RerankResult, error)
}

type ReaderClient interface {
	Answer(question string, context string) (*Answer, error)
}
