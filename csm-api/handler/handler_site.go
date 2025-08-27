package handler

import (
	"csm-api/entity"
	"csm-api/service"
	"csm-api/utils"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

/**
 * @author 작성자: 김진우
 * @created 작성일: 2025-02-12
 * @modified 최종 수정일:
 * @modifiedBy 최종 수정자:
 * @modified description
 * -
 */

type HandlerSite struct {
	Service     service.SiteService
	CodeService service.CodeService
}

// func: 현장 관리 리스트
// @param
// - response: targetDate(현재날짜), pCode(부모코드) - url parameter
func (s *HandlerSite) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// GET 요청에서 파라미터 값 읽기
	targetDateString := r.URL.Query().Get("targetDate")
	if targetDateString == "" {
		BadRequestResponse(ctx, w)
		return
	}
	if targetDateString == "-" {
		targetDateString = time.Now().Format("2006-01-02")
	}
	targetDate, err := time.Parse("2006-01-02", targetDateString)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	pCode := r.URL.Query().Get("pCode")
	if pCode == "" {
		BadRequestResponse(ctx, w)
		return
	}

	isRoleStr := r.URL.Query().Get("isRole")
	isRole, err := strconv.ParseBool(isRoleStr)
	if err != nil {
		BadRequestResponse(ctx, w)
		return
	}

	isNonUseStr := r.URL.Query().Get("isNonUse")
	isNonUse, err := strconv.ParseBool(isNonUseStr)
	if err != nil {
		BadRequestResponse(ctx, w)
		return
	}

	// 현장 관리 리스트 조회
	sites, err := s.Service.GetSiteList(ctx, targetDate, isRole, isNonUse)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	// 현장 관리 코드 조회
	codes, err := s.CodeService.GetCodeList(ctx, pCode)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	values := entity.SiteRes{
		Site: *sites,
		Code: *codes,
	}
	SuccessValuesResponse(ctx, w, values)
}

// func: 현장 관리 수정
// @param
// -
func (s *HandlerSite) Modify(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	site := entity.Site{}
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		BadRequestResponse(ctx, w)
		return
	}

	if err := s.Service.ModifySite(ctx, site); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	SuccessResponse(ctx, w)
}

// func: 현장명 조회
// @param
// -
func (s *HandlerSite) SiteNameList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page := entity.Page{}
	search := entity.Site{}

	pageNum := r.URL.Query().Get(entity.PageNumKey)
	rowSize := r.URL.Query().Get(entity.RowSizeKey)
	order := r.URL.Query().Get(entity.OrderKey)
	nonSite, _ := strconv.Atoi(r.URL.Query().Get("non_site"))

	if pageNum == "" || rowSize == "" {
		BadRequestResponse(ctx, w)
		return
	}
	page.PageNum, _ = strconv.Atoi(pageNum)
	page.RowSize, _ = strconv.Atoi(rowSize)
	page.Order = order

	search.Sno = utils.ParseNullInt(r.URL.Query().Get("sno"))
	search.SiteNm = utils.ParseNullString(r.URL.Query().Get("site_nm"))
	search.Etc = utils.ParseNullString(r.URL.Query().Get("etc"))
	search.LocName = utils.ParseNullString(r.URL.Query().Get("loc_name"))

	// http get paramter를 저장할 구조체 생성
	list, err := s.Service.GetSiteNmList(ctx, page, search, nonSite)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	count, err := s.Service.GetSiteNmCount(ctx, search, nonSite)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	values := struct {
		List  entity.Sites `json:"list"`
		Count int          `json:"count"`
	}{List: *list, Count: count}
	SuccessValuesResponse(ctx, w, values)
}

// func: 현장 상태 조회
// @param
// -
func (s *HandlerSite) StatsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// GET 요청에서 파라미터 값 읽기
	targetDateString := r.URL.Query().Get("targetDate")
	if targetDateString == "" {
		BadRequestResponse(ctx, w)
		return
	}
	targetDate, err := time.Parse("2006-01-02", targetDateString)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	list, err := s.Service.GetSiteStatsList(ctx, targetDate)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	values := struct {
		List entity.Sites `json:"list"`
	}{List: *list}
	SuccessValuesResponse(ctx, w, values)
}

// func: 현장 생성(추가)
// @param
// -
func (h *HandlerSite) Add(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := struct {
		Jno      int64  `json:"jno"`
		Uno      int64  `json:"uno"`
		UserName string `json:"user_name"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		FailResponse(ctx, w, err)
		return
	}
	user := entity.User{}
	user = user.SetUser(request.Uno, request.UserName)

	err := h.Service.AddSite(ctx, request.Jno, user)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	SuccessResponse(ctx, w)
}

// func: 현장 사용안함 변경
// @param
// -
func (h *HandlerSite) ModifyNonUse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqSite := entity.ReqSite{}
	if err := json.NewDecoder(r.Body).Decode(&reqSite); err != nil {
		FailResponse(ctx, w, err)
		return
	}
	if reqSite.Sno.Valid == false {
		BadRequestResponse(ctx, w)
		return
	}

	if err := h.Service.ModifySiteIsNonUse(ctx, reqSite); err != nil {
		FailResponse(ctx, w, err)
	}

	SuccessResponse(ctx, w)
}

// func: 현장 사용으로 변경
// @param
// -
func (h *HandlerSite) ModifyUse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqSite := entity.ReqSite{}
	if err := json.NewDecoder(r.Body).Decode(&reqSite); err != nil {
		FailResponse(ctx, w, err)
		return
	}
	if reqSite.Sno.Valid == false {
		BadRequestResponse(ctx, w)
		return
	}

	if err := h.Service.ModifySiteIsUse(ctx, reqSite); err != nil {
		FailResponse(ctx, w, err)
	}

	SuccessResponse(ctx, w)
}

// func: 현장 프로젝트 사용안함 변경
// @param
// -
func (h *HandlerSite) ModifySiteJobNonUse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqSite := entity.ReqSite{}
	if err := json.NewDecoder(r.Body).Decode(&reqSite); err != nil {
		FailResponse(ctx, w, err)
		return
	}
	if reqSite.Jno.Valid == false || reqSite.Sno.Valid == false {
		BadRequestResponse(ctx, w)
		return
	}

	if err := h.Service.ModifySiteJobNonUse(ctx, reqSite); err != nil {
		FailResponse(ctx, w, err)
	}

	SuccessResponse(ctx, w)
}

// func: 현장 프로젝트 사용으로 변경
// @param
// -
func (h *HandlerSite) ModifySiteJobUse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqSite := entity.ReqSite{}
	if err := json.NewDecoder(r.Body).Decode(&reqSite); err != nil {
		FailResponse(ctx, w, err)
		return
	}
	if reqSite.Jno.Valid == false || reqSite.Sno.Valid == false {
		BadRequestResponse(ctx, w)
		return
	}

	if err := h.Service.ModifySiteJobUse(ctx, reqSite); err != nil {
		FailResponse(ctx, w, err)
	}

	SuccessResponse(ctx, w)
}

// 공정률 수정
func (h *HandlerSite) ModifyWorkRate(w http.ResponseWriter, r *http.Request) {
	var workRate entity.SiteWorkRate
	if err := json.NewDecoder(r.Body).Decode(&workRate); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}

	if err := h.Service.ModifyWorkRate(r.Context(), workRate); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}
	SuccessResponse(r.Context(), w)
}

// 날짜별 공정률 조회
func (h *HandlerSite) SiteWorkRateByDate(w http.ResponseWriter, r *http.Request) {
	searchDate := r.URL.Query().Get("search_date")
	jnoString := r.URL.Query().Get("jno")
	if searchDate == "" || jnoString == "" {
		BadRequestResponse(r.Context(), w)
		return
	}

	jno, _ := strconv.ParseInt(jnoString, 10, 64)
	data, err := h.Service.GetSiteWorkRateByDate(r.Context(), jno, searchDate)
	if err != nil {
		FailResponse(r.Context(), w, err)
		return
	}
	SuccessValuesResponse(r.Context(), w, data)
}

// 월별 공정률 조회
func (h *HandlerSite) SiteWorkRateByMonth(w http.ResponseWriter, r *http.Request) {
	searchDate := r.PathValue("date")
	jnoString := r.PathValue("jno")
	if jnoString == "" {
		BadRequestResponse(r.Context(), w)
		return
	}

	if searchDate == "" {
		BadRequestResponse(r.Context(), w)
		return
	}

	jno, _ := strconv.ParseInt(jnoString, 10, 64)
	data, err := h.Service.GetSiteWorkRateListByMonth(r.Context(), jno, searchDate)
	if err != nil {
		FailResponse(r.Context(), w, err)
		return
	}
	SuccessValuesResponse(r.Context(), w, data)
}

// 공정률 추가
func (h *HandlerSite) AddWorkRate(w http.ResponseWriter, r *http.Request) {
	var workRate entity.SiteWorkRate
	if err := json.NewDecoder(r.Body).Decode(&workRate); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}

	if err := h.Service.AddWorkRate(r.Context(), workRate); err != nil {
		FailResponse(r.Context(), w, err)
		return
	}
	SuccessResponse(r.Context(), w)
}
