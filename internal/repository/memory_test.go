package repository

import (
	"strconv"
	"sync"
	"testing"
)

func newInMemStore(t *testing.T) *InMemory {
	t.Helper()

	store := NewInMemoryStore()
	return store
}

func TestSet(t *testing.T){
	s := newInMemStore(t)

	test_key := "y"
	test_value := "54"

	if s.Set(test_key, test_value); len(s.kvpair) != 1 {
		t.Fatalf("expected %d, got %d", 1, len(s.kvpair))
	}
}

func TestGet(t *testing.T){
	s := newInMemStore(t)

	test_key := "x"
	test_value := "12"
	s.Set(test_key, test_value)

	if got, _ := s.Get(test_key); got != test_value {
		t.Fatalf("want %s, got %s", test_value, got)
	}
}

func TestUpdate(t *testing.T){
	s := newInMemStore(t)

	// seed data
	test_key := "x"
	test_value := "12"
	s.Set(test_key, test_value)

	// test update
	new_value := "66"
	s.Update(test_key, new_value)
	if got, _ := s.Get(test_key); got != new_value {
		t.Fatalf("expected %s value, got %s", new_value, got)
	}
}

func TestDelete(t *testing.T){
	s := newInMemStore(t)

	// seed data
	test_key := "x"
	test_value := "12"
	s.Set(test_key, test_value)

	// test delete
	s.Delete(test_key)
	if _, err := s.Get(test_key); err == nil{
		t.Fatalf("expected not found error, got \"%s\"", err)
	}
}

func TestExistsFalse(t *testing.T){
	s := newInMemStore(t)

	if got := s.Exists("inexistent key"); got != false{
		t.Fatalf("expected false, got %t", got)
	}
}

func TestExistsTrue(t *testing.T){
	s := newInMemStore(t)

	// seed data
	test_key := "x"
	test_value := "12"
	s.Set(test_key, test_value)

	if got := s.Exists(test_key); got != true{
		t.Fatalf("expected true, got %t", got)
	}
}

func TestKeys(t *testing.T){
	s := newInMemStore(t)

	// seed data
	test_data := map[string]string{
		"x" : "2",
		"y" : "34",
		"z" : "1212",
	}
	for k,v := range test_data{
		s.Set(k, v)
	}

	// test
	if got := s.Keys(); len(got) != len(test_data) {
		t.Fatalf("expected %d keys, got %d", len(test_data), len(got))
	}
}

func TestClear(t *testing.T){
	s := newInMemStore(t)

	// seed data
	test_data := map[string]string{
		"x" : "2",
		"y" : "34",
		"z" : "1212",
	}
	for k,v := range test_data{
		s.Set(k, v)
	}

	// test
	s.Clear()
	if got := s.Keys(); len(got) != 0 {
		t.Fatalf("expected %d keys, got %d", 0, len(got))
	}
}

// Concurrent tests
func TestConcurrentWrites(t *testing.T){
	s := newInMemStore(t)

	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)

		go func(i int){
			defer wg.Done()
			s.Set(strconv.Itoa(i), strconv.Itoa(i*5))
		}(i)
	}
	wg.Wait()

	// test whether 100 keys were set
	if got := s.Keys(); len(got) != 100 {
		t.Fatalf("expected 100 keys, got %d", len(got))
	}
	// test each key
	for i := range 100 {
		val,_ := s.Get(strconv.Itoa(i))
		if val != strconv.Itoa(i*5){
			t.Fatalf("expected value %d for key = %d, got %s", i*5, i, val)
		}
	}
}