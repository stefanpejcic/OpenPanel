package emails

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sieveparser"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var pageFiles = []string{
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
}

func loadPage(files ...string) *web.Page {
	return web.MustLoadPage(append(append([]string{}, pageFiles...), files...)...)
}

var (
	accountsPage             = loadPage("emails/accounts.html")
	newEmailPage             = loadPage("emails/new.html")
	singleAccountPage        = loadPage("emails/single_account.html")
	deletePage               = loadPage("emails/delete.html")
	infoPage                 = loadPage("emails/info.html")
	aliasesPage              = loadPage("emails/aliases.html")
	aliasDetailPage          = loadPage("emails/alias_detail.html")
	aliasNewPage             = loadPage("emails/alias_new.html")
	aliasDeletePage          = loadPage("emails/alias_delete.html")
	defaultAddressPage       = loadPage("emails/default_address.html")
	deliverabilityPage       = loadPage("emails/deliverability.html")
	deliverabilityDomainPage = loadPage("emails/deliverability_domain.html")
	filterPage               = loadPage("emails/filter.html")
	importPage               = loadPage("emails/import.html")
	confirmImportPage        = loadPage("emails/confirm_import.html")
)

func addressesOf(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) > 1 {
			out = append(out, parts[1])
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// accounts.html
// ---------------------------------------------------------------------------

// AccountsPageData is emails/accounts.html's template context.
type AccountsPageData struct {
	web.LayoutData
	Rows          []EmailListRow
	TotalCount    int
	QuotaToastID  string
	QuotaToastMsg string
}

// emailQuotaToast mirrors accounts.html's ns.quota_issues +
// quota_toast_issues 1/2/multiple pluralization branch (percent_val > 80).
func emailQuotaToast(rows []EmailListRow) (id, message string) {
	type issue struct {
		address string
		percent int
	}
	var issues []issue
	for _, row := range rows {
		if row.PercentVal > 80 {
			issues = append(issues, issue{row.Address, row.PercentVal})
		}
	}

	maxPercent := 0
	addrs := make([]string, len(issues))
	for i, is := range issues {
		addrs[i] = is.address
		if is.percent > maxPercent {
			maxPercent = is.percent
		}
	}

	switch len(issues) {
	case 0:
		return "", ""
	case 1:
		return "quota:" + addrs[0], "Email " + addrs[0] + " is reaching its quota (" + strconv.Itoa(issues[0].percent) + "%)."
	case 2:
		return "quota:" + strings.Join(addrs, ","), "Emails " + strings.Join(addrs, ", ") + " are reaching their quota."
	default:
		return "quota:multiple", strconv.Itoa(len(issues)) + " emails are reaching their quota."
	}
}

func renderAccountsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, currentEmailsList []string, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Emails")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	rows := make([]EmailListRow, len(currentEmailsList))
	for i, line := range currentEmailsList {
		rows[i] = parseEmailListRow(line)
	}
	toastID, toastMsg := emailQuotaToast(rows)
	data := AccountsPageData{LayoutData: layout, Rows: rows, TotalCount: len(currentEmailsList), QuotaToastID: toastID, QuotaToastMsg: toastMsg}
	if err := accountsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - accounts template render error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// new.html
// ---------------------------------------------------------------------------

// NewEmailPageData is emails/new.html's template context.
type NewEmailPageData struct {
	web.LayoutData
	Domains              []appctx.Domain
	MaxEmailQuotaNumeric float64
	AllocatedUnit        string
}

func renderNewEmailPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain, maxEmailQuotaNumeric float64, allocatedUnit string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Create Account")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := NewEmailPageData{LayoutData: layout, Domains: domains, MaxEmailQuotaNumeric: maxEmailQuotaNumeric, AllocatedUnit: allocatedUnit}
	if err := newEmailPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - new template render error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// single_account.html
// ---------------------------------------------------------------------------

// SingleAccountPageData is emails/single_account.html's template context.
type SingleAccountPageData struct {
	web.LayoutData
	Quota                SingleEmailQuota
	MaxEmailQuotaNumeric float64
	AllocatedUnit        string
	ServerIP             string
	DedicatedIP          string
	SendRestriction      string
	ReceiveRestriction   string
}

func renderSingleAccountPage(a *appctx.App, w http.ResponseWriter, r *http.Request, currentEmailsList string, maxEmailQuotaNumeric float64, allocatedUnit, serverIP, dedicatedIP, sendRestriction, receiveRestriction string) {
	quota := parseSingleEmailQuota(currentEmailsList)
	layout, _, err := web.BuildLayoutData(a, w, r, quota.Address)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := SingleAccountPageData{
		LayoutData: layout, Quota: quota, MaxEmailQuotaNumeric: maxEmailQuotaNumeric, AllocatedUnit: allocatedUnit,
		ServerIP: serverIP, DedicatedIP: dedicatedIP, SendRestriction: sendRestriction, ReceiveRestriction: receiveRestriction,
	}
	if err := singleAccountPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - single account template render error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// delete.html
// ---------------------------------------------------------------------------

// DeletePageData is emails/delete.html's template context.
type DeletePageData struct {
	web.LayoutData
	Address   string
	Addresses []string
}

func renderDeletePage(a *appctx.App, w http.ResponseWriter, r *http.Request, address string, currentEmailsList []string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Delete address")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DeletePageData{LayoutData: layout, Address: address, Addresses: addressesOf(currentEmailsList)}
	if err := deletePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - delete template render error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// info.html
// ---------------------------------------------------------------------------

// InfoPageData is emails/info.html's template context.
type InfoPageData struct {
	web.LayoutData
	Address  string
	Scheme   string
	Hostname string
}

func renderInfoPage(a *appctx.App, w http.ResponseWriter, r *http.Request, address, scheme, hostname string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Connect Devices")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := InfoPageData{LayoutData: layout, Address: address, Scheme: scheme, Hostname: hostname}
	if err := infoPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - info template render error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// aliases.html
// ---------------------------------------------------------------------------

// AliasesPageData is emails/aliases.html's template context.
type AliasesPageData struct {
	web.LayoutData
	AliasList []AliasEntry
	Domains   []appctx.Domain
}

func renderAliasesPage(a *appctx.App, w http.ResponseWriter, r *http.Request, aliasList []AliasEntry, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Aliases")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := AliasesPageData{LayoutData: layout, AliasList: aliasList, Domains: domains}
	if err := aliasesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - aliases template render error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// alias_detail.html
// ---------------------------------------------------------------------------

// AliasDetailPageData is emails/alias_detail.html's template context.
type AliasDetailPageData struct {
	web.LayoutData
	Email string
	Entry *AliasEntry
}

func renderAliasDetailPage(a *appctx.App, w http.ResponseWriter, r *http.Request, email string, entry *AliasEntry) {
	layout, _, err := web.BuildLayoutData(a, w, r, email)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := AliasDetailPageData{LayoutData: layout, Email: email, Entry: entry}
	if err := aliasDetailPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - alias detail template render error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// alias_new.html
// ---------------------------------------------------------------------------

// AliasNewPageData is emails/alias_new.html's template context.
type AliasNewPageData struct {
	web.LayoutData
	Domains []appctx.Domain
}

func renderAliasNewPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "New Alias")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := AliasNewPageData{LayoutData: layout, Domains: domains}
	if err := aliasNewPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - alias new template render error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// alias_delete.html
// ---------------------------------------------------------------------------

// AliasDeletePageData is emails/alias_delete.html's template context.
type AliasDeletePageData struct {
	web.LayoutData
	Email     string
	AliasList []AliasEntry
}

func renderAliasDeletePage(a *appctx.App, w http.ResponseWriter, r *http.Request, email string, aliasList []AliasEntry) {
	title := "Select Alias"
	if email != "" {
		title = "Delete Alias"
	}
	layout, _, err := web.BuildLayoutData(a, w, r, title)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := AliasDeletePageData{LayoutData: layout, Email: email, AliasList: aliasList}
	if err := aliasDeletePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - alias delete template render error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// default_address.html
// ---------------------------------------------------------------------------

// DefaultAddressPageData is emails/default_address.html's template context.
type DefaultAddressPageData struct {
	web.LayoutData
	Domain  string
	Domains []appctx.Domain
	Current string
}

func renderDefaultAddressPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domain string, domains []appctx.Domain, current string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Default Email")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DefaultAddressPageData{LayoutData: layout, Domain: domain, Domains: domains, Current: current}
	if err := defaultAddressPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - default address template render error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// deliverability.html / deliverability_domain.html
// ---------------------------------------------------------------------------

// DeliverabilityPageData is emails/deliverability.html's template context.
type DeliverabilityPageData struct {
	web.LayoutData
	Domains []appctx.Domain
}

func renderDeliverabilityPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Email Deliverability")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DeliverabilityPageData{LayoutData: layout, Domains: domains}
	if err := deliverabilityPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - deliverability template render error: %v", err)
	}
}

// DeliverabilityDomainPageData is emails/deliverability_domain.html's
// template context.
type DeliverabilityDomainPageData struct {
	web.LayoutData
	Domain string
}

func renderDeliverabilityDomainPage(a *appctx.App, w http.ResponseWriter, r *http.Request, domain string) {
	layout, _, err := web.BuildLayoutData(a, w, r, domain)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := DeliverabilityDomainPageData{LayoutData: layout, Domain: domain}
	if err := deliverabilityDomainPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - deliverability domain template render error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// filter.html
// ---------------------------------------------------------------------------

// FilterPageData is emails/filter.html's template context.
type FilterPageData struct {
	web.LayoutData
	Email         string
	Addresses     []string
	ParsedFilters []sieveparser.Filter
	RawContent    string
	SieveFile     string
	ViewMode      string
}

func renderFilterPage(a *appctx.App, w http.ResponseWriter, r *http.Request, email string, currentEmailsList []string, parsedFilters []sieveparser.Filter, rawContent, sieveFile, viewMode string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Filters")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := FilterPageData{
		LayoutData: layout, Email: email, Addresses: addressesOf(currentEmailsList),
		ParsedFilters: parsedFilters, RawContent: rawContent, SieveFile: sieveFile, ViewMode: viewMode,
	}
	if err := filterPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - filter template render error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// import.html / confirm_import.html
// ---------------------------------------------------------------------------

func renderImportPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Import")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := importPage.Render(w, http.StatusOK, struct{ web.LayoutData }{layout}); err != nil {
		log.Printf("EMAILS - import template render error: %v", err)
	}
}

// ConfirmImportPageData is emails/confirm_import.html's template context.
type ConfirmImportPageData struct {
	web.LayoutData
	ValidUsers   []ImportRow
	InvalidUsers []ImportRow
}

func renderConfirmImportPage(a *appctx.App, w http.ResponseWriter, r *http.Request, validUsers, invalidUsers []ImportRow) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Import Emails")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := ConfirmImportPageData{LayoutData: layout, ValidUsers: validUsers, InvalidUsers: invalidUsers}
	if err := confirmImportPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - confirm import template render error: %v", err)
	}
}
