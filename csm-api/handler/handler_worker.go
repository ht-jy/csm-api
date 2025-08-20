package handler

import (
	"csm-api/entity"
	"csm-api/service"
	"csm-api/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

/**
 * @author 작성자: 김진우
 * @created 작성일: 2025-02-17
 * @modified 최종 수정일:
 * @modifiedBy 최종 수정자:
 * @modified description
 * -
 */

type HandlerWorker struct {
	Service service.WorkerService
}

// func: 전체 근로자 조회
// @param
// - response: http get paramter
func (h *HandlerWorker) TotalList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// http get paramter를 저장할 구조체 생성 및 파싱
	page := entity.Page{}
	search := entity.Worker{}

	pageNum := r.URL.Query().Get("page_num")
	rowSize := r.URL.Query().Get("row_size")
	order := r.URL.Query().Get("order")
	rnumOrder := r.URL.Query().Get("rnum_order")
	jobName := r.URL.Query().Get("job_name")
	userId := r.URL.Query().Get("user_id")
	userNm := r.URL.Query().Get("user_nm")
	department := r.URL.Query().Get("department")
	phone := r.URL.Query().Get("phone")
	workerType := r.URL.Query().Get("worker_type")
	discName := r.URL.Query().Get("disc_name")

	retrySearch := r.URL.Query().Get("retry_search")

	if pageNum == "" || rowSize == "" {
		BadRequestResponse(ctx, w)
		return
	}

	page.PageNum, _ = strconv.Atoi(pageNum)
	page.RowSize, _ = strconv.Atoi(rowSize)
	page.Order = order
	page.RnumOrder = rnumOrder
	search.JobName = utils.ParseNullString(jobName)
	search.UserId = utils.ParseNullString(userId)
	search.UserNm = utils.ParseNullString(userNm)
	search.Department = utils.ParseNullString(department)
	search.Phone = utils.ParseNullString(phone)
	search.WorkerType = utils.ParseNullString(workerType)
	search.DiscName = utils.ParseNullString(discName)

	// 조회
	list, err := h.Service.GetWorkerTotalList(ctx, page, search, retrySearch)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	// 개수 조회
	count, err := h.Service.GetWorkerTotalCount(ctx, search, retrySearch)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	values := struct {
		List  entity.Workers `json:"list"`
		Count int            `json:"count"`
	}{List: *list, Count: count}
	SuccessValuesResponse(ctx, w, values)
}

// func: 미출근 근로자 검색
// @param
// -
func (h *HandlerWorker) AbsentList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := entity.Page{}
	search := entity.WorkerDaily{}
	pageNum := r.URL.Query().Get("page_num")
	rowSize := r.URL.Query().Get("row_size")
	retrySearch := r.URL.Query().Get("retry_search")
	searchStartTime := r.URL.Query().Get("search_start_time")
	jno := r.URL.Query().Get("jno")
	sno := r.URL.Query().Get("sno")

	if pageNum == "" || rowSize == "" || jno == "" {
		BadRequestResponse(ctx, w)
		return
	}
	page.PageNum, _ = strconv.Atoi(pageNum)
	page.RowSize, _ = strconv.Atoi(rowSize)
	search.Jno = utils.ParseNullInt(jno)
	search.Sno = utils.ParseNullInt(sno)
	search.SearchStartTime = utils.ParseNullString(searchStartTime)

	list, err := h.Service.GetAbsentWorkerList(ctx, page, search, retrySearch)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	count, err := h.Service.GetAbsentWorkerCount(ctx, search, retrySearch)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	values := struct {
		List  entity.Workers `json:"list"`
		Count int            `json:"count"`
	}{List: *list, Count: count}
	SuccessValuesResponse(ctx, w, values)
}

// 프로젝트에 참여한 회사명 리스트
func (h *HandlerWorker) DepartList(w http.ResponseWriter, r *http.Request) {
	jnoString := r.URL.Query().Get("jno")
	if jnoString == "" {
		BadRequestResponse(r.Context(), w)
		return
	}

	jno, _ := strconv.ParseInt(jnoString, 10, 64)
	list, err := h.Service.GetWorkerDepartList(r.Context(), jno)
	if err != nil {
		FailResponse(r.Context(), w, err)
		return
	}
	SuccessValuesResponse(r.Context(), w, list)
}

// func: 근로자 추가
// @param
// - http method: post
// - param: entity.Worker - json(raw)
func (h *HandlerWorker) Add(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	//데이터 파싱
	worker := entity.Worker{}
	if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	err := h.Service.AddWorker(ctx, worker)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	SuccessResponse(ctx, w)
}

// func: 근로자 수정
// @param
// - http method: put
// - param: entity.Worker - json(raw)
func (h *HandlerWorker) Modify(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	//데이터 파싱
	worker := entity.Worker{}
	if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	err := h.Service.ModifyWorker(ctx, worker)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	SuccessResponse(ctx, w)
}

// 근로자 삭제
func (h *HandlerWorker) Remove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logEntry, worker, err := entity.DecodeItem(r, entity.Worker{})
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	if err = h.Service.RemoveWorker(ctx, worker); err != nil {
		FailResponse(ctx, w, err)
		return
	}
	entity.WriteLog(logEntry)
	SuccessResponse(ctx, w)
}

// func: 현장 근로자 조회
// @param
// - response: http get paramter
func (h *HandlerWorker) SiteBaseList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// http get paramter를 저장할 구조체 생성 및 파싱
	page := entity.Page{}
	search := entity.WorkerDaily{}

	pageNum := r.URL.Query().Get("page_num")
	rowSize := r.URL.Query().Get("row_size")
	order := r.URL.Query().Get("order")
	rnumOrder := r.URL.Query().Get("rnum_order")
	retrySearch := r.URL.Query().Get("retry_search")
	jno := r.URL.Query().Get("jno")
	userId := r.URL.Query().Get("user_id")
	userNm := r.URL.Query().Get("user_nm")
	department := r.URL.Query().Get("department")
	searchStartTime := r.URL.Query().Get("search_start_time")
	searchEndTime := r.URL.Query().Get("search_end_time")

	if pageNum == "" || rowSize == "" || searchStartTime == "" || searchEndTime == "" || jno == "" {
		BadRequestResponse(ctx, w)
		return
	}

	page.PageNum, _ = strconv.Atoi(pageNum)
	page.RowSize, _ = strconv.Atoi(rowSize)
	page.Order = order
	page.RnumOrder = rnumOrder
	search.Jno = utils.ParseNullInt(jno)
	search.UserId = utils.ParseNullString(userId)
	search.UserNm = utils.ParseNullString(userNm)
	search.Department = utils.ParseNullString(department)
	search.SearchStartTime = utils.ParseNullString(searchStartTime)
	search.SearchEndTime = utils.ParseNullString(searchEndTime)

	// 조회
	list, err := h.Service.GetWorkerSiteBaseList(ctx, page, search, retrySearch)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	// 개수 조회
	count, err := h.Service.GetWorkerSiteBaseCount(ctx, search, retrySearch)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	values := struct {
		List  entity.WorkerDailys `json:"list"`
		Count int                 `json:"count"`
	}{List: *list, Count: count}
	SuccessValuesResponse(ctx, w, values)
}

// func: 현장근로자 추가/수정
// @param
// - http method: post
// - param: entity.WorkerDailys - json(raw)
func (h *HandlerWorker) Merge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	//데이터 파싱
	workers := entity.WorkerDailys{}
	if err := json.NewDecoder(r.Body).Decode(&workers); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	err := h.Service.MergeSiteBaseWorker(ctx, workers)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}
	SuccessResponse(ctx, w)
}

// func: 근로자 일괄마감
// @param
// - http method: post
// - param: entity.WorkerDailys - json(raw)
func (h *HandlerWorker) ModifyDeadline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//데이터 파싱
	workers := entity.WorkerDailys{}
	if err := json.NewDecoder(r.Body).Decode(&workers); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	err := h.Service.ModifyWorkerDeadline(ctx, workers)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	SuccessResponse(ctx, w)
}

// func: 현장 근로자 프로젝트 변경
// @param
// - http method: post
// - param: entity.WorkerDailys - json(raw)
func (h *HandlerWorker) ModifyProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//데이터 파싱
	workers := entity.WorkerDailys{}
	if err := json.NewDecoder(r.Body).Decode(&workers); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	err := h.Service.ModifyWorkerProject(ctx, workers)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	SuccessResponse(ctx, w)
}

// func: 현장근로자 삭제
// @param
// - http method: post
// - param: entity.WorkerDailys - json(raw)
func (h *HandlerWorker) SiteBaseRemove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	//데이터 파싱
	workers := entity.WorkerDailys{}
	if err := json.NewDecoder(r.Body).Decode(&workers); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	err := h.Service.RemoveSiteBaseWorkers(ctx, workers)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}
	SuccessResponse(ctx, w)
}

// func: 마감 취소
// @param
// - http method: post
// - param: entity.WorkerDailys - json(raw)
func (h *HandlerWorker) SiteBaseDeadlineCancel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	//데이터 파싱
	workers := entity.WorkerDailys{}
	if err := json.NewDecoder(r.Body).Decode(&workers); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	err := h.Service.ModifyDeadlineCancel(ctx, workers)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}
	SuccessResponse(ctx, w)
}

// 프로젝트, 기간내 모든 현장근로자 근태정보 조회
func (h *HandlerWorker) DailyWorkersByJnoAndDate(w http.ResponseWriter, r *http.Request) {
	jnoString := r.URL.Query().Get("jno")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	if jnoString == "" || startDate == "" || endDate == "" {
		BadRequestResponse(r.Context(), w)
		return
	}

	param := entity.RecordDailyWorkerReq{
		Jno:       utils.ParseNullInt(jnoString),
		StartDate: utils.ParseNullString(startDate),
		EndDate:   utils.ParseNullString(endDate),
	}

	list, err := h.Service.GetDailyWorkersByJnoAndDate(r.Context(), param)
	if err != nil {
		FailResponse(r.Context(), w, err)
		return
	}
	SuccessValuesResponse(r.Context(), w, list)
}

// 현장근로자 일괄 공수 변경
func (h *HandlerWorker) ModifyWorkHours(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logEntry, workers, err := entity.DecodeItem(r, entity.WorkerDailys{})
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	if err = h.Service.ModifyWorkHours(ctx, workers); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	entity.WriteLog(logEntry)
	SuccessResponse(ctx, w)
}

// 변경이력 조회
func (h *HandlerWorker) GetDailyWorkerHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	snoString := r.URL.Query().Get("sno")
	retrySearch := r.URL.Query().Get("retry_search")
	if startDate == "" || endDate == "" || snoString == "" {
		BadRequestResponse(ctx, w)
		return
	}

	userKeys := r.URL.Query()["keys"]

	sno, _ := strconv.ParseInt(snoString, 10, 64)

	list, err := h.Service.GetHistoryDailyWorkers(ctx, startDate, endDate, sno, retrySearch, userKeys)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}
	SuccessValuesResponse(ctx, w, list)
}

// 변경 이력 사유 조회
func (h *HandlerWorker) GetDailyWorkerHistoryReason(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cnoString := r.URL.Query().Get("cno")
	if cnoString == "" {
		BadRequestResponse(ctx, w)
		return
	}
	cno, _ := strconv.ParseInt(cnoString, 10, 64)
	list, err := h.Service.GetHistoryDailyWorkerReason(ctx, cno)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}
	SuccessValuesResponse(ctx, w, list)
}
