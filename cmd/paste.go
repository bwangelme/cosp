package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"bytes"
	"encoding/base64"
	"os/exec"
	"runtime"

	"github.com/atotto/clipboard"
	"github.com/h2non/filetype"
	"github.com/spf13/cobra"
	"github.com/tencentyun/cos-go-sdk-v5"
)

var PasteCmd = &cobra.Command{
	Use:   "paste",
	Short: "上传剪切板中的图片到腾讯云 COS",
	Run: func(cmd *cobra.Command, args []string) {
		var b []byte
		var err error
		if runtime.GOOS == "darwin" {
			// macOS 下用 pbpaste 获取图片二进制
			b, err = exec.Command("pbpaste").Output()
			if err != nil {
				log.Fatalf("读取剪切板失败: %v", err)
			}
		} else {
			// 其他平台降级为 base64 或文本
			data, err2 := clipboard.ReadAll()
			if err2 != nil {
				log.Fatalf("读取剪切板失败: %v", err2)
			}
			b, err = base64.StdEncoding.DecodeString(data)
			if err != nil {
				b = []byte(data)
			}
		}
		if len(b) == 0 {
			log.Fatalf("剪切板为空")
		}
		if !filetype.IsImage(b) {
			log.Fatalf("剪切板内容不是图片，仅支持图片上传")
		}
		bucketURL := os.Getenv("COS_BUCKET_URL")
		secretID := os.Getenv("COS_SECRETID")
		secretKey := os.Getenv("COS_SECRETKEY")
		if bucketURL == "" || secretID == "" || secretKey == "" {
			log.Fatalf("请设置 COS_BUCKET_URL, COS_SECRETID, COS_SECRETKEY 环境变量")
		}
		u, err := url.Parse(bucketURL)
		if err != nil {
			log.Fatalf("COS_BUCKET_URL 解析失败: %v", err)
		}
		baseURL := &cos.BaseURL{BucketURL: u}
		client := cos.NewClient(baseURL, &http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  secretID,
				SecretKey: secretKey,
			},
		})
		ext, _ := filetype.Get(b)
		objectKey := fmt.Sprintf("paste_%d.%s", os.Getpid(), ext.Extension)
		_, err = client.Object.Put(context.Background(), objectKey, bytes.NewReader(b), nil)
		if err != nil {
			log.Fatalf("上传失败: %v", err)
		}
		fmt.Printf("上传成功: %s/%s\n", bucketURL, objectKey)
	},
}
