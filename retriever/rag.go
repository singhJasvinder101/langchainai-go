package retriever

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/singhJasvinder101/agentic-go/llm"
)

const defaultRAGSystemPrompt = "Answer the question using only the provided context. " +
	"If the context doesn't contain the answer, say you don't know.\n\nContext:\n%s"

// defaultK is used when RAGChain.K is unset.
const defaultK = 4

// RAGChain retrieves context documents for a question and answers it with an
// llm.Provider. It implements Run(ctx, vars) (string, error) so it
// structurally satisfies chain.Chain without importing the chain package.
type RAGChain struct {
	Retriever    Retriever
	Provider     llm.Provider
	K            int
	SystemPrompt string
}

// NewRAGChain builds a RAGChain retrieving up to k documents per question. A
// k <= 0 defaults to 4.
func NewRAGChain(retriever Retriever, provider llm.Provider, k int) *RAGChain {
	if k <= 0 {
		k = defaultK
	}
	return &RAGChain{Retriever: retriever, Provider: provider, K: k}
}

// Run retrieves context for vars["question"] and answers it via Provider.
func (c *RAGChain) Run(ctx context.Context, vars map[string]any) (string, error) {
	if c.Retriever == nil {
		return "", errors.New("retriever: retriever is required")
	}
	if c.Provider == nil {
		return "", errors.New("retriever: provider is required")
	}
	question, _ := vars["question"].(string)
	if question == "" {
		return "", errors.New(`retriever: "question" variable is required`)
	}

	k := c.K
	if k <= 0 {
		k = defaultK
	}
	docs, err := c.Retriever.Retrieve(ctx, question, k)
	if err != nil {
		return "", fmt.Errorf("retriever: retrieve: %w", err)
	}

	var contextBuilder strings.Builder
	for i, doc := range docs {
		if i > 0 {
			contextBuilder.WriteString("\n---\n")
		}
		contextBuilder.WriteString(doc.Content)
	}

	system := c.SystemPrompt
	if system == "" {
		system = fmt.Sprintf(defaultRAGSystemPrompt, contextBuilder.String())
	} else {
		system = system + "\n\nContext:\n" + contextBuilder.String()
	}

	resp, err := c.Provider.Generate(ctx, &llm.GenerateRequest{
		Messages: []llm.Message{
			llm.SystemMessage(llm.TextPart(system)),
			llm.UserMessage(llm.TextPart(question)),
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Text(), nil
}
