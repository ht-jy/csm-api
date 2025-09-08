package handler

import (
	"csm-api/entity"
	"csm-api/service"
	"csm-api/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

type HandlerRestSchedule struct {
	Service service.ScheduleService
}

// func: 휴무일 조회
// @param
// -
func (h *HandlerRestSchedule) RestList(w http.ResponseWriter, r *http.Request) {

	jnoString := r.URL.Query().Get("jno")
	year := r.URL.Query().Get("year")
	month := r.URL.Query().Get("month")

	if year == "" {
		BadRequestResponse(r.Context(), w)
		return
	}

	isRoleStr := r.URL.Query().Get("isRole")
	isRole, err := strconv.ParseBool(isRoleStr)
	if err != nil {
		BadRequestResponse(r.Context(), w)
		return
	}

	jno, _ := strconv.ParseInt(jnoString, 10, 64)
	list, err := h.Service.GetRestScheduleList(r.Context(), isRole, jno, year, month)
	if err != nil {
		FailResponse(r.Context(), w, err)
		return
	}

	values := struct {
		List entity.RestSchedules `json:"list"`
	}{List: list}
	SuccessValuesResponse(r.Context(), w, values)
}

// func: 휴무일 추가
// @param
// -
func (h *HandlerRestSchedule) RestAdd(w http.ResponseWriter, r *http.Request) {
	rest := entity.RestSchedules{}

	if err := json.NewDecoder(r.Body).Decode(&rest); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}

	if err := h.Service.AddRestSchedule(r.Context(), rest); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}
	SuccessResponse(r.Context(), w)
}

// func: 휴무일 수정
// @param
// -
func (h *HandlerRestSchedule) RestModify(w http.ResponseWriter, r *http.Request) {
	rest := entity.RestSchedule{}

	if err := json.NewDecoder(r.Body).Decode(&rest); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}
	if err := h.Service.ModifyRestSchedule(r.Context(), rest); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}
	SuccessResponse(r.Context(), w)
}

// func: 휴무일 삭제
// @param
// -
func (h *HandlerRestSchedule) RestRemove(w http.ResponseWriter, r *http.Request) {
	nullCno := utils.ParseNullInt(r.PathValue("cno"))

	if nullCno.Valid == false {
		BadRequestResponse(r.Context(), w)
		return
	}

	cno := nullCno.Int64
	if err := h.Service.RemoveRestSchedule(r.Context(), cno); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}

	SuccessResponse(r.Context(), w)
}
