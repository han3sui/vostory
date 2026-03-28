package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// NormalizeLoudness 使用 ffmpeg loudnorm 滤镜对 WAV 音频进行 EBU R128 响度归一化。
// targetLUFS 推荐值: -16.0（有声书标准）。
// 归一化失败时调用方应降级使用原始音频，不阻断业务流程。
func NormalizeLoudness(inputData []byte, targetLUFS float64) ([]byte, error) {
	if len(inputData) == 0 {
		return inputData, nil
	}

	tmpDir, err := os.MkdirTemp("", "audio-norm-*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inPath := filepath.Join(tmpDir, "input.wav")
	outPath := filepath.Join(tmpDir, "output.wav")

	if err := os.WriteFile(inPath, inputData, 0644); err != nil {
		return nil, fmt.Errorf("写入临时文件失败: %w", err)
	}

	filter := fmt.Sprintf("loudnorm=I=%.1f:TP=-1.5:LRA=11:print_format=summary", targetLUFS)

	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", inPath,
		"-af", filter,
		outPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg 归一化失败: %w\n输出: %s", err, string(output))
	}

	result, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("读取归一化音频失败: %w", err)
	}

	return result, nil
}
