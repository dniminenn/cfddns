package providers

type DNSRecord struct {
	Name        string
	Type        string
	Content     string
	TTL         int
	Proxied     bool
	UpdateToken string
}

type Provider interface {
	CommitRecord(record DNSRecord) error
}
