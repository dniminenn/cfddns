package noip

import (
	"cfddns/providers"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type NoIPProvider struct {
	Username string
	Password string
}

func (p *NoIPProvider) CommitRecord(record providers.DNSRecord) error {
	var endpoint string

	if record.Type == "A" {
		// Update IPv4 address
		endpoint = fmt.Sprintf("https://dynupdate.no-ip.com/nic/update?hostname=%s&myip=%s", record.Name, record.Content)
	} else if record.Type == "AAAA" {
		// Update IPv6 address
		endpoint = fmt.Sprintf("https://dynupdate.no-ip.com/nic/update?hostname=%s&myipv6=%s", record.Name, record.Content)
	} else {
		return fmt.Errorf("unsupported record type: %s", record.Type)
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set Basic Auth header
	auth := p.Username + ":" + p.Password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+encodedAuth)
	req.Header.Set("User-Agent", "cfddns/1.0 (root@dnim.dev)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read HTTP response: %v", err)
	}
	body := string(bodyBytes)

	if strings.HasPrefix(body, "good") || strings.HasPrefix(body, "nochg") {
		logrus.Infof("Updated No-IP record %s to %s (%s)", record.Name, record.Content, record.Type)
		return nil
	} else {
		return fmt.Errorf("failed to update No-IP record, response: %s", body)
	}
}
