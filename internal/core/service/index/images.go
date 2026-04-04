package index

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"trpc.group/trpc-go/trpc-agent-go/model"
)

const describeImagePrompt = "Describe this image concisely in 1-2 sentences. Focus on data, labels, structure, and what information it conveys."

func enrichWithImages(ctx context.Context, llm model.Model, pages *Pages, metrics *Metrics, logger *slog.Logger) error {
	if pages.PDFPath == "" {
		return nil
	}

	images, err := extractImages(pages.PDFPath)
	if err != nil {
		return fmt.Errorf("extract images: %w", err)
	}
	if len(images) == 0 {
		return nil
	}

	logger.Info("describing images", "count", len(images))

	pages.Images = make([][]PageImage, pages.Count())
	for i := range images {
		desc, usage, err := CallVision(ctx, llm, images[i].Data, describeImagePrompt)
		if err != nil {
			logger.Warn("image description failed", "page", images[i].PageNum, "error", err)
			continue
		}
		images[i].Description = desc
		metrics.AddVision(usage)

		idx := images[i].PageNum - 1
		if idx >= 0 && idx < len(pages.Images) {
			pages.Images[idx] = append(pages.Images[idx], images[i])
		}
	}

	return nil
}

func extractImages(pdfPath string) ([]PageImage, error) {
	if _, err := exec.LookPath("pdfimages"); err != nil {
		return nil, nil
	}

	tmpDir, err := os.MkdirTemp("", "logidoc-images-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	prefix := filepath.Join(tmpDir, "img")
	if err := exec.Command("pdfimages", "-png", pdfPath, prefix).Run(); err != nil {
		return nil, fmt.Errorf("pdfimages: %w", err)
	}

	listOut, err := exec.Command("pdfimages", "-list", pdfPath).Output()
	if err != nil {
		return nil, fmt.Errorf("pdfimages list: %w", err)
	}

	pageMap := parseImageList(string(listOut))

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("read images dir: %w", err)
	}
	var images []PageImage
	for idx, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(tmpDir, entry.Name()))
		if err != nil {
			continue
		}
		if len(data) < 1000 {
			continue
		}
		pageNum := 1
		if p, ok := pageMap[idx]; ok {
			pageNum = p
		}
		images = append(images, PageImage{Data: data, PageNum: pageNum})
	}

	return images, nil
}

func parseImageList(listOutput string) map[int]int {
	result := make(map[int]int)
	lines := strings.Split(listOutput, "\n")
	for i, line := range lines {
		if i < 2 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		page, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		result[i-2] = page
	}
	return result
}
