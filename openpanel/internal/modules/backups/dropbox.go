package backups

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// dropboxStore is the remoteStore implementation for the "dropbox"
// backup.env section, talking to the Dropbox API v2 directly over
// net/http/json - no SDK needed for three calls (token refresh, list
// folder, download).
type dropboxStore struct {
	config map[string]string
}

// remotePath normalizes DROPBOX_REMOTE_PATH into what the Dropbox API
// expects: "" for the app's root, or a leading-"/" path otherwise (a bare
// relative path like "backups" is rejected by the API).
func (s *dropboxStore) remotePath() string {
	p := strings.Trim(strings.TrimSpace(s.config["DROPBOX_REMOTE_PATH"]), "/")
	if p == "" {
		return ""
	}
	return "/" + p
}

// accessToken exchanges DROPBOX_REFRESH_TOKEN for a short-lived access
// token via the standard OAuth2 refresh flow - the same one the "backup"
// container itself performs with these exact three env vars.
func (s *dropboxStore) accessToken(ctx context.Context) (string, error) {
	refreshToken := s.config["DROPBOX_REFRESH_TOKEN"]
	appKey := s.config["DROPBOX_APP_KEY"]
	appSecret := s.config["DROPBOX_APP_SECRET"]
	if refreshToken == "" || appKey == "" || appSecret == "" {
		return "", errors.New("DROPBOX_REFRESH_TOKEN/DROPBOX_APP_KEY/DROPBOX_APP_SECRET are not configured")
	}

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {appKey},
		"client_secret": {appSecret},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.dropboxapi.com/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("dropbox oauth2/token: unexpected status %s: %s", resp.Status, string(body))
	}

	var payload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	if payload.AccessToken == "" {
		return "", errors.New("dropbox oauth2/token: empty access_token in response")
	}
	return payload.AccessToken, nil
}

type dropboxEntry struct {
	Tag  string `json:".tag"`
	Name string `json:"name"`
}

type dropboxListFolderResponse struct {
	Entries []dropboxEntry `json:"entries"`
	HasMore bool           `json:"has_more"`
	Cursor  string         `json:"cursor"`
}

func (s *dropboxStore) List(ctx context.Context) ([]string, error) {
	token, err := s.accessToken(ctx)
	if err != nil {
		return nil, err
	}

	var names []string

	body, _ := json.Marshal(map[string]any{"path": s.remotePath()})
	endpoint := "https://api.dropboxapi.com/2/files/list_folder"
	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("dropbox files/list_folder: unexpected status %s: %s", resp.Status, string(respBody))
		}

		var page dropboxListFolderResponse
		if err := json.Unmarshal(respBody, &page); err != nil {
			return nil, err
		}
		for _, entry := range page.Entries {
			if entry.Tag == "file" {
				names = append(names, entry.Name)
			}
		}

		if !page.HasMore {
			break
		}
		endpoint = "https://api.dropboxapi.com/2/files/list_folder/continue"
		body, _ = json.Marshal(map[string]any{"cursor": page.Cursor})
	}

	return names, nil
}

func (s *dropboxStore) Fetch(ctx context.Context, filename string) (localPath string, cleanup func(), err error) {
	token, err := s.accessToken(ctx)
	if err != nil {
		return "", nil, err
	}

	remote := s.remotePath()
	if remote == "" {
		remote = "/" + filename
	} else {
		remote = remote + "/" + filename
	}
	apiArg, _ := json.Marshal(map[string]string{"path": remote})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://content.dropboxapi.com/2/files/download", nil)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Dropbox-API-Arg", string(apiArg))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("dropbox files/download: unexpected status %s: %s", resp.Status, string(body))
	}

	tmpDir, err := os.MkdirTemp("", "backup_restore_")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { _ = os.RemoveAll(tmpDir) }

	localPath = filepath.Join(tmpDir, filename)
	localFile, err := os.Create(localPath)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	defer localFile.Close()

	if _, err := io.Copy(localFile, resp.Body); err != nil {
		cleanup()
		return "", nil, err
	}

	return localPath, cleanup, nil
}

func (s *dropboxStore) Classify(ctx context.Context, filename string) (BackupInfo, error) {
	return classifyViaDownload(ctx, s, filename)
}
