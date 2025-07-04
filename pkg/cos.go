package pkg

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/tencentyun/cos-go-sdk-v5"
	"gopkg.in/ini.v1"
)

// COSConfig 配置结构体
type COSConfig struct {
	SecretID  string
	SecretKey string
	Bucket    string
	Region    string
	MaxThread int
	PartSize  int
	Retry     int
	Timeout   int
	Schema    string
	Verify    string
	Anonymous bool
}

// DefaultConfig 返回默认配置
func DefaultConfig() *COSConfig {
	return &COSConfig{
		MaxThread: 5,
		PartSize:  1,
		Retry:     5,
		Timeout:   60,
		Schema:    "https",
		Verify:    "md5",
		Anonymous: false,
	}
}

// LoadConfig 从配置文件加载配置
func LoadConfig() (*COSConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("无法获取用户主目录: %v", err)
	}

	configPath := filepath.Join(homeDir, ".cos.conf")

	// 如果配置文件不存在，返回错误
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	cfg, err := ini.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	common := cfg.Section("common")
	config := DefaultConfig()

	// 读取必填字段
	config.SecretID = common.Key("secret_id").String()
	config.SecretKey = common.Key("secret_key").String()
	config.Bucket = common.Key("bucket").String()
	config.Region = common.Key("region").String()

	// 验证必填字段
	if config.SecretID == "" || config.SecretKey == "" || config.Bucket == "" || config.Region == "" {
		return nil, fmt.Errorf("配置文件缺少必要字段: secret_id, secret_key, bucket, region")
	}

	// 读取可选字段
	if val, err := common.Key("max_thread").Int(); err == nil {
		config.MaxThread = val
	}
	if val, err := common.Key("part_size").Int(); err == nil {
		config.PartSize = val
	}
	if val, err := common.Key("retry").Int(); err == nil {
		config.Retry = val
	}
	if val, err := common.Key("timeout").Int(); err == nil {
		config.Timeout = val
	}
	if val := common.Key("schema").String(); val != "" {
		config.Schema = val
	}
	if val := common.Key("verify").String(); val != "" {
		config.Verify = val
	}
	if val, err := common.Key("anonymous").Bool(); err == nil {
		config.Anonymous = val
	}

	return config, nil
}

// NewClient 创建COS客户端
func NewClient(config *COSConfig) (*cos.Client, error) {
	// 构建 bucket URL
	bucketURL := fmt.Sprintf("%s://%s.cos.%s.myqcloud.com", config.Schema, config.Bucket, config.Region)

	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("解析 bucket URL 失败: %v", err)
	}

	baseURL := &cos.BaseURL{BucketURL: u}

	// 创建客户端
	client := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.SecretID,
			SecretKey: config.SecretKey,
		},
	})

	return client, nil
}

// GetBucketURL 获取bucket URL
func (c *COSConfig) GetBucketURL() string {
	return fmt.Sprintf("%s://%s.cos.%s.myqcloud.com", c.Schema, c.Bucket, c.Region)
}

// NewClientWithFallback 从配置文件创建客户端
func NewClientWithFallback() (*cos.Client, string, error) {
	// 从配置文件读取
	config, err := LoadConfig()
	if err != nil {
		return nil, "", err
	}

	client, err := NewClient(config)
	if err != nil {
		return nil, "", err
	}
	return client, config.GetBucketURL(), nil
}
