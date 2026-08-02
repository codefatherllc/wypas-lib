package authz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codefatherllc/wypas-lib/jwt"
)

// allScopes is every scope the taxonomy defines; the matrix test asserts each
// one against every group so a mapping change can't slip through unreviewed.
var allScopes = []string{
	ScopeReportsView, ScopeBansManage, ScopeSnifferView, ScopePointsGrant,
	ScopeBlockedNetworks, ScopeMapBlurManage, ScopeTutorReview, ScopeCamsManage,
	ScopeContentPages, ScopeContentShop, ScopeContentMetadata, ScopeCreatorUse,
	ScopeAccountingView, ScopeAccountingManage, ScopeDBRead, ScopeDBWrite,
}

// expected is the owner-approved matrix, hardcoded independently of authz.go so
// this test genuinely locks it (not a re-derivation of the same map).
var expected = map[int][]string{
	1: {},
	2: {},
	3: {ScopeBansManage, ScopeSnifferView},
	4: {ScopeBansManage, ScopeSnifferView, ScopeTutorReview, ScopePointsGrant, ScopeReportsView, ScopeMapBlurManage, ScopeBlockedNetworks, ScopeCamsManage},
	5: {ScopeBansManage, ScopeSnifferView, ScopeTutorReview, ScopePointsGrant, ScopeReportsView, ScopeMapBlurManage, ScopeBlockedNetworks, ScopeCamsManage, ScopeContentPages, ScopeContentShop, ScopeContentMetadata, ScopeCreatorUse},
	6: allScopes,
}

func has(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func TestMatrixEveryCell(t *testing.T) {
	for group := 1; group <= 6; group++ {
		for _, scope := range allScopes {
			want := has(expected[group], scope)
			if got := Has(group, scope); got != want {
				t.Errorf("Has(group=%d, %q) = %v, want %v", group, scope, got, want)
			}
		}
	}
}

func TestScopesForMatchesMatrix(t *testing.T) {
	for group := 1; group <= 6; group++ {
		got := ScopesFor(group)
		if len(got) != len(expected[group]) {
			t.Errorf("ScopesFor(%d) len = %d, want %d (%v)", group, len(got), len(expected[group]), got)
		}
		for _, s := range expected[group] {
			if !has(got, s) {
				t.Errorf("ScopesFor(%d) missing %q", group, s)
			}
		}
	}
	// Player/Tutor hold nothing; Admin holds everything.
	if len(ScopesFor(2)) != 0 {
		t.Errorf("Tutor should hold no scopes, got %v", ScopesFor(2))
	}
	if len(ScopesFor(6)) != len(allScopes) {
		t.Errorf("Admin should hold all %d scopes, got %d", len(allScopes), len(ScopesFor(6)))
	}
}

func TestHasUnknownScope(t *testing.T) {
	if Has(6, "not.a.real.scope") {
		t.Error("unknown scope must never be granted, even to Admin")
	}
}

func TestRequireScope(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	cases := []struct {
		name   string
		claims *jwt.Claims
		scope  string
		want   int
	}{
		{"gm has sniffer", &jwt.Claims{GroupID: 3}, ScopeSnifferView, http.StatusOK},
		{"gm lacks db.read", &jwt.Claims{GroupID: 3}, ScopeDBRead, http.StatusForbidden},
		{"admin has db.write", &jwt.Claims{GroupID: 6}, ScopeDBWrite, http.StatusOK},
		{"cm has reports", &jwt.Claims{GroupID: 4}, ScopeReportsView, http.StatusOK},
		{"gm lacks reports", &jwt.Claims{GroupID: 3}, ScopeReportsView, http.StatusForbidden},
		{"no claims", nil, ScopeBansManage, http.StatusForbidden},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if c.claims != nil {
				r = r.WithContext(context.WithValue(r.Context(), jwt.ClaimsKey, c.claims))
			}
			w := httptest.NewRecorder()
			RequireScope(c.scope)(ok).ServeHTTP(w, r)
			if w.Code != c.want {
				t.Errorf("code = %d, want %d", w.Code, c.want)
			}
		})
	}
}
