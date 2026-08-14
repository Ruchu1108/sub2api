package service

import (
	"context"
	"errors"
	"fmt"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/accountuserbinding"
	"github.com/Wei-Shaw/sub2api/ent/user"
)

// userEntityToServiceAdminView 把 ent.User 映射为管理端展示用的 service.User。
func userEntityToServiceAdminView(u *dbent.User) User {
	out := User{
		ID:                         u.ID,
		Email:                      u.Email,
		Username:                   u.Username,
		Notes:                      u.Notes,
		PasswordHash:               u.PasswordHash,
		Role:                       u.Role,
		Balance:                    u.Balance,
		DefaultAmount:              u.DefaultAmount,
		FrozenBalance:              u.FrozenBalance,
		Concurrency:                u.Concurrency,
		Status:                     u.Status,
		SignupSource:               u.SignupSource,
		LastActiveAt:               u.LastActiveAt,
		TotpEnabled:                u.TotpEnabled,
		TotpEnabledAt:              u.TotpEnabledAt,
		BalanceNotifyEnabled:       u.BalanceNotifyEnabled,
		BalanceNotifyThresholdType: u.BalanceNotifyThresholdType,
		BalanceNotifyThreshold:     u.BalanceNotifyThreshold,
		TotalRecharged:             u.TotalRecharged,
		RPMLimit:                   u.RpmLimit,
		CreatedAt:                  u.CreatedAt,
		UpdatedAt:                  u.UpdatedAt,
		DeletedAt:                  u.DeletedAt,
	}
	if u.BalanceNotifyExtraEmails != "" && u.BalanceNotifyExtraEmails != "[]" {
		out.BalanceNotifyExtraEmails = ParseNotifyEmails(u.BalanceNotifyExtraEmails)
	}
	return out
}

func (s *adminServiceImpl) ListBoundUsers(ctx context.Context, accountID int64) ([]User, error) {
	if accountID <= 0 {
		return []User{}, errors.New("invalid account id")
	}
	if s.entClient == nil {
		return []User{}, errors.New("binding storage unavailable")
	}
	bindings, err := s.entClient.AccountUserBinding.Query().
		Where(accountuserbinding.AccountIDEQ(accountID)).
		WithUser().
		Order(dbent.Asc(accountuserbinding.FieldID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list account user bindings: %w", err)
	}
	users := make([]User, 0, len(bindings))
	for i := range bindings {
		if bindings[i].Edges.User == nil || bindings[i].Edges.User.DeletedAt != nil {
			continue
		}
		users = append(users, userEntityToServiceAdminView(bindings[i].Edges.User))
	}
	return users, nil
}

// SetBoundUsers 全量替换账号绑定的用户集合。
func (s *adminServiceImpl) SetBoundUsers(ctx context.Context, accountID int64, userIDs []int64) (int, error) {
	if accountID <= 0 {
		return 0, errors.New("invalid account id")
	}
	if s.entClient == nil {
		return 0, errors.New("binding storage unavailable")
	}
	if _, err := s.accountRepo.GetByID(ctx, accountID); err != nil {
		return 0, err
	}

	seen := make(map[int64]struct{}, len(userIDs))
	cleaned := make([]int64, 0, len(userIDs))
	for _, uid := range userIDs {
		if uid <= 0 {
			continue
		}
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		cleaned = append(cleaned, uid)
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	txClient := tx.Client()
	txCtx := dbent.NewTxContext(ctx, tx)

	if _, err := txClient.AccountUserBinding.Delete().
		Where(accountuserbinding.AccountIDEQ(accountID)).
		Exec(txCtx); err != nil {
		return 0, fmt.Errorf("clear account user bindings: %w", err)
	}

	validIDs := make([]int64, 0, len(cleaned))
	if len(cleaned) > 0 {
		validIDs, err = txClient.User.Query().Where(user.IDIn(cleaned...)).IDs(txCtx)
		if err != nil {
			return 0, fmt.Errorf("validate bound users: %w", err)
		}
		creates := make([]*dbent.AccountUserBindingCreate, 0, len(validIDs))
		for _, uid := range validIDs {
			creates = append(creates, txClient.AccountUserBinding.Create().
				SetAccountID(accountID).
				SetUserID(uid))
		}
		if len(creates) > 0 {
			if err := txClient.AccountUserBinding.CreateBulk(creates...).Exec(txCtx); err != nil {
				return 0, fmt.Errorf("insert account user bindings: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(validIDs), nil
}

func (s *adminServiceImpl) ResetBoundUserBalances(ctx context.Context, accountID int64) (int, error) {
	if accountID <= 0 || s.entClient == nil {
		return 0, nil
	}
	bindings, err := s.entClient.AccountUserBinding.Query().
		Where(accountuserbinding.AccountIDEQ(accountID)).
		Order(dbent.Asc(accountuserbinding.FieldID)).
		All(ctx)
	if err != nil {
		return 0, fmt.Errorf("list bound user ids: %w", err)
	}
	userIDs := make([]int64, 0, len(bindings))
	for i := range bindings {
		userIDs = append(userIDs, bindings[i].UserID)
	}

	const batchSize = 500
	total := 0
	for start := 0; start < len(userIDs); start += batchSize {
		end := min(start+batchSize, len(userIDs))
		affected, err := s.BatchResetUserBalance(ctx, userIDs[start:end])
		if err != nil {
			return total, err
		}
		total += affected
	}
	return total, nil
}
