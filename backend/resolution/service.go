package resolution

import (
	"goaway/backend/database"
	"goaway/backend/logging"
)

type Service struct {
	repository Repository
}

var log = logging.GetLogger()

func NewService(repo Repository) *Service {
	return &Service{repository: repo}
}

func (s *Service) CreateResolution(ip, domain string) error {
	log.Debug("Creating new resolution '%s' -> '%s'", domain, ip)
	return s.repository.CreateResolution(ip, domain)
}

func (s *Service) GetResolution(domain string) (string, error) {
	log.Debug("Finding resolution for domain: %s", domain)
	return s.repository.FindResolution(domain)
}

func (s *Service) GetResolutions() ([]database.Resolution, error) {
	return s.repository.FindResolutions()
}

func (s *Service) DeleteResolution(ip, domain string) (int, error) {
	return s.repository.DeleteResolution(ip, domain)
}
