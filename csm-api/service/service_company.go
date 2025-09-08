package service

import (
	"context"
	"csm-api/api"
	"csm-api/entity"
	"csm-api/store"
	"csm-api/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/guregu/null"
	"strings"
)

/**
 * @author 작성자: 김진우
 * @created 작성일: 2025-02-18
 * @modified 최종 수정일:
 * @modifiedBy 최종 수정자:
 * @modified description
 * -
 */

type ServiceCompany struct {
	SafeDB        store.Queryer
	TimeSheetDB   store.Queryer
	Store         store.CompanyStore
	UserRoleStore store.UserRoleStore
}

// func: job 정보 조회
// @param
// - jno sql.NullInt64: 프로젝트 고유번호
func (s *ServiceCompany) GetJobInfo(ctx context.Context, jno int64) (*entity.JobInfo, error) {
	jnoSql := entity.ToSQLNulls(jno).(sql.NullInt64)
	data, err := s.Store.GetJobInfo(ctx, s.SafeDB, jnoSql)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}
	return data, nil
}

// func: 현장소장 조회
// @param
// - jno int64: 프로젝트 고유번호
func (s *ServiceCompany) GetSiteManagerList(ctx context.Context, jno int64) (*entity.Managers, error) {
	jnoSql := entity.ToSQLNulls(jno).(sql.NullInt64)
	list, err := s.Store.GetSiteManagerList(ctx, s.TimeSheetDB, jnoSql)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return list, nil
}

// func: 안전관리자 조회
// @param
// - jno int64: 프로젝트 고유번호
func (s *ServiceCompany) GetSafeManagerList(ctx context.Context, jno int64) (*entity.Managers, error) {
	jnoSql := entity.ToSQLNulls(jno).(sql.NullInt64)
	list, err := s.Store.GetSafeManagerList(ctx, s.SafeDB, jnoSql)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return list, nil
}

// func: 관리감독자 조회
// @param
// - jno int64: 프로젝트 고유번호
func (s *ServiceCompany) GetSupervisorList(ctx context.Context, jno int64) (*entity.Supervisors, error) {
	jnoSql := entity.ToSQLNulls(jno).(sql.NullInt64)
	// 안전보건시스템에 등록된 관리감독자
	list, err := s.Store.GetSupervisorList(ctx, s.SafeDB, jnoSql)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	// 조직도에 등록된 construction
	cList, err := s.Store.GetConstruction(ctx, s.TimeSheetDB, jno)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	rList, err := s.UserRoleStore.GetUserRoleListByCodeAndJno(ctx, s.SafeDB, "TEMP_SITE_MANAGER", jno)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	// 현장관리자 권한 셋팅
	siteManagerMap := make(map[int64]struct{})
	for _, role := range rList {
		if role.UserUno.Valid && role.RoleCode.Valid {
			if role.RoleCode.String == "TEMP_SITE_MANAGER" {
				siteManagerMap[role.UserUno.Int64] = struct{}{}
			}
		}
	}
	for _, user := range *cList {
		if user != nil && user.Uno.Valid {
			if _, ok := siteManagerMap[user.Uno.Int64]; ok {
				user.IsTempSiteManager = null.StringFrom("Y")
			} else {
				user.IsTempSiteManager = null.StringFrom("N")
			}
		}
	}

	for _, user := range *cList {
		if user != nil && user.CdNm.Valid && (strings.Contains(user.CdNm.String, "공무") || strings.Contains(user.CdNm.String, "사무보조")) {
			user.IsSiteManager = null.StringFrom("Y")
		} else {
			user.IsSiteManager = null.StringFrom("N")
		}
	}

	// 공종 셋팅
	funcNoMap := make(map[int64]null.String)
	for _, sup := range *list {
		if sup != nil && sup.Uno.Valid {
			funcNoMap[sup.Uno.Int64] = sup.FuncNo
		}
	}
	for _, cs := range *cList {
		if cs != nil && cs.Uno.Valid {
			if fn, exists := funcNoMap[cs.Uno.Int64]; exists {
				cs.FuncNo = fn
				cs.SysSafe = null.StringFrom("Y")
			} else {
				cs.SysSafe = null.StringFrom("N")
			}
		}
	}

	return cList, nil
}

// func: 공종 정보 조회
// @param
func (s *ServiceCompany) GetWorkInfoList(ctx context.Context) (*entity.WorkInfos, error) {
	list, err := s.Store.GetWorkInfoList(ctx, s.SafeDB)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return list, nil
}

// func: 협력업체 정보 조회
// @param
// - jno int64: 프로젝트 고유번호
func (s *ServiceCompany) GetCompanyInfoList(ctx context.Context, jno int64) (*entity.CompanyInfoResList, error) {
	list := &entity.CompanyInfoResList{}
	jnoSql := entity.ToSQLNulls(jno).(sql.NullInt64)

	// 협력업체 정보 조회
	companyList, err := s.Store.GetCompanyInfoList(ctx, s.SafeDB, jnoSql)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}
	for _, item := range *companyList {
		temp := &entity.CompanyInfoRes{}
		temp.Jno = item.Jno.Int64
		temp.Cno = item.Cno.Int64
		temp.Id = item.Id.String
		temp.Cellphone = item.Cellphone.String
		temp.Email = item.Email.String
		*list = append(*list, temp)
	}

	// JOB별 협력업체 리스트 API
	url := fmt.Sprintf("http://wcfservice.hi-techeng.co.kr/apipcs/getcontractinfo?jno=%d&contracttype=C", jno)
	response, err := api.CallGetAPI(url)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}
	companyApiReq := &entity.CompanyApiReq{}
	if err = json.Unmarshal([]byte(response), companyApiReq); err != nil {
		return nil, utils.CustomErrorf(err)
	}
	if companyApiReq.ResultType != "Success" {
		return nil, utils.CustomErrorf(fmt.Errorf("service_conpany;companyInfo/Api ResultType not Success"))
	}

	// 공종 조회
	workInfoList, err := s.Store.GetCompanyWorkInfoList(ctx, s.SafeDB, jnoSql)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	companyApiValues := entity.CompanyApiValues{}
	duplicate := make(map[int64]bool)
	for _, company := range companyApiReq.Value {
		if !duplicate[int64(company.CompCno)] {
			duplicate[int64(company.CompCno)] = true
			temp := &entity.CompanyApiValue{}
			temp.Jno = company.Jno
			temp.CompCno = company.CompCno
			temp.CompNameKr = company.CompNameKr
			temp.WorkerName = company.WorkerName
			companyApiValues = append(companyApiValues, temp)
		}
	}

	matched := make(map[int]bool)
	for _, item := range *list {
		for idx, company := range companyApiValues {
			if item.Jno == int64(company.Jno) && item.Cno == int64(company.CompCno) {
				item.CompNameKr = company.CompNameKr
				item.WorkerName = company.WorkerName
				matched[idx] = true
				break
			}
		}

		for _, work := range *workInfoList {
			if item.Jno == work.Jno.Int64 && item.Cno == work.Cno.Int64 {
				item.WorkInfo = append(item.WorkInfo, work.FuncNo.Int64)
			}
		}
	}

	for idx, company := range companyApiValues {
		if !matched[idx] {
			temp := &entity.CompanyInfoRes{}
			temp.Jno = int64(company.Jno)
			temp.Cno = int64(company.CompCno)
			temp.CompNameKr = company.CompNameKr
			temp.WorkerName = company.WorkerName
			temp.WorkInfo = make([]int64, 0)
			*list = append(*list, temp)
		}
	}

	return list, nil
}
