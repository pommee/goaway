package settings

import "time"

type Status struct {
	PausedAt  time.Time `json:"pausedAt"`
	PauseTime time.Time `json:"pauseTime"`
	Paused    bool      `json:"paused"`
}

type TLSConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Cert    string `yaml:"cert" json:"cert"`
	Key     string `yaml:"key" json:"key"`
}

type UpstreamConfig struct {
	Preferred string   `yaml:"preferred" json:"preferred"`
	Fallback  []string `yaml:"fallback" json:"fallback"`
}

type PortsConfig struct {
	TCPUDP int `yaml:"udptcp" json:"udptcp"`
	DoT    int `yaml:"dot" json:"dot"`
	DoH    int `yaml:"doh" json:"doh"`
}

type DNSConfig struct {
	Status   Status         `yaml:"-" json:"status"`
	Address  string         `yaml:"address" json:"address"`
	Gateway  string         `yaml:"gateway" json:"gateway"`
	CacheTTL int            `yaml:"cacheTTL" json:"cacheTTL"`
	UDPSize  int            `yaml:"udpSize" json:"udpSize"`
	TLS      TLSConfig      `yaml:"tls" json:"tls"`
	Upstream UpstreamConfig `yaml:"upstream" json:"upstream"`
	Ports    PortsConfig    `yaml:"ports" json:"ports"`
}

type RateLimitConfig struct {
	Enabled  bool `yaml:"enabled" json:"enabled"`
	MaxTries int  `yaml:"maxTries" json:"maxTries"`
	Window   int  `yaml:"window" json:"window"`
}

type APIConfig struct {
	Port           int             `yaml:"port" json:"port"`
	Authentication bool            `yaml:"authentication" json:"authentication"`
	JWTSecret      string          `yaml:"jwtSecret" json:"-"`
	RateLimit      RateLimitConfig `yaml:"rateLimit" json:"rateLimit"`
}

type LoggingConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
	Level   int  `yaml:"level" json:"level"`
}

type MiscConfig struct {
	InAppUpdate               bool `yaml:"inAppUpdate" json:"inAppUpdate"`
	StatisticsRetention       int  `yaml:"statisticsRetention" json:"statisticsRetention"`
	Dashboard                 bool `yaml:"dashboard" json:"dashboard"`
	ScheduledBlacklistUpdates bool `yaml:"scheduledBlacklistUpdates" json:"scheduledBlacklistUpdates"`
}

type Config struct {
	BinaryPath string        `yaml:"-" json:"-"`
	DNS        DNSConfig     `yaml:"dns" json:"dns"`
	API        APIConfig     `yaml:"api" json:"api"`
	Logging    LoggingConfig `yaml:"logging" json:"logging"`
	Misc       MiscConfig    `yaml:"misc" json:"misc"`
}
