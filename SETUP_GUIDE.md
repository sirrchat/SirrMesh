# SirrChat 节点搭建教程

## 📋 项目简介

SirrChat 是一个去中心化的加密通讯系统，允许任何人搭建和运行属于自己的通讯节点。通过本教程，你可以部署自己的 SirrChat 节点，拥有完全的数据控制权和隐私保护。

**为什么要搭建自己的 SirrChat 节点？**
- 🔐 **数据主权** - 所有通讯数据存储在你自己的服务器上
- 🌐 **去中心化** - 不依赖任何第三方服务提供商
- 🔒 **隐私保护** - 端到端加密，完全掌控自己的通讯
- ⛓️ **区块链认证** - 基于以太坊钱包的去中心化身份验证
- 🚀 **独立运营** - 构建专属的通讯网络

**核心特性：**
- 完整的 SMTP/IMAP 协议支持，兼容主流邮件客户端
- 区块链钱包签名认证，无需传统密码系统
- 多数据库支持（SQLite、PostgreSQL、MySQL）
- 自动化 TLS 证书管理（ACME）
- 支持多种 DNS 提供商自动配置
- Prometheus 监控集成
- 模块化设计，易于扩展和定制

---

## 🖥️ 系统要求

### 操作系统
- Ubuntu 20.04+
- Debian 11+
- CentOS 8+
- macOS 12+

### 硬件配置
- **CPU**: 2核或以上
- **内存**: 2GB（推荐 4GB）
- **存储**: 20GB SSD
- **网络**: 100Mbps

### 软件依赖
- **Go**: 1.24 或更高版本
- **Git**: 用于克隆代码
- **Make**: 构建工具
- **GCC**: CGO 编译需要（SQLite 支持）

---

## 🚀 快速开始

### 方法一：使用一键安装脚本（推荐）

如果你已获得 SirrChat 的安装脚本，可以使用以下命令快速搭建节点：

```bash
# Download and execute installation script
curl -sSL <YOUR_SCRIPT_URL>/start.sh | bash
```

**脚本功能：**
- 自动检测系统架构（x86_64、arm64）
- 下载对应平台的二进制文件
- 配置 DNS 设置
- 生成 systemd 服务文件
- 一键完成节点部署

### 方法二：从源码构建（完全掌控）

通过源码构建可以让你完全了解和掌控节点的每一个细节。

#### 1. 克隆项目

```bash
# Clone repository from your source
git clone <YOUR_REPOSITORY_URL>
cd mail-chat-chain
```

#### 2. 安装 Go 环境

**Ubuntu/Debian:**
```bash
# Install Go 1.24
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

**macOS:**
```bash
# Using Homebrew
brew install go@1.24

# Verify installation
go version
```

#### 3. 安装系统依赖

**Ubuntu/Debian:**
```bash
# Install build tools
sudo apt-get update
sudo apt-get install -y build-essential git make
```

**macOS:**
```bash
# Install Xcode Command Line Tools
xcode-select --install
```

#### 4. 下载 Go 依赖

```bash
# Download all dependencies
go mod download

# Verify dependencies
go mod verify
```

---

## 🔨 构建项目

### 本地构建

```bash
# Build for current platform
make build

# Output: build/sirrchatd
```

### 交叉编译

```bash
# Build for Linux AMD64
make build-linux

# Build for Linux ARM64
make build-linux-arm64

# Build for macOS AMD64
make build-darwin

# Build for macOS ARM64
make build-darwin-arm64
```

### 安装到系统路径

```bash
# Install to /usr/local/bin (requires sudo)
sudo make install
```

### 使用 Docker 构建

```bash
# Build Docker image
docker build -f Dockerfile.build -t sirrchatd:latest .

# Extract binary from container
docker create --name temp sirrchatd:latest
docker cp temp:/sirrchatd ./build/sirrchatd
docker rm temp
```

---

## ⚙️ 配置节点

配置你的 SirrChat 节点，使其能够独立运行并为你的用户提供服务。所有配置数据都存储在你自己的服务器上，完全由你掌控。

### 1. 设置环境变量

```bash
# Set SIRRCHAT_HOME directory
export SIRRCHAT_HOME=$HOME/.sirrchatd

# Create directory
mkdir -p $SIRRCHAT_HOME
```

**永久配置：**
```bash
# Add to ~/.bashrc or ~/.zshrc
echo 'export SIRRCHAT_HOME=$HOME/.sirrchatd' >> ~/.bashrc
source ~/.bashrc
```

### 2. 生成配置文件

```bash
# Generate default configuration
./build/sirrchatd config init

# Configuration file location: $SIRRCHAT_HOME/sirrchatd.conf
```

### 3. 配置数据库

#### 使用 SQLite（开发环境推荐）

```conf
# Edit $SIRRCHAT_HOME/sirrchatd.conf
storage.imapsql local_mailboxes {
    driver sqlite3
    dsn $SIRRCHAT_HOME/imapsql.db
}
```

#### 使用 PostgreSQL（生产环境推荐）

```conf
storage.imapsql local_mailboxes {
    driver postgres
    dsn postgres://username:password@localhost/sirrchatdb?sslmode=disable
}
```

**创建 PostgreSQL 数据库：**
```bash
# Create database
psql -U postgres -c "CREATE DATABASE sirrchatdb;"
psql -U postgres -c "CREATE USER sirrchat WITH PASSWORD 'your_password';"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE sirrchatdb TO sirrchat;"
```

#### 使用 MySQL

```conf
storage.imapsql local_mailboxes {
    driver mysql
    dsn sirrchat:password@tcp(localhost:3306)/sirrchatdb?parseTime=true
}
```

**创建 MySQL 数据库：**
```bash
# Create database
mysql -u root -p -e "CREATE DATABASE sirrchatdb CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -u root -p -e "CREATE USER 'sirrchat'@'localhost' IDENTIFIED BY 'your_password';"
mysql -u root -p -e "GRANT ALL PRIVILEGES ON sirrchatdb.* TO 'sirrchat'@'localhost';"
mysql -u root -p -e "FLUSH PRIVILEGES;"
```

### 4. 配置存储后端

#### 文件系统存储

```conf
storage.blob.fs local_fs {
    path $SIRRCHAT_HOME/blobs
}
```

#### S3 兼容存储

```conf
storage.blob.s3 s3_storage {
    bucket_name sirrchat-storage
    region us-east-1
    access_key_id YOUR_ACCESS_KEY
    secret_access_key YOUR_SECRET_KEY
}
```

### 5. 配置认证方式

#### 密码表认证（默认）

```bash
# Create user credentials
./build/sirrchatd creds create user@example.com

# Generate password hash
./build/sirrchatd hash mypassword
```

#### 区块链钱包认证（去中心化身份）

```conf
auth.pass_blockchain {
    network mainnet  # or testnet
    chain_id 1       # Ethereum mainnet
}
```

**使用说明：**
这是 SirrChat 的核心特色功能，用户使用以太坊钱包私钥签名消息来完成认证，实现真正的去中心化身份验证：
- 🔑 无需传统密码系统
- 🌐 基于区块链的去中心化身份
- 🔒 私钥由用户自己掌控
- ✅ 无需依赖中心化的身份认证服务

#### LDAP 认证

```conf
auth.ldap {
    url ldap://ldap.example.com:389
    base_dn dc=example,dc=com
    bind_dn cn=admin,dc=example,dc=com
    bind_password admin_password
}
```

---

## 🎯 运行节点

启动你的 SirrChat 节点后，它将成为去中心化通讯网络中的一个独立节点。你的节点将：
- 处理本节点用户的通讯请求
- 与其他 SirrChat 节点互联互通
- 完全由你控制和管理，不受任何第三方干预

### 启动 SirrChat 节点

```bash
# Run in foreground
./build/sirrchatd run

# Run with custom config
./build/sirrchatd run --config /path/to/config.conf

# Run with debug logging
./build/sirrchatd run --debug
```

### 使用 systemd 管理（Linux）

```bash
# Generate systemd service file
./build/sirrchatd systemd generate > /tmp/sirrchatd.service
sudo mv /tmp/sirrchatd.service /etc/systemd/system/

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable sirrchatd
sudo systemctl start sirrchatd

# Check status
sudo systemctl status sirrchatd

# View logs
sudo journalctl -u sirrchatd -f
```

### 后台运行（macOS/Linux）

```bash
# Run in background using nohup
nohup ./build/sirrchatd run > $SIRRCHAT_HOME/sirrchatd.log 2>&1 &

# Check process
ps aux | grep sirrchatd

# Stop process
kill $(pgrep sirrchatd)
```

---

## 🔧 常用命令

### DNS 管理

```bash
# Configure DNS provider
./build/sirrchatd dns setup

# Test DNS configuration
./build/sirrchatd dns verify
```

### 用户管理

```bash
# Create user credentials
./build/sirrchatd creds create user@example.com

# List users
./build/sirrchatd creds list

# Delete user
./build/sirrchatd creds delete user@example.com

# Generate password hash
./build/sirrchatd hash mypassword
```

### IMAP 账户管理

```bash
# Create IMAP account
./build/sirrchatd imap-acct create user@example.com

# List IMAP accounts
./build/sirrchatd imap-acct list

# Delete IMAP account
./build/sirrchatd imap-acct delete user@example.com
```

### IMAP 邮箱管理

```bash
# Create mailbox
./build/sirrchatd imap-mboxes create user@example.com Inbox

# List mailboxes
./build/sirrchatd imap-mboxes list user@example.com

# Delete mailbox
./build/sirrchatd imap-mboxes delete user@example.com Trash
```

### IMAP 消息管理

```bash
# List messages in mailbox
./build/sirrchatd imap-msgs list user@example.com Inbox

# Delete message
./build/sirrchatd imap-msgs delete user@example.com Inbox <message_id>
```

---

## 🧪 开发与测试

### 运行测试

```bash
# Run unit tests
make test

# Run tests with race detection
make test-race

# Run tests with coverage
make test-cover

# View coverage report (HTML)
go tool cover -html=coverage.out
```

### 代码质量检查

```bash
# Run linter
make lint

# Auto-fix linting issues
make lint-fix

# Format code
make format

# Security vulnerability check
make vulncheck
```

### 构建并运行

```bash
# Build and run in one step
make build && ./build/sirrchatd run
```

---

## 📊 监控与日志

### Prometheus 监控

**启用 Prometheus metrics：**
```conf
openmetrics tcp://0.0.0.0:9090 {
    enabled true
}
```

**访问 metrics 端点：**
```bash
# View metrics
curl http://localhost:9090/metrics
```

### 日志配置

```conf
# Log level: debug, info, warn, error
log {
    level info
    format json
    output stdout
}
```

**查看日志：**
```bash
# If using systemd
sudo journalctl -u sirrchatd -f

# If using nohup
tail -f $SIRRCHAT_HOME/sirrchatd.log
```

---

## 🔒 安全配置

### TLS/SSL 证书

#### 使用 ACME 自动获取（Let's Encrypt）

```conf
tls {
    acme_enabled true
    acme_email admin@example.com
    acme_storage $SIRRCHAT_HOME/acme
    dns_provider cloudflare
    dns_api_token YOUR_CLOUDFLARE_TOKEN
}
```

**支持的 DNS 提供商：**
Cloudflare、Route53、DigitalOcean、Google Cloud DNS、Vultr、Hetzner、Gandi、Namecheap 等

#### 使用自签名证书

```bash
# Generate self-signed certificate
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Configure in sirrchatd.conf
tls {
    cert_file /path/to/cert.pem
    key_file /path/to/key.pem
}
```

### 防火墙配置

```bash
# Allow SMTP ports
sudo ufw allow 25/tcp    # SMTP
sudo ufw allow 465/tcp   # SMTPS
sudo ufw allow 587/tcp   # Submission

# Allow IMAP ports
sudo ufw allow 143/tcp   # IMAP
sudo ufw allow 993/tcp   # IMAPS

# Enable firewall
sudo ufw enable
```

---

## 🐛 故障排查

### 常见问题

#### 1. 端口被占用

```bash
# Check port usage
sudo lsof -i :25
sudo lsof -i :143

# Kill process using port
sudo kill -9 $(lsof -t -i:25)
```

#### 2. 数据库连接失败

```bash
# Test PostgreSQL connection
psql -h localhost -U sirrchat -d sirrchatdb

# Test MySQL connection
mysql -h localhost -u sirrchat -p sirrchatdb

# Check SQLite file permissions
ls -la $SIRRCHAT_HOME/imapsql.db
```

#### 3. 权限问题

```bash
# Fix SIRRCHAT_HOME permissions
chmod -R 755 $SIRRCHAT_HOME
chown -R $USER:$USER $SIRRCHAT_HOME
```

#### 4. 依赖下载失败

```bash
# Use Go proxy
export GOPROXY=https://proxy.golang.org,direct

# Clean and retry
go clean -modcache
go mod download
```

### 调试模式

```bash
# Run with verbose logging
./build/sirrchatd run --debug --log-level=debug

# Enable stack traces
./build/sirrchatd run --debug --enable-trace
```

---

## 📚 相关文档

- [README.md](README.md) - 项目英文文档
- [README_ZH.md](README_ZH.md) - 项目中文文档
- [DEPLOYMENT.md](DEPLOYMENT.md) - 部署指南

---

## 🆘 获取帮助

### 查看命令帮助

```bash
# Show all available commands
./build/sirrchatd --help

# Show command-specific help
./build/sirrchatd run --help
./build/sirrchatd dns --help
./build/sirrchatd creds --help
```

### 问题反馈

如遇到问题，请提供以下信息：
- 操作系统版本
- Go 版本
- 完整错误日志
- 配置文件（隐藏敏感信息）

---

## ✅ 验证安装

完成安装后，执行以下检查：

```bash
# 1. Check binary version
./build/sirrchatd version

# 2. Verify configuration
./build/sirrchatd config verify

# 3. Test SMTP connection
telnet localhost 25

# 4. Test IMAP connection
telnet localhost 143

# 5. Check Prometheus metrics
curl http://localhost:9090/metrics
```

如果所有测试通过，说明你的 SirrChat 节点已成功搭建并运行！现在你拥有了一个完全属于自己的去中心化通讯节点。

---

**版本信息:** 0.3.1
**最后更新:** 2025-12-17
