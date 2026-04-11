package resolution

import (
	"errors"
	"goaway/backend/settings"
	"strings"
)

type Repository interface {
	CreateResolution(ip, domain string) error
	FindResolution(domain string) (string, error)
	FindResolutions() (map[string]string, error)
	DeleteResolution(ip, domain string) (int, error)
}

type repository struct {
}

func NewRepository() Repository {
	return &repository{}
}

func (r *repository) CreateResolution(ip, domain string) error {
	cfg, err := settings.LoadSettings()
	if err != nil {
		return err
	}
	if cfg.DNS.Resolutions == nil {
		cfg.DNS.Resolutions = make(map[string]string)
	}

	if _, exists := cfg.DNS.Resolutions[domain]; exists {
		return errors.New("domain already exists, must be unique")
	}

	cfg.DNS.Resolutions[domain] = ip
	cfg.Save()
	return nil
}

func (r *repository) FindResolution(domain string) (string, error) {
	cfg, err := settings.LoadSettings()
	if err != nil {
		return "", err
	}
	if ip, ok := cfg.DNS.Resolutions[domain]; ok {
		return ip, nil
	}

	parts := strings.Split(domain, ".")
	for i := 1; i < len(parts); i++ {
		wildcardDomain := "*." + strings.Join(parts[i:], ".")
		if ip, ok := cfg.DNS.Resolutions[wildcardDomain]; ok {
			return ip, nil
		}
	}

	return "", nil
}

func (r *repository) FindResolutions() (map[string]string, error) {
	cfg, err := settings.LoadSettings()
	if err != nil {
		return nil, err
	}
	if cfg.DNS.Resolutions == nil {
		return map[string]string{}, nil
	}
	return cfg.DNS.Resolutions, nil
}

func (r *repository) DeleteResolution(ip, domain string) (int, error) {
	cfg, err := settings.LoadSettings()
	if err != nil {
		return 0, err
	}
	if cur, ok := cfg.DNS.Resolutions[domain]; ok {
		if cur == ip {
			delete(cfg.DNS.Resolutions, domain)
			cfg.Save()
			return 1, nil
		}
		return 0, nil
	}

	return 0, nil
}
