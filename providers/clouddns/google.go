package clouddns

import (
	"context"
	"fmt"
	"strings"

	"cfddns/providers"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/option"
)

type CloudDNSProvider struct {
	ProjectID       string
	CredentialsJSON []byte
	ZoneName        string
}

func (p *CloudDNSProvider) getService() (*dns.Service, error) {
	ctx := context.Background()
	return dns.NewService(ctx, option.WithCredentialsJSON(p.CredentialsJSON))
}

func (p *CloudDNSProvider) CommitRecord(record providers.DNSRecord) error {
	service, err := p.getService()
	if err != nil {
		return fmt.Errorf("failed to create Cloud DNS service: %v", err)
	}

	ctx := context.Background()
	fqdn := record.Name
	if !strings.HasSuffix(fqdn, ".") {
		fqdn += "."
	}

	// Fetch existing records
	recListCall := service.ResourceRecordSets.List(p.ProjectID, p.ZoneName)
	recListCall.Name(fqdn)
	recListCall.Type(record.Type)
	recList, err := recListCall.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to list DNS records: %v", err)
	}

	// Prepare the change
	rrdata := record.Content
	ttl := int64(record.TTL)
	rrset := &dns.ResourceRecordSet{
		Name:    fqdn,
		Type:    record.Type,
		Ttl:     ttl,
		Rrdatas: []string{rrdata},
	}

	change := &dns.Change{}
	if len(recList.Rrsets) > 0 {
		// Update existing record
		change.Deletions = []*dns.ResourceRecordSet{recList.Rrsets[0]}
	}
	change.Additions = []*dns.ResourceRecordSet{rrset}

	// Apply the change
	changesCall := service.Changes.Create(p.ProjectID, p.ZoneName, change)
	_, err = changesCall.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to apply DNS changes: %v", err)
	}

	logrus.Infof("Record %s -> %s (%s) updated/created successfully", fqdn, rrdata, record.Type)
	return nil
}
