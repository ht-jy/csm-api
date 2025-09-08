package handler

import (
	"csm-api/entity"
	"csm-api/service"
	"csm-api/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

type HandlerProjectDaily struct {
	Service service.ProjectDailyService
}

// 작업내용 조회
func (h *HandlerProjectDaily) List(w http.ResponseWriter, r *http.Request) {

	isRoleStr := r.URL.Query().Get("isRole")
	isRole, err := strconv.ParseBool(isRoleStr)
	if err != nil {
		BadRequestResponse(r.Context(), w)
		return
	}

	targetDate := r.URL.Query().Get("target_date")
	jnoString := r.URL.Query().Get("jno")
	if targetDate == "" {
		BadRequestResponse(r.Context(), w)
		return
	}

	jno, _ := strconv.ParseInt(jnoString, 10, 64)
	list, err := h.Service.GetDailyJobList(r.Context(), isRole, jno, targetDate)
	if err != nil {
		FailResponse(r.Context(), w, err)
		return
	}

	values := struct {
		List entity.ProjectDailys `json:"list"`
	}{List: list}
	SuccessValuesResponse(r.Context(), w, values)
}

// 작업내용 추가
func (h *HandlerProjectDaily) Add(w http.ResponseWriter, r *http.Request) {
	projectDailys := entity.ProjectDailys{}

	if err := json.NewDecoder(r.Body).Decode(&projectDailys); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}

	if err := h.Service.AddDailyJob(r.Context(), projectDailys); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}

	SuccessResponse(r.Context(), w)
}

// 작업내용 수정
func (h *HandlerProjectDaily) Modify(w http.ResponseWriter, r *http.Request) {
	projectDaily := entity.ProjectDaily{}
	if err := json.NewDecoder(r.Body).Decode(&projectDaily); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}
	if err := h.Service.ModifyDailyJob(r.Context(), projectDaily); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}
	SuccessResponse(r.Context(), w)
}

// 작업내용 삭제
func (h *HandlerProjectDaily) Remove(w http.ResponseWriter, r *http.Request) {
	nullIdx := utils.ParseNullInt(r.PathValue("idx"))
	if nullIdx.Valid == false {
		BadRequestResponse(r.Context(), w)
		return
	}
	if err := h.Service.RemoveDailyJob(r.Context(), nullIdx.Int64); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}
	SuccessResponse(r.Context(), w)
}
