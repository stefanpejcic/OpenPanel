package app

import (
	"strings"

	"golang.org/x/net/publicsuffix"
)

// Domain holds the {domain_id, docroot, domain_url, php_version} row shape
// returned by AllDomainsForUser. It lives here (rather than in
// internal/modules/domains) because AllDomainsForUser needs it and
// internal/modules/domains needs to import *App - keeping the type in
// internal/app avoids the import cycle that would create, the same reason
// podman-CLI helpers live in internal/modules/docker instead of here.
type Domain struct {
	DomainID   int
	Docroot    string
	DomainURL  string
	PHPVersion string
}

// MainDomain is a top-level domain, as returned by Categorize.
type MainDomain struct {
	DomainURL string
	TLD       string
}

// Subdomain is a domain classified as belonging under another domain in
// the same list, as returned by Categorize.
type Subdomain struct {
	DomainURL  string
	TLD        string
	MainDomain string
}

// Categorize splits a user's domains into top-level domains and
// subdomains, where "subdomain" means "ends with .<another domain in this
// same list>" - a simple string-suffix heuristic, not a DNS-correctness
// check.
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
