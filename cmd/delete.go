package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/bwangelme/cosp/pkg"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete [文件名...]",
	Short: "根据文件名删除腾讯云 COS 中的文件",
	Long: `根据文件名删除腾讯云 COS 中的文件。

示例:
  cosp delete filename.txt         # 删除指定文件名的文件
  cosp delete file1.txt file2.jpg  # 删除多个文件`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// 创建 COS 客户端
		client, _, err := pkg.NewClientWithFallback()
		if err != nil {
			log.Fatalf("创建COS客户端失败: %v", err)
		}

		// 解析要删除的文件名
		var fileNames []string
		for _, arg := range args {
			fileNames = append(fileNames, arg)
			fmt.Printf("待删除文件: %s\n", arg)
		}

		if len(fileNames) == 0 {
			log.Fatalf("没有找到要删除的文件")
		}

		// 确认删除
		fmt.Printf("\n即将删除 %d 个文件:\n", len(fileNames))
		for _, fileName := range fileNames {
			fmt.Printf("  - %s\n", fileName)
		}

		fmt.Print("\n确认删除？(y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
			fmt.Println("取消删除操作")
			return
		}

		// 删除文件
		successCount := 0
		for _, fileName := range fileNames {
			fmt.Printf("正在删除: %s ... ", fileName)
			resp, err := client.Object.Delete(context.Background(), fileName)
			if err != nil {
				fmt.Printf("失败: %v\n", err)
			} else if resp.StatusCode != 200 && resp.StatusCode != 204 {
				// 删除成功通常返回200或204，其他状态码表示失败
				fmt.Printf("失败 (状态码: %d)\n", resp.StatusCode)
				if resp.Body != nil {
					bodyBytes, readErr := io.ReadAll(resp.Body)
					if readErr == nil {
						fmt.Printf("错误详情: %s\n", string(bodyBytes))
					}
					resp.Body.Close()
				}
			} else {
				fmt.Println("成功")
				successCount++
			}
		}

		fmt.Printf("\n删除完成，成功删除 %d 个文件\n", successCount)
	},
}
