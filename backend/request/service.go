package request

import (
	"context"
	"goaway/backend/api/models"
	model "goaway/backend/dns/server/models"
	"goaway/backend/logging"
	"time"
)

type Service struct {
	repository Repository
}

var log = logging.GetLogger()

func NewService(repo Repository) *Service {
	return &Service{repository: repo}
}

func (s *Service) GetClientNameFromIP(ip string) string {
	return s.repository.GetClientName(ip)
}

func (s *Service) GetDistinctRequestIP() int {
	return s.repository.GetDistinctRequestIP()
}

func (s *Service) GetRequestSummaryByInterval(interval int) ([]model.RequestLogIntervalSummary, error) {
	return s.repository.GetRequestSummaryByInterval(interval)
}

func (s *Service) GetResponseSizeSummaryByInterval(intervalMinutes int) ([]model.ResponseSizeSummary, error) {
	return s.repository.GetResponseSizeSummaryByInterval(intervalMinutes)
}

func (s *Service) GetUniqueQueryTypes() ([]interface{}, error) {
	return s.repository.GetUniqueQueryTypes()
}

func (s *Service) FetchQueries(q models.QueryParams) ([]model.RequestLogEntry, error) {
	return s.repository.FetchQueries(q)
}

func (s *Service) FetchAllClients() (map[string]Client, error) {
	return s.repository.FetchAllClients()
}

func (s *Service) GetUniqueClientNameAndIP() []ClientNameAndIP {
	queryResult := s.repository.GetUniqueClientNameAndIP()
	var uniqueClients []ClientNameAndIP

	for _, client := range queryResult {
		uniqueClients = append(uniqueClients, ClientNameAndIP{
			Name: client.ClientName,
			IP:   client.ClientIP,
		})
	}

	return uniqueClients
}

func (s *Service) GetClientDetailsWithDomains(clientIP string) (ClientRequestDetails, string, map[string]int, error) {
	return s.repository.GetClientDetailsWithDomains(clientIP)
}

func (s *Service) GetTopBlockedDomains(blockedRequests int) ([]map[string]interface{}, error) {
	return s.repository.GetTopBlockedDomains(blockedRequests)
}

func (s *Service) GetTopClients() ([]map[string]interface{}, error) {
	return s.repository.GetTopClients()
}

func (s *Service) CountQueries(search string) (int, error) {
	return s.repository.CountQueries(search)
}

func (s *Service) SaveRequestLog(entries []model.RequestLogEntry) error {
	return s.repository.SaveRequestLog(entries)
}

type vacuumFunc func(ctx context.Context)

func (s *Service) DeleteRequestLogsTimebased(vacuum vacuumFunc, requestThreshold, maxRetries int, retryDelay time.Duration) {
	if err := s.repository.DeleteRequestLogsTimebased(vacuum, requestThreshold, maxRetries, retryDelay); err != nil {
		log.Warning("Error while deleting old request logs: %v", err)
	}
}
