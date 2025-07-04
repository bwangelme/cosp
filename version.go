package main

import (
	"fmt"
	"runtime"
)

// Version 版本信息
const Version = "1.0.0"

// BuildTime 构建时间，会在编译时通过 -ldflags 注入
var BuildTime = "unknown"

// GitCommit Git提交哈希，会在编译时通过 -ldflags 注入
var GitCommit = "unknown"

// PrintVersion 打印版本信息
func PrintVersion() {
	fmt.Printf("COSP - 腾讯云 COS 图片上传工具\n")
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Git Commit: %s\n", GitCommit)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
