package syncdeque

import (
	"sync"
	"testing"
)

func TestDequeBasic(t *testing.T) {
	d := New[int]()
	if d.Size() != 0 {
		t.Fatalf("empty Size = %d, want 0", d.Size())
	}
	if v, ok := d.Shift(); ok || v != 0 {
		t.Fatalf("empty Shift = (%d, %v), want (0, false)", v, ok)
	}

	d.Append(1)
	d.Append(2)
	d.Append(3)
	if d.Size() != 3 {
		t.Fatalf("Size after 3 appends = %d, want 3", d.Size())
	}

	for i, want := range []int{1, 2, 3} {
		v, ok := d.Shift()
		if !ok || v != want {
			t.Fatalf("Shift #%d = (%d, %v), want (%d, true)", i, v, ok, want)
		}
	}
	if d.Size() != 0 {
		t.Fatalf("Size after draining = %d, want 0", d.Size())
	}
	if _, ok := d.Shift(); ok {
		t.Fatalf("Shift on drained deque returned ok=true")
	}
}

func TestDequeConcurrent(t *testing.T) {
	d := New[int]()
	const writers, perWriter = 8, 1000
	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perWriter; i++ {
				d.Append(i)
			}
		}()
	}
	wg.Wait()
	want := writers * perWriter
	if d.Size() != want {
		t.Fatalf("Size after concurrent appends = %d, want %d", d.Size(), want)
	}

	var got int
	for {
		_, ok := d.Shift()
		if !ok {
			break
		}
		got++
	}
	if got != want {
		t.Fatalf("drained %d, want %d", got, want)
	}
}

func TestDequePointerType(t *testing.T) {
	type item struct{ v int }
	d := New[*item]()
	d.Append(&item{v: 7})
	got, ok := d.Shift()
	if !ok || got == nil || got.v != 7 {
		t.Fatalf("Shift = (%+v, %v), want (&{7}, true)", got, ok)
	}
}
