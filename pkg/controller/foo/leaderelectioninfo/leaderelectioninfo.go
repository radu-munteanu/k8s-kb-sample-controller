package leaderelectioninfo

import "sync"

// LeaderElectionInfo - contains information about the leader election
type LeaderElectionInfo struct {
	sync.Mutex
	identity string
	leader   string
}

// New returns a LeaderElectionInfo
func New(identity string) *LeaderElectionInfo {
	return &LeaderElectionInfo{identity: identity}
}

// SetLeader safely sets the leader
func (l *LeaderElectionInfo) SetLeader(leader string) {
	l.Lock()
	l.leader = leader
	l.Unlock()
}

// IsLeader safely checks if the info's identity is equal to the leader
func (l *LeaderElectionInfo) IsLeader() bool {
	l.Lock()
	defer l.Unlock()

	return l.identity == l.leader
}

// GetLeader safely returns the leader's name
func (l *LeaderElectionInfo) GetLeader() string {
	l.Lock()
	leader := l.leader
	l.Unlock()

	return leader
}
