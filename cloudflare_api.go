package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const apiUrl = "https://api.cloudflare.com/client/v4"
const dashUrl = "https://dash.cloudflare.com/api/v4"

type CloudflareAPI struct {
	baseUrl string
	APIKey  string
	Cookie  string
}

type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DNSRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

func NewCloudflareAPI(apiKey string, cookie string) *CloudflareAPI {
	if cookie != "" {
		return &CloudflareAPI{baseUrl: dashUrl, Cookie: cookie}
	}
	return &CloudflareAPI{baseUrl: apiUrl, APIKey: apiKey}
}

func (api *CloudflareAPI) ListZones() ([]Zone, error) {
	req, err := http.NewRequest("GET", api.baseUrl+"/zones", nil)
	if err != nil {
		return nil, err
	}

	if api.Cookie != "" {
		req.Header.Set("Cookie", api.Cookie)
	} else {
		req.Header.Set("Authorization", "Bearer "+api.APIKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch zones: %s", resp.Status)
	}

	var result struct {
		Result []Zone `json:"result"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Result, nil
}

func (api *CloudflareAPI) ListDNSRecords(zoneID string) ([]DNSRecord, error) {
	url := fmt.Sprintf("%s/zones/%s/dns_records", api.baseUrl, zoneID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if api.Cookie != "" {
		req.Header.Set("Cookie", api.Cookie)
	} else {
		req.Header.Set("Authorization", "Bearer "+api.APIKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch DNS records: %s", resp.Status)
	}

	var result struct {
		Result []DNSRecord `json:"result"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Result, nil
}
