# COSP - 腾讯云 COS 图片上传工具

`cosp` 是一个用于上传图片到腾讯云对象存储 (COS) 的命令行工具。支持文件上传、剪切板图片上传、文件列表查看和删除等功能。

## 功能特性

- ✅ **文件上传**: 上传本地图片文件到腾讯云 COS
- ✅ **剪切板上传**: 直接上传剪切板中的图片（支持截图和 SVG 文本）
- ✅ **文件列表**: 列出 COS 中的文件，支持分页和前缀过滤
- ✅ **文件删除**: 根据文件名删除 COS 中的文件
- ✅ **多平台支持**: 支持 macOS、Linux 和 Windows
- ✅ **自动重命名**: 使用时间戳自动生成文件名，避免重名冲突
- ✅ **文件类型检测**: 仅允许上传图片文件和 SVG 文件

## 安装

### 从 github 下载

可以从 [Release](https://github.com/bwangelme/cosp/releases) 页面获取对应平台的可执行文件

### 使用 Go 安装

```bash
go install github.com/bwangelme/cosp@latest
```

**注意**: 确保您的 `$GOPATH/bin` 或 `$HOME/go/bin` 在您的 `$PATH` 中。

## 配置

在使用前，您需要配置腾讯云 COS 的访问凭证：

1. 在用户主目录下创建 `.cos.conf` 文件：
   ```bash
   touch ~/.cos.conf
   ```

2. 编辑配置文件，参考 `example.cos.conf` 的格式：
   ```ini
   [common]
   secret_id = your-secret-id
   secret_key = your-secret-key
   bucket = your-bucket-name
   region = ap-beijing
   max_thread = 5
   part_size = 1
   retry = 5
   timeout = 60
   schema = https
   verify = md5
   anonymous = False
   ```

### 配置说明

可以参考腾讯云文档, 查看如何获取 `secret_id` 和 `secret_key`

https://cloud.tencent.com/document/product/436/7751

- `secret_id`: 腾讯云 API 密钥 ID
- `secret_key`: 腾讯云 API 密钥 Key
- `bucket`: COS 存储桶名称
- `region`: COS 存储桶所在地域（如：ap-beijing、ap-shanghai）
- `max_thread`: 最大并发线程数（默认 5）
- `part_size`: 分块上传大小（默认 1MB）
- `retry`: 重试次数（默认 5）
- `timeout`: 超时时间（默认 60秒）
- `schema`: 协议类型（默认 https）
- `verify`: 校验方式（默认 md5）
- `anonymous`: 是否匿名访问（默认 False）

## 使用方法

### 1. 查看版本信息

```bash
cosp version
```

输出示例：
```
COSP - 腾讯云 COS 图片上传工具
Version: 1.0.0
Build Time: 2024-01-15T10:30:00Z
Git Commit: abc1234
Go Version: go1.21.0
OS/Arch: darwin/arm64
```

### 2. 上传本地图片文件

```bash
cosp upload /path/to/image.jpg
```

支持的图片格式：PNG、JPEG、GIF、BMP、TIFF 等

### 3. 上传剪切板中的图片

```bash
cosp paste
```

支持的格式：
- **普通图片格式**：PNG、JPEG、GIF、BMP、TIFF 等
- **矢量图片格式**：SVG

#### 平台支持

- **macOS**: 使用 `Cmd+Shift+Ctrl+4` 截图到剪切板，然后运行命令。也支持复制 SVG 文本内容后运行命令
- **Linux**: 需要安装 `xclip`，使用 `xclip -selection clipboard -t image/png < image.png` 复制图片。也支持复制 SVG 文本内容后运行命令
- **Windows**: 支持 base64 文本方式。也支持复制 SVG 文本内容后运行命令

#### SVG 支持说明

工具可以智能识别剪切板中的 SVG 内容：
1. 复制任何包含 SVG 代码的文本（从网页、编辑器等）
2. 运行 `cosp paste` 命令
3. 工具会自动检测 SVG 格式并上传

SVG 检测条件：
- 包含 `<svg` 标签
- 包含 SVG 命名空间或 XML 声明
- 是有效的 SVG 文本内容

### 4. 列出 COS 中的文件

```bash
# 列出前 20 个文件
cosp list

# 列出前 50 个文件
cosp list --max-keys 50

# 列出以 "images/" 开头的文件
cosp list --prefix images/

# 从指定文件开始列出
cosp list --marker file.txt

# 从第 10 个文件开始列出
cosp list --marker 10
```

### 5. 删除文件

```bash
# 删除单个文件
cosp delete filename.jpg

# 删除多个文件
cosp delete file1.jpg file2.png file3.gif
```

## 命令详细说明

### `cosp upload`

上传指定路径的图片到腾讯云 COS。

**语法**: `cosp upload <filepath>`

**参数**:
- `<filepath>`: 要上传的图片文件路径

**示例**:
```bash
cosp upload ~/Pictures/screenshot.png
cosp upload /tmp/image.jpg
```

### `cosp paste`

上传剪切板中的图片到腾讯云 COS。

**语法**: `cosp paste`

**支持的格式**:
- **普通图片格式**: PNG、JPEG、GIF、BMP、TIFF 等
- **矢量图片格式**: SVG

**平台差异**:
- **macOS**: 自动检测剪切板中的图片数据，也支持 SVG 文本
- **Linux**: 需要 `xclip` 工具支持，也支持 SVG 文本
- **Windows**: 支持 base64 格式的图片数据，也支持 SVG 文本

**示例**:
```bash
# 先截图到剪切板，然后运行
cosp paste

# 复制 SVG 文本到剪切板，然后运行
cosp paste
```

**SVG 使用说明**:
1. 从任何地方复制 SVG 代码（网页源码、编辑器等）
2. 运行 `cosp paste` 命令
3. 工具会自动识别并上传 SVG 文件

### `cosp list`

列出腾讯云 COS bucket 中的文件。

**语法**: `cosp list [flags]`

**参数**:
- `--max-keys`: 最大返回文件数（默认 20）
- `--prefix`: 文件名前缀过滤
- `--marker`: 分页标记，可以是文件名或编号

**示例**:
```bash
cosp list
cosp list --max-keys 100
cosp list --prefix "2024-01"
cosp list --marker "2024-01-15-120000.png"
```

### `cosp delete`

根据文件名删除腾讯云 COS 中的文件。

**语法**: `cosp delete [文件名...]`

**参数**:
- `[文件名...]`: 要删除的文件名列表

**示例**:
```bash
cosp delete image.jpg
cosp delete file1.png file2.jpg
```

## 文件命名规则

上传的文件会自动重命名为时间戳格式，避免文件名冲突：

- 格式: `2006-01-02-150405.ext`
- 示例: `2024-01-15-143022.png`

## 获取帮助

```bash
# 查看主命令帮助
cosp --help

# 查看子命令帮助
cosp upload --help
cosp paste --help
cosp list --help
cosp delete --help
cosp version --help
```

## 构建与开发

### 本地构建

```bash
# 基本构建
make build

# 构建所有平台
make build-all

# 创建发布包
make release-local

# 查看版本信息
make version

# 查看所有可用命令
make help
```

### 版本管理

版本信息定义在 `version.go` 文件中：

```go
// Version 版本信息
const Version = "1.0.0"
```

构建时会自动注入以下信息：
- 构建时间
- Git 提交哈希
- Go 版本
- 操作系统和架构

### 自动发布

项目配置了 GitHub Actions 自动发布功能：

1. **触发条件**: 推送到 `master` 分支
2. **版本检测**: 自动从 `version.go` 读取版本号
3. **构建**: 自动构建多平台二进制文件
4. **发布**: 创建 GitHub Release 并上传构建文件

支持的平台：
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64, arm64)

### 发布新版本

1. 修改 `version.go` 中的版本号
2. 提交并推送到 `master` 分支
3. GitHub Actions 自动构建并发布新版本

## 常见问题

### Q1: 提示 "配置文件不存在"
**A**: 请确保在用户主目录下创建了 `.cos.conf` 文件，并配置了正确的腾讯云 COS 凭证。

### Q2: Linux 下提示 "xclip 命令不可用"
**A**: 请安装 xclip 工具：
```bash
# Ubuntu/Debian
sudo apt-get install xclip

# CentOS/RHEL
sudo yum install xclip
```

### Q3: 上传失败，提示权限不足
**A**: 请检查：
1. 腾讯云 API 密钥是否正确
2. 存储桶是否存在且有写权限
3. 地域配置是否正确

### Q4: 剪切板图片上传失败
**A**: 请确保：
1. 剪切板中确实有图片数据
2. 图片格式受支持
3. 对应平台的依赖工具已安装

## 许可证

本项目采用 MIT 许可证。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 更新日志

- v1.0.0: 初始版本，支持基本的上传、列表、删除功能
- 支持剪切板图片上传
- 支持多平台（macOS、Linux、Windows）
- 添加文件类型检测和自动重命名功能
