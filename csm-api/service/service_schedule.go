package service

import (
	"context"
	"csm-api/auth"
	"csm-api/entity"
	"csm-api/store"
	"csm-api/txutil"
	"csm-api/utils"
)

type ServiceSchedule struct {
	SafeDB  store.Queryer
	SafeTDB store.Beginner
	Store   store.ScheduleStore
}

// func: 휴무일 조회
// @param
// -
func (s *ServiceSchedule) GetRestScheduleList(ctx context.Context, isRole bool, jno int64, year string, month string) (entity.RestSchedules, error) {

	uno, _ := auth.GetContext(ctx, auth.Uno{})

	list, err := s.Store.GetRestScheduleList(ctx, s.SafeDB, isRole, jno, uno, year, month)
	if err != nil {
		return entity.RestSchedules{}, utils.CustomErrorf(err)
	}

	return list, nil
}

// func: 휴무일 추가
// @param
// -
func (s *ServiceSchedule) AddRestSchedule(ctx context.Context, schedule entity.RestSchedules) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	err = s.Store.AddRestSchedule(ctx, tx, schedule)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	return
}

// func: 휴무일 수정
// @param
// -
func (s *ServiceSchedule) ModifyRestSchedule(ctx context.Context, schedule entity.RestSchedule) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	err = s.Store.ModifyRestSchedule(ctx, tx, schedule)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	return
}

// func: 휴무일 삭제
// @param
// -
func (s *ServiceSchedule) RemoveRestSchedule(ctx context.Context, cno int64) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	err = s.Store.RemoveRestSchedule(ctx, tx, cno)
	if err != nil {
		return utils.CustomErrorf(err)
	}
	return
}
