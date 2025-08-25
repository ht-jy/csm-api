package handler

import (
	"csm-api/entity"
	"csm-api/service"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type HandlerEquip struct {
	Service service.EquipService
}

func (h *HandlerEquip) AllList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	strJno := r.URL.Query().Get("jno")
	strSno := r.URL.Query().Get("sno")
	if strJno == "" || strSno == "" {
		BadRequestResponse(
			ctx,
			w)
	}

	jno, err := strconv.ParseInt(strJno, 10, 64)
	sno, err := strconv.ParseInt(strSno, 10, 64)

	list, err := h.Service.GetEquipList(ctx, jno, sno)
	if err != nil {
		FailResponse(ctx, w, err)
	}

	values := struct {
		List entity.Equips `json:"list"`
	}{List: list}
	SuccessValuesResponse(ctx, w, values)
}

func (h *HandlerEquip) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	strJno := r.URL.Query().Get("jno")
	strSno := r.URL.Query().Get("sno")
	strRecordDate := r.URL.Query().Get("recordDate")

	if strJno == "" || strSno == "" || strRecordDate == "" {
		BadRequestResponse(
			ctx,
			w)
	}

	jno, err := strconv.ParseInt(strJno, 10, 64)
	sno, err := strconv.ParseInt(strSno, 10, 64)

	if strRecordDate == "-" {
		strRecordDate = time.Now().Format("2006-01-02")
	}
	recordDate, err := time.Parse("2006-01-02", strRecordDate)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	equips, err := h.Service.GetEquip(ctx, jno, sno, recordDate)
	if err != nil {
		FailResponse(ctx, w, err)
	}

	values := struct {
		Equips entity.Equips `json:"list"`
	}{Equips: equips}
	SuccessValuesResponse(ctx, w, values)
}

func (h *HandlerEquip) Merge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	equip := entity.Equip{}

	if err := json.NewDecoder(r.Body).Decode(&equip); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	if err := h.Service.MergeEquipCnt(ctx, equip); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	SuccessResponse(ctx, w)
}
