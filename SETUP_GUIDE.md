# MailChat Chain 项目搭建教程

## 📋 项目简介

MailChat Chain 是一个企业级邮件服务器，支持 SMTP/IMAP 协议，并集成了区块链认证特性。主要特点包括：

- 完整的 SMTP/IMAP 邮件服务实现
- 区块链钱包签名认证（基于以太坊）
- 多数据库支持（SQLite、PostgreSQL、MySQL）
- 自动化 TLS 证书管理（ACME）
- 支持多种 DNS 提供商
- Prometheus 监控集成
- 模块化设计，支持扩展

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

```bash
# Download and execute installation script
curl -sSL https://raw.githubusercontent.com/your-repo/mail-chat-chain/main/start.sh | bash
```

**脚本功能：**
- 自动检测系统架构（x86_64、arm64）
- 下载对应平台的二进制文件
- 配置 DNS 设置
- 生成 systemd 服务文件

### 方法二：从源码构建

#### 1. 克隆项目

```bash
# Clone repository
git clone https://github.com/your-repo/mail-chat-chain.git
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

# Output: build/mailchatd
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
docker build -f Dockerfile.build -t mailchatd:latest .

# Extract binary from container
docker create --name temp mailchatd:latest
docker cp temp:/mailchatd ./build/mailchatd
docker rm temp
```

---

## ⚙️ 配置项目

### 1. 设置环境变量

```bash
# Set MAILCHAT_HOME directory
export MAILCHAT_HOME=$HOME/.mailchatd

# Create directory
mkdir -p $MAILCHAT_HOME
```

**永久配置：**
```bash
# Add to ~/.bashrc or ~/.zshrc
echo 'export MAILCHAT_HOME=$HOME/.mailchatd' >> ~/.bashrc
source ~/.bashrc
```

### 2. 生成配置文件

```bash
# Generate default configuration
./build/mailchatd config init

# Configuration file location: $MAILCHAT_HOME/mailchatd.conf
```

### 3. 配置数据库

#### 使用 SQLite（开发环境推荐）

```conf
# Edit $MAILCHAT_HOME/mailchatd.conf
storage.imapsql local_mailboxes {
    driver sqlite3
    dsn $MAILCHAT_HOME/imapsql.db
}
```

#### 使用 PostgreSQL（生产环境推荐）

```conf
storage.imapsql local_mailboxes {
    driver postgres
    dsn postgres://username:password@localhost/mailchatdb?sslmode=disable
}
```

**创建 PostgreSQL 数据库：**
```bash
# Create database
psql -U postgres -c "CREATE DATABASE mailchatdb;"
psql -U postgres -c "CREATE USER mailchat WITH PASSWORD 'your_password';"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE mailchatdb TO mailchat;"
```

#### 使用 MySQL

```conf
storage.imapsql local_mailboxes {
    driver mysql
    dsn mailchat:password@tcp(localhost:3306)/mailchatdb?parseTime=true
}
```

**创建 MySQL 数据库：**
```bash
# Create database
mysql -u root -p -e "CREATE DATABASE mailchatdb CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -u root -p -e "CREATE USER 'mailchat'@'localhost' IDENTIFIED BY 'your_password';"
mysql -u root -p -e "GRANT ALL PRIVILEGES ON mailchatdb.* TO 'mailchat'@'localhost';"
mysql -u root -p -e "FLUSH PRIVILEGES;"
```

### 4. 配置存储后端

#### 文件系统存储

```conf
storage.blob.fs local_fs {
    path $MAILCHAT_HOME/blobs
}
```

#### S3 兼容存储

```conf
storage.blob.s3 s3_storage {
    bucket_name mailchat-storage
    region us-east-1
    access_key_id YOUR_ACCESS_KEY
    secret_access_key YOUR_SECRET_KEY
}
```

### 5. 配置认证方式

#### 密码表认证（默认）

```bash
# Create user credentials
./build/mailchatd creds create user@example.com

# Generate password hash
./build/mailchatd hash mypassword
```

#### 区块链钱包认证（特色功能）

```conf
auth.pass_blockchain {
    network mainnet  # or testnet
    chain_id 1       # Ethereum mainnet
}
```

**使用说明：**
用户使用以太坊钱包私钥签名消息来完成认证，无需传统密码。

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

## 🎯 运行项目

### 启动邮件服务

```bash
# Run in foreground
./build/mailchatd run

# Run with custom config
./build/mailchatd run --config /path/to/config.conf

# Run with debug logging
./build/mailchatd run --debug
```

### 使用 systemd 管理（Linux）

```bash
# Generate systemd service file
./build/mailchatd systemd generate > /tmp/mailchatd.service
sudo mv /tmp/mailchatd.service /etc/systemd/system/

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable mailchatd
sudo systemctl start mailchatd

# Check status
sudo systemctl status mailchatd

# View logs
sudo journalctl -u mailchatd -f
```

### 后台运行（macOS/Linux）

```bash
# Run in background using nohup
nohup ./build/mailchatd run > $MAILCHAT_HOME/mailchatd.log 2>&1 &

# Check process
ps aux | grep mailchatd

# Stop process
kill $(pgrep mailchatd)
```

---

## 🔧 常用命令

### DNS 管理

```bash
# Configure DNS provider
./build/mailchatd dns setup

# Test DNS configuration
./build/mailchatd dns verify
```

### 用户管理

```bash
# Create user credentials
./build/mailchatd creds create user@example.com

# List users
./build/mailchatd creds list

# Delete user
./build/mailchatd creds delete user@example.com

# Generate password hash
./build/mailchatd hash mypassword
```

### IMAP 账户管理

```bash
# Create IMAP account
./build/mailchatd imap-acct create user@example.com

# List IMAP accounts
./build/mailchatd imap-acct list

# Delete IMAP account
./build/mailchatd imap-acct delete user@example.com
```

### IMAP 邮箱管理

```bash
# Create mailbox
./build/mailchatd imap-mboxes create user@example.com Inbox

# List mailboxes
./build/mailchatd imap-mboxes list user@example.com

# Delete mailbox
./build/mailchatd imap-mboxes delete user@example.com Trash
```

### IMAP 消息管理

```bash
# List messages in mailbox
./build/mailchatd imap-msgs list user@example.com Inbox

# Delete message
./build/mailchatd imap-msgs delete user@example.com Inbox <message_id>
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
make build && ./build/mailchatd run
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
sudo journalctl -u mailchatd -f

# If using nohup
tail -f $MAILCHAT_HOME/mailchatd.log
```

---

## 🔒 安全配置

### TLS/SSL 证书

#### 使用 ACME 自动获取（Let's Encrypt）

```conf
tls {
    acme_enabled true
    acme_email admin@example.com
    acme_storage $MAILCHAT_HOME/acme
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

# Configure in mailchatd.conf
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
psql -h localhost -U mailchat -d mailchatdb

# Test MySQL connection
mysql -h localhost -u mailchat -p mailchatdb

# Check SQLite file permissions
ls -la $MAILCHAT_HOME/imapsql.db
```

#### 3. 权限问题

```bash
# Fix MAILCHAT_HOME permissions
chmod -R 755 $MAILCHAT_HOME
chown -R $USER:$USER $MAILCHAT_HOME
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
./build/mailchatd run --debug --log-level=debug

# Enable stack traces
./build/mailchatd run --debug --enable-trace
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
./build/mailchatd --help

# Show command-specific help
./build/mailchatd run --help
./build/mailchatd dns --help
./build/mailchatd creds --help
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
./build/mailchatd version

# 2. Verify configuration
./build/mailchatd config verify

# 3. Test SMTP connection
telnet localhost 25

# 4. Test IMAP connection
telnet localhost 143

# 5. Check Prometheus metrics
curl http://localhost:9090/metrics
```

如果所有测试通过，说明项目已成功搭建！

---

**版本信息:** 0.3.1
**最后更新:** 2025-12-17
