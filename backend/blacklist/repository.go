package blacklist

import (
	"context"
	"errors"
	"fmt"
	"goaway/backend/database"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SourceRepository interface {
	GetSources(ctx context.Context, excludeCustom bool) ([]database.Source, error)
	GetSourceByName(ctx context.Context, name string) (*database.Source, error)
	GetSourceByNameAndURL(ctx context.Context, name, url string) (*database.Source, error)
	GetSourceExists(ctx context.Context, name, url string) bool
	CreateOrUpdateSource(ctx context.Context, source *database.Source) error
	UpdateSourceName(ctx context.Context, oldName, newName, url string) error
	UpdateSourceLastUpdated(ctx context.Context, url string, timestamp time.Time) error
	ToggleSourceActive(ctx context.Context, name string) error
	DeleteSource(ctx context.Context, name, url string) error
	UpsertSource(ctx context.Context, source *database.Source) error
}

type DomainRepository interface {
	GetAllDomains(ctx context.Context) ([]string, error)
	GetDomainsForSource(ctx context.Context, sourceName string) ([]string, error)
	GetPaginatedDomains(ctx context.Context, page, pageSize int, search string) ([]database.Blacklist, int64, error)
	CountDomains(ctx context.Context) (int64, error)
	CreateDomain(ctx context.Context, domain *database.Blacklist) error
	CreateDomainsInBatches(ctx context.Context, domains []database.Blacklist, batchSize int) error
	DeleteDomain(ctx context.Context, domain string) error
	DeleteDomainsBySourceID(ctx context.Context, sourceID uint) error
	DeleteCustomDomain(ctx context.Context, domain string, sourceID uint) error
}

type StatsRepository interface {
	GetAllSourceStats(ctx context.Context) ([]SourceWithCount, error)
	GetSourceStats(ctx context.Context, listname string) (*SourceWithCount, error)
	GetRequestStats(ctx context.Context) ([]RequestStats, error)
}

type TransactionRepository interface {
	WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error
	Vacuum(ctx context.Context) error
}

type Repository interface {
	SourceRepository
	DomainRepository
	StatsRepository
	TransactionRepository
}

type SourceWithCount struct {
	Name         string    `json:"name"`
	URL          string    `json:"url"`
	ID           uint      `json:"id"`
	LastUpdated  time.Time `json:"lastUpdated"`
	BlockedCount int       `json:"blockedCount"`
	Active       bool      `json:"active"`
}

type RequestStats struct {
	Blocked bool `json:"blocked"`
	Cached  bool `json:"cached"`
	Count   int  `json:"count"`
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetSources(ctx context.Context, excludeCustom bool) ([]database.Source, error) {
	var sources []database.Source
	query := r.db.WithContext(ctx)

	if excludeCustom {
		query = query.Where("name != ?", "Custom")
	}

	if err := query.Find(&sources).Error; err != nil {
		return nil, fmt.Errorf("failed to query sources: %w", err)
	}

	return sources, nil
}

func (r *repository) GetSourceByName(ctx context.Context, name string) (*database.Source, error) {
	var source database.Source
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&source).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("source '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to get source: %w", err)
	}
	return &source, nil
}

func (r *repository) GetSourceByNameAndURL(ctx context.Context, name, url string) (*database.Source, error) {
	var source database.Source
	if err := r.db.WithContext(ctx).Where("name = ? AND url = ?", name, url).First(&source).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("source '%s' with URL '%s' not found", name, url)
		}
		return nil, fmt.Errorf("failed to get source: %w", err)
	}
	return &source, nil
}

func (r *repository) GetSourceExists(ctx context.Context, name, url string) bool {
	var count int64
	result := r.db.WithContext(ctx).Model(&database.Source{}).
		Where("name = ? AND url = ?", name, url).
		Count(&count)

	if result.Error != nil || count == 0 {
		return false
	}

	return count > 0
}

func (r *repository) CreateOrUpdateSource(ctx context.Context, source *database.Source) error {
	result := r.db.WithContext(ctx).Where(database.Source{Name: source.Name, URL: source.URL}).FirstOrCreate(source)
	if result.Error != nil {
		return fmt.Errorf("failed to create or update source: %w", result.Error)
	}
	return nil
}

func (r *repository) UpdateSourceName(ctx context.Context, oldName, newName, url string) error {
	if strings.TrimSpace(newName) == "" {
		return fmt.Errorf("new name cannot be empty")
	}

	result := r.db.WithContext(ctx).Model(&database.Source{}).
		Where("name = ? AND url = ?", oldName, url).
		Update("name", newName)

	if result.Error != nil {
		return fmt.Errorf("failed to update source name: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("list with name '%s' not found", oldName)
	}

	return nil
}

func (r *repository) UpdateSourceLastUpdated(ctx context.Context, url string, timestamp time.Time) error {
	result := r.db.WithContext(ctx).Model(&database.Source{}).
		Where("url = ?", url).
		Update("last_updated", timestamp)

	if result.Error != nil {
		return fmt.Errorf("failed to update source: %w", result.Error)
	}

	return nil
}

func (r *repository) ToggleSourceActive(ctx context.Context, name string) error {
	var source database.Source
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&source).Error; err != nil {
		return fmt.Errorf("failed to find source %s: %w", name, err)
	}

	result := r.db.WithContext(ctx).Model(&source).Update("active", !source.Active)
	if result.Error != nil {
		return fmt.Errorf("failed to toggle status for %s: %w", name, result.Error)
	}

	return nil
}

func (r *repository) DeleteSource(ctx context.Context, name, url string) error {
	var source database.Source
	if err := r.db.WithContext(ctx).Where("name = ? AND url = ?", name, url).First(&source).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("source '%s' not found", name)
		}
		return fmt.Errorf("failed to get source: %w", err)
	}

	if err := r.db.WithContext(ctx).Delete(&source).Error; err != nil {
		return fmt.Errorf("failed to remove source '%s': %w", name, err)
	}

	return nil
}

func (r *repository) GetAllDomains(ctx context.Context) ([]string, error) {
	var domains []string
	result := r.db.WithContext(ctx).Model(&database.Blacklist{}).
		Distinct("domain").
		Pluck("domain", &domains)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query blacklist: %w", result.Error)
	}

	return domains, nil
}

func (r *repository) GetDomainsForSource(ctx context.Context, sourceName string) ([]string, error) {
	var blacklistEntries []database.Blacklist
	result := r.db.WithContext(ctx).Select("blacklists.domain").
		Joins("JOIN sources ON blacklists.source_id = sources.id").
		Where("sources.name = ?", sourceName).
		Find(&blacklistEntries)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query domains for list: %w", result.Error)
	}

	domains := make([]string, len(blacklistEntries))
	for i, entry := range blacklistEntries {
		domains[i] = entry.Domain
	}

	return domains, nil
}

func (r *repository) GetPaginatedDomains(ctx context.Context, page, pageSize int, search string) ([]database.Blacklist, int64, error) {
	searchPattern := "%" + search + "%"
	offset := (page - 1) * pageSize

	var blacklistEntries []database.Blacklist
	result := r.db.WithContext(ctx).Select("domain").
		Where("domain LIKE ?", searchPattern).
		Order("domain DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&blacklistEntries)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to query blacklist: %w", result.Error)
	}

	var total int64
	countResult := r.db.WithContext(ctx).Model(&database.Blacklist{}).
		Where("domain LIKE ?", searchPattern).
		Count(&total)

	if countResult.Error != nil {
		return nil, 0, fmt.Errorf("failed to count domains: %w", countResult.Error)
	}

	return blacklistEntries, total, nil
}

func (r *repository) CountDomains(ctx context.Context) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&database.Blacklist{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count domains: %w", result.Error)
	}
	return count, nil
}

func (r *repository) CreateDomain(ctx context.Context, domain *database.Blacklist) error {
	result := r.db.WithContext(ctx).Create(domain)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "UNIQUE constraint failed") ||
			strings.Contains(result.Error.Error(), "duplicate key") {
			return fmt.Errorf("%s is already blacklisted", domain.Domain)
		}
		return fmt.Errorf("failed to add domain to blacklist: %w", result.Error)
	}
	return nil
}

func (r *repository) CreateDomainsInBatches(ctx context.Context, domains []database.Blacklist, batchSize int) error {
	if len(domains) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(domains, batchSize).Error; err != nil {
		if !strings.Contains(err.Error(), "UNIQUE constraint failed") &&
			!strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("failed to add domains: %w", err)
		}
	}
	return nil
}

func (r *repository) DeleteDomain(ctx context.Context, domain string) error {
	result := r.db.WithContext(ctx).Where("domain = ?", domain).Delete(&database.Blacklist{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove domain from blacklist: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("domain not found: %s", domain)
	}
	return nil
}

func (r *repository) DeleteDomainsBySourceID(ctx context.Context, sourceID uint) error {
	if err := r.db.WithContext(ctx).Where("source_id = ?", sourceID).Delete(&database.Blacklist{}).Error; err != nil {
		return fmt.Errorf("failed to remove domains for source ID %d: %w", sourceID, err)
	}
	return nil
}

func (r *repository) DeleteCustomDomain(ctx context.Context, domain string, sourceID uint) error {
	result := r.db.WithContext(ctx).
		Where("domain = ? AND source_id = ?", domain, sourceID).
		Delete(&database.Blacklist{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete domain '%s': %w", domain, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("domain '%s' not found in custom blacklist", domain)
	}

	return nil
}

func (r *repository) GetAllSourceStats(ctx context.Context) ([]SourceWithCount, error) {
	var results []SourceWithCount
	result := r.db.WithContext(ctx).Table("sources s").
		Select("s.id, s.name, s.url, s.last_updated, s.active, COALESCE(bc.blocked_count, 0) as blocked_count").
		Joins("LEFT JOIN (SELECT source_id, COUNT(*) as blocked_count FROM blacklists GROUP BY source_id) bc ON s.id = bc.source_id").
		Order("s.name, s.id").
		Scan(&results)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to query source statistics: %w", result.Error)
	}

	return results, nil
}

func (r *repository) GetSourceStats(ctx context.Context, listname string) (*SourceWithCount, error) {
	var result SourceWithCount
	err := r.db.WithContext(ctx).Table("sources s").
		Select("s.name, s.url, s.last_updated, s.active, COALESCE(bc.blocked_count, 0) as blocked_count").
		Joins("LEFT JOIN (SELECT source_id, COUNT(*) as blocked_count FROM blacklists GROUP BY source_id) bc ON s.id = bc.source_id").
		Where("s.name = ?", listname).
		First(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("list not found")
		}
		return nil, fmt.Errorf("failed to query list statistics: %w", err)
	}

	return &result, nil
}

func (r *repository) GetRequestStats(ctx context.Context) ([]RequestStats, error) {
	var stats []RequestStats
	result := r.db.WithContext(ctx).Model(&database.RequestLog{}).
		Select("blocked, cached, COUNT(*) as count").
		Group("blocked, cached").
		Scan(&stats)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query request_logs: %w", result.Error)
	}
	return stats, nil
}

func (r *repository) Vacuum(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Exec("VACUUM").Error; err != nil {
		return fmt.Errorf("error while vacuuming database: %w", err)
	}
	return nil
}

func (r *repository) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

func (r *repository) UpsertSource(ctx context.Context, source *database.Source) error {
	if err := r.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "url"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "last_updated", "active"}),
		},
	).Create(source).Error; err != nil {
		return fmt.Errorf("failed to upsert source: %w", err)
	}
	return nil
}
