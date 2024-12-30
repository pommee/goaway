package blacklist

import (
	"encoding/json"
	"fmt"
	"os"
)

var blacklistFilepath string

type Blacklist struct {
	Modified string          `json:"modified"`
	Version  string          `json:"version"`
	Entries  int             `json:"entries"`
	Domains  map[string]bool `json:"domains"`
}

// OrderedBlacklist struct ensures "Domains" is at the bottom
type OrderedBlacklist struct {
	Modified string        `json:"modified"`
	Version  string        `json:"version"`
	Entries  int           `json:"entries"`
	Domains  []interface{} `json:"domains"`
}

func LoadBlacklist(filepath string) (Blacklist, error) {
	blacklistFilepath = filepath

	file, err := os.Open(filepath)
	if err != nil {
		return Blacklist{}, fmt.Errorf("could not load %s: %w", blacklistFilepath, err)
	}
	defer file.Close()

	var temp struct {
		Modified string   `json:"modified"`
		Version  string   `json:"version"`
		Entries  int      `json:"entries"`
		Domains  []string `json:"domains"`
	}

	if err := json.NewDecoder(file).Decode(&temp); err != nil {
		return Blacklist{}, fmt.Errorf("%s is invalid: %w", blacklistFilepath, err)
	}

	domainMap := make(map[string]bool)
	for _, domain := range temp.Domains {
		domainMap[domain] = true
	}

	return Blacklist{
		Modified: temp.Modified,
		Version:  temp.Version,
		Entries:  temp.Entries,
		Domains:  domainMap,
	}, nil
}

func (b *Blacklist) AddDomain(domain string) error {
	if _, exists := b.Domains[domain]; exists {
		return fmt.Errorf("domain %s is already in the blacklist", domain)
	}

	if err := b.updateBlacklistJSON(func(data *OrderedBlacklist) {
		data.Domains = append(data.Domains, domain)
	}); err != nil {
		return err
	}

	b.Domains[domain] = true
	return nil
}

func (b *Blacklist) RemoveDomain(domain string) error {
	if _, exists := b.Domains[domain]; !exists {
		return fmt.Errorf("domain %s not found in the blacklist", domain)
	}

	delete(b.Domains, domain)

	return b.updateBlacklistJSON(func(data *OrderedBlacklist) {
		newDomains := []interface{}{}
		for domain := range b.Domains {
			newDomains = append(newDomains, domain)
		}
		data.Domains = newDomains
	})
}

func (b *Blacklist) updateBlacklistJSON(updateFunc func(*OrderedBlacklist)) error {
	file, err := os.ReadFile(blacklistFilepath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", blacklistFilepath, err)
	}

	var data OrderedBlacklist
	if err := json.Unmarshal(file, &data); err != nil {
		return fmt.Errorf("failed to parse %s: %w", blacklistFilepath, err)
	}

	updateFunc(&data)

	updatedFile, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to encode updated data: %w", err)
	}
	if err := os.WriteFile(blacklistFilepath, updatedFile, 0644); err != nil {
		return fmt.Errorf("failed to write to %s: %w", blacklistFilepath, err)
	}

	return nil
}
