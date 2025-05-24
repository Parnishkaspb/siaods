package crdt

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

func simulateNetwork(replicas []*Replica) {
	for _, r := range replicas {
		r.RunProtocol(replicas)
	}
}

func wait() {
	time.Sleep(500 * time.Millisecond)
}

func TestLogoot_SequentialInsertions(t *testing.T) {
	r1 := NewReplica("A")
	r2 := NewReplica("B")
	r3 := NewReplica("C")
	replicas := []*Replica{r1, r2, r3}
	simulateNetwork(replicas)

	for _, ch := range "hello world" {
		r1.LocalInsert(len(r1.Text()), ch)
	}
	wait()

	expected := r1.Text()
	for _, r := range replicas {
		t.Logf("Replica %s final state: %q", r.SiteID, r.Text())
		if r.Text() != expected {
			t.Errorf("Replica %s diverged: got %q, want %q", r.SiteID, r.Text(), expected)
		}
	}
}

func TestLogoot_RandomEdits(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	r1 := NewReplica("A")
	r2 := NewReplica("B")
	r3 := NewReplica("C")
	replicas := []*Replica{r1, r2, r3}
	simulateNetwork(replicas)

	for _, ch := range "distributed" {
		r1.LocalInsert(len(r1.Text()), ch)
	}
	wait()

	for i := 0; i < 30000; i++ {
		pos := rand.Intn(len(r2.Text()) / 2)
		r2.LocalInsert(pos, rune('A'+i%27))
	}
	//wait()

	for i := 0; i < 30000; i++ {
		pos := rand.Intn(len(r3.Text()) / 2)
		r3.LocalInsert(pos, rune('A'+i%27))
	}
	wait()

	reference := r1.Text()
	for _, r := range replicas {
		t.Logf("Replica %s final state: %q", r.SiteID, r.Text())
		if r.Text() != reference {
			t.Errorf("Replica %s diverged: %q vs %q", r.SiteID, r.Text(), reference)
		}
	}
}

func TestLogoot_MultipleWriters(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	r1 := NewReplica("A")
	r2 := NewReplica("B")
	r3 := NewReplica("C")
	replicas := []*Replica{r1, r2, r3}
	simulateNetwork(replicas)

	for _, ch := range "collab" {
		r1.LocalInsert(len(r1.Text()), ch)
	}
	wait()

	for i := 0; i < 3; i++ {
		r1.LocalInsert(rand.Intn(len(r1.Text())+1), '!')
		r2.LocalInsert(rand.Intn(len(r2.Text())+1), '?')
		r3.LocalInsert(rand.Intn(len(r3.Text())+1), '*')
	}
	wait()

	reference := r1.Text()
	for _, r := range replicas {
		t.Logf("Replica %s final state: %q", r.SiteID, r.Text())
		if r.Text() != reference {
			t.Errorf("Replica %s mismatch: %q != %q", r.SiteID, r.Text(), reference)
		}
	}
}

func TestLogoot_ConflictingEdits(t *testing.T) {
	r1 := NewReplica("A")
	r2 := NewReplica("B")
	r3 := NewReplica("C")
	replicas := []*Replica{r1, r2, r3}
	simulateNetwork(replicas)

	for _, ch := range "abcdefgkl" {
		r1.LocalInsert(len(r1.Text()), ch)
	}
	wait()

	if len(r1.Atoms) > 3 {
		target := r1.Atoms[3]
		r1.Outbox <- Message{From: r1.SiteID, Operation: "delete", Atom: target}
		r1.LocalInsert(3, 'D')
		r2.Outbox <- Message{From: r2.SiteID, Operation: "delete", Atom: target}
		r2.LocalInsert(3, 'B')
	}
	wait()

	reference := r1.Text()
	for _, r := range replicas {
		t.Logf("Replica %s final state: %q", r.SiteID, r.Text())
		if r.Text() != reference {
			t.Errorf("Conflict resolution mismatch on %s: %q vs %q", r.SiteID, r.Text(), reference)
		}
	}
}

func TestLogoot_OutOfOrderDelivery(t *testing.T) {
	r1 := NewReplica("A")
	r2 := NewReplica("B")
	r3 := NewReplica("C")
	replicas := []*Replica{r1, r2, r3}

	simulateNetwork([]*Replica{r1, r2})

	op1 := Message{From: r1.SiteID, Operation: "insert", Atom: Atom{Char: 'X', Position: r1.generatePosition(0)}}
	r1.apply(op1)
	r1.Outbox <- op1

	time.Sleep(10 * time.Millisecond)

	op2 := Message{From: r1.SiteID, Operation: "insert", Atom: Atom{Char: 'Y', Position: r1.generatePosition(1)}}
	r1.apply(op2)
	r1.Outbox <- op2

	time.Sleep(10 * time.Millisecond)

	op3 := Message{From: r1.SiteID, Operation: "insert", Atom: Atom{Char: 'Z', Position: r1.generatePosition(2)}}
	r1.apply(op3)
	r1.Outbox <- op3

	r3.Inbox <- op3

	time.Sleep(50 * time.Millisecond)

	r3.Inbox <- op1
	r3.Inbox <- op2

	wait()

	reference := r1.Text()
	for _, r := range replicas {
		if r.Text() != reference {
			t.Errorf("Replica %s state mismatch: %q != %q", r.SiteID, r.Text(), reference)
		}
	}
}

func TestLogoot_WordReplacementConflict(t *testing.T) {
	r1 := NewReplica("A")
	r2 := NewReplica("B")
	replicas := []*Replica{r1, r2}
	simulateNetwork(replicas)

	original := "The quick brown fox jumps fox"
	for _, ch := range original {
		r1.LocalInsert(len(r1.Text()), ch)
	}
	wait() // Ждем, чтобы все символы были доставлены

	text := r1.Text()
	start := strings.Index(text, "fox")
	if start == -1 {
		t.Fatalf("Слово 'fox' не найдено в тексте: %q", text)
	}

	positions := r1.Atoms[start : start+3]

	// Удаление слова "fox" репликой A
	for _, a := range positions {
		r1.Outbox <- Message{From: r1.SiteID, Operation: "delete", Atom: a}
	}

	// Вставка "cat" в конец текста репликой A
	for _, ch := range "cat" {
		r1.LocalInsert(len(r1.Text()), ch)
	}

	// Удаление слова "fox" репликой B
	for _, a := range positions {
		r2.Outbox <- Message{From: r2.SiteID, Operation: "delete", Atom: a}
	}

	// Вставка "dog" в конец текста репликой B
	for _, ch := range "dog" {
		r2.LocalInsert(len(r2.Text()), ch)
	}

	// Повторно запускаем сетевую синхронизацию
	simulateNetwork(replicas)
	wait()

	reference := r1.Text()
	for _, r := range replicas {
		if r.Text() != reference {
			t.Errorf("Replica %s mismatch: %q != %q", r.SiteID, r.Text(), reference)
		}
	}

	if !strings.Contains(reference, "cat") || !strings.Contains(reference, "dog") {
		t.Errorf("Слово 'fox' было удалено, но 'cat' или 'dog' не были вставлены. Финальный текст: %s", reference)
	} else {
		t.Logf("✅ Финальный текст: %q", reference)
	}
}
