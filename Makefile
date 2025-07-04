.PHONY: build install

# 编译可执行文件为 cosp
build:
	go build -o cosp .

# 将执行文件 cosp 复制到 ~/bin 目录下
install: build
	@mkdir -p ~/bin
	cp cosp ~/bin/cosp
	@echo "cosp installed to ~/bin/cosp"

# 清理编译生成的文件
clean:
	rm -f cosp

# 运行测试
test:
	go test ./...

# 格式化代码
fmt:
	go fmt ./...

# 显示帮助信息
help:
	@echo "Available commands:"
	@echo "  build   - Compile executable as 'cosp'"
	@echo "  install - Install executable to ~/bin"
	@echo "  clean   - Remove compiled files"
	@echo "  test    - Run tests"
	@echo "  fmt     - Format code"
	@echo "  help    - Show this help message"
