package app

import "testing"

func TestCategorize(t *testing.T) {
	input := []Domain{
		{DomainURL: "example.com"},
		{DomainURL: "blog.example.com"},
		{DomainURL: "shop.example.com"},
		{DomainURL: "another.org"},
	}

	mains, subs := Categorize(input)

	if len(mains) != 2 {
		t.Fatalf("expected 2 main domains, got %d: %+v", len(mains), mains)
	}
	mainSet := map[string]bool{}
	for _, m := range mains {
		mainSet[m.DomainURL] = true
		if m.TLD == "" {
			t.Errorf("expected a non-empty TLD for %q", m.DomainURL)
		}
	}
	if !mainSet["example.com"] || !mainSet["another.org"] {
		t.Errorf("expected example.com and another.org as main domains, got %+v", mains)
	}

	if len(subs) != 2 {
		t.Fatalf("expected 2 subdomains, got %d: %+v", len(subs), subs)
	}
	for _, s := range subs {
		if s.MainDomain != "example.com" {
			t.Errorf("expected subdomain %q's parent to be example.com, got %q", s.DomainURL, s.MainDomain)
		}
	}
}

func TestCategorizeNoSubdomains(t *testing.T) {
	input := []Domain{{DomainURL: "a.com"}, {DomainURL: "b.com"}}
	mains, subs := Categorize(input)
	if len(mains) != 2 || len(subs) != 0 {
		t.Errorf("expected 2 mains/0 subs, got %d mains, %d subs", len(mains), len(subs))
	}
}

func TestCategorizeEmpty(t *testing.T) {
	mains, subs := Categorize(nil)
	if len(mains) != 0 || len(subs) != 0 {
		t.Errorf("expected empty results for empty input, got %d mains, %d subs", len(mains), len(subs))
	}
}
