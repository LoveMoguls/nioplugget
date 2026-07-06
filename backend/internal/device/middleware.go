package device

import (
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/trollstaven/nioplugget/backend/internal/auth"
)

// EpochGuard evicts profile-selected sessions (the "jwt" cookie) whose
// embedded epoch claim is stale — i.e. the family code was changed (device
// epoch bumped) after the session was minted. This gives "omedelbar
// fjärrutkastning" (immediate remote logout) on a code change.
//
// Sessions without an epoch claim are bootstrap logins (email+password or
// child PIN) and are intentionally left untouched — they were not minted
// from a trusted-device profile pick and are not subject to epoch eviction.
//
// Requests without a jwt cookie, or with one that fails to decode/verify,
// are passed through untouched: this middleware only ever adds an eviction
// on top of whatever the route's own auth middleware decides, it never
// grants access.
func (h *Handler) EpochGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("jwt")
		if err != nil || c.Value == "" {
			next.ServeHTTP(w, r)
			return
		}
		tok, err := jwtauth.VerifyToken(auth.TokenAuth, c.Value)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		var epoch float64
		if err := tok.Get("epoch", &epoch); err != nil {
			// No epoch claim — bootstrap login, not subject to eviction.
			next.ServeHTTP(w, r)
			return
		}
		settings, err := h.store.GetFamilySettings(r.Context())
		if err != nil {
			// Fail open: we can't confirm the current epoch, but the JWT
			// itself is still validly signed and unexpired, so let the
			// request through rather than lock everyone out on a
			// transient DB error.
			next.ServeHTTP(w, r)
			return
		}
		if int32(epoch) != settings.DeviceEpoch {
			auth.ClearAuthCookie(w)
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Utloggad — enheten behöver låsas upp igen"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
