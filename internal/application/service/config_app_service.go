package service

import (
	"catch/internal/application/dto"
	"catch/internal/domain/entity"
	"catch/internal/domain/repository"
	"fmt"
)

type ConfigAppService struct {
	configRepo repository.ConfigRepository
}

func NewConfigAppService(configRepo repository.ConfigRepository) *ConfigAppService {
	return &ConfigAppService{configRepo: configRepo}
}

func (s *ConfigAppService) GetConfig() (*dto.ConfigResponse, error) {
	config, err := s.configRepo.Load()
	if err != nil {
		return nil, fmt.Errorf("无法加载配置: %w", err)
	}

	return s.toConfigResponse(config), nil
}

func (s *ConfigAppService) UpdateConfig(req dto.UpdateConfigRequest) (*dto.ConfigResponse, error) {
	config, err := s.configRepo.Load()
	if err != nil {
		return nil, fmt.Errorf("无法加载配置: %w", err)
	}

	if req.Server != nil {
		config.Server.Port = req.Server.Port
	}
	if req.Trash != nil {
		config.Trash.ExpireDays = req.Trash.ExpireDays
		config.Trash.Path = req.Trash.Path
	}
	if req.Security != nil {
		config.Security.PasswordHint = req.Security.PasswordHint
	}
	if req.SMTP != nil {
		config.SMTP.Host = req.SMTP.Host
		config.SMTP.Port = req.SMTP.Port
		config.SMTP.Username = req.SMTP.Username
		if req.SMTP.Password != "" {
			config.SMTP.Password = req.SMTP.Password
		}
		config.SMTP.To = req.SMTP.To
	}
	if req.Favorites != nil {
		config.Favorites = req.Favorites
	}
	if req.Search != nil {
		config.Search.DefaultPath = req.Search.DefaultPath
	}

	if err := s.configRepo.Save(config); err != nil {
		return nil, fmt.Errorf("无法保存配置: %w", err)
	}

	return s.toConfigResponse(config), nil
}

func (s *ConfigAppService) SetPassword(req dto.SetPasswordRequest) error {
	config, err := s.configRepo.Load()
	if err != nil {
		return fmt.Errorf("无法加载配置: %w", err)
	}

	if config.HasPassword() && config.Security.Password != req.OldPassword {
		return fmt.Errorf("旧密码验证失败")
	}

	config.Security.Password = req.NewPassword
	config.Security.PasswordHint = req.PasswordHint

	return s.configRepo.Save(config)
}

func (s *ConfigAppService) VerifyPassword(req dto.VerifyPasswordRequest) bool {
	config, err := s.configRepo.Load()
	if err != nil {
		return false
	}
	return config.ValidatePassword(req.Password)
}

func (s *ConfigAppService) EnsureConfig() error {
	if !s.configRepo.Exists() {
		defaultConfig := entity.DefaultAppConfig()
		return s.configRepo.Save(defaultConfig)
	}
	return nil
}

func (s *ConfigAppService) toConfigResponse(config *entity.AppConfig) *dto.ConfigResponse {
	return &dto.ConfigResponse{
		Version: config.Version,
		Server: dto.ServerDTO{
			Port: config.Server.Port,
		},
		Trash: dto.TrashDTO{
			ExpireDays: config.Trash.ExpireDays,
			Path:       config.Trash.Path,
		},
		Security: dto.SecurityDTO{
			PasswordHint: config.Security.PasswordHint,
			HasPassword:  config.HasPassword(),
		},
		SMTP: dto.SMTPDTO{
			Host:     config.SMTP.Host,
			Port:     config.SMTP.Port,
			Username: config.SMTP.Username,
			To:       config.SMTP.To,
			HasSMTP:  config.HasSMTP(),
		},
		Favorites: config.Favorites,
		Search: dto.SearchDTO{
			DefaultPath: config.Search.DefaultPath,
		},
		HasPassword: config.HasPassword(),
		HasSMTP:     config.HasSMTP(),
	}
}
