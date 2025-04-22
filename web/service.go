package web

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
)

type Service struct {
	repo       JobRepository
	dataFolder string
}

func NewService(repo JobRepository, dataFolder string) *Service {
	return &Service{
		repo:       repo,
		dataFolder: dataFolder,
	}
}

// isBusyError checks if an error is a SQLite busy or locked error
func isBusyError(err error) bool {
	if err == nil {
		return false
	}

	// Check for mattn/go-sqlite3 error
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.Code == sqlite3.ErrBusy || sqliteErr.Code == sqlite3.ErrLocked
	}

	// Check for modernc.org/sqlite error by string matching
	errStr := err.Error()
	return strings.Contains(errStr, "database is locked") ||
		strings.Contains(errStr, "SQLITE_BUSY") ||
		strings.Contains(errStr, "SQLITE_LOCKED")
}

// executeWithRetry executes a database operation with retry logic
func (s *Service) executeWithRetry(ctx context.Context, operation string, fn func() error) error {
	var err error
	var backoff time.Duration

	// Start with 100ms backoff, max 5 seconds
	initialBackoff := 100 * time.Millisecond
	maxBackoff := 5 * time.Second
	maxAttempts := 10

	for attempts := 0; attempts < maxAttempts; attempts++ {
		// Execute the operation
		err = fn()
		if err == nil {
			return nil
		}

		// Check if it's a busy error
		if isBusyError(err) {
			// Calculate backoff with exponential increase and some jitter
			backoff = initialBackoff * time.Duration(1<<uint(attempts))
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

			// Add some randomness (jitter) to avoid thundering herd
			jitter := time.Duration(int64(float64(backoff) * 0.2 * (0.5 + 0.5*float64(time.Now().Nanosecond())/float64(1<<30))))
			backoff = backoff + jitter

			log.Printf("SQLite busy error on %s (attempt %d/%d), retrying after %v: %v",
				operation, attempts+1, maxAttempts, backoff, err)

			// Check if context is cancelled before sleeping
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				// Continue to next attempt
			}
			continue
		}

		// If it's not a busy error, return immediately
		return err
	}

	return fmt.Errorf("operation %s failed after %d attempts: %w", operation, maxAttempts, err)
}

func (s *Service) Create(ctx context.Context, job *Job) error {
	return s.executeWithRetry(ctx, "Create", func() error {
		return s.repo.Create(ctx, job)
	})
}

func (s *Service) All(ctx context.Context) ([]Job, error) {
	var jobs []Job
	err := s.executeWithRetry(ctx, "Select All", func() error {
		var err error
		jobs, err = s.repo.Select(ctx, SelectParams{})
		return err
	})
	return jobs, err
}

func (s *Service) Get(ctx context.Context, id string) (Job, error) {
	var job Job
	err := s.executeWithRetry(ctx, "Get", func() error {
		var err error
		job, err = s.repo.Get(ctx, id)
		return err
	})
	return job, err
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.Contains(id, "/") || strings.Contains(id, "\\") || strings.Contains(id, "..") {
		return fmt.Errorf("invalid file name")
	}

	datapath := filepath.Join(s.dataFolder, id+".csv")

	if _, err := os.Stat(datapath); err == nil {
		if err := os.Remove(datapath); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	return s.executeWithRetry(ctx, "Delete", func() error {
		return s.repo.Delete(ctx, id)
	})
}

func (s *Service) Update(ctx context.Context, job *Job) error {
	return s.executeWithRetry(ctx, "Update", func() error {
		return s.repo.Update(ctx, job)
	})
}

func (s *Service) SelectPending(ctx context.Context) ([]Job, error) {
	var jobs []Job
	err := s.executeWithRetry(ctx, "Select Pending", func() error {
		var err error
		jobs, err = s.repo.Select(ctx, SelectParams{Status: StatusPending, Limit: 1})
		return err
	})
	return jobs, err
}

func (s *Service) GetCSV(_ context.Context, id string) (string, error) {
	if strings.Contains(id, "/") || strings.Contains(id, "\\") || strings.Contains(id, "..") {
		return "", fmt.Errorf("invalid file name")
	}

	datapath := filepath.Join(s.dataFolder, id+".csv")

	if _, err := os.Stat(datapath); os.IsNotExist(err) {
		return "", fmt.Errorf("csv file not found for job %s", id)
	}

	return datapath, nil
}
