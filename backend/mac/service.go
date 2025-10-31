package mac

import "goaway/backend/logging"

type Service struct {
	repository Repository
}

var log = logging.GetLogger()

func NewService(repo Repository) *Service {
	return &Service{repository: repo}
}

func (s *Service) FindVendor(mac string) (string, error) {
	return s.repository.FindVendor(mac)
}

func (s *Service) SaveMac(clientIP, mac, vendor string) {
	err := s.repository.SaveMac(clientIP, mac, vendor)
	if err != nil {
		log.Warning("Could not save MAC address, %v", err)
	}
}
