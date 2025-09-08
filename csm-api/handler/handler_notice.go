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
 * @author 작성자: 정지영
 * @created 작성일: 2025-02-17
 * @modified 최종 수정일:
 * @modifiedBy 최종 수정자:
 * @modified description
 * -
 */

// struct : 공지사항 조회
type NoticeHandler struct {
	Service service.NoticeService
}

// func : 공지사항 전체조회
// @param
// - response: hhtp get parameter
func (n *NoticeHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := entity.Page{}
	search := entity.Notice{}

	pageNum := r.URL.Query().Get(entity.PageNumKey)
	rowSize := r.URL.Query().Get(entity.RowSizeKey)
	order := r.URL.Query().Get(entity.OrderKey)

	if pageNum == "" || rowSize == "" {
		BadRequestResponse(ctx, w)
		return
	}

	page.PageNum, _ = strconv.Atoi(pageNum)
	page.RowSize, _ = strconv.Atoi(rowSize)
	page.Order = order

	search.Jno = utils.ParseNullInt(r.URL.Query().Get("jno"))
	search.JobLocName = utils.ParseNullString(r.URL.Query().Get("job_loc_name"))
	search.JobName = utils.ParseNullString(r.URL.Query().Get("job_name"))
	search.Title = utils.ParseNullString(r.URL.Query().Get("title"))
	search.UserInfo = utils.ParseNullString(r.URL.Query().Get("user_info"))

	isRoleStr := r.URL.Query().Get("isRole")
	isRole, err := strconv.ParseBool(isRoleStr)

	notices, err := n.Service.GetNoticeList(ctx, page, isRole, search)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	count, err := n.Service.GetNoticeListCount(ctx, isRole, search)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	values := struct {
		Notices entity.Notices `json:"notices"`
		Count   int            `json:"count"`
	}{Notices: *notices, Count: count}
	SuccessValuesResponse(ctx, w, values)
}

// func: 공지사항 추가
// @param
// - request: entity.Notice - json(raw)
func (n *NoticeHandler) Add(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	notice := entity.Notice{}

	if err := json.NewDecoder(r.Body).Decode(&notice); err != nil {
		BadRequestResponse(ctx, w)
		return
	}

	err := n.Service.AddNotice(ctx, notice)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	SuccessResponse(ctx, w)
}

// func: 공지사항 수정
// @param
// - request: entity.Notice - json(raw)
func (n *NoticeHandler) Modify(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// request 데이터 파싱
	notice := entity.Notice{}
	if err := json.NewDecoder(r.Body).Decode(&notice); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	if err := n.Service.ModifyNotice(ctx, notice); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	SuccessResponse(ctx, w)
}

// func: 공지사항 삭제
// @param
// - idx : 공지사항 인덱스
func (n *NoticeHandler) Remove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	nullIdx := utils.ParseNullInt(r.PathValue("idx"))

	if nullIdx.Valid == false {
		BadRequestResponse(ctx, w)
		return
	}

	idx := nullIdx.Int64
	if err := n.Service.RemoveNotice(ctx, idx); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	SuccessResponse(ctx, w)
}
