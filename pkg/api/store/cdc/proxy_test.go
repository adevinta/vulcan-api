/*
Copyright 2021 Adevinta
*/

package cdc

import (
	"errors"
	"testing"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/adevinta/vulcan-api/pkg/api"
)

// mockDB
type mockDB struct {
	DB
	logEntries []Event
}

func (m *mockDB) GetLog(nEntries uint) ([]Event, error) {
	if int(nEntries) <= len(m.logEntries) {
		return m.logEntries[:nEntries], nil
	}
	return m.logEntries, nil
}
func (m *mockDB) FailedEvent(event Event) error {
	return nil
}
func (m *mockDB) CleanEvent(event Event) error {
	for i, e := range m.logEntries {
		if e.ID() == event.ID() {
			m.logEntries = append(m.logEntries[:i], m.logEntries[i+1:]...)
			return nil
		}
	}
	return errors.New("event not found")
}
func (m *mockDB) CleanLog(nEntries uint) error {
	if int(nEntries) <= len(m.logEntries) {
		m.logEntries = m.logEntries[nEntries:]
		return nil
	}
	m.logEntries = []Event{}
	return nil
}
func (m *mockDB) TryGetLock(id uint32) (*Lock, error) {
	return &Lock{Acquired: true}, nil
}
func (m *mockDB) ReleaseLock(l *Lock) error {
	return nil
}

// mockParser
type mockParser struct {
	Parser
	totalParsed   uint
	wantParseErr  bool
	mockParseTime *time.Duration
}

func (m *mockParser) Parse(log []Event) (nParsed uint) {
	if m.mockParseTime != nil {
		time.Sleep(*m.mockParseTime)
	}
	if m.wantParseErr {
		return
	}
	nParsed = uint(len(log))
	m.totalParsed += nParsed
	return
}

// mockStore
type mockStore struct {
	api.VulcanitoStore
}

func (m *mockStore) ListTeams() ([]*api.Team, error) {
	return []*api.Team{}, nil
}

func (m *mockStore) DeleteTeam(teamID string) error {
	return nil
}

// mockLogger
type mockLogger struct {
	log.Logger
}

func TestBrokerProxySync(t *testing.T) {

	// Overwrite default start and err
	// periods
	defErrAwakePeriod = 1 * time.Second

	mockStore := &mockStore{}

	t.Run("Happy path", func(t *testing.T) {
		mockDB := &mockDB{
			logEntries: []Event{
				Outbox{Operation: "1stAction"},
				Outbox{Operation: "2ndAction"},
				Outbox{Operation: "3rdAction"},
				Outbox{Operation: "4thAction"},
				Outbox{Operation: "5thAction"},
				Outbox{Operation: "6thAction"},
				Outbox{Operation: "7thAction"},
				Outbox{Operation: "8thAction"},
				Outbox{Operation: "9thAction"},
				Outbox{Operation: "10thAction"},
				Outbox{Operation: "11thAction"},
			},
		}

		mockParser := &mockParser{}

		brokerProxy := NewBrokerProxy(&mockLogger{},
			mockDB, mockStore, mockParser)

		// Verify that broker proxy is waiting for signal
		wait()
		if mockParser.totalParsed > 0 {
			t.Fatalf("expected no parsed entries yet, but got: %d",
				mockParser.totalParsed)
		}

		// Execute proxied method without awake and verify
		// broker is still waiting
		_, _ = brokerProxy.ListTeams()
		wait()
		if mockParser.totalParsed > 0 {
			t.Fatalf("expected no parsed entries yet, but got: %d",
				mockParser.totalParsed)
		}

		// Execute proxied method with awake so broker is signaled
		// and starts processing entries
		_ = brokerProxy.DeleteTeam("teamID")
		wait()
		if mockParser.totalParsed != defNChanges {
			t.Fatalf("expected %d parsed entries, but got: %d",
				defNChanges, mockParser.totalParsed)
		}

	})

	t.Run("Should retake processing automatically due to parse err", func(t *testing.T) {
		mockDB := &mockDB{
			logEntries: []Event{
				Outbox{Operation: "1stAction"},
			},
		}

		// Make mockparser fail on first processing
		mockParser := &mockParser{wantParseErr: true}

		brokerProxy := NewBrokerProxy(&mockLogger{},
			mockDB, mockStore, mockParser)

		// Verify that broker proxy is waiting for signal
		wait()
		if mockParser.totalParsed > 0 {
			t.Fatalf("expected no parsed entries yet, but got: %d",
				mockParser.totalParsed)
		}

		// Execute proxied method with awake so broker is signaled
		// and starts processing entries, but because mock parser
		// is set to fail, verify no events have been processed
		_ = brokerProxy.DeleteTeam("teamID")
		wait()
		if mockParser.totalParsed > 0 {
			t.Fatalf("expected no parsed entries due to processing err, but got: %d",
				mockParser.totalParsed)
		}

		// Reset mock parser to not produce error and wait for
		// automatic awake after defErrAwakePeriod
		mockParser.wantParseErr = false
		time.Sleep(defErrAwakePeriod)

		// Verify broker awaked and processed remaining events
		if mockParser.totalParsed != 1 {
			t.Fatalf("expected %d parsed entries, but got: %d",
				1, mockParser.totalParsed)
		}
	})

	t.Run("Should handle locking", func(t *testing.T) {
		// Overwrite default time period after
		// errored event parsing for this test case
		defErrAwakePeriod = 10 * time.Second

		mockDB := &mockDB{
			logEntries: []Event{
				Outbox{Operation: "1stAction"},
			},
		}

		// Make mockparser fail on first processing
		mockWorkTime := 500 * time.Millisecond
		mockParser := &mockParser{
			wantParseErr:  true,
			mockParseTime: &mockWorkTime,
		}

		brokerProxy := NewBrokerProxy(&mockLogger{},
			mockDB, mockStore, mockParser)

		// Verify that broker proxy is waiting for signal
		wait()
		if mockParser.totalParsed > 0 {
			t.Fatalf("expected no parsed entries yet, but got: %d",
				mockParser.totalParsed)
		}

		// Execute proxied method which will trigger a 500ms mock
		// parser work time and result as errored event
		_ = brokerProxy.DeleteTeam("teamID")
		// Sleep so we let parser start working after signal from
		// previous proxied method call
		time.Sleep(100 * time.Millisecond)
		// Reset mockParser so no error is produced in parser and
		// no mock work is simulated
		mockParser.wantParseErr = false
		mockParser.mockParseTime = nil
		// Execute proxied method again. This one should be lock
		// meanwhile first parser execution finishes (after the mock
		// time) and then trigger parser to process the log correctly
		_ = brokerProxy.DeleteTeam("teamID")
		wait()

		if mockParser.totalParsed != 1 {
			t.Fatalf("expected %d parsed entries, but got: %d",
				1, mockParser.totalParsed)
		}
	})
}

func wait() {
	time.Sleep(500 * time.Millisecond)
}
