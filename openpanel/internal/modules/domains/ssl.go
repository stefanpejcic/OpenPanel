package domains

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// triggerSSLGeneration reloads Caddy and visits the domain over HTTPS to
// trigger on-demand certificate issuance.
func triggerSSLGeneration(ctx context.Context, domainName string) (bool, string) {
	if err := exec.CommandContext(ctx, "podman", "exec", "caddy", "caddy", "reload", "--config", "/etc/caddy/Caddyfile").Run(); err != nil {
		return false, "Failed to reload Caddy: " + err.Error()
	}

	curlCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	_ = exec.CommandContext(curlCtx, "curl", "-s", "-k", "-o", "/dev/null", "--max-time", "20", "https://"+domainName).Run()

	leCert := "/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/" + domainName + "/" + domainName + ".crt"
	if _, err := os.Stat(leCert); err == nil {
		return true, "SSL certificate generated successfully for " + domainName + "."
	}
	return false, "Certificate was not generated yet. It may take a moment, please try again shortly."
}

// handleDomainCustomSSL installs a custom SSL certificate for a domain.
func handleDomainCustomSSL(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domainsList, _ := a.AllDomainsForUser(ctx, userID)

	var domainName string
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		domainName = r.Form.Get("domain_name")
		if domainName == "" {
			flashAndRedirect(a, w, r, "error", "Invalid request. Domain name must be provided.", "/domains/ssl")
			return
		}
	} else {
		domainName = r.URL.Query().Get("domain_name")
		if domainName == "" {
			renderSSLPage(a, w, r, "", "", "", domainsList)
			return
		}
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		currentUsername, userContext, _ := injected(a, ctx, userID)
		action := r.Form.Get("action")

		switch action {
		case "custom":
			handleCustomSSLUpload(a, w, r, domainName, userContext, currentUsername)
			return

		case "autossl":
			out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-ssl", domainName, "auto").CombinedOutput()
			if cmdErr == nil {
				flashAndRedirect(a, w, r, "success", strings.TrimSpace(string(out)), "/domains/ssl?domain_name="+domainName)
				_ = logger.RecordUserAction(a.Config, currentUsername, "enabled AutoSSL for "+domainName, reqip.ClientIP(r))
			} else {
				flashAndRedirect(a, w, r, "error", strings.TrimSpace(string(out)), "/domains/ssl?domain_name="+domainName)
			}
			return

		case "generate":
			ok, message := triggerSSLGeneration(ctx, domainName)
			if ok {
				flashAndRedirect(a, w, r, "success", message, "/domains/ssl?domain_name="+domainName)
				_ = logger.RecordUserAction(a.Config, currentUsername, "generated SSL certificate for "+domainName, reqip.ClientIP(r))
			} else {
				flashAndRedirect(a, w, r, "error", message, "/domains/ssl?domain_name="+domainName)
			}
			return

		case "switch_and_generate":
			out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-ssl", domainName, "auto").CombinedOutput()
			if cmdErr != nil {
				flashAndRedirect(a, w, r, "error", strings.TrimSpace(string(out)), "/domains/ssl?domain_name="+domainName)
				return
			}
			ok, message := triggerSSLGeneration(ctx, domainName)
			category := "success"
			if !ok {
				category = "error"
			}
			flashAndRedirect(a, w, r, category, message, "/domains/ssl?domain_name="+domainName)
			_ = logger.RecordUserAction(a.Config, currentUsername, "switched "+domainName+" to AutoSSL and triggered certificate generation", reqip.ClientIP(r))
			return

		default:
			flashSess(a, w, r, "error", "Invalid action! Only AutoSSL or Custom are available.")
		}
	}

	currentSetting, keys := "", ""
	out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-ssl", domainName, "status").CombinedOutput()
	if cmdErr == nil {
		currentSetting = strings.ToLower(strings.TrimSpace(string(out)))
		if infoOut, infoErr := exec.CommandContext(ctx, "opencli", "domains-ssl", domainName, "info").CombinedOutput(); infoErr == nil {
			keys = strings.TrimSpace(string(infoOut))
		}
	} else {
		flashSess(a, w, r, "error", strings.TrimSpace(string(out)))
	}

	renderSSLPage(a, w, r, domainName, currentSetting, keys, domainsList)
}

func handleCustomSSLUpload(a *appctx.App, w http.ResponseWriter, r *http.Request, domainName, userContext, currentUsername string) {
	ctx := r.Context()
	certificate := strings.TrimSpace(r.Form.Get("certificate"))
	privateKey := strings.TrimSpace(r.Form.Get("private_key"))

	if certificate == "" || privateKey == "" {
		flashAndRedirect(a, w, r, "error", "Certificate and private key are required.", "/domains/ssl?domain_name="+domainName)
		return
	}
	if !strings.Contains(certificate, "BEGIN CERTIFICATE") {
		flashAndRedirect(a, w, r, "error", "Invalid certificate.", "/domains/ssl?domain_name="+domainName)
		return
	}
	if !strings.Contains(privateKey, "BEGIN") || !strings.Contains(privateKey, "PRIVATE KEY") {
		flashAndRedirect(a, w, r, "error", "Invalid private key.", "/domains/ssl?domain_name="+domainName)
		return
	}

	dataDir := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data"
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		flashAndRedirect(a, w, r, "error", err.Error(), "/domains/ssl?domain_name="+domainName)
		return
	}

	// The domains-ssl script strips "/var/www/html/" off
	// displayCertPath/displayKeyPath and looks for the file under dataDir
	// using what's left - so the on-disk filename must match the "_tmp"
	// name passed to opencli, not a bare "{domain}.crt", or the script can
	// never find it.
	certPath := filepath.Join(dataDir, domainName+"_tmp.crt")
	keyPath := filepath.Join(dataDir, domainName+"_tmp.key")
	displayCertPath := "/var/www/html/" + domainName + "_tmp.crt"
	displayKeyPath := "/var/www/html/" + domainName + "_tmp.key"

	_ = os.WriteFile(certPath, []byte(certificate), 0o644)
	_ = os.WriteFile(keyPath, []byte(privateKey), 0o644)

	out, cmdErr := exec.CommandContext(ctx, "opencli", "domains-ssl", domainName, "custom", displayCertPath, displayKeyPath).CombinedOutput()
	_ = os.Remove(certPath)
	_ = os.Remove(keyPath)

	if cmdErr == nil {
		flashAndRedirect(a, w, r, "success", strings.TrimSpace(string(out)), "/domains/ssl?domain_name="+domainName)
		_ = logger.RecordUserAction(a.Config, currentUsername, "configured custom SSL for "+domainName, reqip.ClientIP(r))
	} else {
		flashAndRedirect(a, w, r, "error", strings.TrimSpace(string(out)), "/domains/ssl?domain_name="+domainName)
	}
}

// handleGetTLSAHash computes the TLSA record hash for a domain's certificate.
func handleGetTLSAHash(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	domain := r.PathValue("domain")

	if !a.CheckDomainBelongsToUser(ctx, userID, domain) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	customCert := "/etc/openpanel/caddy/ssl/custom/" + domain + "/" + domain + "/fullchain.pem"
	leCert := "/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/" + domain + "/" + domain + ".crt"

	certPath, certType := "", ""
	if _, err := os.Stat(customCert); err == nil {
		certPath, certType = customCert, "custom"
	} else if _, err := os.Stat(leCert); err == nil {
		certPath, certType = leCert, "letsencrypt"
	} else {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "No SSL certificate found for this domain."})
		return
	}

	certData, err := os.ReadFile(certPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to decode PEM certificate"})
		return
	}
	cert, parseErr := x509.ParseCertificate(block.Bytes)
	if parseErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": parseErr.Error()})
		return
	}

	pubKeyDER, pubErr := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if pubErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": pubErr.Error()})
		return
	}
	hash311 := sha256.Sum256(pubKeyDER)
	hash301 := sha256.Sum256(cert.Raw)

	writeJSON(w, http.StatusOK, map[string]string{
		"hash_311":  hex.EncodeToString(hash311[:]),
		"hash_301":  hex.EncodeToString(hash301[:]),
		"cert_path": certPath,
		"type":      certType,
	})
}
