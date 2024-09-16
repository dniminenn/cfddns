package dynu

import (
	"cfddns/providers"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type DynuProvider struct {
	Username string
	Password string // password to be hashed using MD5
}

func (p *DynuProvider) md5Hash(input string) string {
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

func (p *DynuProvider) CommitRecord(record providers.DNSRecord) error {
	// Hash the password using MD5
	hashedPassword := p.md5Hash(p.Password)

	var endpoint string
	if record.Type == "A" {
		endpoint = fmt.Sprintf("http://api.dynu.com/nic/update?myip=%s&username=%s&password=%s", record.Content, p.Username, hashedPassword)
	} else if record.Type == "AAAA" {
		endpoint = fmt.Sprintf("http://api.dynu.com/nic/update?myipv6=%s&username=%s&password=%s", record.Content, p.Username, hashedPassword)
	} else {
		return fmt.Errorf("unsupported record type: %s", record.Type)
	}

	resp, err := http.Get(endpoint)
	if err != nil {
		return fmt.Errorf("failed to update DNS record: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read HTTP response: %v", err)
	}
	body := string(bodyBytes)

	if strings.Contains(body, "good") || strings.Contains(body, "nochg") {
		logrus.Infof("Updated Dynu record %s -> %s (%s)", record.Name, record.Content, record.Type)
		return nil
	} else {
		return fmt.Errorf("failed to update Dynu record, response: %s", body)
	}
}
