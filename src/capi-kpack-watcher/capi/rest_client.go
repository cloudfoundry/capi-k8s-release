package capi

import (
	"fmt"
	"io"
	"io/ioutil"
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

	log.Printf("[CAPI/Patch] Sending request Patch %s", url)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))

	resp, err := r.client.Do(req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	return resp, nil
}

func (r *RestClient) Post(url, authToken string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(
		http.MethodPost,
		url,
		body,
	)
	if err != nil {
		return nil, err
	}

	log.Printf("[CAPI/Post] Sending request Post %s", url)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))

	resp, err := r.client.Do(req)
	if err != nil {
		return resp, err
	}
	//defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	fmt.Printf(string(bodyBytes))
	return resp, nil
}

func (r *RestClient) Get(url, authToken string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		url,
		body,
	)
	if err != nil {
		return nil, err
	}

	log.Printf("[CAPI/Get] Sending request Get %s", url)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))

	resp, err := r.client.Do(req)
	if err != nil {
		return resp, err
	}
	//defer resp.Body.Close()
	//buf := new(bytes.Buffer)
	//bodyBytes, err := ioutil.ReadAll(resp.Body)
	//fmt.Printf(string(bodyBytes))

	return resp, nil
}

type RestClient struct {
	client *http.Client
}
