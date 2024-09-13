package duckdns

import (
	"cfddns/providers"
	"fmt"
	"io"
	"net/http"
)

type DuckDNSProvider struct {
	Token  string
	Domain string
}

func (p *DuckDNSProvider) UpdateOrCreateRecord(record providers.DNSRecord) error {
	endpoint := fmt.Sprintf("https://www.duckdns.org/update?domains=%s&token=%s&ip=%s", p.Domain, p.Token, record.Content)

	resp, err := http.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to update DNS record: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "OK" {
		return fmt.Errorf("failed to update DNS record, response: %s", string(body))
	}

	return nil
}
