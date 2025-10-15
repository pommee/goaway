package key

import "time"

type ApiKey struct {
	CreatedAt time.Time `json:"createdAt"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`
}
