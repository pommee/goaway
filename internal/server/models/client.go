package model

type Client struct {
	IP   string `json:"ip"`
	Name string `json:"name"`
	MAC  string `json:"mac"`
}
