package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"sync"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationNewDevice       NotificationType = "new_device"
	NotificationIPLimitReached  NotificationType = "ip_limit_reached"
	NotificationSuspiciousIP    NotificationType = "suspicious_ip"
	NotificationDeviceKicked    NotificationType = "device_kicked"
	NotificationAutoBlacklisted NotificationType = "auto_blacklisted"
)

// NotificationChannel represents the notification channel
type NotificationChannel string

const (
	ChannelEmail    NotificationChannel = "email"
	ChannelTelegram NotificationChannel = "telegram"
)

// NotificationConfig holds the notification configuration
type NotificationConfig struct {
	// Email settings
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	// Telegram settings
	TelegramBotToken string
	TelegramChatID   string

	// Notification preferences
	EnabledTypes    map[NotificationType]bool
	EnabledChannels map[NotificationChannel]bool
}

// IPNotificationData contains data for IP-related notifications
type IPNotificationData struct {
	UserID       uint
	Username     string
	Email        string
	IP           string
	Country      string
	City         string
	DeviceInfo   string
	Reason       string
	CurrentCount int
	MaxCount     int
	Timestamp    time.Time
}

// Service handles sending notifications
type Service struct {
	config *NotificationConfig
	mu     sync.RWMutex
	client *http.Client
}

// NewService creates a new notification service
func NewService(config *NotificationConfig) *Service {
	return &Service{
		config: config,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// UpdateConfig updates the notification configuration
func (s *Service) UpdateConfig(config *NotificationConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}

// NotifyNewDevice sends notification when a new device connects
func (s *Service) NotifyNewDevice(data IPNotificationData) error {
	if !s.isEnabled(NotificationNewDevice) {
		return nil
	}

	subject := "新设备连接通知"
	message := fmt.Sprintf(
		"用户 %s 有新设备连接\n\nIP: %s\n位置: %s %s\n设备: %s\n时间: %s",
		data.Username,
		data.IP,
		data.Country,
		data.City,
		data.DeviceInfo,
		data.Timestamp.Format("2006-01-02 15:04:05"),
	)

	return s.send(data.Email, subject, message)
}

// NotifyIPLimitReached sends notification when IP limit is reached
func (s *Service) NotifyIPLimitReached(data IPNotificationData) error {
	if !s.isEnabled(NotificationIPLimitReached) {
		return nil
	}

	subject := "设备数量已达上限"
	message := fmt.Sprintf(
		"用户 %s 的设备数量已达上限\n\n当前设备数: %d\n最大设备数: %d\n\n新连接 IP: %s\n位置: %s %s\n\n请断开其他设备后重试",
		data.Username,
		data.CurrentCount,
		data.MaxCount,
		data.IP,
		data.Country,
		data.City,
	)

	return s.send(data.Email, subject, message)
}

// NotifySuspiciousActivity sends notification for suspicious IP activity
func (s *Service) NotifySuspiciousActivity(data IPNotificationData) error {
	if !s.isEnabled(NotificationSuspiciousIP) {
		return nil
	}

	subject := "⚠️ 可疑活动告警"
	message := fmt.Sprintf(
		"检测到可疑活动\n\n用户: %s\nIP: %s\n位置: %s %s\n原因: %s\n时间: %s",
		data.Username,
		data.IP,
		data.Country,
		data.City,
		data.Reason,
		data.Timestamp.Format("2006-01-02 15:04:05"),
	)

	// Suspicious activity notifications go to admin
	return s.sendToAdmin(subject, message)
}

// NotifyDeviceKicked sends notification when a device is kicked
func (s *Service) NotifyDeviceKicked(data IPNotificationData) error {
	if !s.isEnabled(NotificationDeviceKicked) {
		return nil
	}

	subject := "设备已被踢出"
	message := fmt.Sprintf(
		"您的设备已被踢出\n\nIP: %s\n位置: %s %s\n原因: %s\n时间: %s\n\n如非本人操作，请检查账号安全",
		data.IP,
		data.Country,
		data.City,
		data.Reason,
		data.Timestamp.Format("2006-01-02 15:04:05"),
	)

	return s.send(data.Email, subject, message)
}

// NotifyAutoBlacklisted sends notification when an IP is auto-blacklisted
func (s *Service) NotifyAutoBlacklisted(data IPNotificationData) error {
	if !s.isEnabled(NotificationAutoBlacklisted) {
		return nil
	}

	subject := "IP 已被自动封禁"
	message := fmt.Sprintf(
		"IP 已被自动加入黑名单\n\nIP: %s\n位置: %s %s\n原因: %s\n时间: %s",
		data.IP,
		data.Country,
		data.City,
		data.Reason,
		data.Timestamp.Format("2006-01-02 15:04:05"),
	)

	return s.sendToAdmin(subject, message)
}

// isEnabled checks if a notification type is enabled
func (s *Service) isEnabled(notifType NotificationType) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.config == nil || s.config.EnabledTypes == nil {
		return false
	}
	return s.config.EnabledTypes[notifType]
}

// send sends notification through all enabled channels
func (s *Service) send(email, subject, message string) error {
	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	if config == nil {
		return nil
	}

	var errs []string

	// Send email if enabled
	if config.EnabledChannels[ChannelEmail] && email != "" {
		if err := s.sendEmail(email, subject, message); err != nil {
			errs = append(errs, fmt.Sprintf("email: %v", err))
		}
	}

	// Send Telegram if enabled
	if config.EnabledChannels[ChannelTelegram] {
		if err := s.sendTelegram(subject + "\n\n" + message); err != nil {
			errs = append(errs, fmt.Sprintf("telegram: %v", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("notification errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// sendToAdmin sends notification to admin only
func (s *Service) sendToAdmin(subject, message string) error {
	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	if config == nil {
		return nil
	}

	// Only send via Telegram for admin notifications
	if config.EnabledChannels[ChannelTelegram] {
		return s.sendTelegram("🔔 " + subject + "\n\n" + message)
	}
	return nil
}

// sendEmail sends an email notification
func (s *Service) sendEmail(to, subject, body string) error {
	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	if config.SMTPHost == "" || config.SMTPUser == "" {
		return fmt.Errorf("SMTP not configured")
	}

	from := config.SMTPFrom
	if from == "" {
		from = config.SMTPUser
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body)

	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)
	auth := smtp.PlainAuth("", config.SMTPUser, config.SMTPPassword, config.SMTPHost)

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

// sendTelegram sends a Telegram notification
func (s *Service) sendTelegram(message string) error {
	s.mu.RLock()
	config := s.config
	s.mu.RUnlock()

	if config.TelegramBotToken == "" || config.TelegramChatID == "" {
		return fmt.Errorf("Telegram not configured")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.TelegramBotToken)

	payload := map[string]interface{}{
		"chat_id":    config.TelegramChatID,
		"text":       message,
		"parse_mode": "HTML",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	return nil
}
