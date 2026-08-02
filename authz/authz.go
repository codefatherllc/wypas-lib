// Package authz is the single source of truth for admin permission scopes and
// the group -> scopes mapping. It is shared by every service that gates admin
// features (wypas-api today; wypas-auth / wypas-creator can adopt the same
// RequireScope seam). Scopes are DERIVED from the JWT group_id at check time, so
// a mapping change takes effect on the next request with no stale 24h token.
package authz

import (
	"net/http"
	"sort"

	"github.com/codefatherllc/wypas-lib/jwt"
)

// Scope names. Keep in sync with the web router meta.requiresScope values.
const (
	ScopeReportsView      = "reports.view"            // list + dismiss player bug reports
	ScopeBansManage       = "bans.manage"             // account/player/ip bans
	ScopeSnifferView      = "sniffer.view"            // character inspect (PII: ip, asn, balances)
	ScopePointsGrant      = "points.grant"            // add/remove premium points
	ScopeBlockedNetworks  = "blocked_networks.manage" // website ASN blocklist
	ScopeMapBlurManage    = "map.blur.manage"         // public-map blur regions
	ScopeCamsManage       = "cams.manage"             // recategorise/describe/hide/delete any player recording
	ScopeTutorReview      = "tutor.review"            // accept/decline tutor applications
	ScopeContentPages     = "content.pages.manage"    // CMS/library pages
	ScopeContentShop      = "content.shop.manage"     // point-shop offers
	ScopeContentMetadata  = "content.metadata.manage" // site metadata KV (backs CMS pages)
	ScopeCreatorUse       = "creator.use"             // map/item/appearance/effect/missile/generator editors
	ScopeAccountingView   = "accounting.view"         // read sales/refunds/payouts/exports
	ScopeAccountingManage = "accounting.manage"       // resync/flag/resolve/export/sync
	ScopeDBRead           = "db.read"                 // schema/table browse (SELECT)
	ScopeDBWrite          = "db.write"                // insert/update/delete + raw query console
)

// scopeMinGroup maps each scope to the LOWEST group_id that holds it. The model
// is cumulative (a higher group holds every scope the lower groups do), matching
// how the platform already treats group_id as a ladder. Groups: 1 Player,
// 2 Tutor (no admin), 3 Gamemaster, 4 Community Manager, 5 Product Manager,
// 6 Administrator (groups.xml). The matrix below is the owner-approved target
// (2026-07-22): GM is limited to enforcement (bans+sniffer); moderation tooling
// lives at CM; content/creator at PM; db+accounting stay Administrator-only.
var scopeMinGroup = map[string]int{
	// GM (3) — enforcement only
	ScopeBansManage:  3,
	ScopeSnifferView: 3,
	// CM (4) — moderation + community
	ScopeTutorReview:     4,
	ScopePointsGrant:     4,
	ScopeReportsView:     4,
	ScopeMapBlurManage:   4,
	ScopeBlockedNetworks: 4,
	ScopeCamsManage:      4,
	// PM (5) — content + creator
	ScopeContentPages:    5,
	ScopeContentShop:     5,
	ScopeContentMetadata: 5,
	ScopeCreatorUse:      5,
	// Admin (6) — database + finance
	ScopeDBRead:           6,
	ScopeDBWrite:          6,
	ScopeAccountingView:   6,
	ScopeAccountingManage: 6,
}

// Has reports whether an account in groupID holds scope.
func Has(groupID int, scope string) bool {
	min, ok := scopeMinGroup[scope]
	return ok && groupID >= min
}

// ScopesFor returns the sorted scope list an account in groupID holds. Empty for
// Player/Tutor (no scope sits at group <= 2). Used by GET /api/me/scopes so the
// SPA can hide nav — the API's per-route RequireScope stays the real gate.
func ScopesFor(groupID int) []string {
	out := make([]string, 0, len(scopeMinGroup))
	for s, min := range scopeMinGroup {
		if groupID >= min {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// RequireScope is HTTP middleware that 403s unless the request's JWT group holds
// scope. Mirrors jwt.RequireAdmin; use in place of the blanket admin gate.
func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := jwt.GetClaims(r)
			if claims == nil || !Has(claims.GroupID, scope) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}
