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

var (
	accountsPage             = web.LoadPage("emails/accounts.html")
	newEmailPage             = web.LoadPage("emails/new.html")
	singleAccountPage        = web.LoadPage("emails/single_account.html")
	deletePage               = web.LoadPage("emails/delete.html")
	infoPage                 = web.LoadPage("emails/info.html")
	aliasesPage              = web.LoadPage("emails/aliases.html")
	aliasNewPage             = web.LoadPage("emails/alias_new.html")
	aliasDetailPage          = web.LoadPage("emails/alias_detail.html")
	aliasDeletePage          = web.LoadPage("emails/alias_delete.html")
	defaultAddressPage       = web.LoadPage("emails/default_address.html")
	deliverabilityPage       = web.LoadPage("emails/deliverability.html")
	deliverabilityDomainPage = web.LoadPage("emails/deliverability_domain.html")
	filterPage               = web.LoadPage("emails/filter.html")
	importPage               = web.LoadPage("emails/import.html")
	confirmImportPage        = web.LoadPage("emails/confirm_import.html")
	migratePage              = web.LoadPage("emails/migrate.html")
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

// accounts.html

// AccountsPageData is emails/accounts.html's template context.
type AccountsPageData struct {
	web.LayoutData
	Rows           []EmailListRow
	TotalCount     int
	QuotaToastID   string
	QuotaToastMsg  string
	EmailLimit     int
	LimitPercent   int
	StorageValue   string
	StorageUnit    string
	NearQuotaCount int
	MailDomains    int
	AliasCount     int
	AliasInboxes   int
	Full           bool
	Connect        AccountsConnect
}

// AccountsConnect drives the "Connect a mail app" card, Created is the mailbox the create form just made
type AccountsConnect struct {
	Created string
	Welcome bool
	Host    string
	Secure  bool
}

// fillAccountsSummary works out the numbers for the stat tiles above the table
func (d *AccountsPageData) fillAccountsSummary(emailLimit int, aliases []AliasEntry) {
	d.EmailLimit = emailLimit
	if emailLimit > 0 {
		d.LimitPercent = min(100, d.TotalCount*100/emailLimit)
		d.Full = d.TotalCount >= emailLimit
	}
	var used float64
	mailDomains := map[string]bool{}
	for _, row := range d.Rows {
		used += row.UsedBytes
		if _, domain, ok := strings.Cut(row.Address, "@"); ok {
			mailDomains[strings.ToLower(domain)] = true
		}
		if row.NearQuota {
			d.NearQuotaCount++
		}
	}
	d.MailDomains = len(mailDomains)
	d.StorageValue, d.StorageUnit, _ = strings.Cut(formatBytes(used), " ")
	d.AliasCount = len(aliases)
	inboxes := map[string]bool{}
	for _, alias := range aliases {
		for _, target := range alias.Targets {
			inboxes[strings.ToLower(target)] = true
		}
	}
	d.AliasInboxes = len(inboxes)
}

// humanQuota turns a doveadm size like "1.7G" or "12M" into "1.7 GB" plus bytes, leaving anything else as is
func humanQuota(s string) (string, float64) {
	units := map[byte]float64{'B': 1, 'K': 1 << 10, 'M': 1 << 20, 'G': 1 << 30, 'T': 1 << 40}
	num, mult := strings.TrimSuffix(strings.ToUpper(s), "B"), 1.0
	if num != "" {
		if m, ok := units[num[len(num)-1]]; ok {
			num, mult = num[:len(num)-1], m
		}
	}
	v, err := strconv.ParseFloat(num, 64)
	if err != nil || v < 0 {
		return s, 0
	}
	return formatBytes(v * mult), v * mult
}

func formatBytes(b float64) string {
	switch {
	case b >= 1<<40:
		return oneDecimal(b/(1<<40)) + " TB"
	case b >= 1<<30:
		return oneDecimal(b/(1<<30)) + " GB"
	case b >= 1<<20:
		return strconv.Itoa(int(b/(1<<20)+0.5)) + " MB"
	case b >= 1<<10:
		return strconv.Itoa(int(b/(1<<10)+0.5)) + " KB"
	}
	return "0 MB"
}

func oneDecimal(v float64) string {
	return strings.TrimSuffix(strconv.FormatFloat(v, 'f', 1, 64), ".0")
}

// emailQuotaToast mirrors accounts.html's ns.quota_issues + quota_toast_issues 1/2/multiple pluralization branch (percent_val > 80)
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

func renderAccountsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, currentEmailsList []string, domains []appctx.Domain, emailLimit int, aliases func() []AliasEntry, connect AccountsConnect) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Emails")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	rows := make([]EmailListRow, len(currentEmailsList))
	for i, line := range currentEmailsList {
		rows[i] = parseEmailListRow(line)
	}
	annotateEmailRows(rows)
	for i := range rows {
		if connect.Created != "" && strings.EqualFold(rows[i].Address, connect.Created) {
			rows[i].IsNew, connect.Created = true, rows[i].Address
		}
	}
	toastID, toastMsg := emailQuotaToast(rows)
	layout.BulkActions = accountsBulkActions(layout.T)
	data := AccountsPageData{LayoutData: layout, Rows: rows, TotalCount: len(currentEmailsList), QuotaToastID: toastID, QuotaToastMsg: toastMsg, Connect: connect}
	var aliasList []AliasEntry
	if layout.UserAllowed["email_aliases"] {
		aliasList = aliases()
	}
	data.fillAccountsSummary(emailLimit, aliasList)
	if err := accountsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - accounts template render error: %v", err)
	}
}

// new.html

// NewEmailPageData is emails/new.html's template context.
type NewEmailPageData struct {
	web.LayoutData
	Domains              []appctx.Domain
	MaxEmailQuotaNumeric float64
	AllocatedUnit        string
	MaxQuotaBytes        float64 // 0 means the plan has no per-mailbox cap
	EmailLimit           int
	EmailCount           int
	Taken                []string
	Domain               string // preselected domain, from ?domain=
	Username             string
	MailHost             string
	MailSecure           bool
	Wait                 bool // opencli gets --wait on plans with a quota cap
	Presets              []QuotaPreset
	MaxQuotaLabel        string
	Full                 bool
	UsagePercent         int
}

// QuotaPreset is one of the size buttons on the create form, Value and Unit go out as the gb and format fields
type QuotaPreset struct {
	Value, Unit, Label, Note string
}

// quotaPresets picks up to four common sizes under the plan's cap, then the cap itself, or Unlimited when there's none
func quotaPresets(maxNumeric float64, maxUnit string, maxBytes float64) []QuotaPreset {
	sizes := []QuotaPreset{{Value: "256", Unit: "M"}, {Value: "512", Unit: "M"}, {Value: "1", Unit: "G"}, {Value: "2", Unit: "G"}, {Value: "5", Unit: "G"}, {Value: "10", Unit: "G"}}
	var out []QuotaPreset
	for _, p := range sizes {
		n, _ := strconv.ParseFloat(p.Value, 64)
		b := n * quotaUnitMultipliers[p.Unit[0]]
		if maxBytes > 0 && b >= maxBytes {
			break
		}
		p.Label = formatBytes(b)
		out = append(out, p)
	}
	if maxBytes == 0 {
		return append(out[:4], QuotaPreset{Value: "0", Unit: "G", Label: "Unlimited"})
	}
	if len(out) > 4 {
		out = out[len(out)-4:]
	}
	value, unit := trimFloat(maxNumeric), maxUnit
	switch unit {
	case "K":
		unit = "k"
	case "P":
		value, unit = trimFloat(maxNumeric*1024), "T"
	}
	return append(out, QuotaPreset{Value: value, Unit: unit, Label: formatBytes(maxBytes), Note: "Plan maximum"})
}

// fill works out the presets, plan-limit state and preselected domain from the raw plan values
func (d *NewEmailPageData) fill() {
	d.Presets = quotaPresets(d.MaxEmailQuotaNumeric, d.AllocatedUnit, d.MaxQuotaBytes)
	if d.MaxQuotaBytes > 0 {
		d.MaxQuotaLabel = formatBytes(d.MaxQuotaBytes)
	}
	if d.EmailLimit > 0 {
		d.Full = d.EmailCount >= d.EmailLimit
		d.UsagePercent = min(100, d.EmailCount*100/d.EmailLimit)
		if d.UsagePercent == 0 && d.EmailCount > 0 {
			d.UsagePercent = 1
		}
	}
	found := false
	for _, dom := range d.Domains {
		found = found || dom.DomainURL == d.Domain
	}
	if !found && len(d.Domains) > 0 {
		d.Domain = d.Domains[0].DomainURL
	}
	if len(d.Username) > 60 {
		d.Username = ""
	}
	for i, addr := range d.Taken {
		d.Taken[i] = strings.ToLower(addr)
	}
}

func renderNewEmailPage(a *appctx.App, w http.ResponseWriter, r *http.Request, data NewEmailPageData) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Create Account")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data.LayoutData = layout
	data.fill()
	if err := newEmailPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - new template render error: %v", err)
	}
}

// single_account.html

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

// delete.html

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

// info.html

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

// aliases.html

// AliasesPageData is emails/aliases.html's template context.
type AliasesPageData struct {
	web.LayoutData
	AliasList []AliasEntry
	Domains   []appctx.Domain
	// ?new=1 opens the create sheet, a rejected submit comes back with its domain and username
	NewOpen            bool
	NewDomain, NewUser string
}

func renderAliasesPage(a *appctx.App, w http.ResponseWriter, r *http.Request, aliasList []AliasEntry, domains []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Aliases")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.BulkActions = aliasesBulkActions(layout.T)
	data := AliasesPageData{LayoutData: layout, AliasList: aliasList, Domains: domains, NewOpen: r.URL.Query().Get("new") == "1", NewUser: r.URL.Query().Get("username")}
	for _, d := range domains {
		if d.DomainURL == r.URL.Query().Get("domain") {
			data.NewDomain = d.DomainURL
		}
	}
	if err := aliasesPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - aliases template render error: %v", err)
	}
}

// alias_detail.html

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

// alias_new.html, only the classic theme still has a separate create page

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

// alias_delete.html

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

// default_address.html

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

// deliverability.html / deliverability_domain.html

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

// DeliverabilityDomainPageData is emails/deliverability_domain.html's template context.
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

// filter.html

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

// import.html / confirm_import.html

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

// MigratePageData is emails/migrate.html's template context.
type MigratePageData struct {
	web.LayoutData
	Accounts []string
	Status   MigrateStatus
}

func renderMigratePage(a *appctx.App, w http.ResponseWriter, r *http.Request, accounts []string, st MigrateStatus) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Migrate Emails")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := MigratePageData{LayoutData: layout, Accounts: accounts, Status: st}
	if err := migratePage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("EMAILS - migrate template render error: %v", err)
	}
}
