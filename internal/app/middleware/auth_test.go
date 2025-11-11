package middleware

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

func TestJWTMiddlewareSuccess(t *testing.T) {
    token := mustMakeToken(t, "user-1", time.Now().Add(time.Hour), "secret")
    mw := JWT("secret")
    called := false
    handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        called = true
    }))

    req := httptest.NewRequest(http.MethodGet, "/", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    rr := httptest.NewRecorder()

    handler.ServeHTTP(rr, req)

    if !called {
        t.Fatalf("expected handler to be called")
    }
    if rr.Code == http.StatusUnauthorized {
        t.Fatalf("expected success, got 401")
    }
}

func TestJWTMiddlewareUnauthorized(t *testing.T) {
    mw := JWT("secret")
    handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

    req := httptest.NewRequest(http.MethodGet, "/", nil)
    rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, req)

    if rr.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401, got %d", rr.Code)
    }
}

func mustMakeToken(t *testing.T, sub string, exp time.Time, secret string) string {
    t.Helper()
    header := map[string]string{"alg": "HS256", "typ": "JWT"}
    payload := map[string]any{"sub": sub}
    if !exp.IsZero() {
        payload["exp"] = exp.Unix()
    }
    headerJSON, _ := json.Marshal(header)
    payloadJSON, _ := json.Marshal(payload)
    headerSegment := base64.RawURLEncoding.EncodeToString(headerJSON)
    payloadSegment := base64.RawURLEncoding.EncodeToString(payloadJSON)
    signingInput := headerSegment + "." + payloadSegment
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(signingInput))
    signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
    return signingInput + "." + signature
}
