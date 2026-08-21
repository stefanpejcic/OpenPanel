package web

import (
	"net/http"

	"github.com/gorilla/csrf"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
)

// BuildLayoutData assembles the shared app-shell data every authenticated
// page needs (nav, flashes, branding, translator, ...), factored out so
// each module (dashboard, docker, ...) doesn't reimplement it. Returns
// the injected user-context map too, since callers need fields like
// current_username/context/hosting_plan for their own page-specific data.
func BuildLayoutData(a *appctx.App, w http.ResponseWriter, r *http.Request, title string) (LayoutData, map[string]any, error) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		return LayoutData{}, nil, err
	}

	sess, _ := a.Sessions.Get(r, session.CookieName)

	userAllowedSlice, _ := injected["user_allowed"].([]string)
	userAllowed := make(map[string]bool, len(userAllowedSlice))
	for _, m := range userAllowedSlice {
		userAllowed[m] = true
	}

	currentUsername, _ := injected["current_username"].(string)
	sessionLocale, _ := sess.Values["locale"].(string)
	userContext, _ := injected["context"].(string)
	userLocale := i18n.UserLocale(userContext)
	locale := a.I18n.ResolveLocale(ctx, sessionLocale, userLocale, r.Header.Get("Accept-Language"))
	t := a.I18n.Translator(locale)

	adminPort, _ := sess.Values["admin_port"].(string)
	if adminPort == "" {
		adminPort = "2087"
	}
	_, impersonating := sess.Values["impersonate"]

	hostingPlanName, _ := injected["hosting_plan_name"].(string)
	avatarType, _ := injected["avatar_type"].(string)
	gravatarURL, _ := injected["gravatar_image_url"].(string)
	panelVersion, _ := injected["panel_version"].(string)
	panelDir, _ := injected["panel_dir"].(string)
	isEnterprise, _ := injected["is_enterprise"].(bool)

	// A reseller can set their own logo, applied instead of the global
	// default throughout the app for every account they own (see
	// RESELLER_LOGO_URL in opencli's user-add/admin-update-branding
	// flow) -- except the login page, which always uses the global
	// default regardless (built via LoginPageData, not this function).
	logo := a.Config.Get("logo", "")
	if resellerLogo := webserver.GetEnvFileValue(userContext, "RESELLER_LOGO_URL"); resellerLogo != "" {
		logo = resellerLogo
	}

	layout := LayoutData{
		Title:            title,
		BrandName:        a.Config.Get("brand_name", ""),
		Logo:             logo,
		Favicon:          a.Config.Get("favicon", ""),
		CSRFToken:        csrf.Token(r),
		PanelDir:         panelDir,
		FoundABugLink:    a.Config.Get("found_a_bug_link", ""),
		PanelVersion:     panelVersion,
		CustomPlugins:    len(a.PluginNames) > 0,
		CustomCSS:        a.CustomCSS,
		CustomJS:         true, // the custom-JS <script> tag is always emitted, whether or not custom.js has real content (see base.html)
		NavGroups:        BuildSidebarNav(userAllowed, NavPath(r)),
		UserAllowed:      userAllowed,
		UserAllowedJSON:  UserAllowedList(userAllowed),
		IsEnterprise:     isEnterprise,
		CurrentUsername:  currentUsername,
		HostingPlanName:  hostingPlanName,
		AvatarType:       avatarType,
		GravatarURL:      gravatarURL,
		RequestPath:      r.URL.Path,
		Flashes:          BuildFlashDisplay(flash.Pop(a.Sessions, w, r, sess)),
		Impersonating:    impersonating,
		AdminPort:        adminPort,
		PasswordStrength: validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50),
		T:                t,
	}

	return layout, injected, nil
}
