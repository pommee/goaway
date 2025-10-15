package whitelist

import "goaway/backend/logging"

type Service struct {
	repository Repository
	Cache      map[string]bool
}

var log = logging.GetLogger()

func NewService(repo Repository) *Service {
	service := &Service{
		repository: repo,
		Cache:      map[string]bool{},
	}

	_, err := service.GetDomains() // Preload cache
	if err != nil {
		log.Warning("Could not preload domains cache, %v", err)
	}

	return service
}

func (s *Service) AddDomain(domain string) error {
	err := s.repository.AddDomain(domain)
	if err != nil {
		return err
	}

	s.Cache[domain] = true
	return nil
}

func (s *Service) GetDomains() (map[string]bool, error) {
	domains, err := s.repository.GetDomains()
	if err != nil {
		return nil, err
	}

	s.Cache = domains
	return domains, nil
}

func (s *Service) RemoveDomain(domain string) error {
	err := s.repository.RemoveDomain(domain)
	if err != nil {
		return err
	}

	delete(s.Cache, domain)
	return nil
}

func (s *Service) IsWhitelisted(domain string) bool {
	_, exists := s.Cache[domain]
	return exists
}
