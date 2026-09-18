package risk

import (
	"fmt"
	"sync"
)

// AddressChecker validates withdrawal destination addresses.
type AddressChecker struct {
	mu        sync.RWMutex
	whitelist map[string]map[string]bool // userID -> address -> trusted
	blacklist map[string]bool
}

func NewAddressChecker() *AddressChecker {
	return &AddressChecker{
		whitelist: make(map[string]map[string]bool),
		blacklist: make(map[string]bool),
	}
}

// WhitelistAddress marks an address as trusted for a user.
func (a *AddressChecker) WhitelistAddress(userID, address string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.whitelist[userID] == nil {
		a.whitelist[userID] = make(map[string]bool)
	}
	a.whitelist[userID][address] = true
}

func (a *AddressChecker) IsWhitelisted(userID, address string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.whitelist[userID] != nil && a.whitelist[userID][address]
}

// Check validates the destination address.
// Non-whitelisted addresses are allowed but flagged (tier may be upgraded).
func (a *AddressChecker) Check(userID, address string) (bool, error) {
	if address == "" {
		return false, fmt.Errorf("destination address required")
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.blacklist[address] {
		return false, fmt.Errorf("address is blacklisted")
	}

	whitelisted := a.whitelist[userID] != nil && a.whitelist[userID][address]
	return whitelisted, nil
}
