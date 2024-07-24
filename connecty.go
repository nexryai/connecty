package connecty

type Request struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type connectionStatus struct {
	dir   string
	err   error
	bytes int64
}
