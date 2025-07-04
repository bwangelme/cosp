package main

import (
	"fmt"
	"os"

	"cos/cmd"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "cosp",
		Short: "腾讯云 COS 图片上传工具",
		Long: `COSP - 腾讯云 COS 图片上传工具

这是一个用于上传图片到腾讯云对象存储 (COS) 的命令行工具。
支持文件上传、剪切板图片上传、文件列表查看和删除等功能。

使用前请先配置您的腾讯云 COS 凭证：
  1. 在用户主目录下创建 .cos.conf 文件
  2. 参考 example.cos.conf 文件配置相关参数

示例:
  cosp upload image.jpg    # 上传本地图片文件
  cosp paste              # 上传剪切板中的图片
  cosp list               # 列出 COS 中的文件
  cosp delete file.jpg    # 删除指定文件`,
	}

	// 添加版本命令
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Long:  "显示 COSP 工具的版本信息，包括版本号、构建时间、Git提交哈希等",
		Run: func(cmd *cobra.Command, args []string) {
			PrintVersion()
		},
	}

	rootCmd.AddCommand(cmd.PasteCmd)
	rootCmd.AddCommand(cmd.UploadCmd)
	rootCmd.AddCommand(cmd.ListCmd)
	rootCmd.AddCommand(cmd.DeleteCmd)
	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
