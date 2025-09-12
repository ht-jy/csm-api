package handler

import (
	"csm-api/entity"
	"csm-api/service"
	"csm-api/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

type HandlerCompare struct {
	Service service.CompareService
}

// 일일근로자비교 리스트
func (h *HandlerCompare) List(w http.ResponseWriter, r *http.Request) {
	snoString := r.URL.Query().Get("sno")
	jnoString := r.URL.Query().Get("jno")
	startDateString := r.URL.Query().Get("start_date")
	order := r.URL.Query().Get("order")
	retrySearch := r.URL.Query().Get("retry_search")

	// 전체 조회할지 본인이 속한 프로젝트만 조회할지 권한
	isRoleStr := r.URL.Query().Get("isRole")
	isRole, err := strconv.ParseBool(isRoleStr)
	if err != nil {
		BadRequestResponse(r.Context(), w)
		return
	}

	if snoString == "" || jnoString == "" || startDateString == "" {
		BadRequestResponse(r.Context(), w)
		return
	}

	compare := entity.Compare{
		Sno:        utils.ParseNullInt(snoString),
		Jno:        utils.ParseNullInt(jnoString),
		RecordDate: utils.ParseNullDate(startDateString),
	}

	list, err := h.Service.GetCompareList(r.Context(), compare, isRole, retrySearch, order)
	if err != nil {
		FailResponse(r.Context(), w, err)
		return
	}

	SuccessValuesResponse(r.Context(), w, list)
}

// 일일근로자 비교 반영
func (h *HandlerCompare) CompareState(w http.ResponseWriter, r *http.Request) {
	var workers entity.WorkerDailys

	if err := json.NewDecoder(r.Body).Decode(&workers); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}

	if err := h.Service.ModifyWorkerCompareApply(r.Context(), workers); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}
	SuccessResponse(r.Context(), w)
}
