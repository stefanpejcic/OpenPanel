package sieveparser

import "testing"

func TestParseEmpty(t *testing.T) {
	if got := Parse(""); got != nil {
		t.Errorf("Parse(\"\") = %#v, want nil", got)
	}
	if got := Parse("   \n\t "); got != nil {
		t.Errorf("Parse(whitespace) = %#v, want nil", got)
	}
}

func TestParseSingleHeaderTest(t *testing.T) {
	src := `require ["fileinto"];

# Spam to junk
if header :contains "X-Spam-Status" "Yes" {
  fileinto "Junk";
}
`
	filters := Parse(src)
	if len(filters) != 1 {
		t.Fatalf("got %d filters, want 1: %+v", len(filters), filters)
	}
	f := filters[0]
	if f.Name != "Spam to junk" {
		t.Errorf("Name = %q", f.Name)
	}
	if f.Logic != "anyof" {
		t.Errorf("Logic = %q, want anyof", f.Logic)
	}
	if len(f.Rules) != 1 || f.Rules[0].Field != "spam_status" || f.Rules[0].Match != "contains" || f.Rules[0].Value != "Yes" {
		t.Errorf("Rules = %+v", f.Rules)
	}
	if len(f.Actions) != 1 || f.Actions[0].Type != "deliver_folder" || f.Actions[0].Value != "Junk" {
		t.Errorf("Actions = %+v", f.Actions)
	}
}

func TestParseAnyofMultipleConditions(t *testing.T) {
	src := `require ["fileinto"];

if anyof (header :contains "Subject" "invoice", header :contains "From" "billing@example.com") {
  fileinto "Finance";
  stop;
}
`
	filters := Parse(src)
	if len(filters) != 1 {
		t.Fatalf("got %d filters, want 1", len(filters))
	}
	f := filters[0]
	if f.Logic != "anyof" {
		t.Errorf("Logic = %q", f.Logic)
	}
	if len(f.Rules) != 2 {
		t.Fatalf("got %d rules, want 2: %+v", len(f.Rules), f.Rules)
	}
	if f.Rules[0].Field != "subject" || f.Rules[0].Value != "invoice" {
		t.Errorf("Rules[0] = %+v", f.Rules[0])
	}
	if f.Rules[1].Field != "from" || f.Rules[1].Value != "billing@example.com" {
		t.Errorf("Rules[1] = %+v", f.Rules[1])
	}
	if len(f.Actions) != 2 || f.Actions[0].Type != "deliver_folder" || f.Actions[1].Type != "stop" {
		t.Errorf("Actions = %+v", f.Actions)
	}
}

func TestParseNegatedAndWildcards(t *testing.T) {
	src := `if not header :matches "Subject" "Re:*" {
  discard;
  stop;
}
`
	filters := Parse(src)
	if len(filters) != 1 {
		t.Fatalf("got %d filters", len(filters))
	}
	r := filters[0].Rules[0]
	if r.Match != "not_begins" || r.Value != "Re:" {
		t.Errorf("Rule = %+v, want match=not_begins value=Re:", r)
	}
	if len(filters[0].Actions) != 2 || filters[0].Actions[0].Type != "discard" || filters[0].Actions[1].Type != "stop" {
		t.Errorf("Actions = %+v", filters[0].Actions)
	}
}

func TestParseRedirectCopy(t *testing.T) {
	src := `if header :is "From" "boss@example.com" {
  redirect :copy "assistant@example.com";
}
`
	filters := Parse(src)
	if len(filters) != 1 {
		t.Fatalf("got %d filters", len(filters))
	}
	a := filters[0].Actions[0]
	if a.Type != "forward" || a.Value != "assistant@example.com" {
		t.Errorf("Actions[0] = %+v", a)
	}
}

func TestParseVacation(t *testing.T) {
	src := `require ["vacation"];

if header :contains "To" "me@example.com" {
  vacation
    :days 3
    :subject "Out of office"
    "I am currently out of office.";
}
`
	filters := Parse(src)
	if len(filters) != 1 {
		t.Fatalf("got %d filters", len(filters))
	}
	a := filters[0].Actions[0]
	if a.Type != "autoresponder" {
		t.Fatalf("Actions[0].Type = %q", a.Type)
	}
	v, ok := a.Value.(*AutoresponderValue)
	if !ok {
		t.Fatalf("Actions[0].Value type = %T", a.Value)
	}
	if v.Days != 3 || v.Subject != "Out of office" || v.Message != "I am currently out of office." {
		t.Errorf("AutoresponderValue = %+v", v)
	}
}

func TestParseAddressAndExists(t *testing.T) {
	src := `if address :contains ["to", "cc"] "list@example.com" {
  fileinto "Lists";
}
if exists "X-Priority" {
  addflag "\\Flagged";
}
`
	filters := Parse(src)
	if len(filters) != 2 {
		t.Fatalf("got %d filters, want 2", len(filters))
	}
	if filters[0].Rules[0].Field != "any_recipient" || filters[0].Rules[0].Value != "list@example.com" {
		t.Errorf("filters[0].Rules[0] = %+v", filters[0].Rules[0])
	}
	if filters[1].Rules[0].Field != "any_header" || filters[1].Rules[0].Value != "X-Priority" {
		t.Errorf("filters[1].Rules[0] = %+v", filters[1].Rules[0])
	}
	if filters[1].Actions[0].Type != "add_flag" || filters[1].Actions[0].Value != `\Flagged` {
		t.Errorf("filters[1].Actions[0] = %+v", filters[1].Actions[0])
	}
}

func TestParseMarkRead(t *testing.T) {
	src := `if header :contains "From" "notifications@example.com" {
  addflag "\\Seen";
}
`
	filters := Parse(src)
	if len(filters) != 1 || filters[0].Actions[0].Type != "mark_read" {
		t.Errorf("filters = %+v", filters)
	}
}

func TestParseElseBlockDropped(t *testing.T) {
	src := `if header :contains "From" "a@example.com" {
  fileinto "A";
} else {
  fileinto "B";
}
`
	filters := Parse(src)
	if len(filters) != 1 {
		t.Fatalf("got %d filters, want 1 (else block has no rules and should be dropped): %+v", len(filters), filters)
	}
	if filters[0].Actions[0].Value != "A" {
		t.Errorf("Actions[0] = %+v", filters[0].Actions[0])
	}
}
