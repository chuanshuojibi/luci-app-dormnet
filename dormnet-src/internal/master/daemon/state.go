package daemon

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/openwrt-dormnet/dormnet/shared/log"
	"go.uber.org/zap"
)

type AccountStatus string

const (
	StatusIdle      AccountStatus = "Idle"
	StatusChecking  AccountStatus = "Checking"
	StatusLoggingIn AccountStatus = "LoggingIn"
	StatusOnline    AccountStatus = "Online"
	StatusBackoff   AccountStatus = "Backoff"
	StatusFailed    AccountStatus = "Failed"
	StatusGivenUp   AccountStatus = "GivenUp"
)

const (
	StatusFilePath = "/var/run/dormnet/status.json"

	MaxRetry           = 5
	BaseBackoffMin     = 1
	MaxBackoffMin      = 16
	LightCheckEverySec = 30
)

type AccountState struct {
	Account       string        `json:"account"`
	Iface         string        `json:"iface"`
	Status        AccountStatus `json:"status"`
	Retry         int           `json:"retry"`
	NextAttemptAt int64         `json:"next_attempt_at"` // unix seconds; 0 means immediately
	LastError     string        `json:"last_error"`
	LastUpdated   int64         `json:"last_updated"` // unix seconds
}

var (
	stateMu sync.RWMutex
	states  = map[string]*AccountState{}
)

func keyOf(account, iface string) string {
	return account + "|" + iface
}

func GetAllStates() []AccountState {
	stateMu.RLock()
	defer stateMu.RUnlock()
	out := make([]AccountState, 0, len(states))
	for _, s := range states {
		out = append(out, *s)
	}
	return out
}

func updateState(account, iface string, mut func(s *AccountState)) {
	stateMu.Lock()
	defer stateMu.Unlock()
	key := keyOf(account, iface)
	s, ok := states[key]
	if !ok {
		s = &AccountState{Account: account, Iface: iface, Status: StatusIdle}
		states[key] = s
	}
	mut(s)
	s.LastUpdated = time.Now().Unix()
	persistLocked()
}

func resetMissingAccounts(activeKeys map[string]struct{}) {
	stateMu.Lock()
	defer stateMu.Unlock()
	changed := false
	for k := range states {
		if _, ok := activeKeys[k]; !ok {
			delete(states, k)
			changed = true
		}
	}
	if changed {
		persistLocked()
	}
}

// persistLocked must be called with stateMu held.
func persistLocked() {
	dir := filepath.Dir(StatusFilePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Warn("failed to create status dir", zap.String("dir", dir), zap.Error(err))
		return
	}
	out := make([]AccountState, 0, len(states))
	for _, s := range states {
		out = append(out, *s)
	}
	body, err := json.Marshal(map[string]interface{}{
		"states":  out,
		"written": time.Now().Unix(),
	})
	if err != nil {
		log.Warn("failed to marshal status", zap.Error(err))
		return
	}
	tmp := StatusFilePath + ".tmp"
	if err := os.WriteFile(tmp, body, 0o644); err != nil {
		log.Warn("failed to write status tmp", zap.Error(err))
		return
	}
	if err := os.Rename(tmp, StatusFilePath); err != nil {
		log.Warn("failed to rename status file", zap.Error(err))
		return
	}
}

// ResetAllStatesOnStartup is called from daemon entry to remove stale states.
func ResetAllStatesOnStartup() {
	stateMu.Lock()
	defer stateMu.Unlock()
	states = map[string]*AccountState{}
	persistLocked()
}
