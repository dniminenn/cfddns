package providers

type DNSRecord struct {
	Name    string
	Type    string
	Content string
	TTL     int
	Proxied bool
}

type Provider interface {
	UpdateOrCreateRecord(record DNSRecord) error
}
