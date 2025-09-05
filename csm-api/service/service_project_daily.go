package service

import (
	"context"
	"csm-api/auth"
	"csm-api/entity"
	"csm-api/store"
	"csm-api/txutil"
	"csm-api/utils"
)

type ServiceProjectDaily struct {
	SafeDB  store.Queryer
	SafeTDB store.Beginner
	Store   store.ProjectDailyStore
}

// 작업내용 조회
func (s *ServiceProjectDaily) GetDailyJobList(ctx context.Context, isRole bool, jno int64, targetDate string) (entity.ProjectDailys, error) {

	uno, _ := auth.GetContext(ctx, auth.Uno{})

	list, err := s.Store.GetDailyJobList(ctx, s.SafeDB, isRole, jno, uno, targetDate)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}
	return list, nil
}

// 작업내용 추가
func (s *ServiceProjectDaily) AddDailyJob(ctx context.Context, project entity.ProjectDailys) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	if err = s.Store.AddDailyJob(ctx, tx, project); err != nil {
		return utils.CustomErrorf(err)
	}
	return
}

// 작업내용 수정
func (s *ServiceProjectDaily) ModifyDailyJob(ctx context.Context, project entity.ProjectDaily) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	if err = s.Store.ModifyDailyJob(ctx, tx, project); err != nil {
		return utils.CustomErrorf(err)
	}
	return
}

// 작업내용 삭제
func (s *ServiceProjectDaily) RemoveDailyJob(ctx context.Context, idx int64) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	if err = s.Store.RemoveDailyJob(ctx, tx, idx); err != nil {
		return utils.CustomErrorf(err)
	}
	return
}
