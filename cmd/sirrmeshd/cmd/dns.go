package cmd

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	parser "github.com/sirrchat/SirrMesh/framework/cfgparser"
	"github.com/sirrchat/SirrMesh/framework/log"
	"github.com/spf13/cobra"
)

type DNSConfig struct {
	Hostname        string
	PrimaryDomain   string
	ServerIP        string
	ServerIPv6      string
	DKIMPublicKey   string
	PostmasterEmail string
	DMARCPolicy     string // none, quarantine, reject
}

func NewDNSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns",
		Short: "DNS configuration guide and checker",
		Long:  `Provides DNS configuration instructions and validates DNS settings for mail server`,
	}

	cmd.AddCommand(
		NewDNSGuideCmd(),
		NewDNSCheckCmd(),
		NewDNSExportCmd(),
	)

	return cmd
}

func NewDNSGuideCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "guide",
		Short: "Show DNS configuration guide",
		Long:  `Display detailed DNS configuration instructions for setting up the mail server`,
		RunE:  runDNSGuide,
	}
}

func NewDNSCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check [domain]",
		Short: "Check DNS configuration",
		Long:  `Verify that DNS records are correctly configured for the mail server`,
		Args:  cobra.MaximumNArgs(1),
		RunE:  runDNSCheck,
	}
}

func NewDNSExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export DNS records",
		Long:  `Export DNS records in various formats (BIND, CloudFlare, etc.)`,
		RunE:  runDNSExport,
	}
}

// ensureDKIMKeys checks if DKIM keys exist and generates them if not
func ensureDKIMKeys(domain, selector string) (string, error) {
	// 使用工作目录（ConfigDirectory）而不是当前目录
	dkimDir := filepath.Join(ConfigDirectory, "dkim_keys")
	if err := os.MkdirAll(dkimDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create dkim_keys directory: %v", err)
	}

	keyPath := filepath.Join(dkimDir, fmt.Sprintf("%s_%s.key", domain, selector))
	dnsPath := filepath.Join(dkimDir, fmt.Sprintf("%s_%s.dns", domain, selector))

	// Check if key already exists
	if _, err := os.Stat(keyPath); err == nil {
		// Key exists, read DNS record
		if dnsData, err := os.ReadFile(dnsPath); err == nil {
			return strings.TrimSpace(string(dnsData)), nil
		}
	}

	// Generate new key
	log.Printf("Generating DKIM key for domain %s with selector %s...", domain, selector)
	
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", fmt.Errorf("failed to generate RSA key: %v", err)
	}

	// Save private key
	keyBlob, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal private key: %v", err)
	}

	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return "", fmt.Errorf("failed to create key file: %v", err)
	}
	defer keyFile.Close()

	if err := pem.Encode(keyFile, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyBlob,
	}); err != nil {
		return "", fmt.Errorf("failed to write private key: %v", err)
	}

	// Generate DNS record
	publicKeyBlob, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %v", err)
	}

	dkimRecord := fmt.Sprintf("v=DKIM1; k=rsa; p=%s", base64.StdEncoding.EncodeToString(publicKeyBlob))
	
	// Save DNS record
	if err := os.WriteFile(dnsPath, []byte(dkimRecord), 0o644); err != nil {
		return "", fmt.Errorf("failed to write DNS record: %v", err)
	}

	log.Printf("Generated DKIM key pair for %s:%s", domain, selector)
	return dkimRecord, nil
}

func loadDNSConfig() (*DNSConfig, error) {
	
	// 首先检查当前工作目录的配置文件
	var configPath string
	workingDirConfig := "sirrmeshd.conf"
	if _, err := os.Stat(workingDirConfig); err == nil {
		configPath = workingDirConfig
	} else {
		// 如果当前目录没有，则使用默认配置目录
		configPath = filepath.Join(ConfigDirectory, "sirrmeshd.conf")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// 错误提示
			log.Printf("Warning: config file not found: %s\n", configPath)
		}
	}
	
	
	f, err := os.Open(configPath)
	if err != nil {
		// 如果都找不到，返回错误
		return nil, fmt.Errorf("failed to open config: %v", err)
	}
	defer f.Close()

	cfg, err := parser.Read(f, configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}
	
	// 同时读取原始配置文件以获取变量定义
	f2, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to reopen config: %v", err)
	}
	defer f2.Close()
	
	variables := make(map[string]string)
	scanner := bufio.NewScanner(f2)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "$(") && strings.Contains(line, ") =") {
			// 解析 $(variable) = value 格式
			parts := strings.SplitN(line, ") =", 2)
			if len(parts) == 2 {
				varName := strings.TrimPrefix(parts[0], "$(")
				value := strings.TrimSpace(parts[1])
				variables[varName] = value
			}
		}
	}

	// 解析配置中的变量定义和普通配置项
	hostname := ""
	primaryDomain := ""
	
	// 解析变量值，处理变量引用 $(other_var)
	resolveVariable := func(value string) string {
		// 简单的变量替换，支持 $(variable) 引用
		for varName, varValue := range variables {
			placeholder := fmt.Sprintf("$(%s)", varName)
			value = strings.ReplaceAll(value, placeholder, varValue)
		}
		return value
	}
	
	// 获取解析后的变量值
	if val, ok := variables["hostname"]; ok {
		hostname = resolveVariable(val)
	}
	if val, ok := variables["primary_domain"]; ok {
		primaryDomain = resolveVariable(val)
	}
	
	
	// 查找已经展开的普通配置节点（解析器已处理了变量）
	for _, node := range cfg {
		if node.Name == "hostname" && len(node.Args) > 0 && hostname == "" {
			hostname = node.Args[0]
		}
		if node.Name == "primary_domain" && len(node.Args) > 0 && primaryDomain == "" {
			primaryDomain = node.Args[0]
		}
	}
	
	// 如果没有找到 primary_domain，尝试从 hostname 推导
	if primaryDomain == "" && hostname != "" {
		// 对于大多数邮件服务器，hostname 通常是 mail.domain.com 或 mx1.domain.com
		// primary_domain 是 domain.com
		parts := strings.Split(hostname, ".")
		if len(parts) > 2 && (parts[0] == "mail" || parts[0] == "mx1" || parts[0] == "mx") {
			primaryDomain = strings.Join(parts[1:], ".")
		} else if len(parts) >= 2 {
			// 如果是普通的子域名，取后面的部分作为主域名
			primaryDomain = strings.Join(parts[1:], ".")
		} else {
			// 如果 hostname 就是域名，直接使用
			primaryDomain = hostname
		}
	}
	
	if hostname == "" {
		hostname = "mail.example.org"
	}
	if primaryDomain == "" {
		primaryDomain = "example.org"
	}

	// 尝试获取服务器IP
	serverIP := getServerIP()
	serverIPv6 := getServerIPv6()

	// 从配置中提取postmaster邮箱，优先从TLS配置中获取
	postmasterEmail := ""
	for _, node := range cfg {
		if node.Name == "tls" && len(node.Children) > 0 {
			for _, child := range node.Children {
				if child.Name == "loader" && child.Args[0] == "acme" && len(child.Children) > 0 {
					for _, acmeChild := range child.Children {
						if acmeChild.Name == "email" && len(acmeChild.Args) > 0 {
							postmasterEmail = acmeChild.Args[0]
							break
						}
					}
				}
			}
		}
	}
	
	if postmasterEmail == "" {
		postmasterEmail = fmt.Sprintf("postmaster@%s", primaryDomain)
	}

	// 确保DKIM密钥存在，使用默认选择器"default"
	dkimPublicKey, err := ensureDKIMKeys(primaryDomain, "default")
	if err != nil {
		log.Printf("Warning: Failed to ensure DKIM keys: %v", err)
		dkimPublicKey = "YOUR_DKIM_PUBLIC_KEY"
	}

	return &DNSConfig{
		Hostname:        hostname,
		PrimaryDomain:   primaryDomain,
		ServerIP:        serverIP,
		ServerIPv6:      serverIPv6,
		DKIMPublicKey:   dkimPublicKey,
		PostmasterEmail: postmasterEmail,
	}, nil
}

func getServerIP() string {
	// Try to get public IP from multiple sources
	ipServices := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ipinfo.io/ip",
		"https://checkip.amazonaws.com",
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, service := range ipServices {
		resp, err := client.Get(service)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		ip := strings.TrimSpace(string(body))
		// Validate IP address
		if net.ParseIP(ip) != nil && !isPrivateIP(ip) {
			return ip
		}
	}

	// Fallback to local connection method
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "YOUR_SERVER_IP"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func getServerIPv6() string {
	// Try to get public IPv6 from multiple sources
	ipv6Services := []string{
		"https://api6.ipify.org",
		"https://v6.ident.me",
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, service := range ipv6Services {
		resp, err := client.Get(service)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		ip := strings.TrimSpace(string(body))
		// Validate IPv6 address
		parsedIP := net.ParseIP(ip)
		if parsedIP != nil && parsedIP.To4() == nil && parsedIP.To16() != nil {
			return ip
		}
	}

	// Fallback to local connection method
	conn, err := net.Dial("udp6", "[2001:4860:4860::8888]:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip := localAddr.IP
	if ip.To4() == nil && ip.To16() != nil {
		return ip.String()
	}
	return ""
}

// isPrivateIP checks if an IP address is in a private range
func isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return true // Invalid IP, consider it private
	}

	// Check for private IPv4 ranges
	if parsedIP.To4() != nil {
		// 10.0.0.0/8
		if parsedIP[12] == 10 {
			return true
		}
		// 172.16.0.0/12
		if parsedIP[12] == 172 && parsedIP[13] >= 16 && parsedIP[13] <= 31 {
			return true
		}
		// 192.168.0.0/16
		if parsedIP[12] == 192 && parsedIP[13] == 168 {
			return true
		}
		// 127.0.0.0/8 (localhost)
		if parsedIP[12] == 127 {
			return true
		}
	}

	// Check for private IPv6 ranges
	if parsedIP.To4() == nil {
		// fc00::/7 (unique local addresses)
		if parsedIP[0] >= 0xfc && parsedIP[0] <= 0xfd {
			return true
		}
		// fe80::/10 (link-local addresses)
		if parsedIP[0] == 0xfe && (parsedIP[1]&0xc0) == 0x80 {
			return true
		}
		// ::1 (localhost)
		if parsedIP.IsLoopback() {
			return true
		}
	}

	return false
}


func runDNSGuide(cmd *cobra.Command, args []string) error {
	cfg, err := loadDNSConfig()
	if err != nil {
		log.Printf("Warning: %v\n", err)
		cfg = &DNSConfig{
			Hostname:      "mx1.example.org",
			PrimaryDomain: "example.org",
			ServerIP:      "YOUR_SERVER_IP",
			DKIMPublicKey: "YOUR_DKIM_PUBLIC_KEY",
			PostmasterEmail: "postmaster@example.org",
		}
	}
	
	// 确保hostname是子域名格式
	if cfg.Hostname == cfg.PrimaryDomain || !strings.Contains(cfg.Hostname, ".") {
		cfg.Hostname = fmt.Sprintf("mx1.%s", cfg.PrimaryDomain)
	}

	fmt.Println("================================================================================")
	fmt.Println("                        邮件服务器 DNS 配置指南")
	fmt.Println("================================================================================")
	fmt.Printf("\n主域名: %s\n", cfg.PrimaryDomain)
	fmt.Printf("邮件服务器主机名: %s\n", cfg.Hostname)
	fmt.Printf("服务器IP: %s\n", cfg.ServerIP)
	if cfg.ServerIPv6 != "" {
		fmt.Printf("服务器IPv6: %s\n", cfg.ServerIPv6)
	}
	fmt.Println("\n================================================================================")
	fmt.Println("DNS区域文件配置 (直接复制到DNS管理面板):")
	fmt.Println("================================================================================")

	// 格式化输出，对齐列
	printDNSRecord := func(name, recordType, value string, comment string) {
		if comment != "" {
			fmt.Printf("; %s\n", comment)
		}
		fmt.Printf("%-40s IN  %-6s %s\n", name+".", recordType, value)
	}
	
	// 基础A/AAAA记录
	fmt.Println("; === 基础IP记录 ===")
	printDNSRecord(cfg.PrimaryDomain, "A", cfg.ServerIP, "主域名IPv4地址")
	if cfg.ServerIPv6 != "" {
		printDNSRecord(cfg.PrimaryDomain, "AAAA", cfg.ServerIPv6, "主域名IPv6地址")
	}
	fmt.Println()
	
	printDNSRecord(cfg.Hostname, "A", cfg.ServerIP, "邮件服务器IPv4地址（重要）")
	if cfg.ServerIPv6 != "" {
		printDNSRecord(cfg.Hostname, "AAAA", cfg.ServerIPv6, "邮件服务器IPv6地址")
	}
	fmt.Println()
	
	// MX记录
	fmt.Println("; === MX记录 - 指定邮件服务器 ===")
	printDNSRecord(cfg.PrimaryDomain, "MX", fmt.Sprintf("10 %s.", cfg.Hostname), "")
	fmt.Println()
	
	// SPF记录
	fmt.Println("; === SPF记录 - 防止邮件伪造 ===")
	printDNSRecord(cfg.PrimaryDomain, "TXT", "\"v=spf1 mx ~all\"", "主域名SPF")
	printDNSRecord(cfg.Hostname, "TXT", "\"v=spf1 a ~all\"", "邮件服务器SPF（推荐）")
	fmt.Println()
	
	// DKIM记录
	fmt.Println("; === DKIM记录 - 邮件签名验证 ===")
	dkimHost := fmt.Sprintf("default._domainkey.%s", cfg.PrimaryDomain)
	printDNSRecord(dkimHost, "TXT", fmt.Sprintf("\"%s\"", cfg.DKIMPublicKey), "")
	fmt.Println()
	
	// DMARC记录
	fmt.Println("; === DMARC记录 - SPF和DKIM策略 ===")
	dmarcHost := fmt.Sprintf("_dmarc.%s", cfg.PrimaryDomain)
	dmarcPolicy := cfg.DMARCPolicy
	if dmarcPolicy == "" {
		dmarcPolicy = "quarantine"
	}
	dmarcValue := fmt.Sprintf("\"v=DMARC1; p=%s; ruf=mailto:%s\"", dmarcPolicy, cfg.PostmasterEmail)
	printDNSRecord(dmarcHost, "TXT", dmarcValue, "")
	fmt.Println()
	
	// MTA-STS和TLSRPT记录
	fmt.Println("; === MTA-STS和TLSRPT记录 - 强制TLS传输（推荐） ===")
	mtastsHost := fmt.Sprintf("_mta-sts.%s", cfg.PrimaryDomain)
	printDNSRecord(mtastsHost, "TXT", "\"v=STSv1; id=1\"", "MTA-STS策略声明")
	
	tlsrptHost := fmt.Sprintf("_smtp._tls.%s", cfg.PrimaryDomain)
	tlsrptValue := fmt.Sprintf("\"v=TLSRPTv1; rua=mailto:%s\"", cfg.PostmasterEmail)
	printDNSRecord(tlsrptHost, "TXT", tlsrptValue, "TLS报告接收地址")
	
	fmt.Println("\n================================================================================")
	fmt.Println("其他重要配置:")
	fmt.Println("================================================================================")
	
	fmt.Println("\n1. 反向DNS (PTR记录):")
	fmt.Printf("   联系服务器提供商设置 IP %s 的PTR记录为: %s\n", cfg.ServerIP, cfg.Hostname)
	
	fmt.Println("\n2. MTA-STS策略文件:")
	fmt.Printf("   在 https://mta-sts.%s/.well-known/mta-sts.txt 提供以下内容:\n", cfg.PrimaryDomain)
	fmt.Println("   ---")
	fmt.Println("   version: STSv1")
	fmt.Println("   mode: enforce")
	fmt.Printf("   mx: %s\n", cfg.Hostname)
	fmt.Println("   max_age: 604800")
	fmt.Println("   ---")
	
	fmt.Println("\n3. 防火墙端口:")
	fmt.Println("   确保开放: 25 (SMTP), 465 (SMTPS), 587 (Submission), 143 (IMAP), 993 (IMAPS)")
	
	fmt.Println("\n================================================================================")
	fmt.Println("验证命令:")
	fmt.Println("================================================================================")
	fmt.Printf("sirrmeshd dns check %s    # 检查DNS配置\n", cfg.PrimaryDomain)
	fmt.Printf("dig MX %s                 # 验证MX记录\n", cfg.PrimaryDomain)
	fmt.Printf("dig TXT %s                # 验证SPF记录\n", cfg.PrimaryDomain)
	fmt.Printf("dig TXT default._domainkey.%s  # 验证DKIM记录\n", cfg.PrimaryDomain)
	
	fmt.Println("\n================================================================================")

	return nil
}

type DNSRecord struct {
	RecordType string   // 记录类型: A, AAAA, MX, TXT等
	Name       string   // 记录名称/主机名
	Content    []string // 记录内容
	Priority   uint16   // MX记录优先级
	Status     string   // 查询状态: OK, NOT_FOUND, ERROR
	Error      string   // 错误信息
}

func runDNSCheck(cmd *cobra.Command, args []string) error {
	var domain string
	
	if len(args) > 0 {
		domain = args[0]
	} else {
		cfg, err := loadDNSConfig()
		if err != nil {
			return fmt.Errorf("请指定域名或确保配置文件存在: %v", err)
		}
		domain = cfg.PrimaryDomain
	}

	fmt.Printf("\n正在检查域名 %s 的DNS配置...\n", domain)
	fmt.Println("================================================================================")

	var records []DNSRecord
	
	// 检查A记录
	fmt.Printf("查询 A 记录...")
	aIPs := []string{}
	ips, err := net.LookupIP(domain)
	if err != nil {
		records = append(records, DNSRecord{
			RecordType: "A",
			Name:       domain,
			Status:     "ERROR",
			Error:      err.Error(),
		})
		fmt.Println(" ❌")
	} else {
		for _, ip := range ips {
			if ip.To4() != nil {
				aIPs = append(aIPs, ip.String())
			}
		}
		if len(aIPs) > 0 {
			records = append(records, DNSRecord{
				RecordType: "A",
				Name:       domain,
				Content:    aIPs,
				Status:     "OK",
			})
			fmt.Println(" ✅")
		} else {
			records = append(records, DNSRecord{
				RecordType: "A",
				Name:       domain,
				Status:     "NOT_FOUND",
				Error:      "未找到IPv4地址",
			})
			fmt.Println(" ❌")
		}
	}
	
	// 检查AAAA记录
	fmt.Printf("查询 AAAA 记录...")
	aaaaIPs := []string{}
	for _, ip := range ips {
		if ip.To4() == nil && ip.To16() != nil {
			aaaaIPs = append(aaaaIPs, ip.String())
		}
	}
	if len(aaaaIPs) > 0 {
		records = append(records, DNSRecord{
			RecordType: "AAAA",
			Name:       domain,
			Content:    aaaaIPs,
			Status:     "OK",
		})
		fmt.Println(" ✅")
	} else {
		fmt.Println(" ⚪ (可选)")
	}

	// 检查MX记录
	fmt.Printf("查询 MX 记录...")
	mxHosts := []string{}
	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		records = append(records, DNSRecord{
			RecordType: "MX",
			Name:       domain,
			Status:     "ERROR",
			Error:      err.Error(),
		})
		fmt.Println(" ❌")
	} else if len(mxRecords) == 0 {
		records = append(records, DNSRecord{
			RecordType: "MX",
			Name:       domain,
			Status:     "NOT_FOUND",
			Error:      "未找到MX记录",
		})
		fmt.Println(" ❌")
	} else {
		for _, mx := range mxRecords {
			mxHost := strings.TrimSuffix(mx.Host, ".")
			mxHosts = append(mxHosts, mxHost)
			records = append(records, DNSRecord{
				RecordType: "MX",
				Name:       domain,
				Content:    []string{mxHost},
				Priority:   mx.Pref,
				Status:     "OK",
			})
		}
		fmt.Println(" ✅")
	}
	
	// 检查MX主机的A/AAAA记录
	for _, mxHost := range mxHosts {
		fmt.Printf("查询 %s 的 A/AAAA 记录...", mxHost)
		mxIPs, err := net.LookupIP(mxHost)
		if err != nil {
			records = append(records, DNSRecord{
				RecordType: "A (MX主机)",
				Name:       mxHost,
				Status:     "ERROR",
				Error:      err.Error(),
			})
			fmt.Println(" ❌")
		} else {
			mxA := []string{}
			mxAAAA := []string{}
			for _, ip := range mxIPs {
				if ip.To4() != nil {
					mxA = append(mxA, ip.String())
				} else if ip.To16() != nil {
					mxAAAA = append(mxAAAA, ip.String())
				}
			}
			if len(mxA) > 0 {
				records = append(records, DNSRecord{
					RecordType: "A (MX主机)",
					Name:       mxHost,
					Content:    mxA,
					Status:     "OK",
				})
			}
			if len(mxAAAA) > 0 {
				records = append(records, DNSRecord{
					RecordType: "AAAA (MX主机)",
					Name:       mxHost,
					Content:    mxAAAA,
					Status:     "OK",
				})
			}
			if len(mxA) > 0 || len(mxAAAA) > 0 {
				fmt.Println(" ✅")
			} else {
				records = append(records, DNSRecord{
					RecordType: "A/AAAA (MX主机)",
					Name:       mxHost,
					Status:     "NOT_FOUND",
					Error:      "MX主机无IP地址",
				})
				fmt.Println(" ❌")
			}
		}
		
		// 检查MX主机的SPF记录
		fmt.Printf("查询 %s 的 SPF 记录...", mxHost)
		mxTxt, _ := net.LookupTXT(mxHost)
		mxSpfFound := false
		for _, txt := range mxTxt {
			if strings.HasPrefix(txt, "v=spf1") {
				records = append(records, DNSRecord{
					RecordType: "TXT (MX SPF)",
					Name:       mxHost,
					Content:    []string{txt},
					Status:     "OK",
				})
				mxSpfFound = true
				break
			}
		}
		if mxSpfFound {
			fmt.Println(" ✅")
		} else {
			fmt.Println(" ⚪ (推荐)")
		}
	}

	// 检查TXT记录（SPF）
	fmt.Printf("查询 SPF 记录...")
	txtRecords, _ := net.LookupTXT(domain)
	spfFound := false
	for _, txt := range txtRecords {
		if strings.HasPrefix(txt, "v=spf1") {
			records = append(records, DNSRecord{
				RecordType: "TXT (SPF)",
				Name:       domain,
				Content:    []string{txt},
				Status:     "OK",
			})
			spfFound = true
		}
	}
	if spfFound {
		fmt.Println(" ✅")
	} else {
		records = append(records, DNSRecord{
			RecordType: "TXT (SPF)",
			Name:       domain,
			Status:     "NOT_FOUND",
			Error:      "未找到SPF记录",
		})
		fmt.Println(" ❌")
	}

	// 检查DKIM记录
	fmt.Printf("查询 DKIM 记录...")
	dkimDomain := fmt.Sprintf("default._domainkey.%s", domain)
	dkimTxt, err := net.LookupTXT(dkimDomain)
	dkimFound := false
	if err == nil && len(dkimTxt) > 0 {
		for _, txt := range dkimTxt {
			if strings.Contains(txt, "DKIM1") || strings.Contains(txt, "k=rsa") {
				// 截断过长的DKIM公钥显示
				displayTxt := txt
				if len(txt) > 100 {
					displayTxt = txt[:97] + "..."
				}
				records = append(records, DNSRecord{
					RecordType: "TXT (DKIM)",
					Name:       dkimDomain,
					Content:    []string{displayTxt},
					Status:     "OK",
				})
				dkimFound = true
				break
			}
		}
	}
	if dkimFound {
		fmt.Println(" ✅")
	} else {
		records = append(records, DNSRecord{
			RecordType: "TXT (DKIM)",
			Name:       dkimDomain,
			Status:     "NOT_FOUND",
			Error:      "未找到DKIM记录",
		})
		fmt.Println(" ❌")
	}

	// 检查DMARC记录
	fmt.Printf("查询 DMARC 记录...")
	dmarcDomain := fmt.Sprintf("_dmarc.%s", domain)
	dmarcTxt, err := net.LookupTXT(dmarcDomain)
	dmarcFound := false
	if err == nil && len(dmarcTxt) > 0 {
		for _, txt := range dmarcTxt {
			if strings.Contains(txt, "DMARC1") {
				records = append(records, DNSRecord{
					RecordType: "TXT (DMARC)",
					Name:       dmarcDomain,
					Content:    []string{txt},
					Status:     "OK",
				})
				dmarcFound = true
				break
			}
		}
	}
	if dmarcFound {
		fmt.Println(" ✅")
	} else {
		records = append(records, DNSRecord{
			RecordType: "TXT (DMARC)",
			Name:       dmarcDomain,
			Status:     "NOT_FOUND",
			Error:      "未找到DMARC记录",
		})
		fmt.Println(" ❌")
	}

	// 检查MTA-STS记录
	fmt.Printf("查询 MTA-STS 记录...")
	mtastsDomain := fmt.Sprintf("_mta-sts.%s", domain)
	mtastsTxt, err := net.LookupTXT(mtastsDomain)
	mtastsFound := false
	if err == nil && len(mtastsTxt) > 0 {
		for _, txt := range mtastsTxt {
			if strings.Contains(txt, "STSv1") {
				records = append(records, DNSRecord{
					RecordType: "TXT (MTA-STS)",
					Name:       mtastsDomain,
					Content:    []string{txt},
					Status:     "OK",
				})
				mtastsFound = true
				break
			}
		}
	}
	if mtastsFound {
		fmt.Println(" ✅")
	} else {
		fmt.Println(" ⚪ (推荐)")
	}
	
	// 检查TLSRPT记录
	fmt.Printf("查询 TLSRPT 记录...")
	tlsrptDomain := fmt.Sprintf("_smtp._tls.%s", domain)
	tlsrptTxt, err := net.LookupTXT(tlsrptDomain)
	tlsrptFound := false
	if err == nil && len(tlsrptTxt) > 0 {
		for _, txt := range tlsrptTxt {
			if strings.Contains(txt, "TLSRPTv1") {
				records = append(records, DNSRecord{
					RecordType: "TXT (TLSRPT)",
					Name:       tlsrptDomain,
					Content:    []string{txt},
					Status:     "OK",
				})
				tlsrptFound = true
				break
			}
		}
	}
	if tlsrptFound {
		fmt.Println(" ✅")
	} else {
		fmt.Println(" ⚪ (推荐)")
	}

	// 检查PTR记录（反向DNS）
	if len(aIPs) > 0 {
		fmt.Printf("查询 PTR 记录...")
		ptrFound := false
		for _, ip := range aIPs {
			names, err := net.LookupAddr(ip)
			if err == nil && len(names) > 0 {
				records = append(records, DNSRecord{
					RecordType: "PTR",
					Name:       ip,
					Content:    names,
					Status:     "OK",
				})
				ptrFound = true
				break
			}
		}
		if ptrFound {
			fmt.Println(" ✅")
		} else {
			records = append(records, DNSRecord{
				RecordType: "PTR",
				Name:       aIPs[0],
				Status:     "NOT_FOUND",
				Error:      "未配置反向DNS",
			})
			fmt.Println(" ⚠️")
		}
	}

	// 打印详细结果表格
	fmt.Println("\n================================================================================")
	fmt.Println("DNS查询详细结果:")
	fmt.Println("================================================================================")
	fmt.Printf("%-12s %-8s %-35s %s\n", "记录类型", "状态", "名称", "内容")
	fmt.Println("--------------------------------------------------------------------------------")
	
	allPass := true
	warningCount := 0
	for _, record := range records {
		statusSymbol := "✅"
		statusText := "正常"
		if record.Status == "NOT_FOUND" {
			statusSymbol = "❌"
			statusText = "缺失"
			if record.RecordType != "PTR" {
				allPass = false
			} else {
				warningCount++
			}
		} else if record.Status == "ERROR" {
			statusSymbol = "❌"
			statusText = "错误"
			allPass = false
		}
		
		// 格式化名称，截断过长的部分
		displayName := record.Name
		if len(displayName) > 33 {
			displayName = displayName[:30] + "..."
		}
		
		// 打印记录
		if len(record.Content) > 0 {
			for i, content := range record.Content {
				if i == 0 {
					if record.RecordType == "MX" && record.Priority > 0 {
						fmt.Printf("%-12s %s %-6s %-35s 优先级:%d %s\n", 
							record.RecordType, statusSymbol, statusText, displayName, record.Priority, content)
					} else {
						fmt.Printf("%-12s %s %-6s %-35s %s\n", 
							record.RecordType, statusSymbol, statusText, displayName, content)
					}
				} else {
					// 多个内容的记录，后续行缩进显示
					fmt.Printf("%-12s %-8s %-35s %s\n", "", "", "", content)
				}
			}
		} else {
			// 没有内容的记录（错误或未找到）
			errorMsg := record.Error
			if errorMsg == "" {
				errorMsg = "无数据"
			}
			fmt.Printf("%-12s %s %-6s %-35s %s\n", 
				record.RecordType, statusSymbol, statusText, displayName, errorMsg)
		}
	}
	
	fmt.Println("================================================================================")
	
	// 打印总结
	fmt.Println("\n检查总结:")
	fmt.Println("--------------------------------------------------------------------------------")
	if allPass && warningCount == 0 {
		fmt.Println("✅ 所有DNS记录配置正确！邮件服务器DNS配置完整。")
	} else if allPass && warningCount > 0 {
		fmt.Println("⚠️  基本DNS记录已配置，但建议配置反向DNS(PTR)记录以提高邮件送达率。")
		fmt.Println("   请联系您的服务器提供商设置PTR记录。")
	} else {
		fmt.Println("❌ 部分必要的DNS记录缺失或配置错误。")
		fmt.Println("   运行 'sirrmeshd dns guide' 查看完整配置指南。")
		fmt.Println("\n缺失的记录:")
		for _, record := range records {
			if record.Status == "NOT_FOUND" && record.RecordType != "PTR" {
				fmt.Printf("   - %s (%s)\n", record.RecordType, record.Name)
			}
		}
	}
	fmt.Println("================================================================================")

	return nil
}

func runDNSExport(cmd *cobra.Command, args []string) error {
	cfg, err := loadDNSConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}
	
	// 确保hostname是子域名格式
	if cfg.Hostname == cfg.PrimaryDomain || !strings.Contains(cfg.Hostname, ".") {
		cfg.Hostname = fmt.Sprintf("mx1.%s", cfg.PrimaryDomain)
	}
	
	fmt.Println("================================================================================")
	fmt.Println("                    DNS 配置导出 - 交互式配置")
	fmt.Println("================================================================================")
	fmt.Printf("当前配置:\n")
	fmt.Printf("  主域名: %s\n", cfg.PrimaryDomain)
	fmt.Printf("  邮件服务器: %s\n", cfg.Hostname)
	fmt.Printf("  检测到的服务器IP: %s\n", cfg.ServerIP)
	if cfg.ServerIPv6 != "" {
		fmt.Printf("  检测到的IPv6: %s\n", cfg.ServerIPv6)
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	
	// 询问并更新配置
	cfg = promptForConfiguration(cfg, reader)

	fmt.Println("\n================================================================================")
	fmt.Println("选择导出格式:")
	fmt.Println("1. 标准DNS区域文件格式 (推荐)")
	fmt.Println("2. BIND Zone File")
	fmt.Println("3. CloudFlare CSV")
	fmt.Println("4. Generic (文本格式)")
	fmt.Print("请选择 (1-4): ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		exportStandardZoneFormat(cfg)
	case "2":
		exportBINDFormat(cfg)
	case "3":
		exportCloudFlareFormat(cfg)
	default:
		exportGenericFormat(cfg)
	}

	return nil
}

func promptForConfiguration(cfg *DNSConfig, reader *bufio.Reader) *DNSConfig {
	newCfg := *cfg // 复制配置
	
	// 询问服务器IP地址
	fmt.Printf("1. 服务器IPv4地址 [当前: %s]: ", cfg.ServerIP)
	if input := readInput(reader); input != "" {
		if isValidIPv4(input) {
			newCfg.ServerIP = input
		} else {
			fmt.Printf("   ⚠️  无效的IPv4地址，使用默认值: %s\n", cfg.ServerIP)
		}
	}
	
	// 询问IPv6地址
	if cfg.ServerIPv6 != "" {
		fmt.Printf("2. 服务器IPv6地址 [当前: %s]: ", cfg.ServerIPv6)
	} else {
		fmt.Printf("2. 服务器IPv6地址 [当前: 无]: ")
	}
	if input := readInput(reader); input != "" {
		if isValidIPv6(input) {
			newCfg.ServerIPv6 = input
		} else if input != "n" && input != "no" && input != "无" {
			fmt.Printf("   ⚠️  无效的IPv6地址，跳过\n")
		}
	}
	
	// 询问邮件服务器主机名
	fmt.Printf("3. 邮件服务器主机名 [当前: %s]: ", cfg.Hostname)
	if input := readInput(reader); input != "" {
		if isValidHostname(input) {
			newCfg.Hostname = input
		} else {
			fmt.Printf("   ⚠️  无效的主机名格式，使用默认值: %s\n", cfg.Hostname)
		}
	}
	
	// 如果域名或主机名发生变化，重新生成DKIM密钥
	if newCfg.PrimaryDomain != cfg.PrimaryDomain || newCfg.Hostname != cfg.Hostname {
		fmt.Printf("4. 检测到域名变化，重新生成DKIM密钥...\n")
		if dkimKey, err := ensureDKIMKeys(newCfg.PrimaryDomain, "default"); err == nil {
			newCfg.DKIMPublicKey = dkimKey
			fmt.Printf("   ✅ 已生成新的DKIM密钥\n")
		} else {
			fmt.Printf("   ⚠️  生成DKIM密钥失败: %v\n", err)
		}
	} else {
		// 询问DKIM公钥
		if len(cfg.DKIMPublicKey) > 100 {
			displayKey := cfg.DKIMPublicKey[:97] + "..."
			fmt.Printf("4. DKIM公钥 [当前: %s]: ", displayKey)
		} else {
			fmt.Printf("4. DKIM公钥 [当前: %s]: ", cfg.DKIMPublicKey)
		}
		if input := readInput(reader); input != "" {
			newCfg.DKIMPublicKey = input
		}
	}
	
	// 询问DMARC策略
	fmt.Println("5. DMARC策略选择:")
	fmt.Println("   1) p=none (仅监控，不阻止)")
	fmt.Println("   2) p=quarantine (隔离可疑邮件) [推荐]")
	fmt.Println("   3) p=reject (直接拒绝)")
	fmt.Print("   选择 (1-3) [默认: 2]: ")
	
	choice := readInput(reader)
	dmarcPolicy := "quarantine" // 默认
	switch choice {
	case "1":
		dmarcPolicy = "none"
	case "3":
		dmarcPolicy = "reject"
	case "2", "":
		dmarcPolicy = "quarantine"
	default:
		fmt.Printf("   ⚠️  无效选择，使用默认: quarantine\n")
	}
	
	// 更新DMARC记录
	newCfg.PostmasterEmail = fmt.Sprintf("postmaster@%s", newCfg.PrimaryDomain)
	
	// 询问postmaster邮箱
	fmt.Printf("6. Postmaster邮箱 [当前: %s]: ", newCfg.PostmasterEmail)
	if input := readInput(reader); input != "" {
		if isValidEmail(input) {
			newCfg.PostmasterEmail = input
		} else {
			fmt.Printf("   ⚠️  无效的邮箱地址，使用默认值: %s\n", newCfg.PostmasterEmail)
		}
	}
	
	// 存储DMARC策略
	newCfg.DMARCPolicy = dmarcPolicy
	
	return &newCfg
}

func readInput(reader *bufio.Reader) string {
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func isValidIPv4(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
		}
		// 简单验证范围
		num := 0
		for _, char := range part {
			num = num*10 + int(char-'0')
		}
		if num > 255 {
			return false
		}
	}
	return true
}

func isValidIPv6(ip string) bool {
	// 简单的IPv6验证
	if !strings.Contains(ip, ":") || len(ip) < 3 {
		return false
	}
	// 至少应该有两个冒号组成的部分
	parts := strings.Split(ip, ":")
	return len(parts) >= 2
}

func isValidHostname(hostname string) bool {
	if len(hostname) == 0 || hostname == "." {
		return false
	}
	// 允许以.结尾的FQDN，也允许不以.结尾的hostname
	hostname = strings.TrimSuffix(hostname, ".")
	// 必须包含至少一个点
	if !strings.Contains(hostname, ".") {
		return false
	}
	// 不能以点开头
	return !strings.HasPrefix(hostname, ".")
}

func isValidEmail(email string) bool {
	if !strings.Contains(email, "@") || len(email) < 5 {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	// 检查域名部分
	domain := parts[1]
	return strings.Contains(domain, ".") && len(parts[0]) > 0
}

func exportStandardZoneFormat(cfg *DNSConfig) {
	fmt.Println("\n; ================================================================================")
	fmt.Printf("; Mail Server DNS Records for %s\n", cfg.PrimaryDomain)
	fmt.Printf("; Mail Server: %s (%s)\n", cfg.Hostname, cfg.ServerIP)
	fmt.Printf("; Generated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(";")
	fmt.Println("; IMPORTANT: This file only contains mail-related records.")
	fmt.Printf("; Do NOT add %s A/AAAA record if your domain already has a website.\n", cfg.PrimaryDomain)
	fmt.Println("; ================================================================================")
	fmt.Println()

	// A Records - Only mail server, not primary domain
	fmt.Println("; A Records (Mail Server)")
	fmt.Printf("%-40s IN  A      %s\n", cfg.Hostname+".", cfg.ServerIP)
	if cfg.ServerIPv6 != "" {
		fmt.Printf("%-40s IN  AAAA   %s\n", cfg.Hostname+".", cfg.ServerIPv6)
	}
	fmt.Println()

	// MX Records
	fmt.Println("; MX Records")
	fmt.Printf("%-40s IN  MX     10 %s.\n", cfg.PrimaryDomain+".", cfg.Hostname)
	fmt.Println()

	// SPF Records
	fmt.Println("; TXT Records (SPF)")
	fmt.Printf("%-40s IN  TXT    \"v=spf1 mx ~all\"\n", cfg.PrimaryDomain+".")
	fmt.Printf("%-40s IN  TXT    \"v=spf1 a ~all\"\n", cfg.Hostname+".")
	fmt.Println()

	// DKIM Records
	fmt.Println("; TXT Records (DKIM)")
	dkimHost := fmt.Sprintf("default._domainkey.%s.", cfg.PrimaryDomain)
	fmt.Printf("%-40s IN  TXT    \"%s\"\n", dkimHost, cfg.DKIMPublicKey)
	fmt.Println()

	// DMARC Records
	fmt.Println("; TXT Records (DMARC)")
	dmarcHost := fmt.Sprintf("_dmarc.%s.", cfg.PrimaryDomain)
	dmarcPolicy := cfg.DMARCPolicy
	if dmarcPolicy == "" {
		dmarcPolicy = "quarantine"
	}
	fmt.Printf("%-40s IN  TXT    \"v=DMARC1; p=%s; ruf=mailto:%s\"\n", dmarcHost, dmarcPolicy, cfg.PostmasterEmail)
	fmt.Println()

	// MTA-STS and TLSRPT Records
	fmt.Println("; TXT Records (MTA-STS & TLSRPT)")
	mtastsHost := fmt.Sprintf("_mta-sts.%s.", cfg.PrimaryDomain)
	fmt.Printf("%-40s IN  TXT    \"v=STSv1; id=1\"\n", mtastsHost)
	tlsrptHost := fmt.Sprintf("_smtp._tls.%s.", cfg.PrimaryDomain)
	fmt.Printf("%-40s IN  TXT    \"v=TLSRPTv1; rua=mailto:%s\"\n", tlsrptHost, cfg.PostmasterEmail)

	fmt.Println("\n; ================================================================================")
}

func exportBINDFormat(cfg *DNSConfig) {
	fmt.Println("\n; BIND Zone File Format")
	fmt.Println("; Copy and paste these records to your zone file")
	fmt.Println()
	fmt.Println("; Basic A/AAAA records")
	fmt.Printf("%s.    IN    A    %s\n", cfg.PrimaryDomain, cfg.ServerIP)
	if cfg.ServerIPv6 != "" {
		fmt.Printf("%s.    IN    AAAA    %s\n", cfg.PrimaryDomain, cfg.ServerIPv6)
	}
	fmt.Printf("%s.    IN    A    %s\n", cfg.Hostname, cfg.ServerIP)
	if cfg.ServerIPv6 != "" {
		fmt.Printf("%s.    IN    AAAA    %s\n", cfg.Hostname, cfg.ServerIPv6)
	}
	fmt.Println()
	fmt.Println("; MX record")
	fmt.Printf("%s.    IN    MX    10    %s.\n", cfg.PrimaryDomain, cfg.Hostname)
	fmt.Println()
	fmt.Println("; SPF records")
	fmt.Printf("%s.    IN    TXT    \"v=spf1 mx ~all\"\n", cfg.PrimaryDomain)
	fmt.Printf("%s.    IN    TXT    \"v=spf1 a ~all\"\n", cfg.Hostname)
	fmt.Println()
	fmt.Println("; DKIM record")
	fmt.Printf("default._domainkey.%s.    IN    TXT    \"%s\"\n", cfg.PrimaryDomain, cfg.DKIMPublicKey)
	fmt.Println()
	fmt.Println("; DMARC record")
	dmarcPolicy := cfg.DMARCPolicy
	if dmarcPolicy == "" {
		dmarcPolicy = "quarantine"
	}
	fmt.Printf("_dmarc.%s.    IN    TXT    \"v=DMARC1; p=%s; ruf=mailto:%s\"\n", cfg.PrimaryDomain, dmarcPolicy, cfg.PostmasterEmail)
	fmt.Println()
	fmt.Println("; MTA-STS and TLSRPT records")
	fmt.Printf("_mta-sts.%s.    IN    TXT    \"v=STSv1; id=1\"\n", cfg.PrimaryDomain)
	fmt.Printf("_smtp._tls.%s.    IN    TXT    \"v=TLSRPTv1; rua=mailto:%s\"\n", cfg.PrimaryDomain, cfg.PostmasterEmail)
}

func exportCloudFlareFormat(cfg *DNSConfig) {
	fmt.Println(";;")
	fmt.Printf(";; Domain:     %s\n", cfg.PrimaryDomain)
	fmt.Printf(";; Exported:   %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(";;")
	fmt.Println(";; This file is intended for use for import into Cloudflare's DNS service.  If you are")
	fmt.Println(";; having trouble importing this file, please reach out to support.")
	fmt.Println(";;")
	fmt.Println()

	// A Records - Only mail server
	fmt.Println(";; A Records")
	fmt.Printf("%s.\t1\tIN\tA\t%s\n", cfg.Hostname, cfg.ServerIP)
	if cfg.ServerIPv6 != "" {
		fmt.Printf("%s.\t1\tIN\tAAAA\t%s\n", cfg.Hostname, cfg.ServerIPv6)
	}
	fmt.Println()

	// MX Records
	fmt.Println(";; MX Records")
	fmt.Printf("%s.\t1\tIN\tMX\t10 %s.\n", cfg.PrimaryDomain, cfg.Hostname)
	fmt.Println()

	// TXT Records - All in one section
	fmt.Println(";; TXT Records")
	// DKIM
	fmt.Printf("default._domainkey.%s.\t1\tIN\tTXT\t\"%s\"\n", cfg.PrimaryDomain, cfg.DKIMPublicKey)
	// SPF
	fmt.Printf("%s.\t1\tIN\tTXT\t\"v=spf1 mx ~all\"\n", cfg.PrimaryDomain)
	fmt.Printf("%s.\t1\tIN\tTXT\t\"v=spf1 a ~all\"\n", cfg.Hostname)
	// DMARC
	dmarcPolicy := cfg.DMARCPolicy
	if dmarcPolicy == "" {
		dmarcPolicy = "quarantine"
	}
	fmt.Printf("_dmarc.%s.\t1\tIN\tTXT\t\"v=DMARC1; p=%s; ruf=mailto:%s\"\n", cfg.PrimaryDomain, dmarcPolicy, cfg.PostmasterEmail)
	// MTA-STS and TLSRPT
	fmt.Printf("_mta-sts.%s.\t1\tIN\tTXT\t\"v=STSv1; id=1\"\n", cfg.PrimaryDomain)
	fmt.Printf("_smtp._tls.%s.\t1\tIN\tTXT\t\"v=TLSRPTv1; rua=mailto:%s\"\n", cfg.PrimaryDomain, cfg.PostmasterEmail)
	fmt.Println()

	// Print reminder after the records
	fmt.Println(";; IMPORTANT: After import, edit the A record for the mail server")
	fmt.Println(";; and DISABLE the Cloudflare proxy (change orange cloud to grey).")
}

func exportGenericFormat(cfg *DNSConfig) {
	fmt.Println("\n通用文本格式:")
	fmt.Println("================================================================================")
	fmt.Println("基础记录:")
	fmt.Printf("A记录 (主域名):\n  主机: %s\n  IP: %s\n", cfg.PrimaryDomain, cfg.ServerIP)
	if cfg.ServerIPv6 != "" {
		fmt.Printf("\nAAAA记录 (主域名):\n  主机: %s\n  IPv6: %s\n", cfg.PrimaryDomain, cfg.ServerIPv6)
	}
	fmt.Printf("\nA记录 (邮件服务器):\n  主机: %s\n  IP: %s\n", cfg.Hostname, cfg.ServerIP)
	if cfg.ServerIPv6 != "" {
		fmt.Printf("\nAAAA记录 (邮件服务器):\n  主机: %s\n  IPv6: %s\n", cfg.Hostname, cfg.ServerIPv6)
	}
	fmt.Println("\n--------------------------------------------------------------------------------")
	fmt.Printf("MX记录:\n  域名: %s\n  邮件服务器: %s\n  优先级: 10\n", cfg.PrimaryDomain, cfg.Hostname)
	fmt.Println("\n--------------------------------------------------------------------------------")
	fmt.Println("安全相关记录:")
	fmt.Printf("SPF记录 (主域名TXT):\n  主机: %s\n  值: v=spf1 mx ~all\n", cfg.PrimaryDomain)
	fmt.Printf("\nSPF记录 (MX主机TXT):\n  主机: %s\n  值: v=spf1 a ~all\n", cfg.Hostname)
	fmt.Printf("\nDKIM记录 (TXT):\n  主机: default._domainkey.%s\n  值: %s\n", cfg.PrimaryDomain, cfg.DKIMPublicKey)
	dmarcPolicy := cfg.DMARCPolicy
	if dmarcPolicy == "" {
		dmarcPolicy = "quarantine"
	}
	fmt.Printf("\nDMARC记录 (TXT):\n  主机: _dmarc.%s\n  值: v=DMARC1; p=%s; ruf=mailto:%s\n", cfg.PrimaryDomain, dmarcPolicy, cfg.PostmasterEmail)
	fmt.Println("\n--------------------------------------------------------------------------------")
	fmt.Println("TLS安全记录 (推荐):")
	fmt.Printf("MTA-STS记录 (TXT):\n  主机: _mta-sts.%s\n  值: v=STSv1; id=1\n", cfg.PrimaryDomain)
	fmt.Printf("\nTLSRPT记录 (TXT):\n  主机: _smtp._tls.%s\n  值: v=TLSRPTv1; rua=mailto:%s\n", cfg.PrimaryDomain, cfg.PostmasterEmail)
	fmt.Println("================================================================================")
}