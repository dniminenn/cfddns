package duckdns

import (
	"cfddns/providers"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type DuckDNSProvider struct {
    Token string
}

func (p *DuckDNSProvider) UpdateOrCreateRecord(record providers.DNSRecord) error {
    subdomain := record.Name
    // Remove '.duckdns.org' if present
    if strings.HasSuffix(subdomain, ".duckdns.org") {
        subdomain = strings.TrimSuffix(subdomain, ".duckdns.org")
    }

    var endpoint string

    if record.Type == "A" {
        // Update IPv4 address
        endpoint = fmt.Sprintf(
            "https://www.duckdns.org/update?domains=%s&token=%s&ip=%s",
            subdomain,
            p.Token,
            record.Content,
        )
    } else if record.Type == "AAAA" {
        // Update IPv6 address
        endpoint = fmt.Sprintf(
            "https://www.duckdns.org/update?domains=%s&token=%s&ipv6=%s",
            subdomain,
            p.Token,
            record.Content,
        )
    } else {
        return fmt.Errorf("unsupported record type: %s", record.Type)
    }

    resp, err := http.Get(endpoint)
    if err != nil {
        return fmt.Errorf("failed to update DNS record: %v", err)
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    if string(body) != "OK" {
        return fmt.Errorf("failed to update DNS record, response: %s", string(body))
    }

    logrus.Infof("Updated DuckDNS record %s to %s (%s)", record.Name, record.Content, record.Type)
    return nil
}
