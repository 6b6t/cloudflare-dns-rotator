package dns

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
)

// Client wraps the Cloudflare API for DNS operations
type Client struct {
	api    *cloudflare.API
	zoneID string
	domain string
}

// NewClient creates a new Cloudflare DNS client
func NewClient(apiToken, domain string) (*Client, error) {
	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudflare client: %w", err)
	}

	// Get zone ID for the domain
	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone ID for domain %s: %w", domain, err)
	}

	return &Client{
		api:    api,
		zoneID: zoneID,
		domain: domain,
	}, nil
}

// Record represents a DNS record
type Record struct {
	ID      string
	Name    string
	Type    string
	Content string
	Proxied bool
	TTL     int
}

// GetCNAMERecord retrieves a CNAME record by name
func (c *Client) GetCNAMERecord(ctx context.Context, recordName string) (*Record, error) {
	fullName := recordName + "." + c.domain

	records, _, err := c.api.ListDNSRecords(
		ctx,
		cloudflare.ZoneIdentifier(c.zoneID),
		cloudflare.ListDNSRecordsParams{
			Type: "CNAME",
			Name: fullName,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CNAME record %s not found", fullName)
	}

	r := records[0]
	proxied := false
	if r.Proxied != nil {
		proxied = *r.Proxied
	}

	return &Record{
		ID:      r.ID,
		Name:    r.Name,
		Type:    r.Type,
		Content: r.Content,
		Proxied: proxied,
		TTL:     r.TTL,
	}, nil
}

// UpdateCNAMERecord updates an existing CNAME record with a new target
func (c *Client) UpdateCNAMERecord(ctx context.Context, recordID, recordName, target string, proxied bool, ttl int) error {
	fullName := recordName + "." + c.domain

	_, err := c.api.UpdateDNSRecord(
		ctx,
		cloudflare.ZoneIdentifier(c.zoneID),
		cloudflare.UpdateDNSRecordParams{
			ID:      recordID,
			Type:    "CNAME",
			Name:    fullName,
			Content: target,
			Proxied: &proxied,
			TTL:     ttl,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update DNS record: %w", err)
	}

	return nil
}

// CreateCNAMERecord creates a new CNAME record
func (c *Client) CreateCNAMERecord(ctx context.Context, recordName, target string, proxied bool, ttl int) (*Record, error) {
	fullName := recordName + "." + c.domain

	record, err := c.api.CreateDNSRecord(
		ctx,
		cloudflare.ZoneIdentifier(c.zoneID),
		cloudflare.CreateDNSRecordParams{
			Type:    "CNAME",
			Name:    fullName,
			Content: target,
			Proxied: &proxied,
			TTL:     ttl,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create DNS record: %w", err)
	}

	proxiedValue := false
	if record.Proxied != nil {
		proxiedValue = *record.Proxied
	}

	return &Record{
		ID:      record.ID,
		Name:    record.Name,
		Type:    record.Type,
		Content: record.Content,
		Proxied: proxiedValue,
		TTL:     record.TTL,
	}, nil
}

// ZoneID returns the zone ID
func (c *Client) ZoneID() string {
	return c.zoneID
}

// Domain returns the domain name
func (c *Client) Domain() string {
	return c.domain
}
