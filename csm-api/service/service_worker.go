package service

import (
	"context"
	"csm-api/entity"
	"csm-api/store"
	"fmt"
)

/**
 * @author 작성자: 김진우
 * @created 작성일: 2025-02-17
 * @modified 최종 수정일:
 * @modifiedBy 최종 수정자:
 * @modified description
 * -
 */

type ServiceWorker struct {
	DB    store.Queryer
	TDB   store.Beginner
	Store store.WorkerStore
}

// func: 전체 근로자 조회
// @param
// - page entity.PageSql: 정렬, 리스트 수
// - search entity.WorkerSql: 검색 단어
// - retry string: 통합검색 텍스트
func (s *ServiceWorker) GetWorkerTotalList(ctx context.Context, page entity.Page, search entity.Worker, retry string) (*entity.Workers, error) {
	// regular type ->  sql type 변환
	pageSql := entity.PageSql{}
	pageSql, err := pageSql.OfPageSql(page)
	if err != nil {
		//TODO: 에러 아카이브
		return nil, fmt.Errorf("service_worker;total/OfPageSql err: %v", err)
	}

	// 조회
	list, err := s.Store.GetWorkerTotalList(ctx, s.DB, pageSql, search, retry)
	if err != nil {
		//TODO: 에러 아카이브
		return nil, fmt.Errorf("service_worker/GetWorkerTotalList err: %v", err)
	}

	return list, nil
}

// func: 전체 근로자 개수 조회
// @param
// - searchTime string: 조회 날짜
// - retry string: 통합검색 텍스트
func (s *ServiceWorker) GetWorkerTotalCount(ctx context.Context, search entity.Worker, retry string) (int, error) {
	count, err := s.Store.GetWorkerTotalCount(ctx, s.DB, search, retry)
	if err != nil {
		//TODO: 에러 아카이브
		return 0, fmt.Errorf("service_worker/GetWorkerTotalCount err: %v", err)
	}
	return count, nil
}

// func: 근로자 검색(현장근로자 추가시 사용)
// @param
// - userId string
func (s *ServiceWorker) GetWorkerListByUserId(ctx context.Context, page entity.Page, search entity.WorkerDaily, retry string) (*entity.Workers, error) {
	// regular type ->  sql type 변환
	pageSql := entity.PageSql{}
	pageSql, err := pageSql.OfPageSql(page)
	if err != nil {
		//TODO: 에러 아카이브
		return nil, fmt.Errorf("service_worker;ByUserId/OfPageSql err: %v", err)
	}

	// 조회
	list, err := s.Store.GetWorkerListByUserId(ctx, s.DB, pageSql, search, retry)
	if err != nil {
		//TODO: 에러 아카이브
		return nil, fmt.Errorf("service_worker/GetWorkerListByUserId err: %v", err)
	}

	return list, nil
}

// func: 근로자 개수 검색(현장근로자 추가시 사용)
// @param
// - userId string
func (s *ServiceWorker) GetWorkerCountByUserId(ctx context.Context, search entity.WorkerDaily, retry string) (int, error) {
	count, err := s.Store.GetWorkerCountByUserId(ctx, s.DB, search, retry)
	if err != nil {
		//TODO: 에러 아카이브
		return 0, fmt.Errorf("service_worker;ByUserId/StoreGetWorkerCountByUserId err: %v", err)
	}
	return count, nil
}

// func: 근로자 추가
// @param
// -
func (s *ServiceWorker) AddWorker(ctx context.Context, worker entity.Worker) error {
	err := s.Store.AddWorker(ctx, s.TDB, worker)
	if err != nil {
		//TODO: 에러 아카이브
		return fmt.Errorf("service_worker/AddWorker err: %v", err)
	}
	return nil
}

// func: 근로자 수정
// @param
// -
func (s *ServiceWorker) ModifyWorker(ctx context.Context, worker entity.Worker) error {
	err := s.Store.ModifyWorker(ctx, s.TDB, worker)
	if err != nil {
		//TODO: 에러 아카이브
		return fmt.Errorf("service_worker/ModifyWorker err: %v", err)
	}
	return nil
}

// func: 현장 근로자 조회
// @param
// - page entity.PageSql: 정렬, 리스트 수
// - search entity.WorkerSql: 검색 단어
func (s *ServiceWorker) GetWorkerSiteBaseList(ctx context.Context, page entity.Page, search entity.WorkerDaily, retry string) (*entity.WorkerDailys, error) {
	// regular type ->  sql type 변환
	pageSql := entity.PageSql{}
	pageSql, err := pageSql.OfPageSql(page)
	if err != nil {
		//TODO: 에러 아카이브
		return nil, fmt.Errorf("service_worker;site_base/OfPageSql err: %v", err)
	}

	// 조회
	list, err := s.Store.GetWorkerSiteBaseList(ctx, s.DB, pageSql, search, retry)
	if err != nil {
		//TODO: 에러 아카이브
		return nil, fmt.Errorf("service_worker/GetWorkerSiteBaseList err: %v", err)
	}

	return list, nil
}

// func: 현장 근로자 개수 조회
// @param
// - searchTime string: 조회 날짜
func (s *ServiceWorker) GetWorkerSiteBaseCount(ctx context.Context, search entity.WorkerDaily, retry string) (int, error) {
	count, err := s.Store.GetWorkerSiteBaseCount(ctx, s.DB, search, retry)
	if err != nil {
		//TODO: 에러 아카이브
		return 0, fmt.Errorf("service_worker/GetWorkerSiteBaseCount err: %v", err)
	}
	return count, nil
}

// func: 현장 근로자 추가/수정
// @param
// -
func (s *ServiceWorker) MergeSiteBaseWorker(ctx context.Context, workers entity.WorkerDailys) error {
	if err := s.Store.MergeSiteBaseWorker(ctx, s.TDB, workers); err != nil {
		//TODO: 에러 아카이브
		return fmt.Errorf("service_worker/MergeSiteBaseWorker err: %v", err)
	}

	return nil
}

// func: 현장 근로자 일괄마감
// @param
// -
func (s *ServiceWorker) ModifyWorkerDeadline(ctx context.Context, workers entity.WorkerDailys) error {
	if err := s.Store.ModifyWorkerDeadline(ctx, s.TDB, workers); err != nil {
		//TODO: 에러 아카이브
		return fmt.Errorf("service_worker/ModifyWorkerDeadline err: %v", err)
	}
	return nil
}

// func: 현장 근로자 프로젝트 변경
// @param
// -
func (s *ServiceWorker) ModifyWorkerProject(ctx context.Context, workers entity.WorkerDailys) error {
	if err := s.Store.ModifyWorkerProject(ctx, s.TDB, workers); err != nil {
		//TODO: 에러 아카이브
		return fmt.Errorf("service_worker/ModifyWorkerProject err: %v", err)
	}
	return nil
}
