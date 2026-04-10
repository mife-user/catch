package service

import (
	"catch/internal/application/dto"
	"catch/internal/domain/repository"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"runtime"
	"strings"
	"time"
)

type FeedbackAppService struct {
	configRepo repository.ConfigRepository
}

func NewFeedbackAppService(configRepo repository.ConfigRepository) *FeedbackAppService {
	return &FeedbackAppService{configRepo: configRepo}
}

func (s *FeedbackAppService) SendFeedback(req dto.FeedbackRequest) (*dto.FeedbackResponse, error) {
	config, err := s.configRepo.Load()
	if err != nil {
		return nil, fmt.Errorf("无法加载配置: %w", err)
	}

	if !config.HasSMTP() {
		return &dto.FeedbackResponse{
			Success: false,
			Message: "未配置SMTP，请先在设置中配置SMTP信息",
		}, nil
	}

	subject := fmt.Sprintf("【Catch反馈】%s - %s", req.Type, time.Now().Format("2006-01-02"))
	body := fmt.Sprintf(
		"反馈类型：%s\n反馈内容：\n%s\n\n---\n系统环境：\n操作系统：%s %s\nCatch版本：1.0.0\n提交时间：%s",
		req.Type,
		req.Content,
		runtime.GOOS,
		runtime.GOARCH,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	if err := s.sendEmail(config.SMTP.Host, config.SMTP.Port, config.SMTP.Username, config.SMTP.Password, config.SMTP.To, subject, body); err != nil {
		return &dto.FeedbackResponse{
			Success: false,
			Message: fmt.Sprintf("邮件发送失败: %s", err.Error()),
		}, nil
	}

	return &dto.FeedbackResponse{
		Success: true,
		Message: "反馈已发送",
	}, nil
}

func (s *FeedbackAppService) TestSMTP(req dto.SMTPTestRequest) (*dto.SMTPTestResponse, error) {
	if err := s.sendEmail(req.Host, req.Port, req.Username, req.Password, req.To, "Catch SMTP测试", "这是一封测试邮件，用于验证SMTP配置是否正确。"); err != nil {
		return &dto.SMTPTestResponse{
			Success: false,
			Message: fmt.Sprintf("SMTP测试失败: %s", err.Error()),
		}, nil
	}

	return &dto.SMTPTestResponse{
		Success: true,
		Message: "SMTP配置正确，测试邮件已发送",
	}, nil
}

func (s *FeedbackAppService) GetSMTPTemplates() []dto.SMTPTemplate {
	return []dto.SMTPTemplate{
		{Name: "QQ邮箱", Host: "smtp.qq.com", Port: 465},
		{Name: "163邮箱", Host: "smtp.163.com", Port: 465},
		{Name: "Gmail", Host: "smtp.gmail.com", Port: 587},
		{Name: "Outlook", Host: "smtp.office365.com", Port: 587},
	}
}

func (s *FeedbackAppService) sendEmail(host string, port int, username, password, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", host, port)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		username, to, subject, body,
	)

	if port == 465 {
		return s.sendWithTLS(addr, username, password, to, []byte(msg))
	}
	return s.sendWithSTARTTLS(addr, username, password, to, []byte(msg))
}

func (s *FeedbackAppService) sendWithTLS(addr, username, password, to string, msg []byte) error {
	host := strings.Split(addr, ":")[0]
	tlsConfig := &tls.Config{
		ServerName: host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS连接失败: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("创建SMTP客户端失败: %w", err)
	}
	defer client.Close()

	if err := client.Auth(smtp.PlainAuth("", username, password, host)); err != nil {
		return fmt.Errorf("SMTP认证失败: %w", err)
	}

	if err := client.Mail(username); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("设置收件人失败: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("准备发送数据失败: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("关闭数据写入失败: %w", err)
	}

	return client.Quit()
}

func (s *FeedbackAppService) sendWithSTARTTLS(addr, username, password, to string, msg []byte) error {
	host := strings.Split(addr, ":")[0]

	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("连接SMTP服务器失败: %w", err)
	}
	defer client.Close()

	tlsConfig := &tls.Config{ServerName: host}
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("STARTTLS失败: %w", err)
	}

	if err := client.Auth(smtp.PlainAuth("", username, password, host)); err != nil {
		return fmt.Errorf("SMTP认证失败: %w", err)
	}

	if err := client.Mail(username); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("设置收件人失败: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("准备发送数据失败: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("关闭数据写入失败: %w", err)
	}

	return client.Quit()
}
