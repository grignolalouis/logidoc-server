package index

import (
	"context"
	"fmt"
	"log/slog"

	"trpc.group/trpc-go/trpc-agent-go/model"
)

const defaultMaxPagesPerNode = 20


// SubdivideLargeNodes splits nodes that span too many pages by asking the LLM
// to detect sub-structure within the page range.
func SubdivideLargeNodes(ctx context.Context, nodes []treeNode, pages *Pages, llm model.Model, logger *slog.Logger, maxPages int) []treeNode {
	if maxPages <= 0 {
		maxPages = defaultMaxPagesPerNode
	}

	result := make([]treeNode, len(nodes))
	for i, n := range nodes {
		result[i] = n

		pageSpan := n.EndPage - n.StartPage + 1
		if pageSpan > maxPages && len(n.Children) == 0 {
			logger.Debug("subdividing large node", "id", n.ID, "pages", pageSpan)

			children, err := detectSubStructure(ctx, llm, pages, n.StartPage, n.EndPage, n.Level+1)
			if err != nil {
				logger.Warn("subdivision failed, keeping as-is", "id", n.ID, "error", err)
			} else if len(children) > 1 {
				result[i].Children = children
			}
		}

		result[i].Children = SubdivideLargeNodes(ctx, result[i].Children, pages, llm, logger, maxPages)
	}
	return result
}

func detectSubStructure(ctx context.Context, llm model.Model, pages *Pages, start, end, level int) ([]treeNode, error) {
	text := pages.Read(start, end)
	if len(text) > 15000 {
		text = string([]rune(text)[:15000])
	}

	prompt := fmt.Sprintf(SubdividePrompt, level, start, end, text)

	req := &model.Request{
		Messages: []model.Message{
			model.NewUserMessage(prompt),
		},
		GenerationConfig: model.GenerationConfig{
			Temperature: float64Ptr(0.1),
			MaxTokens:   intPtr(4000),
			Stream:      false,
		},
	}

	respChan, err := llm.GenerateContent(ctx, req)
	if err != nil {
		return nil, err
	}

	var content string
	for resp := range respChan {
		if resp.Error != nil {
			return nil, fmt.Errorf("llm: %s", resp.Error.Message)
		}
		if len(resp.Choices) > 0 {
			content += resp.Choices[0].Message.Content
			content += resp.Choices[0].Delta.Content
		}
	}

	sections, err := parseJSONArray[FlatSection](content)
	if err != nil || len(sections) <= 1 {
		return nil, fmt.Errorf("no sub-structure found")
	}

	totalPages := end
	subTree := BuildTree(sections, totalPages)
	return subTree, nil
}
