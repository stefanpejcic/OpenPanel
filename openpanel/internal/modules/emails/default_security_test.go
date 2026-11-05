package emails

import "testing"

func TestSetDefaultAliasRejectsInjectedDomain(t *testing.T) {
	for _, d := range []string{"mine.com/ a@mine.com\n/*@victim.com/ thief@mine.com\n#", "mine.com\nvictim.com", "mine.com/x", ""} {
		if err := setDefaultAliasForDomain(d, "a@mine.com"); err == nil {
			t.Errorf("accepted domain %q", d)
		}
	}
}
