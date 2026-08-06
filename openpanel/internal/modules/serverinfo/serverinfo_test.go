package serverinfo

import "testing"

func TestBuildPageEntries(t *testing.T) {
	entries := buildPageEntries(5, 10)
	var numbers []int
	ellipses := 0
	for _, e := range entries {
		if e.IsEllipsis {
			ellipses++
		} else {
			numbers = append(numbers, e.Number)
		}
	}
	want := []int{1, 3, 4, 5, 6, 7, 10}
	if len(numbers) != len(want) {
		t.Fatalf("expected numbers %v, got %v", want, numbers)
	}
	for i, n := range want {
		if numbers[i] != n {
			t.Errorf("position %d: expected %d, got %d", i, n, numbers[i])
		}
	}
	if ellipses != 2 {
		t.Errorf("expected 2 ellipses, got %d", ellipses)
	}
}

func TestUnameToString(t *testing.T) {
	var b [65]int8
	for i, c := range "Linux" {
		b[i] = int8(c)
	}
	if got := unameToString(b); got != "Linux" {
		t.Errorf("got %q, want %q", got, "Linux")
	}
	if got := unameToString([65]int8{}); got != "" {
		t.Errorf("empty array: got %q, want empty", got)
	}
}
