package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"cfddns/providers"

	"github.com/sirupsen/logrus"
)

type CloudflareProvider struct {
	Email        string
	GlobalAPIKey string
	APIToken     string
	ZoneName     string
	ZoneID       string
}

type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DnsRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	TTL     int    `json:"ttl"`
}

func (p *CloudflareProvider) addAuthHeaders(req *http.Request) {
	if p.APIToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIToken)
	} else if p.Email != "" && p.GlobalAPIKey != "" {
		req.Header.Set("X-Auth-Email", p.Email)
		req.Header.Set("X-Auth-Key", p.GlobalAPIKey)
	}
	req.Header.Set("Content-Type", "application/json")
}

func (p *CloudflareProvider) fetchZoneID() (string, error) {
	if p.ZoneID != "" {
		return p.ZoneID, nil
	}
	apiUrl := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones?name=%s", p.ZoneName)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return "", err
	}
	p.addAuthHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool       `json:"success"`
		Errors  []struct{} `json:"errors"`
		Result  []Zone     `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	if !result.Success || len(result.Result) == 0 {
		return "", fmt.Errorf("failed to fetch zone ID for domain: %s", p.ZoneName)
	}

	p.ZoneID = result.Result[0].ID
	return p.ZoneID, nil
}

func (p *CloudflareProvider) fetchDNSRecord(recordName, recordType string) (*DnsRecord, error) {
	zoneID, err := p.fetchZoneID()
	if err != nil {
		return nil, err
	}

	apiUrl := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?type=%s&name=%s", zoneID, recordType, recordName)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}
	p.addAuthHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Success bool        `json:"success"`
		Errors  []struct{}  `json:"errors"`
		Result  []DnsRecord `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to fetch DNS records")
	}

	if len(result.Result) > 0 {
		return &result.Result[0], nil
	}

	return nil, nil
}

func (p *CloudflareProvider) UpdateDNSRecord(record DnsRecord) error {
	zoneID, err := p.fetchZoneID()
	if err != nil {
		return err
	}

	apiUrl := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneID, record.ID)

	payloadBytes, err := json.Marshal(record)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", apiUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	p.addAuthHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	success, ok := responseBody["success"].(bool)
	if !ok || !success {
		return fmt.Errorf("failed to update DNS record: %v (status code: %d)", responseBody, resp.StatusCode)
	}

	return nil
}

func (p *CloudflareProvider) CreateDNSRecord(record DnsRecord) error {
	zoneID, err := p.fetchZoneID()
	if err != nil {
		return err
	}

	apiUrl := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", zoneID)

	payloadBytes, err := json.Marshal(record)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	p.addAuthHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	success, ok := responseBody["success"].(bool)
	if !ok || !success {
		return fmt.Errorf("failed to create DNS record: %v (status code: %d)", responseBody, resp.StatusCode)
	}

	return nil
}

func (p *CloudflareProvider) CommitRecord(record providers.DNSRecord) error {
	existingRecord, err := p.fetchDNSRecord(record.Name, record.Type)
	if err != nil {
		return err
	}

	dnsRecord := DnsRecord{
		Type:    record.Type,
		Name:    record.Name,
		Content: record.Content,
		TTL:     record.TTL,
		Proxied: record.Proxied,
	}

	if existingRecord != nil {
		dnsRecord.ID = existingRecord.ID
		if existingRecord.Content != dnsRecord.Content || existingRecord.TTL != dnsRecord.TTL || existingRecord.Proxied != dnsRecord.Proxied {
			err := p.UpdateDNSRecord(dnsRecord)
			if err != nil {
				return err
			}
			logrus.Infof("Updated DNS record: %s -> %s (TTL: %d, Proxied: %v)", dnsRecord.Name, dnsRecord.Content, dnsRecord.TTL, dnsRecord.Proxied)
		} else {
			logrus.Infof("Already up-to-date: %s -> %s (TTL: %d, Proxied: %v)", dnsRecord.Name, dnsRecord.Content, dnsRecord.TTL, dnsRecord.Proxied)
		}
	} else {
		err := p.CreateDNSRecord(dnsRecord)
		if err != nil {
			return err
		}
		logrus.Infof("Created new DNS record: %s -> %s (TTL: %d, Proxied: %v)", dnsRecord.Name, dnsRecord.Content, dnsRecord.TTL, dnsRecord.Proxied)
	}
	return nil
}
