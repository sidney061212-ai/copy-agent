package agent

import (
	"errors"
	"testing"
)

func TestSessionTurnQueueStartsFirstMessageImmediately(t *testing.T) {
	queue := NewSessionTurnQueue(2)
	started, err := queue.BeginOrQueue("feishu:c1:u1", &Message{Content: "first"})
	if err != nil {
		t.Fatalf("BeginOrQueue returned error: %v", err)
	}
	if !started {
		t.Fatal("expected first message to start immediately")
	}
	if !queue.IsBusy("feishu:c1:u1") {
		t.Fatal("expected session to be busy")
	}
}

func TestSessionTurnQueueQueuesWhileBusyAndDequeuesInOrder(t *testing.T) {
	queue := NewSessionTurnQueue(2)
	mustStart(t, queue, "feishu:c1:u1", "first")
	mustQueue(t, queue, "feishu:c1:u1", "second")
	mustQueue(t, queue, "feishu:c1:u1", "third")

	if got := queue.PendingLen("feishu:c1:u1"); got != 2 {
		t.Fatalf("PendingLen() = %d", got)
	}
	next, ok := queue.CompleteAndDequeue("feishu:c1:u1")
	if !ok || next.Content != "second" {
		t.Fatalf("first dequeue = %#v, %v", next, ok)
	}
	next, ok = queue.CompleteAndDequeue("feishu:c1:u1")
	if !ok || next.Content != "third" {
		t.Fatalf("second dequeue = %#v, %v", next, ok)
	}
	next, ok = queue.CompleteAndDequeue("feishu:c1:u1")
	if ok || next != nil {
		t.Fatalf("expected queue drained, got %#v, %v", next, ok)
	}
	if queue.IsBusy("feishu:c1:u1") {
		t.Fatal("expected session no longer busy")
	}
}

func TestSessionTurnQueueRejectsWhenFull(t *testing.T) {
	queue := NewSessionTurnQueue(1)
	mustStart(t, queue, "feishu:c1:u1", "first")
	mustQueue(t, queue, "feishu:c1:u1", "second")

	started, err := queue.BeginOrQueue("feishu:c1:u1", &Message{Content: "third"})
	if started {
		t.Fatal("did not expect full queue message to start")
	}
	if !errors.Is(err, ErrSessionQueueFull) {
		t.Fatalf("expected ErrSessionQueueFull, got %v", err)
	}
}

func TestSessionTurnQueueKeepsSessionKeysIsolated(t *testing.T) {
	queue := NewSessionTurnQueue(1)
	mustStart(t, queue, "feishu:c1:u1", "first")
	started, err := queue.BeginOrQueue("feishu:c1:u2", &Message{Content: "other user"})
	if err != nil || !started {
		t.Fatalf("second session should start independently: started=%v err=%v", started, err)
	}
}

func mustStart(t *testing.T, queue *SessionTurnQueue, sessionKey string, content string) {
	t.Helper()
	started, err := queue.BeginOrQueue(sessionKey, &Message{Content: content})
	if err != nil || !started {
		t.Fatalf("expected start for %q: started=%v err=%v", content, started, err)
	}
}

func mustQueue(t *testing.T, queue *SessionTurnQueue, sessionKey string, content string) {
	t.Helper()
	started, err := queue.BeginOrQueue(sessionKey, &Message{Content: content})
	if err != nil || started {
		t.Fatalf("expected queue for %q: started=%v err=%v", content, started, err)
	}
}
