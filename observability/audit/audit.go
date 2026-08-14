// Package audit provides an append-only audit log with hash-chaining so that
// tampering with past entries is detectable.
//
// Each entry stores the previous entry's hash; recomputing the chain verifies
// integrity. This mirrors the property we required for financial-grade trails.
package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

// Entry is one immutable audit record.
type Entry struct {
	Seq       int       `json:"seq"`
	Timestamp time.Time `json:"timestamp"`
	Actor     string    `json:"actor"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	PrevHash  string    `json:"prev_hash"`
	Hash      string    `json:"hash"`
}

// Log is an append-only, hash-chained log (in-memory for dev; swap storage).
type Log struct {
	mu      sync.Mutex
	entries []Entry
}

// New builds an empty Log.
func New() *Log {
	return &Log{}
}

func hashEntry(e Entry) string {
	e.Hash = ""
	payload, _ := json.Marshal(e)
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

// Append adds a new entry and returns it.
func (l *Log) Append(actor, action, resource string) Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	prev := ""
	if len(l.entries) > 0 {
		prev = l.entries[len(l.entries)-1].Hash
	}
	e := Entry{
		Seq:       len(l.entries) + 1,
		Timestamp: time.Now(),
		Actor:     actor,
		Action:    action,
		Resource:  resource,
		PrevHash:  prev,
	}
	e.Hash = hashEntry(e)
	l.entries = append(l.entries, e)
	return e
}

// Entries returns a snapshot of all entries.
func (l *Log) Entries() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Entry, len(l.entries))
	copy(out, l.entries)
	return out
}

// Verify recomputes the chain and reports whether integrity holds.
func (l *Log) Verify() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, e := range l.entries {
		if hashEntry(e) != e.Hash {
			return false
		}
		if i == 0 {
			if e.PrevHash != "" {
				return false
			}
			continue
		}
		if e.PrevHash != l.entries[i-1].Hash {
			return false
		}
	}
	return true
}

// EnforceAppendOnly returns a copy that rejects entries with explicit seq/hash
// via a Writer guard (inspection helper).
func EnforceAppendOnly(entries []Entry) bool {
	for i := 1; i < len(entries); i++ {
		if entries[i].Seq <= entries[i-1].Seq {
			return false
		}
		if entries[i].Timestamp.Before(entries[i-1].Timestamp) {
			return false
		}
	}
	return true
}

// CSV is a naive exporter for reporting.
func (l *Log) CSV() string {
	var sb strings.Builder
	sb.WriteString("seq,timestamp,actor,action,resource,hash\n")
	for _, e := range l.Entries() {
		sb.WriteString(e.CSVRow())
	}
	return sb.String()
}

// CSVRow renders one row (single-line) for the CSV export.
func (e Entry) CSVRow() string {
	return e.Timestamp.Format(time.RFC3339) + "," + e.Actor + "," + e.Action + "," + e.Resource + "," + e.Hash + "\n"
}
