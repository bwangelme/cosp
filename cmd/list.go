package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"cos/pkg"

	"github.com/spf13/cobra"
	"github.com/tencentyun/cos-go-sdk-v5"
)

var (
	maxKeys int
	prefix  string
	marker  string
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出腾讯云 COS bucket 中的文件",
	Long: `列出腾讯云 COS bucket 中的文件，支持分页和前缀过滤。

示例:
  cos list                    # 列出前20个文件
  cos list --max-keys 50      # 列出前50个文件
  cos list --prefix images/   # 列出以 "images/" 开头的文件
  cos list --marker file.txt  # 从指定文件开始列出
  cos list --marker 10        # 从第10个文件开始列出`,
	Run: func(cmd *cobra.Command, args []string) {
		// 创建 COS 客户端
		client, bucketURL, err := pkg.NewClientWithFallback()
		if err != nil {
			log.Fatalf("创建COS客户端失败: %v", err)
		}

		// 处理 marker 参数：如果是数字，则从临时文件中读取对应的文件名
		var actualMarker string
		if marker != "" {
			if markerNum, err := strconv.Atoi(marker); err == nil && markerNum > 0 {
				// marker 是数字，从临时文件中读取对应的文件名
				actualMarker, err = getFileNameFromTempFile(markerNum)
				if err != nil {
					log.Fatalf("无法从临时文件中找到第%d个文件: %v", markerNum, err)
				}
				fmt.Printf("使用编号 %d，对应的文件名为: %s\n", markerNum, actualMarker)
			} else {
				// marker 是文件名，直接使用
				actualMarker = marker
				fmt.Printf("使用文件名作为 marker: %s\n", actualMarker)
			}
		}

		// 设置列表选项
		opts := &cos.BucketGetOptions{
			MaxKeys: maxKeys,
		}

		if prefix != "" {
			opts.Prefix = prefix
		}

		if actualMarker != "" {
			opts.Marker = actualMarker
		}

		// 获取文件列表
		result, _, err := client.Bucket.Get(context.Background(), opts)
		if err != nil {
			log.Fatalf("获取文件列表失败: %v", err)
		}

		// 检查是否有文件
		if len(result.Contents) == 0 {
			fmt.Println("没有找到文件")
			return
		}

		// 创建表格输出
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		// 输出表头
		fmt.Fprintln(w, "编号\t文件名\t大小\t最后修改时间\t文件地址")
		fmt.Fprintln(w, "----\t----\t----\t--------\t----")

		// 计算起始编号
		startIndex := 1
		if actualMarker != "" {
			// 如果有 marker，需要计算起始编号
			startIndex, _ = getStartIndexFromTempFile(actualMarker)
		}

		// 准备要写入临时文件的数据
		var tempFileData []string

		// 输出文件信息
		for i, obj := range result.Contents {
			// 解析时间
			lastModified, err := time.Parse(time.RFC3339, obj.LastModified)
			if err != nil {
				lastModified = time.Time{}
			}

			// 格式化文件大小
			sizeStr := formatSize(obj.Size)

			// 格式化时间
			timeStr := lastModified.Format("2006-01-02 15:04:05")

			// 构建完整的文件地址
			fileURL := fmt.Sprintf("%s/%s", bucketURL, obj.Key)

			// 当前文件的编号
			currentIndex := startIndex + i

			// 输出带编号的文件信息
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", currentIndex, obj.Key, sizeStr, timeStr, fileURL)

			// 准备写入临时文件的数据
			tempFileData = append(tempFileData, fmt.Sprintf("%d\t%s", currentIndex, obj.Key))
		}

		w.Flush()

		// 将文件名和编号存储到临时文件中
		err = saveToTempFile(tempFileData, true) // 每次都覆盖临时文件
		if err != nil {
			log.Printf("保存临时文件失败: %v", err)
		}

		// 输出分页信息
		fmt.Printf("\n总共 %d 个文件（按创建时间反向排序）", len(result.Contents))
		if result.IsTruncated {
			nextIndex := startIndex + len(result.Contents)
			fmt.Printf("，还有更多文件，使用 --marker %d 继续查看", nextIndex)
		}
		fmt.Println()
	},
}

// formatSize 格式化文件大小
func formatSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	} else {
		return fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024))
	}
}

// getTempFilePath 获取临时文件路径
func getTempFilePath() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join("/tmp", fmt.Sprintf("%s_coslist", currentUser.Uid)), nil
}

// saveToTempFile 将文件名和编号保存到临时文件中
func saveToTempFile(data []string, isFirstPage bool) error {
	tempFilePath, err := getTempFilePath()
	if err != nil {
		return err
	}

	var file *os.File
	if isFirstPage {
		// 如果是第一页，创建或覆盖文件
		file, err = os.Create(tempFilePath)
	} else {
		// 如果不是第一页，追加到文件
		file, err = os.OpenFile(tempFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}
	if err != nil {
		return err
	}
	defer file.Close()

	for _, line := range data {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return nil
}

// getFileNameFromTempFile 从临时文件中根据编号获取文件名
func getFileNameFromTempFile(index int) (string, error) {
	tempFilePath, err := getTempFilePath()
	if err != nil {
		return "", err
	}

	file, err := os.Open(tempFilePath)
	if err != nil {
		return "", fmt.Errorf("无法打开临时文件，请先运行 cos list 命令")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 {
			if lineIndex, err := strconv.Atoi(parts[0]); err == nil && lineIndex == index {
				return parts[1], nil
			}
		}
	}

	return "", fmt.Errorf("未找到编号为 %d 的文件", index)
}

// getStartIndexFromTempFile 从临时文件中根据文件名获取起始编号
func getStartIndexFromTempFile(fileName string) (int, error) {
	tempFilePath, err := getTempFilePath()
	if err != nil {
		return 1, err
	}

	file, err := os.Open(tempFilePath)
	if err != nil {
		return 1, nil // 如果临时文件不存在，从1开始
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 && parts[1] == fileName {
			if index, err := strconv.Atoi(parts[0]); err == nil {
				return index + 1, nil // 返回下一个位置的索引
			}
		}
	}

	return 1, nil // 如果没找到，从1开始
}

func init() {
	// 添加命令行参数
	ListCmd.Flags().IntVarP(&maxKeys, "max-keys", "n", 20, "最大返回文件数量")
	ListCmd.Flags().StringVarP(&prefix, "prefix", "p", "", "文件名前缀过滤")
	ListCmd.Flags().StringVarP(&marker, "marker", "m", "", "从指定文件名或编号开始列出")
}
