package service

import (
	"context"
	"csm-api/entity"
	"csm-api/store"
	"csm-api/txutil"
	"csm-api/utils"
	"fmt"
	"github.com/guregu/null"
	"time"
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
	SafeDB  store.Queryer
	SafeTDB store.Beginner
	Store   store.WorkerStore
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
		return nil, utils.CustomErrorf(err)
	}

	// 조회
	list, err := s.Store.GetWorkerTotalList(ctx, s.SafeDB, pageSql, search, retry)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return list, nil
}

// func: 전체 근로자 개수 조회
// @param
// - searchTime string: 조회 날짜
// - retry string: 통합검색 텍스트
func (s *ServiceWorker) GetWorkerTotalCount(ctx context.Context, search entity.Worker, retry string) (int, error) {
	count, err := s.Store.GetWorkerTotalCount(ctx, s.SafeDB, search, retry)
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}
	return count, nil
}

// func: 미출근 근로자 검색(현장근로자 추가시 사용)
// @param
// - userId string
func (s *ServiceWorker) GetAbsentWorkerList(ctx context.Context, page entity.Page, search entity.WorkerDaily, retry string) (*entity.Workers, error) {
	// regular type ->  sql type 변환
	pageSql := entity.PageSql{}
	pageSql, err := pageSql.OfPageSql(page)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	// 조회
	list, err := s.Store.GetAbsentWorkerList(ctx, s.SafeDB, pageSql, search, retry)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return list, nil
}

// func: 근로자 개수 검색(현장근로자 추가시 사용)
// @param
// - userId string
func (s *ServiceWorker) GetAbsentWorkerCount(ctx context.Context, search entity.WorkerDaily, retry string) (int, error) {
	count, err := s.Store.GetAbsentWorkerCount(ctx, s.SafeDB, search, retry)
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}
	return count, nil
}

// 프로젝트에 참여한 회사명 리스트
func (s *ServiceWorker) GetWorkerDepartList(ctx context.Context, jno int64) ([]string, error) {
	list, err := s.Store.GetWorkerDepartList(ctx, s.SafeDB, jno)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}
	return list, nil
}

// func: 근로자 추가
// @param
// -
func (s *ServiceWorker) AddWorker(ctx context.Context, worker entity.Worker) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	var count int64
	count, err = s.Store.AddWorker(ctx, tx, worker)
	if err != nil {
		return utils.CustomErrorf(err)
	} else if count == 0 {
		return utils.CustomErrorf(fmt.Errorf("중복데이터 존재"))
	}
	return
}

// func: 근로자 수정
// @param
// -
func (s *ServiceWorker) ModifyWorker(ctx context.Context, worker entity.Worker) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	err = s.Store.ModifyWorker(ctx, tx, worker)
	if err != nil {
		return utils.CustomErrorf(err)
	}
	return
}

// 근로자 삭제
func (s *ServiceWorker) RemoveWorker(ctx context.Context, worker entity.Worker) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}
	defer txutil.DeferTx(tx, &err)

	err = s.Store.RemoveWorker(ctx, tx, worker)
	if err != nil {
		return utils.CustomErrorf(err)
	}
	return
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
		return nil, utils.CustomErrorf(err)
	}

	// 조회
	list, err := s.Store.GetWorkerSiteBaseList(ctx, s.SafeDB, pageSql, search, retry)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return list, nil
}

// func: 현장 근로자 개수 조회
// @param
// - searchTime string: 조회 날짜
func (s *ServiceWorker) GetWorkerSiteBaseCount(ctx context.Context, search entity.WorkerDaily, retry string) (int, error) {
	count, err := s.Store.GetWorkerSiteBaseCount(ctx, s.SafeDB, search, retry)
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}
	return count, nil
}

// func: 현장 근로자 추가/수정
// @param
// -
func (s *ServiceWorker) MergeSiteBaseWorker(ctx context.Context, workers entity.WorkerDailys) (err error) {
	// 변경전 데이터 조회
	beforeList, err := s.Store.GetDailyWorkerBeforeList(ctx, s.SafeDB, workers)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// 추가/수정
	if err = s.Store.MergeSiteBaseWorker(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 변경사항 로그 저장
	if err = s.Store.MergeSiteBaseWorkerLog(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 변경이력 저장
	regDate := null.NewTime(time.Now(), true)
	// 변경전
	for i := range beforeList {
		beforeList[i].HisStatus = utils.ParseNullString("BEFORE")
		beforeList[i].RegDate = regDate
	}
	if err = s.Store.AddHistoryDailyWorkers(ctx, tx, beforeList); err != nil {
		return utils.CustomErrorf(err)
	}
	// 변경후
	for i := range workers {
		workers[i].HisStatus = utils.ParseNullString("AFTER")
		workers[i].RegDate = regDate
	}
	if err = s.Store.AddHistoryDailyWorkers(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	return
}

// func: 현장 근로자 일괄마감
// @param
// -
func (s *ServiceWorker) ModifyWorkerDeadline(ctx context.Context, workers entity.WorkerDailys) (err error) {
	// 변경전 데이터 조회
	beforeList, err := s.Store.GetDailyWorkerBeforeList(ctx, s.SafeDB, workers)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// 마감처리
	if err = s.Store.ModifyWorkerDeadline(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 마감 로그 저장
	if err = s.Store.MergeSiteBaseWorkerLog(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 변경이력 저장
	regDate := null.NewTime(time.Now(), true)
	// 변경전
	for i := range beforeList {
		beforeList[i].HisStatus = utils.ParseNullString("BEFORE")
		beforeList[i].RegDate = regDate
	}
	if err = s.Store.AddHistoryDailyWorkers(ctx, tx, beforeList); err != nil {
		return utils.CustomErrorf(err)
	}
	// 변경후
	for i := range workers {
		workers[i].HisStatus = utils.ParseNullString("AFTER")
		workers[i].RegDate = regDate
	}
	if err = s.Store.AddHistoryDailyWorkers(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	return
}

// func: 현장 근로자 프로젝트 변경
// @param
// -
func (s *ServiceWorker) ModifyWorkerProject(ctx context.Context, workers entity.WorkerDailys) (err error) {
	// 변경전 데이터 조회
	beforeList, err := s.Store.GetDailyWorkerBeforeList(ctx, s.SafeDB, workers)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// 전체 근로자 프로젝트 변경
	if err = s.Store.ModifyWorkerDefaultProject(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 현장 근로자 프로젝트 변경
	if err = s.Store.ModifyWorkerProject(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 프로젝트 변경 로그 저장
	if err = s.Store.MergeSiteBaseWorkerLog(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 변경이력 저장
	regDate := null.NewTime(time.Now(), true)
	// 변경전
	for i := range beforeList {
		beforeList[i].HisStatus = utils.ParseNullString("BEFORE")
		beforeList[i].RegDate = regDate
	}
	if err = s.Store.AddHistoryDailyWorkers(ctx, tx, beforeList); err != nil {
		return utils.CustomErrorf(err)
	}
	// 변경후
	for i := range workers {
		workers[i].HisStatus = utils.ParseNullString("AFTER")
		workers[i].RegDate = regDate
		workers[i].Jno = workers[i].AfterJno
	}
	if err = s.Store.AddHistoryDailyWorkers(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	return
}

// func: 현장 근로자 일일 마감처리
// @param
// -
func (s *ServiceWorker) ModifyWorkerDeadlineInit(ctx context.Context) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	if err = s.Store.ModifyWorkerDeadlineInit(ctx, tx); err != nil {
		return utils.CustomErrorf(err)
	}
	return
}

// func: 현장 근로자 철야 처리
// @param
// -
func (s *ServiceWorker) ModifyWorkerOverTime(ctx context.Context) (count int, err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// 철야 근로자 존재 여부 확인
	workerOverTimes := &entity.WorkerOverTimes{}
	workerOverTimes, err = s.Store.GetWorkerOverTime(ctx, s.SafeDB)
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}
	count = len(*workerOverTimes)

	for _, workerOverTime := range *workerOverTimes {

		// 철야 근로자 철야 표시 및 퇴근시간 합치기.
		if err = s.Store.ModifyWorkerOverTime(ctx, tx, *workerOverTime); err != nil {
			return 0, utils.CustomErrorf(err)
		}

		// 다음날 퇴근 표시 삭제
		if err = s.Store.DeleteWorkerOverTime(ctx, tx, (*workerOverTime).AfterCno); err != nil {
			return 0, utils.CustomErrorf(err)
		}
	}
	return
}

// 현장 근로자 삭제
func (s *ServiceWorker) RemoveSiteBaseWorkers(ctx context.Context, workers entity.WorkerDailys) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// 현장 근로자 삭제
	if err = s.Store.RemoveSiteBaseWorkers(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 삭제 로그 저장
	if err = s.Store.MergeSiteBaseWorkerLog(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 변경이력 저장
	regDate := null.NewTime(time.Now(), true)
	// 변경전
	for i := range workers {
		workers[i].HisStatus = utils.ParseNullString("BEFORE")
		workers[i].RegDate = regDate
	}
	if err = s.Store.AddHistoryDailyWorkers(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}
	// 변경후
	for i := range workers {
		workers[i].HisStatus = utils.ParseNullString("AFTER")
		workers[i].Sno = null.NewInt(0, false)
		workers[i].Jno = null.NewInt(0, false)
		workers[i].RecordDate = null.NewTime(time.Time{}, false)
		workers[i].InRecogTime = null.NewTime(time.Time{}, false)
		workers[i].OutRecogTime = null.NewTime(time.Time{}, false)
		workers[i].IsDeadline = null.NewString("", false)
		workers[i].WorkState = null.NewString("", false)
		workers[i].IsOvertime = null.NewString("", false)
		workers[i].WorkHour = null.NewFloat(0, false)
	}
	if err = s.Store.AddHistoryDailyWorkers(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	return
}

// 마감 취소
func (s *ServiceWorker) ModifyDeadlineCancel(ctx context.Context, workers entity.WorkerDailys) (err error) {
	// 변경전 데이터 조회
	beforeList, err := s.Store.GetDailyWorkerBeforeList(ctx, s.SafeDB, workers)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// 마감 취소
	if err = s.Store.ModifyDeadlineCancel(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 마감 취소 로그 저장
	if err = s.Store.MergeSiteBaseWorkerLog(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 변경이력 저장
	regDate := null.NewTime(time.Now(), true)
	// 변경전
	for i := range beforeList {
		beforeList[i].HisStatus = utils.ParseNullString("BEFORE")
		beforeList[i].RegDate = regDate
	}
	if err = s.Store.AddHistoryDailyWorkers(ctx, tx, beforeList); err != nil {
		return utils.CustomErrorf(err)
	}
	// 변경후
	for i := range workers {
		workers[i].HisStatus = utils.ParseNullString("AFTER")
		workers[i].RegDate = regDate
	}
	if err = s.Store.AddHistoryDailyWorkers(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	return
}

// 프로젝트, 기간내 모든 현장근로자 근태정보 조회
func (s *ServiceWorker) GetDailyWorkersByJnoAndDate(ctx context.Context, param entity.RecordDailyWorkerReq) ([]entity.RecordDailyWorkerRes, error) {
	list, err := s.Store.GetDailyWorkersByJnoAndDate(ctx, s.SafeDB, param)
	if err != nil {
		return []entity.RecordDailyWorkerRes{}, utils.CustomErrorf(err)
	}
	return list, nil
}

// 현장근로자 일괄 공수 변경
func (s *ServiceWorker) ModifyWorkHours(ctx context.Context, workers entity.WorkerDailys) (err error) {
	// 변경전 데이터 조회
	beforeList, err := s.Store.GetDailyWorkerBeforeList(ctx, s.SafeDB, workers)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// 공수 변경
	if err = s.Store.ModifyWorkHours(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 공수 변경 로그 저장
	if err = s.Store.MergeSiteBaseWorkerLog(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 변경이력 저장
	regDate := null.NewTime(time.Now(), true)
	// 변경전
	for i := range beforeList {
		beforeList[i].HisStatus = utils.ParseNullString("BEFORE")
		beforeList[i].RegDate = regDate
	}
	if err = s.Store.AddHistoryDailyWorkers(ctx, tx, beforeList); err != nil {
		return utils.CustomErrorf(err)
	}
	// 변경후
	for i := range workers {
		workers[i].HisStatus = utils.ParseNullString("AFTER")
		workers[i].RegDate = regDate
	}
	if err = s.Store.AddHistoryDailyWorkers(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	return
}

// 홍채인식기데이터 전체근로자 테이블(IRIS_WORKER_SET)에 반영::스케줄 용도
func (s *ServiceWorker) MergeRecdWorker(ctx context.Context) (err error) {
	// 미반영 데이터 조회
	recdList, err := s.Store.GetRecdWorkerList(ctx, s.SafeDB)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	if len(recdList) == 0 {
		return
	}

	// 근로자 키 조회 및 생성
	for i := range recdList {
		userKey, err := s.Store.GetRecdWorkerUserKey(ctx, s.SafeDB, recdList[i])
		if err != nil {
			return utils.CustomErrorf(err)
		}
		recdList[i].UserKey = utils.ParseNullString(userKey)
	}

	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// 전체근로자 테이블에 반영
	if err = s.Store.MergeRecdWorker(ctx, tx, recdList); err != nil {
		return utils.CustomErrorf(err)
	}

	return
}

// 홍채인식기데이터 현장근로자 테이블(IRIS_WORKER_DAILY_SET)에 반영::스케줄 용도
func (s *ServiceWorker) MergeRecdDailyWorker(ctx context.Context) (err error) {
	// 미방영 데이터 조회
	recdList, err := s.Store.GetRecdDailyWorkerList(ctx, s.SafeDB)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	if len(recdList) == 0 {
		return
	}

	for i := range recdList {
		// 근로자 키 조회 및 생성
		temp := entity.Worker{
			UserId: recdList[i].UserId,
			UserNm: recdList[i].UserNm,
			RegNo:  recdList[i].RegNo,
		}
		userKey, err := s.Store.GetRecdWorkerUserKey(ctx, s.SafeDB, temp)
		if err != nil {
			return utils.CustomErrorf(err)
		}
		recdList[i].UserKey = utils.ParseNullString(userKey)

		// 출퇴근 시간
		if !recdList[i].RecordDate.Valid { // 시간 기록이 없는 경우 패스
			continue
		} else {
			// 출근 기록 확인
			isChk, err := s.Store.GetRecdDailyWorkerChk(ctx, s.SafeDB, userKey, recdList[i].RecordDate)
			if err != nil {
				return utils.CustomErrorf(err)
			}
			if isChk { // 출근을 한 경우
				recdList[i].OutRecogTime = recdList[i].RecordDate
			} else { // 출근이 없는 경우
				cmp := time.Date(recdList[i].RecordDate.Time.Year(), recdList[i].RecordDate.Time.Month(), recdList[i].RecordDate.Time.Day(), 15, 0, 0, 0, time.Local)
				if recdList[i].RecordDate.Time.Before(cmp) {
					recdList[i].InRecogTime = recdList[i].RecordDate
				} else {
					recdList[i].OutRecogTime = recdList[i].RecordDate
				}
			}

			// 출퇴근 상태
			if recdList[i].InRecogTime.Valid {
				recdList[i].WorkState = utils.ParseNullString("01")
			} else {
				recdList[i].WorkState = utils.ParseNullString("02")
			}
		}
	}

	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}
	defer txutil.DeferTx(tx, &err)

	// 현장근로자 테이블에 반영
	if err = s.Store.MergeRecdDailyWorker(ctx, tx, recdList); err != nil {
		return utils.CustomErrorf(err)
	}
	return
}

// 변경 이력 조회
func (s *ServiceWorker) GetHistoryDailyWorkers(ctx context.Context, startDate string, endDate string, sno int64, retry string, userKeys []string) (entity.WorkerDailys, error) {
	list, err := s.Store.GetHistoryDailyWorkers(ctx, s.SafeDB, startDate, endDate, sno, retry, userKeys)
	if err != nil {
		return entity.WorkerDailys{}, utils.CustomErrorf(err)
	}
	return list, nil
}

// 변경 이력 사유 조회
func (s *ServiceWorker) GetHistoryDailyWorkerReason(ctx context.Context, cno int64) (string, error) {
	reason, err := s.Store.GetHistoryDailyWorkerReason(ctx, s.SafeDB, cno)
	if err != nil {
		return "", utils.CustomErrorf(err)
	}
	return reason, nil
}
