package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// WriteOperation represents a queued write operation
type WriteOperation struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // "add_to_library", "update_progress", "create_review", etc.
	UserID    int64                  `json:"user_id"`
	MangaID   int64                  `json:"manga_id"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	Retries   int                    `json:"retries"`
}

// WriteQueue manages queued write operations
type WriteQueue struct {
	mu          sync.RWMutex
	operations  []WriteOperation
	maxSize     int
	maxRetries  int
	processFunc func(ctx context.Context, op WriteOperation) error
}

// NewWriteQueue creates a new write queue
func NewWriteQueue(maxSize, maxRetries int, processFunc func(ctx context.Context, op WriteOperation) error) *WriteQueue {
	return &WriteQueue{
		operations:  make([]WriteOperation, 0),
		maxSize:     maxSize,
		maxRetries:  maxRetries,
		processFunc: processFunc,
	}
}

// Enqueue adds a write operation to the queue
func (q *WriteQueue) Enqueue(opType string, userID, mangaID int64, data map[string]interface{}) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.operations) >= q.maxSize {
		return fmt.Errorf("write queue is full")
	}

	op := WriteOperation{
		ID:        fmt.Sprintf("%d-%d-%d", userID, mangaID, time.Now().UnixNano()),
		Type:      opType,
		UserID:    userID,
		MangaID:   mangaID,
		Data:      data,
		CreatedAt: time.Now(),
		Retries:   0,
	}

	q.operations = append(q.operations, op)
	return nil
}

// Dequeue removes and returns the oldest operation
func (q *WriteQueue) Dequeue() *WriteOperation {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.operations) == 0 {
		return nil
	}

	op := q.operations[0]
	q.operations = q.operations[1:]
	return &op
}

// Peek returns the oldest operation without removing it
func (q *WriteQueue) Peek() *WriteOperation {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if len(q.operations) == 0 {
		return nil
	}

	op := q.operations[0]
	return &op
}

// Size returns the current queue size
func (q *WriteQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.operations)
}

// IsEmpty returns whether the queue is empty
func (q *WriteQueue) IsEmpty() bool {
	return q.Size() == 0
}

// ProcessNext attempts to process the next operation in the queue
func (q *WriteQueue) ProcessNext(ctx context.Context) error {
	op := q.Dequeue()
	if op == nil {
		return nil // Queue is empty
	}

	err := q.processFunc(ctx, *op)
	if err != nil {
		// Re-queue if retries available
		if op.Retries < q.maxRetries {
			op.Retries++
			q.mu.Lock()
			q.operations = append(q.operations, *op)
			q.mu.Unlock()
		}
		return err
	}

	return nil
}

// ProcessAll attempts to process all operations in the queue
func (q *WriteQueue) ProcessAll(ctx context.Context) (processed, failed int) {
	for !q.IsEmpty() {
		err := q.ProcessNext(ctx)
		if err != nil {
			failed++
		} else {
			processed++
		}
	}
	return processed, failed
}

// GetAll returns all operations (for persistence/recovery)
func (q *WriteQueue) GetAll() []WriteOperation {
	q.mu.RLock()
	defer q.mu.RUnlock()

	result := make([]WriteOperation, len(q.operations))
	copy(result, q.operations)
	return result
}

// Clear removes all operations from the queue
func (q *WriteQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.operations = make([]WriteOperation, 0)
}

// MarshalJSON implements json.Marshaler for persistence
func (q *WriteQueue) MarshalJSON() ([]byte, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return json.Marshal(q.operations)
}

// UnmarshalJSON implements json.Unmarshaler for recovery
func (q *WriteQueue) UnmarshalJSON(data []byte) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	return json.Unmarshal(data, &q.operations)
}
