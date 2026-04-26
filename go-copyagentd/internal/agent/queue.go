package agent

import (
	"errors"
	"sync"
)

var ErrSessionQueueFull = errors.New("agent session queue is full")

type PendingMessage struct {
	Message Message
}

type SessionTurnQueue struct {
	mu         sync.Mutex
	maxPending int
	sessions   map[string]*sessionTurnState
}

type sessionTurnState struct {
	busy    bool
	pending []PendingMessage
}

func NewSessionTurnQueue(maxPending int) *SessionTurnQueue {
	if maxPending < 0 {
		maxPending = 0
	}
	return &SessionTurnQueue{maxPending: maxPending, sessions: make(map[string]*sessionTurnState)}
}

func (queue *SessionTurnQueue) BeginOrQueue(sessionKey string, msg *Message) (bool, error) {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	state := queue.state(sessionKey)
	if !state.busy {
		state.busy = true
		return true, nil
	}
	if len(state.pending) >= queue.maxPending {
		return false, ErrSessionQueueFull
	}
	state.pending = append(state.pending, PendingMessage{Message: cloneMessage(msg)})
	return false, nil
}

func (queue *SessionTurnQueue) CompleteAndDequeue(sessionKey string) (*Message, bool) {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	state := queue.sessions[sessionKey]
	if state == nil {
		return nil, false
	}
	if len(state.pending) == 0 {
		state.busy = false
		delete(queue.sessions, sessionKey)
		return nil, false
	}
	next := state.pending[0].Message
	copy(state.pending, state.pending[1:])
	state.pending = state.pending[:len(state.pending)-1]
	state.busy = true
	return &next, true
}

func (queue *SessionTurnQueue) IsBusy(sessionKey string) bool {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	state := queue.sessions[sessionKey]
	return state != nil && state.busy
}

func (queue *SessionTurnQueue) PendingLen(sessionKey string) int {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	state := queue.sessions[sessionKey]
	if state == nil {
		return 0
	}
	return len(state.pending)
}

func (queue *SessionTurnQueue) state(sessionKey string) *sessionTurnState {
	state := queue.sessions[sessionKey]
	if state == nil {
		state = &sessionTurnState{}
		queue.sessions[sessionKey] = state
	}
	return state
}

func cloneMessage(msg *Message) Message {
	if msg == nil {
		return Message{}
	}
	clone := *msg
	clone.Images = append([]ImageAttachment(nil), msg.Images...)
	clone.Files = append([]FileAttachment(nil), msg.Files...)
	return clone
}
