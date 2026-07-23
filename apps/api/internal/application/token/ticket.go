package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const RealtimeAudience = "realtime"

const RealtimeTicketTTL = 30 * time.Second

var ErrInvalidTicket = errors.New("invalid realtime ticket")

func NewRealtimeTicket(secret string, userID uuid.UUID, now time.Time) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		Audience:  jwt.ClaimStrings{RealtimeAudience},
		ExpiresAt: jwt.NewNumericDate(now.Add(RealtimeTicketTTL)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func ParseRealtimeTicket(secret, ticket string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	tok, err := jwt.ParseWithClaims(ticket, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidTicket
		}
		return []byte(secret), nil
	}, jwt.WithAudience(RealtimeAudience), jwt.WithExpirationRequired())
	if err != nil || !tok.Valid {
		return uuid.Nil, ErrInvalidTicket
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidTicket
	}
	return userID, nil
}
