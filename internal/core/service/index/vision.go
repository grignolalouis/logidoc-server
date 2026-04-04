package index

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"trpc.group/trpc-go/trpc-agent-go/model"

	"github.com/logidoc/logidoc-server/pkg/ptr"
)

// CallVision sends an image to the VLM with a prompt and returns the text response.
func CallVision(ctx context.Context, llm model.Model, imageData []byte, prompt string) (string, *model.Usage, error) {
	msg := model.NewUserMessage(prompt)
	msg.AddImageData(imageData, "low", "png")

	req := &model.Request{
		Messages:         []model.Message{msg},
		GenerationConfig: model.GenerationConfig{Temperature: ptr.Float64(0.2), MaxTokens: ptr.Int(500), Stream: false},
	}

	respChan, err := llm.GenerateContent(ctx, req)
	if err != nil {
		return "", nil, err
	}

	var content string
	var usage *model.Usage
	for resp := range respChan {
		if resp.Error != nil {
			return "", nil, fmt.Errorf("vlm: %s", resp.Error.Message)
		}
		usage = resp.Usage
		if len(resp.Choices) > 0 {
			if resp.Choices[0].Delta.Content != "" {
				content += resp.Choices[0].Delta.Content
			} else if resp.Choices[0].Message.Content != "" {
				content = resp.Choices[0].Message.Content
			}
		}
	}

	return strings.TrimSpace(content), usage, nil
}

// RenderPageAsImage renders a PDF page as PNG using pdftoppm.
func RenderPageAsImage(pdfPath string, page int) ([]byte, error) {
	return renderPagePNG(pdfPath, page)
}

// ocrVLM scans pages with little or no text and uses the vision model to extract content.
func ocrVLM(ctx context.Context, llm model.Model, pages *Pages, logger *slog.Logger) {
	for i, text := range pages.Content {
		if len(strings.TrimSpace(text)) >= 50 {
			continue
		}

		imgData, err := renderPagePNG(pages.PDFPath, i+1)
		if err != nil {
			continue
		}

		content, _, err := CallVision(ctx, llm, imgData, "Extract all text from this page. Preserve structure, headings, tables, and lists. Output the text only.")
		if err != nil {
			logger.Warn("vlm ocr failed", "page", i+1, "error", err)
			continue
		}

		if len(content) > len(text) {
			pages.Content[i] = content
			logger.Debug("vlm ocr extracted", "page", i+1, "chars", len(content))
		}
	}
}
