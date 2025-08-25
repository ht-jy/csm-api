package service

import (
	"context"
	"csm-api/entity"
	"csm-api/store"
	"csm-api/txutil"
	"csm-api/utils"
	"time"
)

type ServiceEquip struct {
	SafeDB  store.Queryer
	SafeTDB store.Beginner
	Store   store.EquipStore
}

func (s *ServiceEquip) GetEquipList(ctx context.Context, jno int64, sno int64) (entity.Equips, error) {
	list, err := s.Store.GetEquipList(ctx, s.SafeDB, jno, sno)
	if err != nil {
		return entity.Equips{}, utils.CustomErrorf(err)
	}
	return list, nil
}
func (s *ServiceEquip) GetEquip(ctx context.Context, jno int64, sno int64, recordDate time.Time) (entity.Equips, error) {

	equips, err := s.Store.GetEquip(ctx, s.SafeDB, jno, sno, recordDate)
	if err != nil {
		return entity.Equips{}, utils.CustomErrorf(err)
	}
	return equips, nil
}

func (s *ServiceEquip) MergeEquipCnt(ctx context.Context, equip entity.Equip) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	if err = s.Store.MergeEquipCnt(ctx, tx, equip); err != nil {
		return utils.CustomErrorf(err)
	}
	return
}
