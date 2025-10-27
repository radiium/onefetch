package client

type OneFichierInfoResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	// Date        time.Time `json:"date"`
	Checksum    string  `json:"checksum"`
	ContentType string  `json:"content-type"`
	Description *string `json:"description,omitempty"`
	Pass        int     `json:"pass"`
	Path        string  `json:"path"`
	FolderID    string  `json:"folder_id"`
}

type OneFichierTokenResponse struct {
	URL     string  `json:"url"`
	Status  string  `json:"status"`
	Message *string `json:"message,omitempty"`
}
