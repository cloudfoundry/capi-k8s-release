package cf

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func (r *RestClient) Patch(url, authToken string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(
		http.MethodPatch,
		url,
		body,
	)
	if err != nil {
		return nil, err
	}

	log.Printf("[CF API/Patch] Sending request Patch %s", url)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))

	resp, err := r.Do(req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	return resp, nil
}

type RestClient struct {
	*http.Client
}
