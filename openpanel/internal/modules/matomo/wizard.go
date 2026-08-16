package matomo

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// matomoWizardParams is everything runMatomoInstallWizard needs to drive
// Matomo's browser installation wizard end to end.
type matomoWizardParams struct {
	SiteURL       string // e.g. "https://example.com/matomo/" - must already resolve to the freshly-extracted docroot
	DBHost        string // podman network hostname of the MySQL/MariaDB container, e.g. "mariadb"
	DBName        string
	DBUser        string
	DBPassword    string
	DBSchema      string // "Mysql" or "Mariadb"
	AdminLogin    string
	AdminPassword string
	AdminEmail    string
	SiteName      string
}

// runMatomoInstallWizard drives plugins/Installation/Controller.php's step
// sequence (databaseSetup -> tablesCreation -> setupSuperUser ->
// firstWebsiteSetup -> finished) as a plain HTTP request sequence -
// field-for-field matched against that controller/its Form classes' source
// (confirmed against a real 5.12.0 release) and verified live end to end,
// including a subsequent nonce-based login. Matomo ships no non-interactive
// CLI installer, so this is the only mechanism that reaches a fully
// installed, ready-to-log-into state without reimplementing Matomo's
// internal DB-schema/superuser-creation logic by hand.
func runMatomoInstallWizard(ctx context.Context, p matomoWizardParams) (string, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 90 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	base := strings.TrimSuffix(p.SiteURL, "/")

	// Step 1: hit the base URL once first so Matomo's welcome step runs
	// (sets the initial session cookie / selected language) before the
	// database form is submitted - mirrors what a real browser does.
	if _, _, err := doGet(ctx, client, base+"/index.php"); err != nil {
		return "", fmt.Errorf("could not reach installer welcome step: %w", err)
	}

	// Step 2: Database Set-up.
	dbStepURL := base + "/index.php?module=Installation&action=databaseSetup"
	finalURL, body, err := doPost(ctx, client, dbStepURL, url.Values{
		"host":          {p.DBHost},
		"username":      {p.DBUser},
		"password":      {p.DBPassword},
		"dbname":        {p.DBName},
		"tables_prefix": {"matomo_"},
		"adapter":       {`PDO\MYSQL`},
		"schema":        {p.DBSchema},
		"type":          {"InnoDB"},
		"submit":        {"Next »"},
	})
	if err != nil {
		return "", fmt.Errorf("database setup step failed: %w", err)
	}
	if !strings.Contains(finalURL, "action=tablesCreation") {
		return "", fmt.Errorf("database setup step did not advance (still on %s): %s", finalURL, extractInstallerError(body))
	}

	// Step 3: Table Creation runs synchronously as part of the GET the
	// databaseSetup redirect already followed above (Controller.tablesCreation()
	// creates every table inline before rendering) - just verify it
	// actually reports success before continuing.
	if !strings.Contains(body, `name="setupSuperUser"`) && !strings.Contains(body, "action=setupSuperUser") {
		return "", fmt.Errorf("table creation step did not report success: %s", extractInstallerError(body))
	}

	// Step 4: Super User setup.
	suStepURL := base + "/index.php?module=Installation&action=setupSuperUser"
	finalURL, body, err = doPost(ctx, client, suStepURL, url.Values{
		"login":        {p.AdminLogin},
		"password":     {p.AdminPassword},
		"password_bis": {p.AdminPassword},
		"email":        {p.AdminEmail},
		"submit":       {"Next »"},
	})
	if err != nil {
		return "", fmt.Errorf("super user setup step failed: %w", err)
	}
	if !strings.Contains(finalURL, "action=firstWebsiteSetup") {
		return "", fmt.Errorf("super user setup step did not advance (still on %s): %s", finalURL, extractInstallerError(body))
	}

	// Step 5: First website setup.
	wsStepURL := base + "/index.php?module=Installation&action=firstWebsiteSetup"
	finalURL, body, err = doPost(ctx, client, wsStepURL, url.Values{
		"siteName":  {p.SiteName},
		"url":       {p.SiteURL},
		"timezone":  {"UTC"},
		"ecommerce": {"0"},
		"submit":    {"Next »"},
	})
	if err != nil {
		return "", fmt.Errorf("first website setup step failed: %w", err)
	}
	if !strings.Contains(finalURL, "action=trackingCode") {
		return "", fmt.Errorf("first website setup step did not advance (still on %s): %s", finalURL, extractInstallerError(body))
	}
	siteIDMatch := regexp.MustCompile(`site_idSite=(\d+)`).FindStringSubmatch(finalURL)
	siteID := "1"
	if siteIDMatch != nil {
		siteID = siteIDMatch[1]
	}

	// Step 6: Finish - writes installation_in_progress=0 into config.ini.php,
	// completing the install.
	finishStepURL := base + "/index.php?module=Installation&action=finished&site_idSite=" + siteID + "&site_name=" + url.QueryEscape(p.SiteName)
	finalURL, body, err = doPost(ctx, client, finishStepURL, url.Values{
		"setup_geoip2": {"1"},
		"anonymise_ip": {"1"},
		"submit":       {"Continue to Matomo »"},
	})
	if err != nil {
		return "", fmt.Errorf("finish step failed: %w", err)
	}
	if strings.Contains(finalURL, "module=Installation") {
		return "", fmt.Errorf("finish step did not complete (still on %s): %s", finalURL, extractInstallerError(body))
	}

	return siteID, nil
}

func doGet(ctx context.Context, client *http.Client, u string) (finalURL, body string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.Request.URL.String(), string(b), nil
}

func doPost(ctx context.Context, client *http.Client, u string, form url.Values) (finalURL, body string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.Request.URL.String(), string(b), nil
}

var installerErrorRE = regexp.MustCompile(`(?s)class="error"[^>]*>(.*?)</`)

// extractInstallerError pulls a human-readable snippet out of an installer
// step's HTML when it didn't advance, so a failed install.go run surfaces
// something more useful than "step did not advance".
func extractInstallerError(body string) string {
	if m := installerErrorRE.FindStringSubmatch(body); m != nil {
		text := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(m[1], "")
		text = strings.TrimSpace(text)
		if text != "" {
			return text
		}
	}
	if len(body) > 300 {
		return body[:300]
	}
	return body
}
