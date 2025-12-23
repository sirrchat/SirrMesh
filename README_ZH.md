# SirrMesh

**SirrMesh** 是一个去中心化的加密通讯系统，允许任何人搭建和运行属于自己的通讯节点。通过集成区块链认证和企业级邮件协议，SirrMesh 为用户提供完全自主可控的通讯平台，实现真正的数据主权和隐私保护。

**搭建你自己的 SirrMesh 节点，掌控你的通讯数据。**

[![许可证](https://img.shields.io/badge/license-GPL%203.0-blue)](LICENSE)
[![Go版本](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org)

[English](README.md) | [中文](README_ZH.md)

## 为什么要搭建自己的 SirrMesh 节点？

- 🔐 **数据主权** - 所有通讯数据存储在你自己的服务器上
- 🌐 **去中心化** - 不依赖任何第三方服务提供商
- 🔒 **隐私保护** - 端到端加密，完全掌控自己的通讯
- ⛓️ **区块链认证** - 基于以太坊钱包的去中心化身份验证
- 🚀 **独立运营** - 构建专属的通讯网络

## 功能特性

### 核心能力
- **去中心化节点**: 完全自主的通讯节点，独立运行
- **区块链身份**: 基于 EVM 钱包签名的去中心化身份认证
- **企业级协议**: 完整的 SMTP/IMAP 支持，兼容主流邮件客户端
- **自动化证书**: 15+ 种 DNS 提供商的自动 TLS 证书集成
- **安全防护**: DKIM、SPF、DMARC 验证与信誉评分
- **灵活存储**: SQL 数据库后端（PostgreSQL、MySQL、SQLite）和 S3 兼容对象存储

### 技术规格

| 功能 | 规格 |
|------|------|
| **邮件协议** | SMTP、IMAP、Submission |
| **认证方式** | EVM 钱包、LDAP、PAM、SASL |
| **TLS** | 自动 ACME 证书 |
| **存储** | SQLite、PostgreSQL、MySQL、S3 |
| **DNS 提供商** | 15+ 种支持 |

## 快速开始

### 一键搭建节点

使用单条命令快速搭建你的 SirrMesh 节点：

```bash
# 下载并运行安装脚本
curl -sSL https://raw.githubusercontent.com/mail-chat-chain/sirrmeshd/main/start.sh | bash
```

自动安装程序将执行以下操作，帮你快速搭建独立节点：

1. **下载和安装** `sirrmeshd` 节点程序
2. **域名配置** - 设置你的节点域名
3. **DNS 提供商设置** - 从 15 种支持的提供商中选择
4. **TLS 证书** - 自动 ACME DNS-01 挑战设置
5. **服务管理** - 创建并启动节点服务
6. **节点上线** - 你的去中心化通讯节点开始运行

### 支持的 DNS 提供商

| 提供商 | 类型 | 认证方式 |
|--------|------|----------|
| **Cloudflare** | 全球 CDN | API Token |
| Amazon Route53 | AWS DNS | Access Key + Secret |
| DigitalOcean | 云端 DNS | API Token |
| Google Cloud DNS | GCP DNS | 服务账户 JSON |
| Vultr | 云端 DNS | API Key |
| Hetzner | 欧洲 DNS | API Token |
| Gandi | 域名注册商 | API Token |
| Namecheap | 域名注册商 | API 凭证 |
| **+ 7 个更多** | 各种 | 各种 |

## 手动安装

### 系统要求

```yaml
系统要求:
  操作系统: Ubuntu 20.04+ / macOS 12+ / CentOS 8+
  CPU: 2核以上
  内存: 2GB 最低（推荐 4GB）
  存储: 20GB SSD
  网络: 100Mbps

软件依赖:
  Go: 1.24+
  Git: 最新版
  Make: 最新版
```

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/mail-chat-chain/sirrmeshd.git
cd sirrmeshd

# 构建二进制文件
make build

# 验证安装
./build/sirrmeshd --help
```

### 启动节点

```bash
# 启动你的 SirrMesh 节点
./sirrmeshd run

# 或者使用 start.sh 脚本自动配置并启动节点
./start.sh
```

## 节点功能

### 去中心化通讯能力

- **独立节点**: 完全自主运行的通讯节点，不依赖第三方
- **区块链身份**: 通过 EVM 钱包签名实现去中心化身份认证
- **数据主权**: 所有数据存储在你的节点上，完全由你掌控
- **协议兼容**: 完整的 SMTP/IMAP 支持，兼容各种邮件客户端
- **分布式存储**: IMAP 邮箱存储在 SQL 中，可选 S3 blob 存储
- **安全防护**: DKIM、SPF、DMARC 验证与信誉评分

### 节点配置示例

```
# sirrmeshd.conf - 你的节点配置文件
$(hostname) = mx1.example.com
$(primary_domain) = example.com

tls {
    loader acme {
        hostname $(hostname)
        email postmaster@$(hostname)
        agreed
        challenge dns-01
        dns cloudflare {
            api_token YOUR_API_TOKEN
        }
    }
}

storage.imapsql local_mailboxes {
    driver sqlite3
    dsn imapsql.db
}

auth.pass_blockchain blockchain_auth {
    blockchain &sirrmeshd
    storage &local_mailboxes
}

smtp tcp://0.0.0.0:8825 {
    hostname $(hostname)

    source $(primary_domain) {
        deliver_to &local_mailboxes
    }
}

imap tls://0.0.0.0:993 {
    auth &blockchain_auth
    storage &local_mailboxes
}
```

### DNS 管理命令

```bash
# 配置 DNS 设置
sirrmeshd dns config

# 检查 DNS 配置
sirrmeshd dns check

# 导出域名设置的 DNS 记录
sirrmeshd dns export

# 获取 A 记录的公网 IP
sirrmeshd dns ip
```

## 节点管理命令

```
sirrmeshd [command]

可用命令:
  run          启动 SirrMesh 节点
  creds        节点用户凭证管理
  dns          DNS 配置指南和检查器
  hash         生成用于 pass_table 的密码哈希
  imap-acct    IMAP 存储账户管理
  imap-mboxes  IMAP 邮箱（文件夹）管理
  imap-msgs    IMAP 消息管理
  help         关于任何命令的帮助
```

## 系统架构

### 系统组件

```
┌─────────────────┐     ┌─────────────────┐
│  邮件客户端     │────▶│   SMTP/IMAP     │
│  (Thunderbird,  │     │   端点          │
│   Outlook 等)   │     └────────┬────────┘
└─────────────────┘              │
                                 ▼
                    ┌─────────────────────┐
                    │       认证          │
                    │  (区块链/LDAP)      │
                    └────────┬────────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
     ┌─────────────┐  ┌───────────┐  ┌──────────┐
     │    存储     │  │   检查    │  │   修改   │
     │  (SQL/S3)   │  │(DKIM/SPF) │  │  (DKIM)  │
     └─────────────┘  └───────────┘  └──────────┘
```

### 可用模块

**认证模块:**
- `auth.pass_blockchain` - 区块链钱包签名认证
- `auth.pass_table` - 密码表认证
- `auth.ldap` - LDAP 目录认证
- `auth.pam` - Linux PAM 认证
- `auth.external` - 外部脚本认证

**存储模块:**
- `storage.imapsql` - SQL 数据库 IMAP 后端
- `storage.blob.fs` - 文件系统 blob 存储
- `storage.blob.s3` - S3 兼容对象存储

**检查模块:**
- `check.dkim` - DKIM 签名验证
- `check.spf` - SPF 发件人策略验证
- `check.dnsbl` - DNS 黑名单检查
- `check.rspamd` - Rspamd 垃圾邮件检查

**端点模块:**
- `smtp` - SMTP 服务器
- `imap` - IMAP 服务器
- `submission` - 邮件提交

## 配置

### 性能调优

```
# sirrmeshd.conf

smtp tcp://0.0.0.0:8825 {
    limits {
        all rate 20 1s
        all concurrency 10
    }
}

imap tls://0.0.0.0:993 {
    io_debug no
}
```

## 文档

- **[完整技术文档](DOCUMENTATION.md)** - 全面的设置和配置指南
- **[部署指南](DEPLOYMENT.md)** - 服务器部署和管理

## 贡献

我们欢迎贡献！请查看我们的[贡献指南](CONTRIBUTING.md)了解详情。

### 开发环境设置

```bash
# 构建项目
make build

# 运行测试
make test

# 运行代码检查
make lint
```

## 许可证

本项目采用 GPL 3.0 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 链接

- **官网**: https://mailcoin.org
- **文档**: https://docs.mailcoin.org

## 支持

- **GitHub Issues**: 用于 bug 报告和功能请求
- **文档**: 用于设置和配置帮助

---

**SirrMesh** - 去中心化加密通讯系统，搭建属于你自己的通讯节点
