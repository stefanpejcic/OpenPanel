package backups

import (
	"os"
	"path/filepath"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

// envFixture points the account home, core/users dir and admin default at a temp dir and returns it
func envFixture(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	home, core := filepath.Join(tmp, "home"), filepath.Join(tmp, "core")
	for _, d := range []string{filepath.Join(home, "acct"), filepath.Join(core, "acct")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for name, bucket := range map[string]string{"admin.env": "admin-bucket", "other.env": "other-bucket"} {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte("AWS_S3_BUCKET_NAME="+bucket+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	oldHome, oldCore, oldEnv := homeRoot, coreUsersRoot, adminBackupEnvPath
	homeRoot, coreUsersRoot, adminBackupEnvPath = home+"/", core+"/", filepath.Join(tmp, "admin.env")
	t.Cleanup(tenantfs.SetHomeRootForTest(home))
	t.Cleanup(func() { homeRoot, coreUsersRoot, adminBackupEnvPath = oldHome, oldCore, oldEnv })
	return tmp
}

func TestReadEnvFileFollowsOnlyTheAdminLink(t *testing.T) {
	cases := []struct {
		name    string
		managed bool
		link    string
		want    string
	}{
		{"own file", false, "", "own-bucket"},
		{"admin link on an admin-managed account", true, "admin.env", "admin-bucket"},
		{"admin link without the marker", false, "admin.env", ""},
		{"other link on an admin-managed account", true, "other.env", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tmp := envFixture(t)
			envPath := filepath.Join(tmp, "home", "acct", "backup.env")
			if c.link == "" {
				if err := os.WriteFile(envPath, []byte("AWS_S3_BUCKET_NAME='own-bucket'\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Symlink(filepath.Join(tmp, c.link), envPath); err != nil {
				t.Fatal(err)
			}
			if c.managed {
				if err := os.WriteFile(filepath.Join(tmp, "core", "acct", "admin.backups"), nil, 0o644); err != nil {
					t.Fatal(err)
				}
			}

			config, _ := readBackupEnv("acct")
			if got := config["AWS_S3_BUCKET_NAME"]; got != c.want {
				t.Errorf("readBackupEnv bucket = %q, want %q", got, c.want)
			}
			wantTarget := ""
			if c.want != "" {
				wantTarget = "s3"
			}
			if got := ActiveBackupTarget("acct"); got != wantTarget {
				t.Errorf("ActiveBackupTarget = %q, want %q", got, wantTarget)
			}
		})
	}
}
