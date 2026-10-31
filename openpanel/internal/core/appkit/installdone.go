package appkit

// InstallInfo is what the install page shows once the app is up, empty fields are just left out
type InstallInfo struct {
	AdminPath     string
	AdminUsername string
	AdminEmail    string
	DBName        string
	DBUser        string
}

// InstallDoneEvent builds the last NDJSON event of an install, the page keeps these on screen instead of redirecting
func InstallDoneEvent(status, domain string, info InstallInfo) map[string]any {
	siteURL := "https://" + domain
	ev := map[string]any{"status": status, "domain": domain, "site_url": siteURL}
	add := func(k, v string) {
		if v != "" {
			ev[k] = v
		}
	}
	if info.AdminPath != "" {
		add("admin_url", siteURL+info.AdminPath)
	}
	add("admin_username", info.AdminUsername)
	add("admin_email", info.AdminEmail)
	add("db_name", info.DBName)
	add("db_user", info.DBUser)
	return ev
}
