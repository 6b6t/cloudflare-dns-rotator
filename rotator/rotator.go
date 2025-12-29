package rotator

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"cloudflare-dns-rotator/config"
	"cloudflare-dns-rotator/dns"
)

// Rotator handles the DNS record rotation logic
type Rotator struct {
	client   *dns.Client
	config   *config.Config
	recordID string
	current  string
}

// New creates a new Rotator instance
func New(cfg *config.Config) (*Rotator, error) {
	client, err := dns.NewClient(cfg.APIToken, cfg.Domain)
	if err != nil {
		return nil, fmt.Errorf("failed to create DNS client: %w", err)
	}

	return &Rotator{
		client: client,
		config: cfg,
	}, nil
}

// Initialize fetches the current record state or creates it if it doesn't exist
func (r *Rotator) Initialize(ctx context.Context) error {
	record, err := r.client.GetCNAMERecord(ctx, r.config.RecordName)
	if err != nil {
		// Record doesn't exist, create it with the first target
		log.Printf("Record %s.%s not found, creating with first target", r.config.RecordName, r.config.Domain)

		newRecord, err := r.client.CreateCNAMERecord(
			ctx,
			r.config.RecordName,
			r.config.Targets[0],
			r.config.Proxied,
			r.config.TTL,
		)
		if err != nil {
			return fmt.Errorf("failed to create initial record: %w", err)
		}

		r.recordID = newRecord.ID
		r.current = newRecord.Content
		log.Printf("Created record %s -> %s", newRecord.Name, newRecord.Content)
		return nil
	}

	r.recordID = record.ID
	r.current = record.Content
	log.Printf("Found existing record %s -> %s", record.Name, record.Content)
	return nil
}

// Rotate changes the CNAME target to a random different target
func (r *Rotator) Rotate(ctx context.Context) error {
	// Select a random target that's different from the current one
	newTarget := r.selectRandomTarget()

	log.Printf("Rotating %s.%s: %s -> %s", r.config.RecordName, r.config.Domain, r.current, newTarget)

	err := r.client.UpdateCNAMERecord(
		ctx,
		r.recordID,
		r.config.RecordName,
		newTarget,
		r.config.Proxied,
		r.config.TTL,
	)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	r.current = newTarget
	log.Printf("Successfully rotated to %s", newTarget)
	return nil
}

// selectRandomTarget picks a random target from the list, preferring one different from current
func (r *Rotator) selectRandomTarget() string {
	targets := r.config.Targets

	// If we have more than one target, exclude the current one
	if len(targets) > 1 && r.current != "" {
		available := make([]string, 0, len(targets)-1)
		for _, t := range targets {
			if t != r.current {
				available = append(available, t)
			}
		}
		if len(available) > 0 {
			targets = available
		}
	}

	return targets[rand.Intn(len(targets))]
}

// Run starts the rotation loop with the configured interval
func (r *Rotator) Run(ctx context.Context) error {
	interval, err := r.config.Duration()
	if err != nil {
		return fmt.Errorf("invalid interval: %w", err)
	}

	log.Printf("Starting rotation loop with interval %s", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Rotation loop stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := r.Rotate(ctx); err != nil {
				log.Printf("Rotation failed: %v", err)
				// Continue running even if a single rotation fails
			}
		}
	}
}

// Current returns the current target
func (r *Rotator) Current() string {
	return r.current
}

// RecordID returns the record ID
func (r *Rotator) RecordID() string {
	return r.recordID
}
