.PHONY: build install clean test fmt help build-all release-local

# 获取版本信息
VERSION := $(shell grep 'const Version = ' version.go | sed 's/.*"\(.*\)".*/\1/')
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS := -ldflags "-s -w -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

# 编译可执行文件为 cosp
build:
	go build ${LDFLAGS} -o cosp .

# 将执行文件 cosp 复制到 ~/bin 目录下
install: build
	@mkdir -p ~/bin
	cp cosp ~/bin/cosp
	@echo "cosp installed to ~/bin/cosp"

# 构建所有平台的二进制文件
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p dist
	
	# Linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o dist/cosp-v${VERSION}-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o dist/cosp-v${VERSION}-linux-arm64 .
	
	# macOS
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o dist/cosp-v${VERSION}-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o dist/cosp-v${VERSION}-darwin-arm64 .
	
	# Windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o dist/cosp-v${VERSION}-windows-amd64.exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build ${LDFLAGS} -o dist/cosp-v${VERSION}-windows-arm64.exe .
	
	@echo "All binaries built in dist/ directory"

# 本地发布（创建压缩包）
release-local: build-all
	@echo "Creating release archives..."
	cd dist && \
	for file in cosp-v${VERSION}-*; do \
		if [[ $$file == *.exe ]]; then \
			zip "$${file%.exe}.zip" "$$file"; \
		else \
			tar -czf "$$file.tar.gz" "$$file"; \
		fi; \
	done
	@echo "Release archives created in dist/ directory"

# 清理编译生成的文件
clean:
	rm -f cosp
	rm -rf dist/

# 运行测试
test:
	go test ./...

# 格式化代码
fmt:
	go fmt ./...

# 显示版本信息
version:
	@echo "Version: ${VERSION}"
	@echo "Build Time: ${BUILD_TIME}"
	@echo "Git Commit: ${GIT_COMMIT}"

# 显示帮助信息
help:
	@echo "Available commands:"
	@echo "  build        - Compile executable as 'cosp' with version info"
	@echo "  install      - Install executable to ~/bin"
	@echo "  build-all    - Build for all supported platforms"
	@echo "  release-local - Build all platforms and create archives"
	@echo "  clean        - Remove compiled files and dist directory"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  version      - Show version information"
	@echo "  help         - Show this help message"
