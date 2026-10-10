package stores

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/dvtkrlbs/market/internal/ent"
	"github.com/dvtkrlbs/market/internal/ent/predicate"
	"github.com/dvtkrlbs/market/internal/ent/store"
	"github.com/dvtkrlbs/market/internal/ent/storeadmin"
)

var (
	ErrNameRequired = errors.New("name is required")
	ErrInvalidMode  = errors.New("for must be admin, manage or shop")
	ErrEmptyUpdate  = errors.New("at least one field is required")
	ErrNotFound     = errors.New("store not found")
	ErrForbidden    = errors.New("store administration required")
)

type Update struct {
	Name     *string
	Approved *bool
}

type Service struct{ client *ent.Client }

func New(client *ent.Client) *Service { return &Service{client: client} }

func administeredBy(userID int) predicate.Store {
	return store.HasAdminsWith(storeadmin.UserIDEQ(userID))
}

func (s *Service) List(ctx context.Context, userID int, mode string) ([]*ent.Store, error) {
	query := s.client.Store.Query().Order(ent.Asc(store.FieldID))
	switch mode {
	case "", "shop":
		query.Where(store.DeletedAtIsNil(), store.ApprovedEQ(true))
	case "manage":
		query.Where(store.DeletedAtIsNil(), administeredBy(userID))
	case "admin":
	default:
		return nil, ErrInvalidMode
	}
	return query.All(ctx)
}

func (s *Service) Create(ctx context.Context, userID int, name string) (*ent.Store, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrNameRequired
	}
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	created, err := tx.Store.Create().SetName(name).SetApproved(false).Save(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := tx.StoreAdmin.Create().SetUserID(userID).SetStoreID(created.ID).Save(ctx); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return created.Unwrap(), nil
}

func (s *Service) Get(ctx context.Context, userID, storeID int) (*ent.Store, error) {
	item, err := s.client.Store.Query().Where(store.IDEQ(storeID), store.DeletedAtIsNil(), store.Or(store.ApprovedEQ(true), administeredBy(userID))).Only(ctx)
	if ent.IsNotFound(err) {
		return nil, ErrNotFound
	}
	return item, err
}

func (s *Service) Update(ctx context.Context, userID, storeID int, input Update) (*ent.Store, error) {
	if input.Name == nil && input.Approved == nil {
		return nil, ErrEmptyUpdate
	}
	name := ""
	if input.Name != nil {
		name = strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, ErrNameRequired
		}
	}
	query := s.client.Store.UpdateOneID(storeID).Where(store.DeletedAtIsNil(), administeredBy(userID))
	if input.Name != nil {
		query.SetName(name)
	}
	if input.Approved != nil {
		query.SetApproved(*input.Approved)
	}
	item, err := query.Save(ctx)
	if ent.IsNotFound(err) {
		return nil, s.mutationError(ctx, storeID)
	}
	return item, err
}

func (s *Service) mutationError(ctx context.Context, storeID int) error {
	exists, err := s.client.Store.Query().Where(store.IDEQ(storeID), store.DeletedAtIsNil()).Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return ErrForbidden
}

// Approve is called only after the API has verified global admin membership.
func (s *Service) Approve(ctx context.Context, storeID int) (*ent.Store, error) {
	item, err := s.client.Store.UpdateOneID(storeID).Where(store.DeletedAtIsNil()).SetApproved(true).Save(ctx)
	if ent.IsNotFound(err) {
		return nil, ErrNotFound
	}
	return item, err
}

func (s *Service) Delete(ctx context.Context, userID, storeID int) error {
	// Preserve the store and its memberships; only admin listings include it.
	count, err := s.client.Store.Update().Where(store.IDEQ(storeID), store.DeletedAtIsNil(), administeredBy(userID)).SetDeletedAt(time.Now()).Save(ctx)
	if err != nil {
		return err
	}
	if count == 0 {
		return s.mutationError(ctx, storeID)
	}
	return nil
}
