package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/dvtkrlbs/market/internal/ent"
	"github.com/dvtkrlbs/market/internal/ent/session"
	"github.com/dvtkrlbs/market/internal/ent/user"
)

const sessionLifetime = 30 * 24 * time.Hour

var (
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrInvalidPassword    = errors.New("password must be between 8 and 128 characters")
	ErrEmailAlreadyExists = errors.New("email is already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidSession     = errors.New("invalid or expired session")
)

type SessionMetadata struct {
	UserAgent *string
	IP        *string
}

type SessionToken struct {
	Token     string
	ExpiresAt time.Time
}

type Principal struct {
	User    *ent.User
	Session *ent.Session
	Token   string
}

type Service struct {
	client    *ent.Client
	now       func() time.Time
	dummyHash string
}

func New(client *ent.Client) *Service {
	return &Service{
		client: client,
		now:    time.Now,
		dummyHash: encodePassword(
			"not-a-real-password",
			make([]byte, argonSaltLength),
			argonMemory,
			argonIterations,
			argonParallelism,
			argonKeyLength,
		),
	}
}

func (s *Service) Register(ctx context.Context, email, password string) (*ent.User, error) {
	email, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}
	if err := validatePassword(password); err != nil {
		return nil, err
	}
	passwordHash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	created, err := s.client.User.Create().
		SetEmail(email).
		SetPasswordHash(passwordHash).
		Save(ctx)
	if ent.IsConstraintError(err) {
		return nil, ErrEmailAlreadyExists
	}
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return created, nil
}

func (s *Service) Login(ctx context.Context, email, password string, metadata SessionMetadata) (SessionToken, error) {
	normalized, normalizeErr := normalizeEmail(email)
	var account *ent.User
	var err error
	if normalizeErr == nil {
		account, err = s.client.User.Query().
			Where(user.Email(normalized), user.DeletedAtIsNil()).
			Only(ctx)
	}
	encoded := s.dummyHash
	if err == nil && account != nil {
		encoded = account.PasswordHash
	}
	valid, verifyErr := verifyPassword(encoded, password)
	if verifyErr != nil && account != nil {
		return SessionToken{}, fmt.Errorf("verify password hash: %w", verifyErr)
	}
	if err != nil && !ent.IsNotFound(err) {
		return SessionToken{}, fmt.Errorf("find user: %w", err)
	}
	if normalizeErr != nil || err != nil || account == nil || !valid {
		return SessionToken{}, ErrInvalidCredentials
	}
	return s.createSession(ctx, s.client.Session, account.ID, metadata)
}

func (s *Service) Authenticate(ctx context.Context, token string) (*Principal, error) {
	id, secret, err := parseToken(token)
	if err != nil {
		return nil, ErrInvalidSession
	}
	stored, err := s.client.Session.Query().
		Where(session.ID(id)).
		WithUser().
		Only(ctx)
	if err != nil {
		return nil, ErrInvalidSession
	}
	want := secretDigest(secret)
	if subtle.ConstantTimeCompare([]byte(stored.SessionSecret), []byte(want)) != 1 ||
		stored.RevokedAt != nil || !stored.ExpiresAt.After(s.now()) {
		return nil, ErrInvalidSession
	}
	account, err := stored.Edges.UserOrErr()
	if err != nil || account.DeletedAt != nil {
		return nil, ErrInvalidSession
	}
	return &Principal{User: account, Session: stored, Token: token}, nil
}

func (s *Service) Logout(ctx context.Context, principal *Principal) error {
	now := s.now()
	updated, err := s.client.Session.Update().
		Where(
			session.ID(principal.Session.ID),
			session.RevokedAtIsNil(),
		).
		SetRevokedAt(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if updated != 1 {
		return ErrInvalidSession
	}
	return nil
}

func (s *Service) Refresh(ctx context.Context, principal *Principal) (SessionToken, error) {
	now := s.now()
	expiresAt := now.Add(sessionLifetime)
	updated, err := s.client.Session.Update().
		Where(
			session.ID(principal.Session.ID),
			session.RevokedAtIsNil(),
			session.ExpiresAtGT(now),
		).
		SetExpiresAt(expiresAt).
		Save(ctx)
	if err != nil {
		return SessionToken{}, fmt.Errorf("refresh session: %w", err)
	}
	if updated != 1 {
		return SessionToken{}, ErrInvalidSession
	}
	principal.Session.ExpiresAt = expiresAt
	return SessionToken{Token: principal.Token, ExpiresAt: expiresAt}, nil
}

type sessionCreator interface {
	Create() *ent.SessionCreate
}

func (s *Service) createSession(ctx context.Context, sessions sessionCreator, userID int, metadata SessionMetadata) (SessionToken, error) {
	id, err := randomTokenPart(16)
	if err != nil {
		return SessionToken{}, err
	}
	secret, err := randomTokenPart(32)
	if err != nil {
		return SessionToken{}, err
	}
	expiresAt := s.now().Add(sessionLifetime)
	_, err = sessions.Create().
		SetID(id).
		SetSessionSecret(secretDigest(secret)).
		SetUserID(userID).
		SetNillableUserAgent(metadata.UserAgent).
		SetNillableIP(metadata.IP).
		SetExpiresAt(expiresAt).
		Save(ctx)
	if err != nil {
		return SessionToken{}, fmt.Errorf("create session: %w", err)
	}
	return SessionToken{Token: id + "." + secret, ExpiresAt: expiresAt}, nil
}

func normalizeEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || len(value) > 320 {
		return "", ErrInvalidEmail
	}
	parsed, err := mail.ParseAddress(value)
	if err != nil || parsed.Address != value {
		return "", ErrInvalidEmail
	}
	return value, nil
}

func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 128 {
		return ErrInvalidPassword
	}
	return nil
}

func randomTokenPart(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func parseToken(token string) (string, string, error) {
	id, secret, ok := strings.Cut(token, ".")
	if !ok || id == "" || secret == "" || strings.Contains(secret, ".") {
		return "", "", ErrInvalidSession
	}
	idBytes, idErr := base64.RawURLEncoding.Strict().DecodeString(id)
	secretBytes, secretErr := base64.RawURLEncoding.Strict().DecodeString(secret)
	if idErr != nil || secretErr != nil || len(idBytes) != 16 || len(secretBytes) != 32 {
		return "", "", ErrInvalidSession
	}
	return id, secret, nil
}

func secretDigest(secret string) string {
	digest := sha256.Sum256([]byte(secret))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}
