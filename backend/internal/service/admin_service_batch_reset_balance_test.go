//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type batchResetUserRepoStub struct {
	*userRepoStub
	changes []BatchBalanceChange
	err     error
}

func (s *batchResetUserRepoStub) BatchResetBalance(_ context.Context, userIDs []int64) ([]BatchBalanceChange, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.userRepoStub == nil || s.userRepoStub.user == nil {
		return nil, ErrUserNotFound
	}
	changes := make([]BatchBalanceChange, 0, len(userIDs))
	for _, id := range userIDs {
		u := s.userRepoStub.user
		changes = append(changes, BatchBalanceChange{UserID: id, Old: u.Balance, New: u.DefaultAmount})
		u.Balance = u.DefaultAmount
	}
	s.changes = changes
	return changes, nil
}

func newBatchResetService(repo *batchResetUserRepoStub, redeem *balanceRedeemRepoStub, invalidator *authCacheInvalidatorStub) *adminServiceImpl {
	return &adminServiceImpl{userRepo: repo, redeemCodeRepo: redeem, authCacheInvalidator: invalidator}
}

func TestAdminService_BatchResetUserBalance_ResetsEachUserToOwnDefaultAmount(t *testing.T) {
	repo := &batchResetUserRepoStub{userRepoStub: &userRepoStub{user: &User{ID: 7, Balance: 10, DefaultAmount: 2}}}
	redeem := &balanceRedeemRepoStub{redeemRepoStub: &redeemRepoStub{}}
	invalidator := &authCacheInvalidatorStub{}
	svc := newBatchResetService(repo, redeem, invalidator)

	affected, err := svc.BatchResetUserBalance(context.Background(), []int64{7})
	require.NoError(t, err)
	require.Equal(t, 1, affected)
	require.Equal(t, []BatchBalanceChange{{UserID: 7, Old: 10, New: 2}}, repo.changes)
	require.Equal(t, []int64{7}, invalidator.userIDs)
	require.Len(t, redeem.created, 1)
	record := redeem.created[0]
	require.Equal(t, AdjustmentTypeAdminBalance, record.Type)
	require.Equal(t, -8.0, record.Value)
	require.NotNil(t, record.UsedBy)
	require.Equal(t, int64(7), *record.UsedBy)
	require.Equal(t, "batch reset to default amount", record.Notes)
	require.NotNil(t, record.UsedAt)
}

func TestAdminService_BatchResetUserBalance_NoChangeNoAuditNoInvalidate(t *testing.T) {
	repo := &batchResetUserRepoStub{userRepoStub: &userRepoStub{user: &User{ID: 7, Balance: 5, DefaultAmount: 5}}}
	redeem := &balanceRedeemRepoStub{redeemRepoStub: &redeemRepoStub{}}
	invalidator := &authCacheInvalidatorStub{}
	svc := newBatchResetService(repo, redeem, invalidator)

	affected, err := svc.BatchResetUserBalance(context.Background(), []int64{7})
	require.NoError(t, err)
	require.Equal(t, 1, affected)
	require.Empty(t, redeem.created)
	require.Empty(t, invalidator.userIDs)
}

func TestAdminService_BatchResetUserBalance_DedupesAndSkipsInvalidIDs(t *testing.T) {
	repo := &batchResetUserRepoStub{userRepoStub: &userRepoStub{user: &User{ID: 7, Balance: 10, DefaultAmount: 2}}}
	redeem := &balanceRedeemRepoStub{redeemRepoStub: &redeemRepoStub{}}
	invalidator := &authCacheInvalidatorStub{}
	svc := newBatchResetService(repo, redeem, invalidator)

	affected, err := svc.BatchResetUserBalance(context.Background(), []int64{7, 7, 0, -3, 7})
	require.NoError(t, err)
	require.Equal(t, 1, affected)
	require.Equal(t, []BatchBalanceChange{{UserID: 7, Old: 10, New: 2}}, repo.changes)
	require.Len(t, redeem.created, 1)
}

func TestAdminService_BatchResetUserBalance_EmptyInput(t *testing.T) {
	repo := &batchResetUserRepoStub{userRepoStub: &userRepoStub{user: &User{ID: 7, Balance: 10, DefaultAmount: 2}}}
	redeem := &balanceRedeemRepoStub{redeemRepoStub: &redeemRepoStub{}}
	svc := newBatchResetService(repo, redeem, nil)

	affected, err := svc.BatchResetUserBalance(context.Background(), nil)
	require.NoError(t, err)
	require.Zero(t, affected)
	require.Nil(t, repo.changes)
	require.Empty(t, redeem.created)
}
