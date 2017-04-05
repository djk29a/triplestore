package triplestore_test

import (
	"sync"
	"testing"

	tstore "github.com/wallix/triplestore"
)

func TestQueries(t *testing.T) {
	all := []tstore.Triple{
		tstore.SubjPred("one", "two").StringLiteral("three"),
		tstore.SubjPred("four", "two").IntegerLiteral(42),
		tstore.SubjPred("one", "two").Resource("four"),
	}

	s := tstore.New()
	s.Add(all...)

	g := s.Snapshot()

	if got, want := g.Count(), len(all); got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
	if got, want := tstore.Triples(g.Triples()), tstore.Triples(all); !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	exp := tstore.Triples{all[0], all[2]}
	if got, want := tstore.Triples(g.WithSubject("one")), exp; !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	exp = tstore.Triples{all[0], all[1], all[2]}
	if got, want := tstore.Triples(g.WithPredicate("two")), exp; !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	exp = tstore.Triples{all[1]}
	if got, want := tstore.Triples(g.WithObject(tstore.IntegerLiteral(42))), exp; !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	exp = tstore.Triples{all[2]}
	if got, want := tstore.Triples(g.WithSubjObj("one", tstore.Resource("four"))), exp; !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	exp = tstore.Triples{all[0], all[2]}
	if got, want := tstore.Triples(g.WithSubjPred("one", "two")), exp; !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	exp = tstore.Triples{all[0]}
	if got, want := tstore.Triples(g.WithPredObj("two", tstore.StringLiteral("three"))), exp; !got.Equal(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestStore(t *testing.T) {
	s := tstore.New()
	s.Add(
		tstore.SubjPred("one", "two").StringLiteral("three"),
		tstore.SubjPred("one", "two").Resource("four"),
		tstore.SubjPred("four", "two").IntegerLiteral(42),
		tstore.SubjPred("one", "two").Resource("four"),
	)
	g := s.Snapshot()
	expected := []tstore.Triple{
		tstore.SubjPred("one", "two").StringLiteral("three"),
		tstore.SubjPred("one", "two").Resource("four"),
		tstore.SubjPred("four", "two").IntegerLiteral(42),
	}
	if got, want := g.Count(), len(expected); got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
	for _, tr := range expected {
		if got, want := g.Contains(tr), true; got != want {
			t.Fatalf("%v: got %t, want %t", tr, got, want)
		}
	}
	s.Remove(tstore.SubjPred("one", "two").Resource("four"))
	newG := s.Snapshot()

	t.Run("old snapshot unmodified", func(t *testing.T) {
		if got, want := g.Count(), len(expected); got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
		for _, tr := range expected {
			if got, want := g.Contains(tr), true; got != want {
				t.Fatalf("%v: got %t, want %t", tr, got, want)
			}
		}
	})

	t.Run("triple 1 removed in new snapshot", func(t *testing.T) {
		if got, want := newG.Count(), 2; got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
		if got, want := newG.Contains(expected[0]), true; got != want {
			t.Fatalf("%v: got %t, want %t", expected[0], got, want)
		}
		if got, want := newG.Contains(expected[1]), false; got != want {
			t.Fatalf("%v: got %t, want %t", expected[1], got, want)
		}
		if got, want := newG.Contains(expected[2]), true; got != want {
			t.Fatalf("%v: got %t, want %t", expected[2], got, want)
		}
	})

}

func TestStoreConcurrentAccess(t *testing.T) {
	s := tstore.New()
	any := tstore.SubjPred("", "").StringLiteral("")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			s.Add(any)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			s.Add(any)
		}
	}()

	wg.Wait()
}
