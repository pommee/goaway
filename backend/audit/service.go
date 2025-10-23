package audit

import (
	"goaway/backend/logging"
	"time"
)

type Topic string

const (
	TopicServer     Topic = "server"
	TopicDNS        Topic = "dns"
	TopicAPI        Topic = "api"
	TopicResolution Topic = "resolution"
	TopicPrefetch   Topic = "prefetch"
	TopicUpstream   Topic = "upstream"
	TopicUser       Topic = "user"
	TopicList       Topic = "list"
	TopicLogs       Topic = "logs"
	TopicSettings   Topic = "settings"
	TopicDatabase   Topic = "database"
)

type Entry struct {
	CreatedAt time.Time `json:"createdAt"`
	Topic     Topic     `json:"topic"`
	Message   string    `json:"message"`
	ID        uint      `json:"id"`
}

var log = logging.GetLogger()

type Service struct {
	repository Repository
}

func NewService(repo Repository) *Service {
	return &Service{repository: repo}
}

func (s *Service) CreateAudit(entry *Entry) {
	err := s.repository.CreateAudit(entry)
	if err != nil {
		log.Warning("Could not create audit: %v", err)
	}
}

func (s *Service) ReadAudits() ([]Entry, error) {
	return s.repository.ReadAudits()
}
