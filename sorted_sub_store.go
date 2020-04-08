package sls

// SortedSubStoreKey key define
type SortedSubStoreKey struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// IsValid ...
func (s *SortedSubStoreKey) IsValid() bool {
	if len(s.Name) == 0 {
		return false
	}
	if s.Type != "text" &&
		s.Type != "long" &&
		s.Type != "double" {
		return false
	}
	return true
}

// SortedSubStore define
type SortedSubStore struct {
	Name           string              `json:"name,omitempty"`
	TTL            int                 `json:"ttl"`
	SortedKeyCount int                 `json:"sortedKeyCount"`
	TimeIndex      int                 `json:"timeIndex"`
	Keys           []SortedSubStoreKey `json:"keys"`
}

// NewSortedSubStore create a new sorted sub store
func NewSortedSubStore(name string,
	ttl int,
	sortedKeyCount int,
	timeIndex int,
	keys []SortedSubStoreKey) *SortedSubStore {
	sss := &SortedSubStore{
		Name:           name,
		TTL:            ttl,
		SortedKeyCount: sortedKeyCount,
		TimeIndex:      timeIndex,
		Keys:           keys,
	}
	if sss.IsValid() {
		return sss
	}
	return nil
}

// IsValid ...
func (s *SortedSubStore) IsValid() bool {
	if s.SortedKeyCount <= 0 || s.SortedKeyCount >= len(s.Keys) {
		return false
	}
	if s.TimeIndex < 0 || s.TimeIndex <= s.SortedKeyCount {
		return false
	}
	if s.TTL <= 0 || s.TTL > 3650 {
		return false
	}
	for index, key := range s.Keys {
		if !key.IsValid() {
			return false
		}
		if index == s.TimeIndex && key.Type != "long" {
			return false
		}
		if index < s.SortedKeyCount && key.Type == "double" {
			return false
		}
	}
	return true
}
