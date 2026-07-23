// Package emailnorm canonicalises e-mail addresses for uniqueness checks, so one
// mailbox cannot register many accounts via aliasing. Shared by wypas-auth
// (registration) and wypas-api (e-mail change): both compare and store the
// normalized form in accounts.email_normalized while keeping the user's original
// address in accounts.email for actual delivery.
package emailnorm

import "strings"

// gmailDomains are the domains Google treats as one mailbox namespace: dots in
// the local part are ignored and googlemail.com is an alias of gmail.com.
var gmailDomains = map[string]bool{
	"gmail.com":      true,
	"googlemail.com": true,
}

// Normalize returns the canonical identity of an e-mail address:
//   - trimmed and lowercased
//   - the local part's "+label" suffix removed (sub-addressing: gmail, outlook,
//     fastmail and most providers deliver user+anything to user)
//   - for gmail/googlemail: dots removed from the local part and the domain
//     folded to gmail.com (Google ignores both)
//
// A value with no "@" (or an empty local/domain part) is returned trimmed and
// lowercased only — callers validate address format separately.
func Normalize(email string) string {
	s := strings.ToLower(strings.TrimSpace(email))
	at := strings.LastIndex(s, "@")
	if at <= 0 || at == len(s)-1 {
		return s
	}
	local, domain := s[:at], s[at+1:]
	if plus := strings.Index(local, "+"); plus >= 0 {
		local = local[:plus]
	}
	if gmailDomains[domain] {
		local = strings.ReplaceAll(local, ".", "")
		domain = "gmail.com"
	}
	return local + "@" + domain
}
