package v1

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	errs "eshkere/internal/errors"
	"eshkere/internal/handler"
)

const (
	vkidStateCookieName      = "vkid_oauth_state"
	vkidVerifierCookieName   = "vkid_oauth_verifier"
	vkidRedirectCookieName   = "vkid_oauth_redirect"
	vkidCodeChallengeMethod  = "s256"
	vkidDefaultScope         = "email phone"
	vkidDefaultDomain        = "id.vk.ru"
	vkidFallbackRedirectPath = "/"
)

func (a *API) BeginVKIDLogin(w http.ResponseWriter, r *http.Request) {
	if a.vkidConfig.ClientID <= 0 || strings.TrimSpace(a.vkidConfig.RedirectURI) == "" {
		handler.HandleError(w, r, "start vk id auth", errs.NotImplementedError)
		return
	}

	state, err := generateRandomBase64URL(32)
	if err != nil {
		handler.HandleError(w, r, "generate vk id state", err)
		return
	}

	codeVerifier, err := generateRandomBase64URL(48)
	if err != nil {
		handler.HandleError(w, r, "generate vk id verifier", err)
		return
	}

	returnURL := a.resolveVKIDReturnURL()
	a.setOAuthCookie(w, vkidStateCookieName, state)
	a.setOAuthCookie(w, vkidVerifierCookieName, codeVerifier)
	a.setOAuthCookie(w, vkidRedirectCookieName, base64.RawURLEncoding.EncodeToString([]byte(returnURL)))

	http.Redirect(w, r, buildVKIDAuthorizeURL(a.vkidConfig, state, codeVerifier), http.StatusFound)
}

func (a *API) LoginVKIDCallback(w http.ResponseWriter, r *http.Request) {
	returnURL := a.readVKIDReturnURL(r)
	clearVKIDOAuthCookies(w, a.cookieConfig.Secure)

	state := strings.TrimSpace(r.URL.Query().Get("state"))
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	deviceID := strings.TrimSpace(r.URL.Query().Get("device_id"))
	if state == "" || code == "" || deviceID == "" {
		a.redirectVKIDError(w, r, returnURL, "missing callback params")
		return
	}

	expectedState, err := readCookieValue(r, vkidStateCookieName)
	if err != nil || state != expectedState {
		a.redirectVKIDError(w, r, returnURL, "state mismatch")
		return
	}

	codeVerifier, err := readCookieValue(r, vkidVerifierCookieName)
	if err != nil || strings.TrimSpace(codeVerifier) == "" {
		a.redirectVKIDError(w, r, returnURL, "missing code verifier")
		return
	}

	advID, sessionID, expiresAt, err := a.authClient.LoginVKID(r.Context(), code, deviceID, codeVerifier)
	if err != nil {
		a.redirectVKIDError(w, r, returnURL, err.Error())
		return
	}

	email, _, err := a.authClient.GetCredentials(r.Context(), advID)
	if err != nil {
		_ = a.authClient.Logout(r.Context(), sessionID)
		a.redirectVKIDError(w, r, returnURL, err.Error())
		return
	}

	if err := a.ensureAdvertiserProfile(r.Context(), advID, email); err != nil {
		_ = a.authClient.Logout(r.Context(), sessionID)
		a.redirectVKIDError(w, r, returnURL, err.Error())
		return
	}

	a.setSessionCookie(w, sessionID, time.Unix(expiresAt, 0))
	http.Redirect(w, r, returnURL, http.StatusFound)
}

func (a *API) resolveVKIDReturnURL() string {
	if strings.TrimSpace(a.vkidConfig.DefaultRedirectURL) != "" {
		return a.vkidConfig.DefaultRedirectURL
	}
	return vkidFallbackRedirectPath
}

func (a *API) readVKIDReturnURL(r *http.Request) string {
	raw, err := readCookieValue(r, vkidRedirectCookieName)
	if err != nil || raw == "" {
		return a.resolveVKIDReturnURL()
	}

	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(decoded) == 0 {
		return a.resolveVKIDReturnURL()
	}

	return string(decoded)
}

func (a *API) redirectVKIDError(w http.ResponseWriter, r *http.Request, returnURL, message string) {
	target := strings.TrimSpace(a.vkidConfig.ErrorRedirectURL)
	if target == "" {
		target = returnURL
	}
	http.Redirect(w, r, appendQueryParam(target, "vk_auth_error", message), http.StatusFound)
}

func (a *API) setOAuthCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   a.cookieConfig.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearVKIDOAuthCookies(w http.ResponseWriter, secure bool) {
	for _, name := range []string{vkidStateCookieName, vkidVerifierCookieName, vkidRedirectCookieName} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		})
	}
}

func buildVKIDAuthorizeURL(cfg VKIDConfig, state, codeVerifier string) string {
	domain := strings.TrimSpace(cfg.AuthDomain)
	if domain == "" {
		domain = vkidDefaultDomain
	}
	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		domain = "https://" + domain
	}

	scope := strings.TrimSpace(cfg.Scope)
	if scope == "" {
		scope = vkidDefaultScope
	}

	query := url.Values{
		"client_id":             {fmt.Sprintf("%d", cfg.ClientID)},
		"redirect_uri":          {cfg.RedirectURI},
		"response_type":         {"code"},
		"state":                 {state},
		"scope":                 {scope},
		"code_challenge":        {buildCodeChallenge(codeVerifier)},
		"code_challenge_method": {vkidCodeChallengeMethod},
	}

	return strings.TrimRight(domain, "/") + "/authorize?" + query.Encode()
}

func buildCodeChallenge(codeVerifier string) string {
	sum := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func generateRandomBase64URL(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func readCookieValue(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	if cookie == nil || cookie.Value == "" {
		return "", http.ErrNoCookie
	}
	return cookie.Value, nil
}

func appendQueryParam(rawURL, key, value string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
	return u.String()
}
