package mac

type Service struct {
	repository Repository
}

func NewService(repo Repository) *Service {
	return &Service{repository: repo}
}

func (s *Service) FindVendor(mac string) (string, error) {
	return s.repository.FindVendor(mac)
}

func (s *Service) SaveMac(clientIP, mac, vendor string) error {
	return s.repository.SaveMac(clientIP, mac, vendor)
}
