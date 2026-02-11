// Package certificate provides TLS certificate management functionality.
package certificate

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"v/internal/database/repository"
	"v/internal/logger"
)

// Service provides certificate management operations.
type Service struct {
	certRepo repository.CertificateRepository
	logger   logger.Logger
	certDir  string // 证书存储目录
}

// NewService creates a new certificate service.
func NewService(
	certRepo repository.CertificateRepository,
	log logger.Logger,
	certDir string,
) *Service {
	return &Service{
		certRepo: certRepo,
		logger:   log,
		certDir:  certDir,
	}
}

// ApplyRequest represents a certificate application request.
type ApplyRequest struct {
	Domain   string
	Email    string
	Provider string // "letsencrypt" or "zerossl"
	Method   string // "http" or "dns"
}

// Apply applies for a new certificate using acme.sh.
func (s *Service) Apply(ctx context.Context, req *ApplyRequest) (*repository.Certificate, error) {
	s.logger.Info("申请证书",
		logger.F("domain", req.Domain),
		logger.F("provider", req.Provider))

	// 检查 acme.sh 是否安装
	if !s.isAcmeInstalled() {
		return nil, fmt.Errorf("acme.sh 未安装，请先安装: curl https://get.acme.sh | sh")
	}

	// 设置默认值
	if req.Provider == "" {
		req.Provider = "letsencrypt"
	}
	if req.Method == "" {
		req.Method = "http"
	}

	// 创建证书记录
	cert := &repository.Certificate{
		Domain:    req.Domain,
		Provider:  req.Provider,
		AutoRenew: true,
		Status:    "pending",
	}

	if err := s.certRepo.Create(ctx, cert); err != nil {
		return nil, fmt.Errorf("创建证书记录失败: %w", err)
	}

	// 异步申请证书
	go func() {
		if err := s.issueWithAcme(req, cert); err != nil {
			s.logger.Error("证书申请失败",
				logger.F("domain", req.Domain),
				logger.F("error", err.Error()))
			
			cert.Status = "failed"
			cert.ErrorMessage = err.Error()
			s.certRepo.Update(context.Background(), cert)
		}
	}()

	return cert, nil
}

// isAcmeInstalled checks if acme.sh is installed.
func (s *Service) isAcmeInstalled() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	
	acmePath := filepath.Join(homeDir, ".acme.sh", "acme.sh")
	_, err = os.Stat(acmePath)
	return err == nil
}

// issueWithAcme issues a certificate using acme.sh.
func (s *Service) issueWithAcme(req *ApplyRequest, cert *repository.Certificate) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败: %w", err)
	}

	acmePath := filepath.Join(homeDir, ".acme.sh", "acme.sh")

	// 构建命令参数
	args := []string{
		"--issue",
		"-d", req.Domain,
		"--keylength", "ec-256", // 使用 ECC 证书
	}

	// 设置 CA
	if req.Provider == "zerossl" {
		args = append(args, "--server", "zerossl")
	} else {
		args = append(args, "--server", "letsencrypt")
	}

	// 设置验证方式
	if req.Method == "dns" {
		args = append(args, "--dns")
	} else {
		// HTTP 验证需要 webroot
		args = append(args, "-w", "/var/www/html")
	}

	// 执行申请命令
	s.logger.Info("执行 acme.sh 申请", logger.F("args", strings.Join(args, " ")))
	
	cmd := exec.Command(acmePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("acme.sh 申请失败: %s, output: %s", err.Error(), string(output))
	}

	s.logger.Info("证书申请成功", logger.F("domain", req.Domain))

	// 安装证书到指定目录
	certPath := filepath.Join(s.certDir, req.Domain)
	if err := os.MkdirAll(certPath, 0755); err != nil {
		return fmt.Errorf("创建证书目录失败: %w", err)
	}

	certFile := filepath.Join(certPath, "fullchain.pem")
	keyFile := filepath.Join(certPath, "privkey.pem")

	installArgs := []string{
		"--installcert",
		"-d", req.Domain,
		"--fullchain-file", certFile,
		"--key-file", keyFile,
		"--ecc",
	}

	cmd = exec.Command(acmePath, installArgs...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("安装证书失败: %s, output: %s", err.Error(), string(output))
	}

	// 读取证书信息
	certData, err := os.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("读取证书文件失败: %w", err)
	}

	// 解析证书获取过期时间
	block, _ := pem.Decode(certData)
	if block == nil {
		return fmt.Errorf("解析证书失败")
	}

	x509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("解析证书失败: %w", err)
	}

	// 更新证书记录
	now := time.Now()
	cert.CertPath = certFile
	cert.KeyPath = keyFile
	cert.IssueDate = &now
	cert.ExpireDate = &x509Cert.NotAfter
	cert.Status = "active"
	cert.ErrorMessage = ""

	return s.certRepo.Update(context.Background(), cert)
}

// Upload uploads a certificate manually.
func (s *Service) Upload(ctx context.Context, domain string, certData, keyData []byte) (*repository.Certificate, error) {
	s.logger.Info("上传证书", logger.F("domain", domain))

	// 验证证书格式
	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, fmt.Errorf("无效的证书格式")
	}

	x509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	// 验证私钥格式
	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil {
		return nil, fmt.Errorf("无效的私钥格式")
	}

	// 保存证书文件
	certPath := filepath.Join(s.certDir, domain)
	if err := os.MkdirAll(certPath, 0755); err != nil {
		return nil, fmt.Errorf("创建证书目录失败: %w", err)
	}

	certFile := filepath.Join(certPath, "fullchain.pem")
	keyFile := filepath.Join(certPath, "privkey.pem")

	if err := os.WriteFile(certFile, certData, 0644); err != nil {
		return nil, fmt.Errorf("保存证书文件失败: %w", err)
	}

	if err := os.WriteFile(keyFile, keyData, 0600); err != nil {
		return nil, fmt.Errorf("保存私钥文件失败: %w", err)
	}

	// 创建证书记录
	now := time.Now()
	cert := &repository.Certificate{
		Domain:      domain,
		Provider:    "manual",
		CertPath:    certFile,
		KeyPath:     keyFile,
		IssueDate:   &now,
		ExpireDate:  &x509Cert.NotAfter,
		AutoRenew:   false,
		Status:      "active",
	}

	if err := s.certRepo.Create(ctx, cert); err != nil {
		return nil, fmt.Errorf("创建证书记录失败: %w", err)
	}

	s.logger.Info("证书上传成功", logger.F("domain", domain))
	return cert, nil
}

// Renew renews a certificate.
func (s *Service) Renew(ctx context.Context, id int64) error {
	cert, err := s.certRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("获取证书失败: %w", err)
	}

	if cert.Provider == "manual" {
		return fmt.Errorf("手动上传的证书不支持自动续期")
	}

	s.logger.Info("续期证书", logger.F("domain", cert.Domain))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败: %w", err)
	}

	acmePath := filepath.Join(homeDir, ".acme.sh", "acme.sh")

	// 执行续期命令
	args := []string{
		"--renew",
		"-d", cert.Domain,
		"--ecc",
		"--force",
	}

	cmd := exec.Command(acmePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("续期失败: %s, output: %s", err.Error(), string(output))
	}

	// 重新安装证书
	certFile := cert.CertPath
	keyFile := cert.KeyPath

	installArgs := []string{
		"--installcert",
		"-d", cert.Domain,
		"--fullchain-file", certFile,
		"--key-file", keyFile,
		"--ecc",
	}

	cmd = exec.Command(acmePath, installArgs...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("安装证书失败: %s, output: %s", err.Error(), string(output))
	}

	// 读取新证书信息
	certData, err := os.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("读取证书文件失败: %w", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return fmt.Errorf("解析证书失败")
	}

	x509Cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("解析证书失败: %w", err)
	}

	// 更新证书记录
	now := time.Now()
	cert.IssueDate = &now
	cert.ExpireDate = &x509Cert.NotAfter

	if err := s.certRepo.Update(ctx, cert); err != nil {
		return fmt.Errorf("更新证书记录失败: %w", err)
	}

	s.logger.Info("证书续期成功", logger.F("domain", cert.Domain))
	return nil
}

// Delete deletes a certificate.
func (s *Service) Delete(ctx context.Context, id int64) error {
	cert, err := s.certRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("获取证书失败: %w", err)
	}

	// 删除证书文件
	if cert.CertPath != "" {
		certDir := filepath.Dir(cert.CertPath)
		if err := os.RemoveAll(certDir); err != nil {
			s.logger.Warn("删除证书文件失败", logger.F("error", err.Error()))
		}
	}

	// 删除数据库记录
	if err := s.certRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("删除证书记录失败: %w", err)
	}

	s.logger.Info("证书删除成功", logger.F("domain", cert.Domain))
	return nil
}

// List lists all certificates.
func (s *Service) List(ctx context.Context) ([]*repository.Certificate, error) {
	return s.certRepo.List(ctx, 1000, 0)
}

// GetByID gets a certificate by ID.
func (s *Service) GetByID(ctx context.Context, id int64) (*repository.Certificate, error) {
	return s.certRepo.GetByID(ctx, id)
}

// GetByDomain gets a certificate by domain.
func (s *Service) GetByDomain(ctx context.Context, domain string) (*repository.Certificate, error) {
	return s.certRepo.GetByDomain(ctx, domain)
}

// UpdateAutoRenew updates auto-renew setting.
func (s *Service) UpdateAutoRenew(ctx context.Context, id int64, autoRenew bool) error {
	cert, err := s.certRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("获取证书失败: %w", err)
	}

	cert.AutoRenew = autoRenew
	if err := s.certRepo.Update(ctx, cert); err != nil {
		return fmt.Errorf("更新证书失败: %w", err)
	}

	s.logger.Info("更新自动续期设置",
		logger.F("domain", cert.Domain),
		logger.F("auto_renew", autoRenew))
	
	return nil
}

// CheckExpiring checks for expiring certificates and renews them if auto-renew is enabled.
func (s *Service) CheckExpiring(ctx context.Context) error {
	certs, err := s.certRepo.List(ctx, 1000, 0)
	if err != nil {
		return fmt.Errorf("获取证书列表失败: %w", err)
	}

	now := time.Now()
	renewThreshold := 30 * 24 * time.Hour // 30 天内过期

	for _, cert := range certs {
		if cert.ExpireDate == nil {
			continue
		}

		timeUntilExpiry := cert.ExpireDate.Sub(now)
		
		// 检查是否即将过期
		if timeUntilExpiry < renewThreshold && timeUntilExpiry > 0 {
			s.logger.Info("证书即将过期",
				logger.F("domain", cert.Domain),
				logger.F("days_left", int(timeUntilExpiry.Hours()/24)))

			// 如果启用了自动续期，则续期
			if cert.AutoRenew && cert.Provider != "manual" {
				s.logger.Info("自动续期证书", logger.F("domain", cert.Domain))
				if err := s.Renew(ctx, cert.ID); err != nil {
					s.logger.Error("自动续期失败",
						logger.F("domain", cert.Domain),
						logger.F("error", err.Error()))
				}
			}
		}
	}

	return nil
}

// GenerateSelfSigned generates a self-signed certificate for testing.
func (s *Service) GenerateSelfSigned(ctx context.Context, domain string) (*repository.Certificate, error) {
	s.logger.Info("生成自签名证书", logger.F("domain", domain))

	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("生成私钥失败: %w", err)
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: domain,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
	}

	// 生成证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("生成证书失败: %w", err)
	}

	// 编码证书
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// 编码私钥
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// 保存文件
	certPath := filepath.Join(s.certDir, domain)
	if err := os.MkdirAll(certPath, 0755); err != nil {
		return nil, fmt.Errorf("创建证书目录失败: %w", err)
	}

	certFile := filepath.Join(certPath, "fullchain.pem")
	keyFile := filepath.Join(certPath, "privkey.pem")

	if err := os.WriteFile(certFile, certPEM, 0644); err != nil {
		return nil, fmt.Errorf("保存证书文件失败: %w", err)
	}

	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		return nil, fmt.Errorf("保存私钥文件失败: %w", err)
	}

	// 创建证书记录
	now := time.Now()
	expireDate := now.Add(365 * 24 * time.Hour)
	
	cert := &repository.Certificate{
		Domain:      domain,
		Provider:    "self-signed",
		CertPath:    certFile,
		KeyPath:     keyFile,
		IssueDate:   &now,
		ExpireDate:  &expireDate,
		AutoRenew:   false,
		Status:      "active",
	}

	if err := s.certRepo.Create(ctx, cert); err != nil {
		return nil, fmt.Errorf("创建证书记录失败: %w", err)
	}

	s.logger.Info("自签名证书生成成功", logger.F("domain", domain))
	return cert, nil
}
