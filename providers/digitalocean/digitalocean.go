package digitalocean

import (
	"context"
	"fmt"

	"cfddns/providers"

	"github.com/digitalocean/godo"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type DigitalOceanProvider struct {
	APIToken string
	Domain   string
}

func (p *DigitalOceanProvider) getClient() *godo.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: p.APIToken})
	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	return godo.NewClient(oauthClient)
}

func (p *DigitalOceanProvider) CommitRecord(record providers.DNSRecord) error {
	client := p.getClient()
	ctx := context.Background()

	// List existing records to check if the record exists
	records, _, err := client.Domains.Records(ctx, p.Domain, nil)
	if err != nil {
		return fmt.Errorf("failed to list DNS records: %v", err)
	}

	var existingRecord *godo.DomainRecord
	for _, r := range records {
		if r.Type == record.Type && r.Name == record.Name {
			existingRecord = &r
			break
		}
	}

	if existingRecord != nil {
		// Update the existing record if necessary
		if existingRecord.Data != record.Content || existingRecord.TTL != record.TTL {
			editRequest := &godo.DomainRecordEditRequest{
				Type: record.Type,
				Name: record.Name,
				Data: record.Content,
				TTL:  record.TTL,
			}
			_, _, err := client.Domains.EditRecord(ctx, p.Domain, existingRecord.ID, editRequest)
			if err != nil {
				return fmt.Errorf("failed to update DNS record: %v", err)
			}
			logrus.Infof("Updated DNS record: %s -> %s (TTL: %d)", record.Name, record.Content, record.TTL)
		} else {
			logrus.Debugf("DNS record already up-to-date: %s -> %s (TTL: %d)", record.Name, record.Content, record.TTL)
		}
	} else {
		// Create a new record
		createRequest := &godo.DomainRecordEditRequest{
			Type: record.Type,
			Name: record.Name,
			Data: record.Content,
			TTL:  record.TTL,
		}
		_, _, err := client.Domains.CreateRecord(ctx, p.Domain, createRequest)
		if err != nil {
			return fmt.Errorf("failed to create DNS record: %v", err)
		}
		logrus.Infof("Created new DNS record: %s -> %s (TTL: %d)", record.Name, record.Content, record.TTL)
	}

	return nil
}
