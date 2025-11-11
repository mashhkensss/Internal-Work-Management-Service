package auth

import (
    "context"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "errors"
    "strings"
    "time"
)

var (
    ErrMalformedToken = errors.New("invalid jwt format")
    ErrUnsupportedAlg = errors.New("unsupported jwt algorithm")
    ErrInvalidSignature = errors.New("invalid jwt signature")
    ErrExpiredToken    = errors.New("jwt expired")
    ErrMissingSubject  = errors.New("jwt subject missing")
)

type Claims struct {
    Subject   string
    ExpiresAt time.Time
}

type claimsContextKey struct{}

var ctxKey claimsContextKey

func WithClaims(ctx context.Context, claims Claims) context.Context {
    return context.WithValue(ctx, ctxKey, claims)
}

func ClaimsFromContext(ctx context.Context) (Claims, bool) {
    val, ok := ctx.Value(ctxKey).(Claims)
    return val, ok
}

type jwtHeader struct {
    Alg string `json:"alg"`
    Typ string `json:"typ"`
}

type jwtPayload struct {
    Sub string `json:"sub"`
    Exp int64  `json:"exp"`
}

func Parse(token string, secret string) (Claims, error) {
    parts := strings.Split(token, ".")
    if len(parts) != 3 {
        return Claims{}, ErrMalformedToken
    }

    headerBytes, err := decodeSegment(parts[0])
    if err != nil {
        return Claims{}, ErrMalformedToken
    }
    var header jwtHeader
    if err := json.Unmarshal(headerBytes, &header); err != nil {
        return Claims{}, ErrMalformedToken
    }
    if header.Alg != "HS256" {
        return Claims{}, ErrUnsupportedAlg
    }

    payloadBytes, err := decodeSegment(parts[1])
    if err != nil {
        return Claims{}, ErrMalformedToken
    }
    var payload jwtPayload
    if err := json.Unmarshal(payloadBytes, &payload); err != nil {
        return Claims{}, ErrMalformedToken
    }

    if payload.Sub == "" {
        return Claims{}, ErrMissingSubject
    }

    signingInput := parts[0] + "." + parts[1]
    signature, err := decodeSegment(parts[2])
    if err != nil {
        return Claims{}, ErrMalformedToken
    }

    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write([]byte(signingInput))
    if !hmac.Equal(signature, mac.Sum(nil)) {
        return Claims{}, ErrInvalidSignature
    }

    claims := Claims{Subject: payload.Sub}
    if payload.Exp != 0 {
        exp := time.Unix(payload.Exp, 0)
        if time.Now().After(exp) {
            return Claims{}, ErrExpiredToken
        }
        claims.ExpiresAt = exp
    }

    return claims, nil
}

func decodeSegment(seg string) ([]byte, error) {
    switch len(seg) % 4 {
    case 2:
        seg += "=="
    case 3:
        seg += "="
    }
    return base64.URLEncoding.DecodeString(seg)
}
