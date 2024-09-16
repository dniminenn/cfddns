package freedns

import (
	"cfddns/providers"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type FreeDNSProvider struct {
}

func (p *FreeDNSProvider) CommitRecord(record providers.DNSRecord) error {
	if record.UpdateToken == "" {
		return fmt.Errorf("UpdateToken is required for FreeDNS record")
	}

	endpoint := fmt.Sprintf("https://freedns.afraid.org/dynamic/update.php?%s", record.UpdateToken)

	if record.Content != "" {
		if record.Type == "A" || record.Type == "AAAA" {
			endpoint += "&address=" + record.Content
		} else {
			return fmt.Errorf("unsupported record type: %s", record.Type)
		}
	}

	resp, err := http.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to update FreeDNS record: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response from FreeDNS: %v", err)
	}
	body := string(bodyBytes)

	if strings.Contains(body, "has not changed.") || strings.Contains(body, "Updated") {
		logrus.Infof("Updated FreeDNS record %s -> %s (%s)", record.Name, record.Content, record.Type)
		return nil
	} else if strings.Contains(body, "ERROR") {
		return fmt.Errorf("failed to update FreeDNS record, response: %s", body)
	} else {
		logrus.Warnf("Unexpected response from FreeDNS: %s", body)
		return nil
	}
}
