package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"bytes"
	"encoding/base64"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/bwangelme/cosp/pkg"

	"github.com/atotto/clipboard"
	"github.com/h2non/filetype"
	"github.com/spf13/cobra"
)

var PasteCmd = &cobra.Command{
	Use:   "paste",
	Short: "上传剪切板中的图片到腾讯云 COS",
	Long: `上传剪切板中的图片到腾讯云 COS。

支持的格式：
- 普通图片格式：PNG、JPEG、GIF、BMP、TIFF 等
- 矢量图片格式：SVG

支持的平台：
- macOS: 使用 Cmd+Shift+Ctrl+4 截图到剪切板，或复制 SVG 文本
- Linux: 使用 xclip 复制图片到剪切板，或复制 SVG 文本
- Windows: 降级为 base64 文本方式，或复制 SVG 文本`,
	Run: func(cmd *cobra.Command, args []string) {
		var b []byte
		var err error
		var isSVG bool
		var fileExtension string

		// 首先尝试读取剪切板的文本内容，检查是否为 SVG
		fmt.Println("正在读取剪切板中的内容...")
		textContent, textErr := clipboard.ReadAll()
		if textErr == nil && len(textContent) > 0 {
			// 检查是否为 SVG 格式
			if isSVGContent(textContent) {
				fmt.Println("检测到 SVG 格式内容")
				b = []byte(textContent)
				isSVG = true
				fileExtension = "svg"
			}
		}

		// 如果不是 SVG，按照原有逻辑处理图片
		if !isSVG {
			fmt.Println("尝试读取剪切板中的图片...")

			switch runtime.GOOS {
			case "darwin":
				// macOS 下使用 AppleScript 读取图片
				b, err = readImageFromClipboardMacOS()
				if err != nil {
					log.Fatalf("读取剪切板失败: %v", err)
				}

			case "linux":
				// Linux 下使用 xclip 读取图片
				b, err = readImageFromClipboardLinux()
				if err != nil {
					log.Fatalf("读取剪切板失败: %v", err)
				}

			default:
				// 其他平台降级为 base64 或文本
				if textErr != nil {
					log.Fatalf("读取剪切板失败: %v", textErr)
				}

				// 尝试解析为 base64
				b, err = base64.StdEncoding.DecodeString(textContent)
				if err != nil {
					// 如果不是 base64，直接使用原始数据
					b = []byte(textContent)
				}
			}

			if len(b) == 0 {
				log.Fatalf("剪切板为空")
			}

			// 验证是否为图片格式
			if !filetype.IsImage(b) {
				log.Fatalf("剪切板内容不是图片或 SVG，仅支持图片和 SVG 上传")
			}

			// 检测文件类型
			ext, err := filetype.Get(b)
			if err != nil {
				log.Fatalf("检测文件类型失败: %v", err)
			}
			fileExtension = ext.Extension
		}

		fmt.Printf("成功读取到 %d 字节的%s数据\n", len(b), func() string {
			if isSVG {
				return "SVG"
			}
			return "图片"
		}())

		// 使用新的客户端初始化方式
		client, bucketURL, err := pkg.NewClientWithFallback()
		if err != nil {
			log.Fatalf("创建COS客户端失败: %v", err)
		}

		// 使用时间戳格式生成文件名
		timestamp := time.Now().Format("2006-01-02-150405")
		objectKey := fmt.Sprintf("%s.%s", timestamp, fileExtension)

		_, err = client.Object.Put(context.Background(), objectKey, bytes.NewReader(b), nil)
		if err != nil {
			log.Fatalf("上传失败: %v", err)
		}
		fmt.Printf("上传成功: %s/%s\n", bucketURL, objectKey)
	},
}

// isSVGContent 检查文本内容是否为 SVG 格式
func isSVGContent(content string) bool {
	// 移除前后空白字符
	content = strings.TrimSpace(content)

	// 检查是否为空
	if len(content) == 0 {
		return false
	}

	// 转换为小写进行比较
	lowerContent := strings.ToLower(content)

	// 检查是否以 <?xml 开头（可选的 XML 声明）
	xmlStart := strings.HasPrefix(lowerContent, "<?xml")

	// 检查是否包含 <svg 标签
	svgTagExists := strings.Contains(lowerContent, "<svg")

	// 检查是否包含 SVG 命名空间
	svgNamespace := strings.Contains(lowerContent, "xmlns") &&
		(strings.Contains(lowerContent, "http://www.w3.org/2000/svg") ||
			strings.Contains(lowerContent, "svg"))

	// 检查是否以 </svg> 结束
	svgEnd := strings.HasSuffix(lowerContent, "</svg>")

	// 必须包含 <svg 标签，并且满足以下条件之一：
	// 1. 有 XML 声明和 SVG 命名空间
	// 2. 有 SVG 命名空间和正确的结束标签
	// 3. 有 XML 声明和正确的结束标签
	return svgTagExists &&
		((xmlStart && svgNamespace) ||
			(svgNamespace && svgEnd) ||
			(xmlStart && svgEnd) ||
			(svgNamespace)) // 简化检查，只要有 svg 标签和命名空间即可
}

// readImageFromClipboardMacOS 从 macOS 剪切板读取图片
func readImageFromClipboardMacOS() ([]byte, error) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "clipboard_image_*")
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 尝试读取不同格式的图片
	formats := []string{"png", "tiff"}

	for _, format := range formats {
		// 使用 AppleScript 将剪切板图片写入临时文件
		script := fmt.Sprintf(`
			try
				set clipboardData to the clipboard as «class %s»
				set fileRef to (open for access POSIX file "%s" with write permission)
				write clipboardData to fileRef
				close access fileRef
			on error
				try
					close access fileRef
				end try
				return "error"
			end try
		`, getAppleScriptFormat(format), tmpFile.Name())

		cmd := exec.Command("osascript", "-e", script)
		output, err := cmd.Output()
		if err == nil && string(output) != "error\n" {
			// 读取临时文件内容
			data, err := os.ReadFile(tmpFile.Name())
			if err == nil && len(data) > 0 {
				return data, nil
			}
		}
	}

	return nil, fmt.Errorf("无法从剪切板读取图片数据，请确保剪切板中有图片")
}

// readImageFromClipboardLinux 从 Linux 剪切板读取图片
func readImageFromClipboardLinux() ([]byte, error) {
	// 检查 xclip 是否可用
	if _, err := exec.LookPath("xclip"); err != nil {
		return nil, fmt.Errorf("xclip 命令不可用，请安装 xclip: sudo apt-get install xclip 或 sudo yum install xclip")
	}

	// 尝试读取不同格式的图片
	formats := []string{"image/png", "image/jpeg", "image/gif", "image/bmp", "image/tiff"}

	for _, format := range formats {
		cmd := exec.Command("xclip", "-selection", "clipboard", "-t", format, "-o")
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			// 验证是否为有效图片
			if filetype.IsImage(output) {
				return output, nil
			}
		}
	}

	return nil, fmt.Errorf("无法从剪切板读取图片数据，请确保剪切板中有图片。\n\n使用示例：\n  xclip -selection clipboard -t image/png < image.png")
}

// getAppleScriptFormat 获取 AppleScript 格式标识符
func getAppleScriptFormat(format string) string {
	switch format {
	case "png":
		return "PNGf"
	case "tiff":
		return "TIFF"
	case "jpeg", "jpg":
		return "JPEG"
	default:
		return "PNGf"
	}
}
