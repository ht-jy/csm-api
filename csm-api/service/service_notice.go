package service

import (
	"context"
	"csm-api/entity"
	"csm-api/store"
	"fmt"
	"github.com/guregu/null"
)

type ServiceNotice struct {
	DB    store.Queryer
	TDB   store.Beginner
	Store store.NoticeStore
}

// func: 공지사항 전체 조회
// @param
// - page entity.PageSql : 현재 페이지번호, 리스트 목록 개수
func (s *ServiceNotice) GetNoticeList(ctx context.Context, uno null.Int, role null.String, page entity.Page, search entity.Notice) (*entity.Notices, error) {

	pageSql := entity.PageSql{}
	pageSql, err := pageSql.OfPageSql(page)

	if err != nil {
		//TODO: 에러 아카이브 처리
		return nil, fmt.Errorf("service_notice/GetNoticeList err : %w", err)
	}

	var roleInt int
	if role.String == "ADMIN" {
		roleInt = 1
	} else {
		roleInt = 0
	}
	notices, err := s.Store.GetNoticeList(ctx, s.DB, uno, roleInt, pageSql, search)
	if err != nil {
		//TODO: 에러 아카이브 처리
		return &entity.Notices{}, fmt.Errorf("fail to list notice: %w", err)
	}

	return notices, nil
}

// func: 공지사항 전체 개수 조회
// @param
// -
func (s *ServiceNotice) GetNoticeListCount(ctx context.Context, uno null.Int, role null.String, search entity.Notice) (int, error) {

	var roleInt int
	if role.String == "ADMIN" {
		roleInt = 1
	} else {
		roleInt = 0
	}

	count, err := s.Store.GetNoticeListCount(ctx, s.DB, uno, roleInt, search)
	if err != nil {
		//TODO: 에러 아카이브 처리
		return 0, fmt.Errorf("service_notice/GetNoticeListCount err : %w", err)
	}

	return count, nil

}

// func: 공지사항 추가
// @param
// - notice entity.Notice: JNO, TITLE, CONTENT, SHOW_YN, PERIOD_CODE, REG_UNO, REG_USER
func (s *ServiceNotice) AddNotice(ctx context.Context, notice entity.Notice) error {

	if err := s.Store.AddNotice(ctx, s.TDB, notice); err != nil {
		//TODO: 에러 아카이브 처리
		return fmt.Errorf("service_notice/AddNotice err : %w", err)
	}

	return nil
}

// func: 공지사항 수정
// @param
// -notice entity.Notice: IDX, JNO, TITLE, CONTENT, SHOW_YN, MOD_UNO, MOD_USER
func (s *ServiceNotice) ModifyNotice(ctx context.Context, notice entity.Notice) error {

	if err := s.Store.ModifyNotice(ctx, s.TDB, notice); err != nil {
		//TODO: 에러 아카이브 처리
		return fmt.Errorf("service_notice/ModifyNotice err: %w", err)
	}

	return nil
}

// func: 공지사항 삭제
// @param
// - IDX: 공지사항 인덱스
func (s *ServiceNotice) RemoveNotice(ctx context.Context, idx null.Int) error {

	if err := s.Store.RemoveNotice(ctx, s.TDB, idx); err != nil {
		//TODO: 에러 아카이브 처리
		return fmt.Errorf("service_notice/RemomveNotice err: %w", err)
	}

	return nil
}
