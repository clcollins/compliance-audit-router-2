package alerts

import "net/http"

type ProcessResponse struct {
	StatusCode int
	Body       string
}

func ProcessAlert(r *http.Request) (ProcessResponse, error) {
	var err error
	return ProcessResponse{StatusCode: http.StatusOK, Body: "Processed"}, err

}
