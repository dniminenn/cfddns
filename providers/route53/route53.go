package route53

import (
	"fmt"
	"strings"

	"cfddns/providers"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/sirupsen/logrus"
)

type Route53Provider struct {
	ZoneName        string
	ZoneID          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

func (p *Route53Provider) getSession() (*session.Session, error) {
	creds := credentials.NewStaticCredentials(p.AccessKeyID, p.SecretAccessKey, "")
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(p.Region),
		Credentials: creds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}
	return sess, nil
}

func (p *Route53Provider) getZoneID() (string, error) {
	if p.ZoneID != "" {
		return p.ZoneID, nil
	}

	sess, err := p.getSession()
	if err != nil {
		return "", err
	}

	svc := route53.New(sess)

	// List hosted zones
	result, err := svc.ListHostedZones(nil)
	if err != nil {
		return "", fmt.Errorf("failed to list hosted zones: %v", err)
	}

	// Find the correct zone by name
	for _, zone := range result.HostedZones {
		if strings.TrimSuffix(*zone.Name, ".") == p.ZoneName {
			p.ZoneID = strings.TrimPrefix(*zone.Id, "/hostedzone/")
			return p.ZoneID, nil
		}
	}

	return "", fmt.Errorf("hosted zone for %s not found", p.ZoneName)
}

func (p *Route53Provider) CommitRecord(record providers.DNSRecord) error {
	sess, err := p.getSession()
	if err != nil {
		return err
	}

	svc := route53.New(sess)

	zoneID, err := p.getZoneID()
	if err != nil {
		return err
	}

	// Prepare the record change
	change := &route53.Change{
		Action: aws.String("UPSERT"),
		ResourceRecordSet: &route53.ResourceRecordSet{
			Name: aws.String(record.Name),
			Type: aws.String(record.Type),
			TTL:  aws.Int64(int64(record.TTL)),
			ResourceRecords: []*route53.ResourceRecord{
				{
					Value: aws.String(record.Content),
				},
			},
		},
	}

	// Make the request to update the record
	_, err = svc.ChangeResourceRecordSets(&route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{change},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update or create record: %v", err)
	}

	logrus.Infof("Record %s -> %s (%s) updated/created successfully", record.Name, record.Content, record.Type)
	return nil
}
