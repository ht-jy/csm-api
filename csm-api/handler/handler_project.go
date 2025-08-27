package handler

import (
	"csm-api/entity"
	"csm-api/service"
	"csm-api/utils"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"time"
)

/**
 * @author 작성자: 김진우
 * @created 작성일: 2025-02-20
 * @modified 최종 수정일:
 * @modifiedBy 최종 수정자:
 * @modified description
 * -
 */

type HandlerProject struct {
	Service service.ProjectService
}

// func: 프로젝트별 근로자 수 조회
// @param
//   - get parameter
//     targetDate: 현재시간
func (h *HandlerProject) WorkerCountList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	list, err := h.Service.GetProjectWorkerCountList(ctx, targetDate)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	values := struct {
		List entity.ProjectInfos `json:"list"`
	}{List: *list}
	SuccessValuesResponse(ctx, w, values)
}

// func: 프로젝트 이름 조회
// @param
// -
func (h *HandlerProject) JobNameList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	isRoleStr := r.URL.Query().Get("isRole")

	isRole, err := strconv.ParseBool(isRoleStr)

	if err != nil {
		BadRequestResponse(ctx, w)
		return
	}

	list, err := h.Service.GetProjectNmList(ctx, isRole)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	values := struct {
		List entity.ProjectInfos `json:"list"`
	}{List: *list}
	SuccessValuesResponse(ctx, w, values)
}

// func: 공사관리시스템 등록 프로젝트 전체 조회
// @param
// -
func (h *HandlerProject) RegList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// http get paramter를 저장할 구조체 생성 및 파싱
	page := entity.Page{}
	search := entity.JobInfo{}

	pageNum := r.URL.Query().Get("page_num")
	rowSize := r.URL.Query().Get("row_size")
	retry_search := r.URL.Query().Get("retry_search")
	order := r.URL.Query().Get("order")
	jobNo := r.URL.Query().Get("job_no")
	compName := r.URL.Query().Get("comp_name")
	orderCompName := r.URL.Query().Get("order_comp_name")
	jobName := r.URL.Query().Get("job_name")
	jobPmName := r.URL.Query().Get("job_pm_name")
	jobSd := r.URL.Query().Get("job_sd")
	jobEd := r.URL.Query().Get("job_ed")
	cdNm := r.URL.Query().Get("cd_nm")
	includeJno := r.URL.Query().Get("include_jno")
	snoString := r.URL.Query().Get("sno")
	if pageNum == "" || rowSize == "" {
		BadRequestResponse(ctx, w)
		return
	}
	page.PageNum, _ = strconv.Atoi(pageNum)
	page.RowSize, _ = strconv.Atoi(rowSize)
	page.Order = order
	search.JobNo = utils.ParseNullString(jobNo)
	search.CompName = utils.ParseNullString(compName)
	search.OrderCompName = utils.ParseNullString(orderCompName)
	search.JobName = utils.ParseNullString(jobName)
	search.JobPmName = utils.ParseNullString(jobPmName)
	search.JobSd = utils.ParseNullString(jobSd)
	search.JobEd = utils.ParseNullString(jobEd)
	search.CdNm = utils.ParseNullString(cdNm)

	// 프로젝트 조회
	list, err := h.Service.GetUsedProjectList(ctx, page, search, retry_search, includeJno, snoString)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	// 개수
	count, err := h.Service.GetUsedProjectCount(ctx, search, retry_search, includeJno, snoString)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	values := struct {
		List  entity.JobInfos `json:"list"`
		Count int             `json:"count"`
	}{List: *list, Count: count}
	SuccessValuesResponse(ctx, w, values)
}

// func: 프로젝트 전체 조회
// @param
// - all : 프로젝트 선택 시 "전체"도 넣을 것인지 여부
func (h *HandlerProject) EnterpriseList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := entity.Page{}
	search := entity.JobInfo{}

	// 프로젝트 "전체" 넣을 것인지 여부
	all := r.URL.Query().Get("all")
	retrySearch := r.URL.Query().Get("retry_search")
	pageNum := r.URL.Query().Get("page_num")
	rowSize := r.URL.Query().Get("row_size")
	order := r.URL.Query().Get("order")
	jobNo := r.URL.Query().Get("job_no")
	compName := r.URL.Query().Get("comp_name")
	orderCompName := r.URL.Query().Get("order_comp_name")
	jobName := r.URL.Query().Get("job_name")
	jobPmName := r.URL.Query().Get("job_pm_name")
	jobSd := r.URL.Query().Get("job_sd")
	jobEd := r.URL.Query().Get("job_ed")
	cdNm := r.URL.Query().Get("cd_nm")

	if pageNum == "" || rowSize == "" {
		BadRequestResponse(ctx, w)
		return
	}

	page.PageNum, _ = strconv.Atoi(pageNum)
	page.RowSize, _ = strconv.Atoi(rowSize)
	page.Order = order
	search.JobNo = utils.ParseNullString(jobNo)
	search.CompName = utils.ParseNullString(compName)
	search.OrderCompName = utils.ParseNullString(orderCompName)
	search.JobName = utils.ParseNullString(jobName)
	search.JobPmName = utils.ParseNullString(jobPmName)
	search.JobSd = utils.ParseNullString(jobSd)
	search.JobEd = utils.ParseNullString(jobEd)
	search.CdNm = utils.ParseNullString(cdNm)

	isAll, _ := strconv.Atoi(all)

	list, err := h.Service.GetAllProjectList(ctx, page, search, isAll, retrySearch)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	count, err := h.Service.GetAllProjectCount(ctx, search, isAll, retrySearch)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	values := struct {
		List  entity.JobInfos `json:"list"`
		Count int             `json:"count"`
	}{List: *list, Count: count}
	SuccessValuesResponse(ctx, w, values)

}

// func: 본인이 속한 조직도의 프로젝트 조회
// @param
// -
func (h *HandlerProject) MyOrgList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uno := r.PathValue("uno")

	int64UNO, err := strconv.ParseInt(uno, 10, 64)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	page := entity.Page{}
	search := entity.JobInfo{}

	pageNum := r.URL.Query().Get("page_num")
	rowSize := r.URL.Query().Get("row_size")
	order := r.URL.Query().Get("order")
	jobNo := r.URL.Query().Get("job_no")
	compName := r.URL.Query().Get("comp_name")
	orderCompName := r.URL.Query().Get("order_comp_name")
	jobName := r.URL.Query().Get("job_name")
	jobPmName := r.URL.Query().Get("job_pm_name")
	jobSd := r.URL.Query().Get("job_sd")
	jobEd := r.URL.Query().Get("job_ed")
	cdNm := r.URL.Query().Get("cd_nm")
	retrySearch := r.URL.Query().Get("retry_search")

	if pageNum == "" || rowSize == "" {
		BadRequestResponse(ctx, w)
		return
	}

	page.PageNum, _ = strconv.Atoi(pageNum)
	page.RowSize, _ = strconv.Atoi(rowSize)
	page.Order = order
	search.JobNo = utils.ParseNullString(jobNo)
	search.CompName = utils.ParseNullString(compName)
	search.OrderCompName = utils.ParseNullString(orderCompName)
	search.JobName = utils.ParseNullString(jobName)
	search.JobPmName = utils.ParseNullString(jobPmName)
	search.JobSd = utils.ParseNullString(jobSd)
	search.JobEd = utils.ParseNullString(jobEd)
	search.CdNm = utils.ParseNullString(cdNm)

	list, err := h.Service.GetStaffProjectList(ctx, page, search, int64UNO, retrySearch)

	if err != nil {
		BadRequestResponse(ctx, w)
		return
	}

	count, err := h.Service.GetStaffProjectCount(ctx, search, int64UNO, retrySearch)

	if err != nil {
		BadRequestResponse(ctx, w)
		return
	}

	values := struct {
		List  entity.JobInfos `json:"list"`
		Count int             `json:"count"`
	}{List: *list, Count: count}
	SuccessValuesResponse(ctx, w, values)
}

func (h *HandlerProject) ProjectBySite(w http.ResponseWriter, r *http.Request) {
	snoString := r.URL.Query().Get("sno")
	if snoString == "" {
		BadRequestResponse(r.Context(), w)
		return
	}

	sno, _ := strconv.ParseInt(snoString, 10, 64)
	list, err := h.Service.GetProjectBySite(r.Context(), sno)
	if err != nil {
		FailResponse(r.Context(), w, err)
		return
	}
	SuccessValuesResponse(r.Context(), w, list)
}

// func: 본인이 속한 프로젝트 이름 목록
// @param
// -
func (h *HandlerProject) MyJobNameList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	uno := r.PathValue("uno")
	role := r.URL.Query().Get("role")

	int64UNO, err := strconv.ParseInt(uno, 10, 64)

	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	projectNm, err := h.Service.GetProjectNmUnoList(ctx, int64UNO, role)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	values := struct {
		ProjectNm entity.ProjectInfos `json:"project_nm"`
	}{ProjectNm: *projectNm}
	SuccessValuesResponse(ctx, w, values)
}

// func: 현장근태 사용되지 않은 프로젝트
// @param
// -
func (h *HandlerProject) NonRegList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := entity.Page{}
	search := entity.NonUsedProject{}

	pageNum := r.URL.Query().Get("page_num")
	rowSize := r.URL.Query().Get("row_size")
	order := r.URL.Query().Get("order")
	rnumOrder := r.URL.Query().Get("rnum_order")
	retrySearch := r.URL.Query().Get("retry_search")
	jno := r.URL.Query().Get("jno")
	jobNo := r.URL.Query().Get("job_no")
	JobName := r.URL.Query().Get("job_name")
	JobYear := r.URL.Query().Get("job_year")
	JobSd := r.URL.Query().Get("job_sd")
	JobEd := r.URL.Query().Get("job_ed")
	UserName := r.URL.Query().Get("job_pm_nm")

	if pageNum == "" || rowSize == "" {
		BadRequestResponse(ctx, w)
		return
	}

	page.PageNum, _ = strconv.Atoi(pageNum)
	page.RowSize, _ = strconv.Atoi(rowSize)
	page.Order = order
	page.RnumOrder = rnumOrder

	search.Jno = utils.ParseNullInt(jno)
	search.JobNo = utils.ParseNullString(jobNo)
	search.JobName = utils.ParseNullString(JobName)
	search.JobYear = utils.ParseNullInt(JobYear)
	search.JobSd = utils.ParseNullString(JobSd)
	search.JobEd = utils.ParseNullString(JobEd)
	search.JobPmNm = utils.ParseNullString(UserName)

	list, err := h.Service.GetNonUsedProjectList(ctx, page, search, retrySearch)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	// 개수 조회
	count, err := h.Service.GetNonUsedProjectCount(ctx, search, retrySearch)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	values := struct {
		List  entity.NonUsedProjects `json:"list"`
		Count int                    `json:"count"`
	}{List: *list, Count: count}
	SuccessValuesResponse(ctx, w, values)
}

// func: 현장근태 사용되지 않은 프로젝트
// @param
// -
func (h *HandlerProject) NonRegListByType(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := entity.Page{}
	search := entity.NonUsedProject{}

	pageNum := r.URL.Query().Get("page_num")
	rowSize := r.URL.Query().Get("row_size")
	order := r.URL.Query().Get("order")
	rnumOrder := r.URL.Query().Get("rnum_order")
	retrySearch := r.URL.Query().Get("retry_search")
	jno := r.URL.Query().Get("jno")
	jobNo := r.URL.Query().Get("job_no")
	JobName := r.URL.Query().Get("job_name")
	JobYear := r.URL.Query().Get("job_year")
	JobSd := r.URL.Query().Get("job_sd")
	JobEd := r.URL.Query().Get("job_ed")
	UserName := r.URL.Query().Get("job_pm_nm")

	typeValue := r.PathValue("type")
	if typeValue == "" {
		BadRequestResponse(ctx, w)
		return
	}

	if pageNum == "" || rowSize == "" {
		BadRequestResponse(ctx, w)
		return
	}

	page.PageNum, _ = strconv.Atoi(pageNum)
	page.RowSize, _ = strconv.Atoi(rowSize)
	page.Order = order
	page.RnumOrder = rnumOrder

	search.Jno = utils.ParseNullInt(jno)
	search.JobNo = utils.ParseNullString(jobNo)
	search.JobName = utils.ParseNullString(JobName)
	search.JobYear = utils.ParseNullInt(JobYear)
	search.JobSd = utils.ParseNullString(JobSd)
	search.JobEd = utils.ParseNullString(JobEd)
	search.JobPmNm = utils.ParseNullString(UserName)

	list, err := h.Service.GetNonUsedProjectListByType(ctx, page, search, retrySearch, typeValue)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	// 개수 조회
	count, err := h.Service.GetNonUsedProjectCountByType(ctx, search, retrySearch, typeValue)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	values := struct {
		List  entity.NonUsedProjects `json:"list"`
		Count int                    `json:"count"`
	}{List: *list, Count: count}
	SuccessValuesResponse(ctx, w, values)
}

// func: 현장 프로젝트 추가
// @param
// -
func (h *HandlerProject) Add(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	project := entity.ReqProject{}

	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	err := h.Service.AddProject(ctx, project)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	SuccessResponse(ctx, w)
}

// func: 현장 기본 프로젝트 변경
// @param
// -
func (h *HandlerProject) ModifyDefault(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	project := entity.ReqProject{}
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	err := h.Service.ModifyDefaultProject(ctx, project)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	SuccessResponse(ctx, w)
}

// func: 현장 프로젝트 사용여부 변경
// @param
// -
func (h *HandlerProject) ModifyIsUse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	project := entity.ReqProject{}
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		FailResponse(ctx, w, err)
		return
	}

	err := h.Service.ModifyUseProject(ctx, project)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	SuccessResponse(ctx, w)
}

// func: 현장 프로젝트 삭제
// @param
// -
func (h *HandlerProject) Remove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	snoString := chi.URLParam(r, "sno")
	jnoString := chi.URLParam(r, "jno")
	if snoString == "" || jnoString == "" {
		BadRequestResponse(ctx, w)
		return
	}
	sno, _ := strconv.ParseInt(snoString, 10, 64)
	jno, _ := strconv.ParseInt(jnoString, 10, 64)

	err := h.Service.RemoveProject(ctx, sno, jno)
	if err != nil {
		FailResponse(ctx, w, err)
		return
	}

	SuccessResponse(ctx, w)
}
