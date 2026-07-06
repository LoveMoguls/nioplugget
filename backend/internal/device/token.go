package device

import (
	"net"
	"net/http"
	"time"

	"github.com/trollstaven/nioplugget/backend/internal/auth"
)

const deviceCookieName = "device"

// GenerateDeviceToken creates a 365-day JWT with role "device" and the
// current device epoch. Signed with the same secret as session JWTs.
func GenerateDeviceToken(epoch int32) (string, time.Time, error) {
	expiry := time.Now().Add(365 * 24 * time.Hour)
	claims := map[string]any{
		"role":  "device",
		"epoch": epoch,
		"exp":   expiry.Unix(),
	}
	_, tokenStr, err := auth.TokenAuth.Encode(claims)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenStr, expiry, nil
}

// generateProfileToken creates a 30-day session JWT for a profile chosen
// from a trusted device, stamping the current device epoch. This lets
// EpochGuard evict the session immediately if the family code is later
// changed (epoch bump), unlike bootstrap password/PIN logins which carry
// no epoch claim and are unaffected by epoch bumps.
func generateProfileToken(userID, role string, epoch int32) (string, time.Time, error) {
	expiry := time.Now().Add(30 * 24 * time.Hour)
	claims := map[string]any{
		"sub":   userID,
		"role":  role,
		"epoch": epoch,
		"exp":   expiry.Unix(),
	}
	_, tokenStr, err := auth.TokenAuth.Encode(claims)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenStr, expiry, nil
}

// TokenFromDeviceCookie extracts the device JWT for jwtauth.Verify.
func TokenFromDeviceCookie(r *http.Request) string {
	c, err := r.Cookie(deviceCookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

// clientIP prefers the Cloudflare header (all public traffic arrives via
// the tunnel from localhost), falling back to RemoteAddr.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
