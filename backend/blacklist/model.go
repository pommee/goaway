package blacklist

type ListUpdateAvailable struct {
	RemoteChecksum  string   `json:"remoteChecksum"`
	DBChecksum      string   `json:"dbChecksum"`
	RemoteDomains   []string `json:"remoteDomains"`
	DBDomains       []string `json:"dbDomains"`
	DiffAdded       []string `json:"diffAdded"`
	DiffRemoved     []string `json:"diffRemoved"`
	UpdateAvailable bool     `json:"updateAvailable"`
}

type BlocklistSource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
