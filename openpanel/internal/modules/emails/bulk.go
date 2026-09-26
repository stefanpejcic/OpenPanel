package emails

import (
	"net/http"
	"net/url"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func accountsBulkActions(t i18n.Translator) []web.BulkAction {
	return []web.BulkAction{
		{Key: "suspend_in", Label: t.Get("Suspend incoming"), Confirm: t.Get("Stop the selected accounts from receiving emails?")},
		{Key: "unsuspend_in", Label: t.Get("Unsuspend incoming"), Confirm: t.Get("Let the selected accounts receive emails again?")},
		{Key: "suspend_out", Label: t.Get("Suspend outgoing"), Confirm: t.Get("Stop the selected accounts from sending emails?")},
		{Key: "unsuspend_out", Label: t.Get("Unsuspend outgoing"), Confirm: t.Get("Let the selected accounts send emails again?")},
		{Key: "quota", Label: t.Get("Change quota"), Confirm: t.Get("New mailbox size (GB) for the selected accounts:"),
			Input: &web.BulkInput{Type: "number", Min: "0", Step: "any", Placeholder: "GB", Hint: t.Get("0 = unlimited")}},
		{Key: "password", Label: t.Get("Change password"), Confirm: t.Get("Set this password for the selected accounts:"), Input: &web.BulkInput{Type: "password", Placeholder: t.Get("New password")}},
		{Key: "delete", Label: t.Get("Delete"), Confirm: t.Get("Permanently delete the selected email accounts and all their emails? This cannot be undone."), Danger: true},
	}
}

func aliasesBulkActions(t i18n.Translator) []web.BulkAction {
	return []web.BulkAction{
		{Key: "add_target", Label: t.Get("Add destination"), Confirm: t.Get("Also deliver the selected aliases to:"), Input: &web.BulkInput{Type: "email", Placeholder: "name@example.com"}},
		{Key: "delete", Label: t.Get("Delete"), Confirm: t.Get("Delete the selected aliases?"), Danger: true},
	}
}

// accountsBulkRoute sends each account through the Manage page's POST (or DELETE) /emails/edit/{email}
func accountsBulkRoute(action, value, email string) (*web.BulkCall, web.BulkResult) {
	path := "/emails/edit/" + url.PathEscape(email)
	form := url.Values{}
	switch action {
	case "suspend_in":
		form.Set("incoming", "suspend")
	case "unsuspend_in":
		form.Set("incoming", "allow")
	case "suspend_out":
		form.Set("outgoing", "suspend")
	case "unsuspend_out":
		form.Set("outgoing", "allow")
	case "quota":
		form.Set("gb", value)
		form.Set("format", "G")
	case "password":
		form.Set("email_password", value)
	case "delete":
		return web.Call(web.BulkCall{Method: http.MethodDelete, Path: path})
	default:
		return web.Skip("Unknown bulk action.")
	}
	return web.Call(web.BulkCall{Method: http.MethodPost, Path: path, Form: form})
}

func aliasesBulkRoute(action, value, alias string) (*web.BulkCall, web.BulkResult) {
	path := "/emails/aliases/" + url.PathEscape(alias)
	switch action {
	case "add_target":
		return web.Call(web.BulkCall{Method: http.MethodPost, Path: path, Form: url.Values{"target": {value}}})
	case "delete":
		return web.Call(web.BulkCall{Method: http.MethodDelete, Path: path, JSON: map[string]any{"delete_all": true}})
	}
	return web.Skip("Unknown bulk action.")
}
