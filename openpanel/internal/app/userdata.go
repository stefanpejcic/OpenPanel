package app

import (
	"context"
	"crypto/md5" //nolint:gosec // matches gravatar's md5-of-email protocol, not used for security
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sysinfo"
)

// UserDetails holds the user/plan fields returned by
// GetUserDetailsWithPlan.
type UserDetails struct {
	Username string
	Context  string
	Email    string
	PlanID   int
	PlanName string
	Found    bool
}

// GetUserDetailsWithPlan looks up the user's details and plan, cached for
// 1h. Should be evicted early when the user changes email/username, but
// that account-management flow isn't implemented yet, so no eviction path
// exists here either.
func (a *App) GetUserDetailsWithPlan(ctx context.Context, userID int) (UserDetails, error) {
	key := "get_user_details_with_plan:" + strconv.Itoa(userID)
	return cache.Memoize(ctx, a.Cache, key, time.Hour, func() (UserDetails, error) {
		var (
			d       UserDetails
			context sql.NullString // users.server: NULL until a container/domain is provisioned
			email   sql.NullString
		)
		row := a.DB.QueryRowContext(ctx, `
			SELECT users.username, users.server, users.email, plans.id, plans.name
			FROM users
			JOIN plans ON users.plan_id = plans.id
			WHERE users.id = ?`, userID)

		err := row.Scan(&d.Username, &context, &email, &d.PlanID, &d.PlanName)
		if err == sql.ErrNoRows {
			return UserDetails{}, nil
		}
		if err != nil {
			return UserDetails{}, err
		}
		d.Context = context.String
		d.Email = email.String
		d.Found = true
		return d, nil
	})
}

// TwoFAStatus holds the (twofa_enabled, otp_secret) pair returned by
// Get2FAStatusForUser.
type TwoFAStatus struct {
	Enabled   bool
	OTPSecret string
}

// Get2FAStatusForUser looks up a user's 2FA enrollment status, cached for
// 6h. Only this single lookup lives here, for the login/dashboard 2FA-nag
// check; the rest of 2FA (enrollment, verification) belongs to its own
// later phase.
func (a *App) Get2FAStatusForUser(ctx context.Context, userID int) (TwoFAStatus, error) {
	key := "get_2fa_status_for_user:" + strconv.Itoa(userID)
	return cache.Memoize(ctx, a.Cache, key, 6*time.Hour, func() (TwoFAStatus, error) {
		var (
			enabled sql.NullBool
			secret  sql.NullString
		)
		row := a.DB.QueryRowContext(ctx, `SELECT twofa_enabled, otp_secret FROM users WHERE id = ?`, userID)
		if err := row.Scan(&enabled, &secret); err != nil {
			if err == sql.ErrNoRows {
				return TwoFAStatus{}, nil
			}
			return TwoFAStatus{}, err
		}
		return TwoFAStatus{Enabled: enabled.Bool, OTPSecret: secret.String}, nil
	})
}

// GetFeatureSetOnPlan looks up the feature set name for a user's plan,
// cached for 24h. Returns the literal string "None" (not an empty string)
// when the user or plan isn't found - callers use this to build a
// deliberately nonexistent
// /etc/openpanel/openpanel/features/None.txt path that falls through to
// the default feature set.
func (a *App) GetFeatureSetOnPlan(ctx context.Context, username string) (string, error) {
	key := "get_feature_set_on_plan:" + username
	return cache.Memoize(ctx, a.Cache, key, 24*time.Hour, func() (string, error) {
		var featureSet sql.NullString
		row := a.DB.QueryRowContext(ctx, `
			SELECT plans.feature_set
			FROM users
			JOIN plans ON users.plan_id = plans.id
			WHERE users.username = ?`, username)

		if err := row.Scan(&featureSet); err != nil {
			if err == sql.ErrNoRows {
				return "None", nil
			}
			return "", err
		}
		if !featureSet.Valid {
			return "None", nil
		}
		return featureSet.String, nil
	})
}

// QueryContextByUsername looks up a user's server context, cached 24h.
func (a *App) QueryContextByUsername(ctx context.Context, username string) (string, error) {
	key := "query_context_by_username:" + username
	return cache.Memoize(ctx, a.Cache, key, 24*time.Hour, func() (string, error) {
		var server sql.NullString
		row := a.DB.QueryRowContext(ctx, `SELECT server FROM users WHERE username = ?`, username)
		if err := row.Scan(&server); err != nil {
			if err == sql.ErrNoRows {
				return "", nil
			}
			return "", err
		}
		return server.String, nil
	})
}

// baselineFeatures are always granted regardless of the user's/plan's
// feature file.
var baselineFeatures = []string{
	"dashboard", "helpers", "websites", "databases_size_info",
	"screenshots", "favicons", "logout", "errors", "search", "app",
}

// LoadUserFeatures returns the feature list for a user: user-specific
// features.txt, else the plan's feature set file, else default.txt, else
// empty - plus the always-on baseline and any enabled plugins. Cached for
// 24h.
func (a *App) LoadUserFeatures(ctx context.Context, username, userContext string) ([]string, error) {
	if userContext == "" {
		var err error
		userContext, err = a.QueryContextByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
	}

	key := "load_user_features:" + username + ":" + userContext
	return cache.Memoize(ctx, a.Cache, key, 24*time.Hour, func() ([]string, error) {
		features, _ := loadFeaturesFromFile(fmt.Sprintf("/home/%s/features.txt", userContext))

		if features == nil {
			planFeatureSet, err := a.GetFeatureSetOnPlan(ctx, username)
			if err != nil {
				return nil, err
			}
			features, _ = loadFeaturesFromFile(fmt.Sprintf("/etc/openpanel/openpanel/features/%s.txt", planFeatureSet))
			if features == nil {
				features, _ = loadFeaturesFromFile("/etc/openpanel/openpanel/features/default.txt")
			}
		}

		result := append([]string{}, features...)
		result = append(result, baselineFeatures...)
		for name := range a.PluginNames {
			result = append(result, name)
		}
		return result, nil
	})
}

func loadFeaturesFromFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines, nil
}

func gravatarURL(avatarType, email string) string {
	if !strings.EqualFold(avatarType, "gravatar") || email == "" {
		return ""
	}
	sum := md5.Sum([]byte(strings.ToLower(email))) //nolint:gosec
	return "https://www.gravatar.com/avatar/" + hex.EncodeToString(sum[:]) + "?s=150&d=identicon"
}

// InjectData computes the per-request user/branding data every page
// template needs. Returns an empty, non-nil map for anonymous requests
// (userID == 0).
func (a *App) InjectData(ctx context.Context, userID int) (map[string]any, error) {
	if userID == 0 {
		return map[string]any{}, nil
	}

	details, err := a.GetUserDetailsWithPlan(ctx, userID)
	if err != nil {
		return nil, err
	}

	userFeatures, err := a.LoadUserFeatures(ctx, details.Username, details.Context)
	if err != nil {
		return nil, err
	}
	featureSet := make(map[string]bool, len(userFeatures))
	for _, f := range userFeatures {
		featureSet[f] = true
	}

	allowed := map[string]bool{}
	for _, m := range a.EnabledModules {
		if featureSet[m] {
			allowed[m] = true
		}
	}
	for name := range a.PluginNames {
		if featureSet[name] {
			allowed[name] = true
		}
	}
	userAllowed := make([]string, 0, len(allowed))
	for m := range allowed {
		userAllowed = append(userAllowed, m)
	}

	return map[string]any{
		"user_id":           userID,
		"current_username":  details.Username,
		"context":           details.Context,
		"current_email":     details.Email,
		"hosting_plan":      details.PlanID,
		"hosting_plan_name": details.PlanName,
		"panel_version":     sysinfo.GetOpenPanelVersion(ctx, a.Cache),
		// "dir" here is the config key for UI text direction (ltr/rtl),
		// unrelated to any filesystem path despite the "panel_dir" name below.
		"panel_dir":          a.Config.Get("dir", "ltr"),
		"avatar_type":        a.AvatarType,
		"gravatar_image_url": gravatarURL(a.AvatarType, details.Email),
		"user_allowed":       userAllowed,
		"is_enterprise":      strings.HasPrefix(a.LicenseKey, "enterprise"),
	}, nil
}
