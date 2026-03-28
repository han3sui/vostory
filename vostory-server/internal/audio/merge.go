package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MergeAudioFiles 使用 ffmpeg concat demuxer 将多个音频文件按顺序合并为一个文件。
// inputPaths 必须按播放顺序排列，outputPath 为输出文件路径。
// format 支持 "wav" 和 "mp3"。
func MergeAudioFiles(inputPaths []string, outputPath string, format string) error {
	if len(inputPaths) == 0 {
		return fmt.Errorf("没有可合并的音频文件")
	}

	if len(inputPaths) == 1 {
		srcExt := strings.ToLower(filepath.Ext(inputPaths[0]))
		dstExt := "." + format
		if srcExt == dstExt {
			data, err := os.ReadFile(inputPaths[0])
			if err != nil {
				return fmt.Errorf("读取音频文件失败: %w", err)
			}
			return os.WriteFile(outputPath, data, 0644)
		}
	}

	tmpDir, err := os.MkdirTemp("", "audio-merge-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	var listLines []string
	for _, p := range inputPaths {
		absPath, _ := filepath.Abs(p)
		escaped := strings.ReplaceAll(absPath, "'", "'\\''")
		listLines = append(listLines, fmt.Sprintf("file '%s'", escaped))
	}
	listPath := filepath.Join(tmpDir, "filelist.txt")
	if err := os.WriteFile(listPath, []byte(strings.Join(listLines, "\n")), 0644); err != nil {
		return fmt.Errorf("写入文件列表失败: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	args := []string{"-y", "-f", "concat", "-safe", "0", "-i", listPath}
	switch format {
	case "mp3":
		args = append(args, "-ac", "1", "-ar", "24000", "-codec:a", "libmp3lame", "-b:a", "128k")
	default:
		// WAV: 源文件已经是 24kHz/mono/pcm_s16le，直接 stream copy 避免重编码
		args = append(args, "-c", "copy")
	}
	args = append(args, outputPath)

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg 合并失败: %w\n输出: %s", err, string(output))
	}

	return nil
}
