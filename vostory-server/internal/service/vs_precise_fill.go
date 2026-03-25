package service

import (
	"context"
	"fmt"
	"strings"

	"iot-alert-center/internal/repository"
)

type VsPreciseFillService interface {
	AlignChapter(ctx context.Context, chapterID uint64) (int, error)
}

func NewVsPreciseFillService(
	service *Service,
	chapterRepo repository.VsChapterRepository,
	segmentRepo repository.VsScriptSegmentRepository,
) VsPreciseFillService {
	return &vsPreciseFillService{
		Service:     service,
		chapterRepo: chapterRepo,
		segmentRepo: segmentRepo,
	}
}

type vsPreciseFillService struct {
	*Service
	chapterRepo repository.VsChapterRepository
	segmentRepo repository.VsScriptSegmentRepository
}

// AlignChapter 将章节下所有脚本片段的 Content 对齐回章节原文。
// 对于每个片段，在原文中查找最佳匹配位置，用原文片段替换 LLM 输出。
// 替换后将原 Content 保存到 OriginalContent，用原文匹配结果覆盖 Content。
func (s *vsPreciseFillService) AlignChapter(ctx context.Context, chapterID uint64) (int, error) {
	chapter, err := s.chapterRepo.FindByID(ctx, chapterID)
	if err != nil {
		return 0, fmt.Errorf("章节不存在")
	}

	segments, err := s.segmentRepo.FindByChapterID(ctx, chapterID)
	if err != nil {
		return 0, fmt.Errorf("获取片段失败: %w", err)
	}

	if len(segments) == 0 {
		return 0, nil
	}

	originalText := chapter.Content
	alignedCount := 0
	loginName := ctx.Value("login_name").(string)

	searchFrom := 0

	for _, seg := range segments {
		content := strings.TrimSpace(seg.Content)
		if content == "" {
			continue
		}

		idx := strings.Index(originalText[searchFrom:], content)
		if idx >= 0 {
			matchStart := searchFrom + idx
			matchEnd := matchStart + len(content)
			matched := originalText[matchStart:matchEnd]

			if matched != seg.Content {
				seg.OriginalContent = seg.Content
				seg.Content = matched
				seg.UpdatedBy = loginName
				if err := s.segmentRepo.Update(ctx, seg); err != nil {
					return alignedCount, fmt.Errorf("更新片段失败: %w", err)
				}
				alignedCount++
			}
			searchFrom = matchEnd
			continue
		}

		bestIdx, bestLen := fuzzyMatch(originalText, searchFrom, content)
		if bestLen > len(content)/2 {
			matched := originalText[bestIdx : bestIdx+bestLen]
			seg.OriginalContent = seg.Content
			seg.Content = matched
			seg.UpdatedBy = loginName
			if err := s.segmentRepo.Update(ctx, seg); err != nil {
				return alignedCount, fmt.Errorf("更新片段失败: %w", err)
			}
			alignedCount++
			searchFrom = bestIdx + bestLen
		}
	}

	return alignedCount, nil
}

// fuzzyMatch 在 text[from:] 中查找与 pattern 最相似的子串，返回起始位置和匹配长度。
// 使用滑动窗口 + 公共子序列长度作为相似度度量。
func fuzzyMatch(text string, from int, pattern string) (bestStart, bestMatchLen int) {
	textRunes := []rune(text[from:])
	patternRunes := []rune(pattern)
	pLen := len(patternRunes)

	if pLen == 0 || len(textRunes) == 0 {
		return from, 0
	}

	windowSize := pLen + pLen/2
	if windowSize > len(textRunes) {
		windowSize = len(textRunes)
	}

	bestScore := 0

	for i := 0; i <= len(textRunes)-pLen/2; i++ {
		end := i + windowSize
		if end > len(textRunes) {
			end = len(textRunes)
		}
		window := textRunes[i:end]
		score := lcsLength(window, patternRunes)
		if score > bestScore {
			bestScore = score
			bestStart = from + i*3
			bestMatchLen = (end - i) * 3
		}
	}

	bestStart = from
	for i := 0; i <= len(textRunes)-pLen/2; i++ {
		end := i + windowSize
		if end > len(textRunes) {
			end = len(textRunes)
		}
		window := textRunes[i:end]
		score := lcsLength(window, patternRunes)
		if score == bestScore {
			byteStart := len(string(textRunes[:i]))
			byteEnd := len(string(textRunes[:end]))
			bestStart = from + byteStart
			bestMatchLen = byteEnd - byteStart
			break
		}
	}

	return bestStart, bestMatchLen
}

func lcsLength(a, b []rune) int {
	m, n := len(a), len(b)
	if m == 0 || n == 0 {
		return 0
	}

	prev := make([]int, n+1)
	curr := make([]int, n+1)

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				curr[j] = prev[j-1] + 1
			} else {
				curr[j] = prev[j]
				if curr[j-1] > curr[j] {
					curr[j] = curr[j-1]
				}
			}
		}
		prev, curr = curr, prev
		for k := range curr {
			curr[k] = 0
		}
	}

	return prev[n]
}
