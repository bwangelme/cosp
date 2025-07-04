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

	logger "github.com/bwangelme/cosp/log"
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
		textContent, textErr := clipboard.ReadAll()
		if textErr == nil && len(textContent) > 0 {
			logger.L.Debugf("读取到剪切板文本内容，长度: %d 字节", len(textContent))
			logger.L.Debugf("剪切板内容预览: %s", func() string {
				if len(textContent) > 100 {
					return textContent[:100] + "..."
				}
				return textContent
			}())

			// 检查是否为 SVG 格式
			if isSVGContent(textContent) {
				fmt.Println("✅ 检测到 SVG 格式内容")
				b = []byte(textContent)
				isSVG = true
				fileExtension = "svg"
			} else {
				logger.L.Debug("SVG 检测失败，尝试按图片格式处理")
			}
		} else {
			logger.L.Debugf("读取剪切板文本失败: %v", textErr)
		}

		// 如果不是 SVG，按照原有逻辑处理图片
		if !isSVG {
			switch runtime.GOOS {
			case "darwin":
				// macOS 下使用 AppleScript 读取图片
				logger.L.Debug("使用 macOS AppleScript 方式读取剪切板图片")
				b, err = readImageFromClipboardMacOS()
				if err != nil {
					logger.L.Errorf("macOS AppleScript 读取失败: %v", err)
				}

			case "linux":
				// Linux 下使用 xclip 读取图片
				fmt.Println("使用 Linux xclip 读取图片...")
				logger.L.Debug("使用 Linux xclip 方式读取剪切板图片")
				b, err = readImageFromClipboardLinux()
				if err != nil {
					logger.L.Errorf("Linux xclip 读取失败: %v", err)
					log.Fatalf("读取剪切板失败: %v", err)
				}

			default:
				// 其他平台降级为 base64 或文本
				fmt.Println("使用通用文本方式读取...")
				logger.L.Debug("使用通用文本方式处理剪切板内容")
				if textErr != nil {
					logger.L.Errorf("通用文本方式读取失败: %v", textErr)
					log.Fatalf("读取剪切板失败: %v", textErr)
				}

				// 尝试解析为 base64
				b, err = base64.StdEncoding.DecodeString(textContent)
				if err != nil {
					fmt.Println("内容不是 base64 格式，使用原始文本数据")
					logger.L.Debug("内容不是 base64 格式，直接使用原始文本数据")
					// 如果不是 base64，直接使用原始数据
					b = []byte(textContent)
				} else {
					fmt.Println("检测到 base64 格式内容")
					logger.L.Debug("成功解析 base64 格式内容")
				}
			}

			if len(b) == 0 {
				logger.L.Error("剪切板内容为空")
				log.Fatalf("剪切板为空")
			}

			// 验证是否为图片格式
			if !filetype.IsImage(b) {
				fmt.Printf("❌ 剪切板内容不是图片格式，内容前 100 字符: %s\n",
					func() string {
						if len(textContent) > 100 {
							return textContent[:100] + "..."
						}
						return textContent
					}())
				logger.L.Errorf("剪切板内容不是有效的图片格式，数据长度: %d", len(b))
				log.Fatalf("剪切板内容不是图片或 SVG，仅支持图片和 SVG 上传")
			}

			// 检测文件类型
			ext, err := filetype.Get(b)
			if err != nil {
				logger.L.Errorf("检测文件类型失败: %v", err)
				log.Fatalf("检测文件类型失败: %v", err)
			}
			fileExtension = ext.Extension
			logger.L.Debugf("检测到图片格式: %s", fileExtension)
		}

		logger.L.Debugf("准备上传 %s 文件，数据大小: %d 字节", func() string {
			if isSVG {
				return "SVG"
			}
			return "图片"
		}(), len(b))

		// 使用新的客户端初始化方式
		logger.L.Debug("开始初始化腾讯云 COS 客户端")
		client, bucketURL, err := pkg.NewClientWithFallback()
		if err != nil {
			logger.L.Errorf("创建 COS 客户端失败: %v", err)
			log.Fatalf("创建COS客户端失败: %v", err)
		}
		logger.L.Debugf("成功连接到 COS，bucket URL: %s", bucketURL)

		// 使用时间戳格式生成文件名
		timestamp := time.Now().Format("2006-01-02-150405")
		objectKey := fmt.Sprintf("%s.%s", timestamp, fileExtension)
		logger.L.Debugf("生成文件名: %s，准备开始上传", objectKey)

		_, err = client.Object.Put(context.Background(), objectKey, bytes.NewReader(b), nil)
		if err != nil {
			logger.L.Errorf("上传到 COS 失败: %v", err)
			log.Fatalf("上传失败: %v", err)
		}
		fmt.Printf("✅ 上传成功: %s/%s\n", bucketURL, objectKey)
		logger.L.Debugf("成功上传文件: %s/%s，文件大小: %d 字节", bucketURL, objectKey, len(b))
	},
}

// isSVGContent 检查文本内容是否为 SVG 格式
func isSVGContent(content string) bool {
	// 移除前后空白字符
	content = strings.TrimSpace(content)

	// 检查是否为空
	if len(content) == 0 {
		logger.L.Debug("SVG 检测: 内容为空")
		return false
	}

	// 转换为小写进行比较
	lowerContent := strings.ToLower(content)
	logger.L.Debugf("SVG 检测: 开始检测内容，长度: %d", len(content))

	// 检查是否包含 <svg 标签（这是最基本的要求）
	svgTagExists := strings.Contains(lowerContent, "<svg")
	logger.L.Debugf("SVG 检测: <svg 标签存在: %v", svgTagExists)
	if !svgTagExists {
		logger.L.Debug("SVG 检测: 未找到 <svg 标签，不是 SVG 格式")
		return false
	}

	// 检查是否以 <?xml 开头（可选的 XML 声明）
	xmlStart := strings.HasPrefix(lowerContent, "<?xml")
	logger.L.Debugf("SVG 检测: XML 声明存在: %v", xmlStart)

	// 检查是否包含 SVG 命名空间
	svgNamespace := strings.Contains(lowerContent, "http://www.w3.org/2000/svg") ||
		strings.Contains(lowerContent, `xmlns="http://www.w3.org/2000/svg"`) ||
		strings.Contains(lowerContent, "xmlns:svg") ||
		strings.Contains(lowerContent, "xmlns=")
	logger.L.Debugf("SVG 检测: SVG 命名空间存在: %v", svgNamespace)

	// 检查是否以 </svg> 结束
	svgEnd := strings.HasSuffix(lowerContent, "</svg>")
	logger.L.Debugf("SVG 检测: </svg> 结束标签存在: %v", svgEnd)

	// 检查是否包含常见的 SVG 元素
	commonSVGElements := []string{"<path", "<rect", "<circle", "<ellipse", "<line", "<polygon", "<polyline", "<g", "<text", "<defs"}
	hasCommonElements := false
	foundElements := []string{}
	for _, element := range commonSVGElements {
		if strings.Contains(lowerContent, element) {
			hasCommonElements = true
			foundElements = append(foundElements, element)
		}
	}
	logger.L.Debugf("SVG 检测: 常见 SVG 元素存在: %v, 找到的元素: %v", hasCommonElements, foundElements)

	// 宽松的检测条件：
	// 1. 必须包含 <svg 标签
	// 2. 满足以下任一条件：
	//    - 有 SVG 命名空间
	//    - 有 XML 声明和结束标签
	//    - 有结束标签和常见 SVG 元素
	//    - 有 XML 声明和常见 SVG 元素
	result := svgTagExists &&
		(svgNamespace ||
			(xmlStart && svgEnd) ||
			(svgEnd && hasCommonElements) ||
			(xmlStart && hasCommonElements))

	logger.L.Debugf("SVG 检测: 主要条件检测结果: %v", result)

	// 如果以上都不满足，但包含 <svg 标签，也认为是 SVG（最宽松的检测）
	if !result && svgTagExists {
		// 检查内容长度，如果太短可能不是真正的 SVG
		if len(content) > 20 {
			result = true
			logger.L.Debug("SVG 检测: 通过宽松条件检测（包含 <svg 标签且长度 > 20）")
		} else {
			logger.L.Debug("SVG 检测: 内容太短，可能不是真正的 SVG")
		}
	}

	logger.L.Debugf("SVG 检测: 最终结果: %v", result)
	return result
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
