package service

import (
	"context"
	"csm-api/auth"
	"csm-api/entity"
	"csm-api/store"
	"csm-api/txutil"
	"csm-api/utils"
	"fmt"
)

type ServiceNotice struct {
	SafeDB  store.Queryer
	SafeTDB store.Beginner
	Store   store.NoticeStore
}

// func: 공지사항 전체 조회
// @param
// - page entity.PageSql : 현재 페이지번호, 리스트 목록 개수
func (s *ServiceNotice) GetNoticeList(ctx context.Context, page entity.Page, isRole bool, search entity.Notice) (*entity.Notices, error) {

	// 사용자 정보 가져오기
	uno, _ := auth.GetContext(ctx, auth.Uno{})

	// 페이지 변환
	pageSql := entity.PageSql{}
	pageSql, err := pageSql.OfPageSql(page)

	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	notices, err := s.Store.GetNoticeList(ctx, s.SafeDB, uno, isRole, pageSql, search)
	if err != nil {
		return &entity.Notices{}, utils.CustomErrorf(err)
	}

	return notices, nil
}

// func: 공지사항 전체 개수 조회
// @param
// -
func (s *ServiceNotice) GetNoticeListCount(ctx context.Context, isRole bool, search entity.Notice) (int, error) {

	uno, _ := auth.GetContext(ctx, auth.Uno{})

	count, err := s.Store.GetNoticeListCount(ctx, s.SafeDB, uno, isRole, search)
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}

	return count, nil

}

// func: 공지사항 추가
// @param
// - notice entity.Notice: JNO, TITLE, CONTENT, SHOW_YN, PERIOD_CODE, REG_UNO, REG_USER
func (s *ServiceNotice) AddNotice(ctx context.Context, notice entity.Notice) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	if err = s.Store.AddNotice(ctx, tx, notice); err != nil {
		return fmt.Errorf("service_notice/AddNotice err : %w", err)
	}

	return
}

// func: 공지사항 수정
// @param
// -notice entity.Notice: IDX, JNO, TITLE, CONTENT, SHOW_YN, MOD_UNO, MOD_USER
func (s *ServiceNotice) ModifyNotice(ctx context.Context, notice entity.Notice) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	if err = s.Store.ModifyNotice(ctx, tx, notice); err != nil {
		return utils.CustomErrorf(err)
	}

	return
}

// func: 공지사항 삭제
// @param
// - IDX: 공지사항 인덱스
func (s *ServiceNotice) RemoveNotice(ctx context.Context, idx int64) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	if err = s.Store.RemoveNotice(ctx, tx, idx); err != nil {
		return utils.CustomErrorf(err)
	}

	return
}
