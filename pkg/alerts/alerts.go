package alerts

type ProcessResponse struct {
	StatusCode int
	Body       string
}

type Alert struct {
	//  TODO: convert time strings to time.Date
	FirstEvent string
	LastEvent  string
	ClusterID  string
	UserName   string
	Summary    string
	SessionID  string
	Raw        string
}
