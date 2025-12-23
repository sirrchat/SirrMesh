# SirrMesh 部署指南

## 目录

1. [系统要求](#系统要求)
2. [快速部署](#快速部署)
3. [手动部署](#手动部署)
4. [配置说明](#配置说明)
5. [服务管理](#服务管理)
6. [监控与维护](#监控与维护)
7. [故障排查](#故障排查)
8. [安全最佳实践](#安全最佳实践)

---

## 系统要求

### 硬件要求

```yaml
最低配置:
  CPU: 2 核
  RAM: 2GB
  存储: 20GB SSD
  网络: 100Mbps

推荐配置:
  CPU: 4 核
  RAM: 4GB
  存储: 100GB SSD
  网络: 1Gbps
```

### 软件要求

```yaml
操作系统:
  - Ubuntu 20.04+
  - Debian 11+
  - CentOS 8+
  - macOS 12+

依赖:
  - Go 1.24+（仅编译需要）
  - Git
  - Make
```

### 端口要求

| 端口 | 服务 | 说明 |
|------|------|------|
| 25 | SMTP | 邮件接收（可选） |
| 587 | Submission | 邮件提交 |
| 465 | SMTPS | 加密邮件提交 |
| 993 | IMAPS | 加密 IMAP |
| 143 | IMAP | IMAP（可选） |
| 8825 | SMTP Alt | 替代 SMTP 端口 |

---

## 快速部署

### 一键部署脚本

使用自动化脚本快速部署：

```bash
# 下载并执行部署脚本
curl -sSL https://raw.githubusercontent.com/sirrchat/sirrmesh/main/start.sh | bash

# 或者下载后执行
wget https://raw.githubusercontent.com/sirrchat/sirrmesh/main/start.sh
chmod +x start.sh
sudo ./start.sh
```

脚本将自动：
1. 检测系统架构并下载正确的二进制文件
2. 初始化配置目录
3. 配置 DNS 和 TLS 证书（支持 15 种 DNS 提供商）
4. 创建并启动 systemd 服务

---

## 手动部署

### 1. 系统准备

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装依赖
sudo apt install -y build-essential git curl wget

# 创建工作目录
export SIRRMESH_HOME="${SIRRMESH_HOME:-$HOME/.sirrmeshd}"
mkdir -p $SIRRMESH_HOME
```

### 2. 下载二进制文件

```bash
# 自动检测系统架构
get_system_arch() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)

    case "$arch" in
        x86_64|amd64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) arch="amd64" ;;
    esac

    echo "${os}-${arch}"
}

SYSTEM_ARCH=$(get_system_arch)
VERSION="v0.2.3"

# 下载对应版本
wget https://download.sirr.chat/sirrmeshd-${SYSTEM_ARCH}-${VERSION}
sudo mv sirrmeshd-${SYSTEM_ARCH}-${VERSION} /usr/local/bin/sirrmeshd
sudo chmod +x /usr/local/bin/sirrmeshd

# 验证安装
sirrmeshd --help
```

### 3. 从源码构建（可选）

```bash
# 克隆仓库
git clone https://github.com/sirrchat/sirrmesh.git
cd sirrmeshd

# 构建
make build

# 安装
sudo cp build/sirrmeshd /usr/local/bin/
```

---

## 配置说明

### 基本配置

创建 `$SIRRMESH_HOME/sirrmeshd.conf`:

```
# 域名配置
$(hostname) = mx1.example.com
$(primary_domain) = example.com
$(local_domains) = $(primary_domain)

# TLS 证书配置
tls {
    loader acme {
        hostname $(hostname)
        email postmaster@$(hostname)
        agreed
        challenge dns-01
        dns cloudflare {
            api_token YOUR_CLOUDFLARE_API_TOKEN
        }
    }
}

# 存储配置
storage.imapsql local_mailboxes {
    driver sqlite3
    dsn $SIRRMESH_HOME/imapsql.db
}

# 认证配置
auth.pass_table local_auth {
    table file $SIRRMESH_HOME/users
}

# SMTP 服务
smtp tcp://0.0.0.0:8825 {
    hostname $(hostname)

    limits {
        all rate 20 1s
        all concurrency 10
    }

    dmarc yes

    check {
        require_mx_record
        dkim
        spf
    }

    source $(local_domains) {
        deliver_to &local_mailboxes
    }
}

# Submission 服务
submission tls://0.0.0.0:465 tcp://0.0.0.0:587 {
    hostname $(hostname)

    auth &local_auth

    source $(local_domains) {
        default_destination {
            modify {
                dkim $(primary_domain) $(local_domains) default
            }
            deliver_to &remote_queue
        }
    }
}

# IMAP 服务
imap tls://0.0.0.0:993 tcp://0.0.0.0:143 {
    auth &local_auth
    storage &local_mailboxes
}
```

### DNS 提供商配置

支持的 DNS 提供商及配置：

```
# Cloudflare
dns cloudflare {
    api_token YOUR_API_TOKEN
}

# Amazon Route53
dns route53 {
    access_key_id YOUR_ACCESS_KEY
    secret_access_key YOUR_SECRET_KEY
}

# DigitalOcean
dns digitalocean {
    api_token YOUR_API_TOKEN
}

# Google Cloud DNS
dns googleclouddns {
    service_account_json /path/to/service-account.json
}

# Vultr
dns vultr {
    api_token YOUR_API_TOKEN
}

# Hetzner
dns hetzner {
    api_token YOUR_API_TOKEN
}
```

---

## 服务管理

### Systemd 服务配置

#### 创建邮件服务

```bash
sudo tee /etc/systemd/system/sirrmeshd.service > /dev/null <<EOF
[Unit]
Description=Sirr Mesh Server
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Environment="SIRRMESH_HOME=/root/.sirrmeshd"
ExecStart=/usr/local/bin/sirrmeshd run
Restart=always
RestartSec=3
LimitNOFILE=65535
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
```

#### 启动和管理服务

```bash
# 重载 systemd 配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start sirrmeshd

# 设置开机自启
sudo systemctl enable sirrmeshd

# 查看服务状态
sudo systemctl status sirrmeshd

# 查看日志
sudo journalctl -u sirrmeshd -f

# 重启服务
sudo systemctl restart sirrmeshd
```

---

## 监控与维护

### 日志查看

```bash
# 查看实时日志
sudo journalctl -u sirrmeshd -f

# 查看最近的错误日志
sudo journalctl -u sirrmeshd -p err -n 100

# 查看今日日志
sudo journalctl -u sirrmeshd --since today
```

### 健康检查脚本

创建 `~/check_sirrmesh_health.sh`:

```bash
#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo "=== SirrMesh Health Check ==="
echo "Time: $(date)"
echo "=============================="

# 检查进程状态
if pgrep -x sirrmeshd > /dev/null; then
    echo -e "${GREEN}✓${NC} Process is running"
else
    echo -e "${RED}✗${NC} Process is NOT running"
    exit 1
fi

# 检查端口
for port in 587 993 8825; do
    if netstat -tlnp 2>/dev/null | grep -q ":$port "; then
        echo -e "${GREEN}✓${NC} Port $port is listening"
    else
        echo -e "${RED}✗${NC} Port $port is NOT listening"
    fi
done

# 检查磁盘空间
DISK_USAGE=$(df -h $SIRRMESH_HOME | awk 'NR==2 {print $5}' | tr -d '%')
if [ "$DISK_USAGE" -lt 80 ]; then
    echo -e "${GREEN}✓${NC} Disk usage: ${DISK_USAGE}%"
else
    echo -e "${RED}⚠${NC} Disk usage high: ${DISK_USAGE}%"
fi

echo "=============================="
```

---

## 故障排查

### 常见问题

#### 1. 服务无法启动

```bash
# 检查配置文件语法
sirrmeshd run --config $SIRRMESH_HOME/sirrmeshd.conf

# 查看详细错误
sudo journalctl -u sirrmeshd -n 50
```

#### 2. TLS 证书问题

```bash
# 检查 DNS 配置
sirrmeshd dns check

# 手动测试 DNS 挑战
sirrmeshd dns export
```

#### 3. 邮件无法发送/接收

```bash
# 检查端口是否开放
netstat -tlnp | grep -E '25|587|993'

# 检查防火墙
sudo ufw status
```

---

## 安全最佳实践

### 防火墙配置

```bash
# 基础防火墙规则
sudo ufw default deny incoming
sudo ufw default allow outgoing

# SSH 访问
sudo ufw allow 22/tcp

# 邮件服务端口
sudo ufw allow 587/tcp comment 'Submission'
sudo ufw allow 993/tcp comment 'IMAPS'
sudo ufw allow 8825/tcp comment 'SMTP Alt'

# 启用防火墙
sudo ufw enable
```

### 安全检查清单

- [ ] 防火墙规则已正确配置
- [ ] SSH 使用密钥认证
- [ ] TLS 证书已配置并自动续期
- [ ] 系统自动安全更新已启用
- [ ] 日志轮转已配置
- [ ] 定期备份已实施

### 备份建议

```bash
# 备份配置和数据
BACKUP_DIR="/backup/sirrmeshd/$(date +%Y%m%d)"
mkdir -p $BACKUP_DIR

# 备份配置
cp $SIRRMESH_HOME/sirrmeshd.conf $BACKUP_DIR/

# 备份数据库
cp $SIRRMESH_HOME/*.db $BACKUP_DIR/

# 保留最近 7 天的备份
find /backup/sirrmeshd -type d -mtime +7 -exec rm -rf {} +
```

---

## 有用的命令

```bash
# 用户管理
sirrmeshd creds list
sirrmeshd creds add user@example.com

# 邮箱管理
sirrmeshd imap-acct list
sirrmeshd imap-mboxes list user@example.com

# DNS 配置
sirrmeshd dns config
sirrmeshd dns check
sirrmeshd dns export
sirrmeshd dns ip

# 密码哈希
sirrmeshd hash --password mypassword
```

---

## 相关资源

- **官方网站**: https://sirr.chat
- **GitHub 仓库**: https://github.com/sirrchat/SirrMesh

---

*最后更新时间：2025年12月*
