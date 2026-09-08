package app

import (
	"strings"

	"golang.org/x/net/publicsuffix"
)

// Domain is the row shape returned by AllDomainsForUser. It lives here instead of internal/modules/domains to avoid an import cycle (that package needs to import *App).
type Domain struct {
	DomainID   int
	Docroot    string
	DomainURL  string
	PHPVersion string
}

// MainDomain is a top-level domain returned by Categorize.
type MainDomain struct {
	DomainURL string
	TLD       string
}

// Subdomain is a domain classified as belonging under another domain in the same list, returned by Categorize.
type Subdomain struct {
	DomainURL  string
	TLD        string
	MainDomain string
}

// Categorize splits a user's domains into top-level domains and subdomains, using a simple "ends with .<other domain>" string check, not real DNS lookups.
func Categorize(userDomains []Domain) ([]MainDomain, []Subdomain) {
	var mains []MainDomain
	var subs []Subdomain

	for _, d := range userDomains {
		tld, _ := publicsuffix.PublicSuffix(d.DomainURL)

		var parent string
		isSubdomain := false
		for _, other := range userDomains {
			if other.DomainURL == d.DomainURL {
				continue
			}
			if strings.HasSuffix(d.DomainURL, "."+other.DomainURL) {
				isSubdomain = true
				parent = other.DomainURL
				break
			}
		}

		if isSubdomain {
			subs = append(subs, Subdomain{DomainURL: d.DomainURL, TLD: tld, MainDomain: parent})
		} else {
			mains = append(mains, MainDomain{DomainURL: d.DomainURL, TLD: tld})
		}
	}

	return mains, subs
}
