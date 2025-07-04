package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"cos/pkg"

	"github.com/h2non/filetype"
	"github.com/spf13/cobra"
)

var UploadCmd = &cobra.Command{
	Use:   "upload <filepath>",
	Short: "上传指定路径的图片到腾讯云 COS",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatalf("无法打开文件: %v", err)
		}
		defer file.Close()

		buf := make([]byte, 261)
		_, err = file.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatalf("读取文件失败: %v", err)
		}
		if !filetype.IsImage(buf) {
			log.Fatalf("只支持图片类型文件上传")
		}
		file.Seek(0, io.SeekStart)

		// 使用新的客户端初始化方式
		client, bucketURL, err := pkg.NewClientWithFallback()
		if err != nil {
			log.Fatalf("创建COS客户端失败: %v", err)
		}

		// 获取原文件的扩展名
		originalExt := filepath.Ext(filePath)

		// 使用时间戳格式生成新的文件名，保留原文件的扩展名
		timestamp := time.Now().Format("2006-01-02-150405")
		objectKey := timestamp + originalExt

		_, err = client.Object.Put(context.Background(), objectKey, file, nil)
		if err != nil {
			log.Fatalf("上传失败: %v", err)
		}
		fmt.Printf("上传成功: %s/%s\n", bucketURL, objectKey)
	},
}
