package ftp

import (
	"log"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
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

var ftpAccountsPage = web.MustLoadPage(append(append([]string{}, pageFiles...), "files/ftp.html")...)
var ftpConnectionsPage = web.MustLoadPage(append(append([]string{}, pageFiles...), "files/ftp_connections.html")...)
var ftpNewPage = web.MustLoadPage(append(append([]string{}, pageFiles...), "files/ftp_new.html")...)
var ftpPasswordPage = web.MustLoadPage(append(append([]string{}, pageFiles...), "files/ftp_password.html")...)
var ftpPathPage = web.MustLoadPage(append(append([]string{}, pageFiles...), "files/ftp_path.html")...)

// FTPAccountsPageData is ftp.html's template context.
type FTPAccountsPageData struct {
	web.LayoutData
	ServerIP, DedicatedIP string
	FTPHost, FTPPort      string
	Accounts              []Account
}

func renderFTPAccountsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, serverIP, dedicatedIP, ftpHost, ftpPort string, accounts []Account) {
	layout, _, err := web.BuildLayoutData(a, w, r, "FTP Accounts")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := FTPAccountsPageData{LayoutData: layout, ServerIP: serverIP, DedicatedIP: dedicatedIP, FTPHost: ftpHost, FTPPort: ftpPort, Accounts: accounts}
	if err := ftpAccountsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("FTP - accounts template render error: %v", err)
	}
}

// FTPConnectionsPageData is ftp_connections.html's template context.
type FTPConnectionsPageData struct {
	web.LayoutData
	ConnectionsOutput string
	HasConnections    bool
}

func renderFTPConnectionsPage(a *appctx.App, w http.ResponseWriter, r *http.Request, output string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "FTP Connections")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := FTPConnectionsPageData{LayoutData: layout, ConnectionsOutput: output, HasConnections: strings.TrimSpace(output) != ""}
	if err := ftpConnectionsPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("FTP - connections template render error: %v", err)
	}
}

// FTPNewPageData is ftp_new.html's template context.
type FTPNewPageData struct {
	web.LayoutData
	Domains []domainOption
}

func renderNewFTPAccountPage(a *appctx.App, w http.ResponseWriter, r *http.Request, list []appctx.Domain) {
	layout, _, err := web.BuildLayoutData(a, w, r, "New FTP Account")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := FTPNewPageData{LayoutData: layout, Domains: domainOptions(list)}
	if err := ftpNewPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("FTP - new account template render error: %v", err)
	}
}

// FTPUsernamePageData is ftp_password.html's/ftp_path.html's shared shape.
type FTPUsernamePageData struct {
	web.LayoutData
	Username string
}

func renderFTPPasswordPage(a *appctx.App, w http.ResponseWriter, r *http.Request, username string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Change FTP Password")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := FTPUsernamePageData{LayoutData: layout, Username: username}
	if err := ftpPasswordPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("FTP - password template render error: %v", err)
	}
}

func renderFTPPathPage(a *appctx.App, w http.ResponseWriter, r *http.Request, username string) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Change FTP Path")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := FTPUsernamePageData{LayoutData: layout, Username: username}
	if err := ftpPathPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("FTP - path template render error: %v", err)
	}
}
