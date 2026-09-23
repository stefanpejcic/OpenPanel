package backups

import (
	"fmt"
	"net/http"
	"os"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// adminManagedMarkerPath is the per-account switch that hands backup destination/settings management to the admin - the account can still list and restore backups but can't view or change where/how they're stored. Lives under core/users like krompir.lock and every other per-account marker, not /home, since it's provisioned via OpenAdmin's /etc/openpanel/skeleton/ template (copied into core/users/<username>/ by opencli's user-add) rather than backup.env's direct /home copy
func adminManagedMarkerPath(userContext string) string {
	return fmt.Sprintf("/etc/openpanel/openpanel/core/users/%s/admin.backups", userContext)
}

// AdminManagedBackups reports whether backup destinations/settings are admin-managed for this account: locked whenever adminManagedMarkerPath exists
func AdminManagedBackups(userContext string) bool {
	_, err := os.Stat(adminManagedMarkerPath(userContext))
	return err == nil
}

func init() {
	web.BackupsAdminManaged = AdminManagedBackups
}

// respondBackupsAdminManaged is what handleBackupSettings/handleBackupTarget return instead of their normal view/update when AdminManagedBackups is true, for both the HTML page and the JSON API
func respondBackupsAdminManaged(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	msg := "Backup destinations and settings are managed by your administrator. You can still list and restore backups."
	if r.URL.Query().Get("output") == "json" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": msg})
		return
	}
	flashAndRedirect(a, w, r, "error", msg, "/backups")
}
