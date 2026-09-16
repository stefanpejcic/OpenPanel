package plugins

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// CaptchaWidget is what the login page needs to render the active captcha provider's widget - Provider is "" when no captcha plugin is installed or none is configured, which callers treat as "don't show a captcha".
type CaptchaWidget struct {
	Provider  string `json:"provider"`
	FieldName string `json:"field_name"`
	SiteKey   string `json:"site_key"`
}

// captchaBinaryPath returns the captcha plugin's binary path under baseDir if it's installed, or "" otherwise - the plugin contract puts a plugin's own binary at <baseDir>/<name>/<name> (see the captcha plugin's README).
func captchaBinaryPath(baseDir string) string {
	path := filepath.Join(baseDir, "captcha", "captcha")
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return path
	}
	return ""
}

// GetCaptchaWidget calls the installed captcha plugin's "widget" subcommand. Any failure (not installed, exec error, bad output) is treated as "no captcha configured" - rendering the login page is never blocked by a broken plugin.
func GetCaptchaWidget(ctx context.Context, baseDir string) CaptchaWidget {
	bin := captchaBinaryPath(baseDir)
	if bin == "" {
		return CaptchaWidget{}
	}

	cctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(cctx, bin, "widget").Output()
	if err != nil {
		log.Printf("captcha plugin: widget call failed: %v", err)
		return CaptchaWidget{}
	}

	var widget CaptchaWidget
	if err := json.Unmarshal(out, &widget); err != nil {
		log.Printf("captcha plugin: could not parse widget output: %v", err)
		return CaptchaWidget{}
	}
	return widget
}

// VerifyCaptcha calls the installed captcha plugin's "verify" subcommand and trusts its {"success": bool} JSON output regardless of exit code (the plugin exits non-zero exactly when success is false, but the JSON is the source of truth here). Returns true when no plugin is installed - nothing to enforce - or when the plugin's output can't be parsed at all, since an infrastructure fault in an optional bot-mitigation plugin shouldn't lock every customer out of the login page.
func VerifyCaptcha(ctx context.Context, baseDir, token, ip string) bool {
	bin := captchaBinaryPath(baseDir)
	if bin == "" {
		return true
	}

	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	out, _ := exec.CommandContext(cctx, bin, "verify", "--token="+token, "--ip="+ip).Output()

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		log.Printf("captcha plugin: could not parse verify output: %v", err)
		return true
	}
	return result.Success
}
