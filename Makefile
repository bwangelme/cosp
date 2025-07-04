.PHONY: build install

# 编译可执行文件为 cos
build:
	go build -o cos .

# 将执行文件 cos 复制到 ~/bin 目录下
install: build
	@mkdir -p ~/bin
	cp cos ~/bin/cos
	@echo "cos installed to ~/bin/cos"

# 清理编译生成的文件
clean:
	rm -f cos

# 运行测试
test:
	go test ./...

# 格式化代码
fmt:
	go fmt ./...

# 显示帮助信息
help:
	@echo "Available commands:"
	@echo "  build   - Compile executable as 'cos'"
	@echo "  install - Install executable to ~/bin"
	@echo "  clean   - Remove compiled files"
	@echo "  test    - Run tests"
	@echo "  fmt     - Format code"
	@echo "  help    - Show this help message" 