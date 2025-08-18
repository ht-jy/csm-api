package service

import (
	"context"
	"csm-api/auth"
	"csm-api/entity"
	"csm-api/store"
	"csm-api/txutil"
	"csm-api/utils"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ServiceProject struct {
	SafeDB      store.Queryer
	SafeTDB     store.Beginner
	Store       store.ProjectStore
	UserStore   store.UserStore
	UserService UserService
}

// 현장 고유번호로 현장에 해당하는 프로젝트 리스트 조회 비즈니스
//
// @param sno: 현장 고유번호, , targetDate time.Time: 현재시간
func (p *ServiceProject) GetProjectList(ctx context.Context, sno int64, targetDate time.Time) (*entity.ProjectInfos, error) {
	projectInfos, err := p.Store.GetProjectList(ctx, p.SafeDB, sno, targetDate)
	if err != nil {
		return &entity.ProjectInfos{}, utils.CustomErrorf(err)
	}

	// 안전관리자 수 조회
	safeInfos, err := p.Store.GetProjectSafeWorkerCountList(ctx, p.SafeDB, targetDate)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	// 프로젝트 정보 객체에 pm, pe 정보 삽입
	for _, projectInfo := range *projectInfos {
		var unoList []int

		// pe uno 조회
		if &projectInfo.JobPe != nil && projectInfo.JobPe.String != "" {
			jobPeList := strings.Split(projectInfo.JobPe.String, ",")
			for _, jonPe := range jobPeList {
				uno, err := strconv.Atoi(jonPe)
				if err != nil {
					return &entity.ProjectInfos{}, utils.CustomErrorf(fmt.Errorf("service_project/strconv.Atoi(jonPe) parse err"))
				}
				unoList = append(unoList, uno)
			}
		}

		// pe 정보 일괄 조회
		userPeList, err := p.UserStore.GetUserInfoPeList(ctx, p.SafeDB, unoList)
		if err != nil {
			return &entity.ProjectInfos{}, utils.CustomErrorf(err)
		}
		projectInfo.ProjectPeList = userPeList

		// 안전, 공사 근로자 수
		for _, safe := range *safeInfos {
			if projectInfo.Sno == safe.Sno && projectInfo.Jno == safe.Jno {
				projectInfo.WorkerCountSafe = safe.SafeCount
				projectInfo.WorkerCountWork.Int64 = projectInfo.WorkerCountHtenc.Int64 - safe.SafeCount.Int64
				projectInfo.WorkerCountWork.Valid = true
				break
			}
		}
	}

	return projectInfos, nil
}

// func: 프로젝트 근로자 수 조회
// @param
// - sno int64 현장 번호, targetDate time.Time: 현재시간
func (p *ServiceProject) GetProjectWorkerCountList(ctx context.Context, targetDate time.Time) (*entity.ProjectInfos, error) {
	// 근로자 수 조회
	projectInfos, err := p.Store.GetProjectWorkerCountList(ctx, p.SafeDB, targetDate)
	if err != nil {
		return &entity.ProjectInfos{}, utils.CustomErrorf(err)
	}

	// 안전관리자 수 조회
	safeInfos, err := p.Store.GetProjectSafeWorkerCountList(ctx, p.SafeDB, targetDate)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	// 안전, 공사 근로자 수
	for _, project := range *projectInfos {
		for _, safe := range *safeInfos {
			if project.Sno == safe.Sno && project.Jno == safe.Jno {
				project.WorkerCountSafe = safe.SafeCount
				project.WorkerCountWork.Int64 = project.WorkerCountHtenc.Int64 - safe.SafeCount.Int64
				project.WorkerCountWork.Valid = true
				break
			}
		}
	}

	return projectInfos, nil
}

// func: 프로젝트 조회(이름)
// @param
// -
func (p *ServiceProject) GetProjectNmList(ctx context.Context, isRole bool) (*entity.ProjectInfos, error) {
	unoString, _ := auth.GetContext(ctx, auth.Uno{})

	var roleInt int
	if isRole { // 권한이 있는 경우
		roleInt = 1
	} else {
		roleInt = 0
	}

	uno, err := strconv.ParseInt(unoString, 10, 64)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	nmList, err := p.Store.GetProjectNmList(ctx, p.SafeDB, roleInt, uno)
	if err != nil {
		return &entity.ProjectInfos{}, utils.CustomErrorf(err)
	}

	return nmList, nil
}

// func: 공사관리시스템 등록 프로젝트 전체 조회
// @param
// -
func (p *ServiceProject) GetUsedProjectList(ctx context.Context, page entity.Page, search entity.JobInfo, retry string, includeJno string, snoString string) (*entity.JobInfos, error) {
	pageSql := entity.PageSql{}
	pageSql, err := pageSql.OfPageSql(page)
	if err != nil {
		return &entity.JobInfos{}, utils.CustomErrorf(err)
	}

	jobInfos, err := p.Store.GetUsedProjectList(ctx, p.SafeDB, pageSql, search, retry, includeJno, snoString)
	if err != nil {
		return &entity.JobInfos{}, utils.CustomErrorf(err)
	}

	return jobInfos, nil
}

// func: 프로젝트 전체 조회 개수
// @param
// -
func (p *ServiceProject) GetUsedProjectCount(ctx context.Context, search entity.JobInfo, retry string, includeJno string, snoString string) (int, error) {
	count, err := p.Store.GetUsedProjectCount(ctx, p.SafeDB, search, retry, includeJno, snoString)
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}

	return count, nil
}

// func: 프로젝트 전체 조회
// @param
// -
func (p *ServiceProject) GetAllProjectList(ctx context.Context, page entity.Page, search entity.JobInfo, isAll int, retry string) (*entity.JobInfos, error) {
	pageSql := entity.PageSql{}
	pageSql, err := pageSql.OfPageSql(page)
	if err != nil {
		return &entity.JobInfos{}, utils.CustomErrorf(err)
	}

	jobInfos, err := p.Store.GetAllProjectList(ctx, p.SafeDB, pageSql, search, isAll, retry)
	if err != nil {
		return &entity.JobInfos{}, utils.CustomErrorf(err)
	}

	return jobInfos, nil

}

// func: 프로젝트 개수 조회
// @param
// -
func (p *ServiceProject) GetAllProjectCount(ctx context.Context, search entity.JobInfo, isAll int, retry string) (int, error) {
	count, err := p.Store.GetAllProjectCount(ctx, p.SafeDB, search, retry)
	if isAll == 1 {
		count += 1
	}
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}

	return count, nil
}

// func: 본인이 속한 프로젝트 조회
// @param
// - UNO
func (p *ServiceProject) GetStaffProjectList(ctx context.Context, page entity.Page, search entity.JobInfo, uno int64, retry string) (*entity.JobInfos, error) {
	var unoSql sql.NullInt64

	if uno != 0 {
		unoSql = sql.NullInt64{Valid: true, Int64: uno}
	} else {
		unoSql = sql.NullInt64{Valid: false}
	}

	pageSql := entity.PageSql{}
	pageSql, err := pageSql.OfPageSql(page)
	if err != nil {
		return &entity.JobInfos{}, utils.CustomErrorf(err)
	}

	jobInfos, err := p.Store.GetStaffProjectList(ctx, p.SafeDB, pageSql, search, unoSql, retry)
	if err != nil {
		return &entity.JobInfos{}, utils.CustomErrorf(err)
	}
	return jobInfos, nil

}

// func: 본인이 속한 프로젝트 개수
// @param
// - UNO
func (p *ServiceProject) GetStaffProjectCount(ctx context.Context, search entity.JobInfo, uno int64, retry string) (int, error) {
	var unoSql sql.NullInt64

	if uno != 0 {
		unoSql = sql.NullInt64{Valid: true, Int64: uno}
	} else {
		unoSql = sql.NullInt64{Valid: false}
	}

	count, err := p.Store.GetStaffProjectCount(ctx, p.SafeDB, search, unoSql, retry)
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}

	return count, nil

}

// 본인이 속한 프로젝트 이름 목록
func (p *ServiceProject) GetProjectNmUnoList(ctx context.Context, uno int64, role string) (*entity.ProjectInfos, error) {

	var unoSql sql.NullInt64
	if uno != 0 {
		unoSql = sql.NullInt64{Valid: true, Int64: uno}
	} else {
		unoSql = sql.NullInt64{Valid: false}
	}

	var roleInt int
	authorizationList := []string{"ADMIN", "SUPER_ADMIN", "SYSTEM_ADMIN"}
	if utils.AuthorizationListCheck(authorizationList, utils.ParseNullString(role)) {
		roleInt = 1
	} else {
		roleInt = 0
	}
	projectInfos, err := p.Store.GetProjectNmUnoList(ctx, p.SafeDB, unoSql, roleInt)

	if err != nil {
		return &entity.ProjectInfos{}, utils.CustomErrorf(err)
	}

	return projectInfos, nil
}

// func: 현장근태 사용되지 않은 프로젝트
// @param
// -
func (s *ServiceProject) GetNonUsedProjectList(ctx context.Context, page entity.Page, search entity.NonUsedProject, retry string) (*entity.NonUsedProjects, error) {
	pageSql := entity.PageSql{}
	pageSql, err := pageSql.OfPageSql(page)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	list, err := s.Store.GetNonUsedProjectList(ctx, s.SafeDB, pageSql, search, retry)

	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return list, nil
}

// func: 현장근태 사용되지 않은 프로젝트 수
// @param
// -
func (s *ServiceProject) GetNonUsedProjectCount(ctx context.Context, search entity.NonUsedProject, retry string) (int, error) {
	count, err := s.Store.GetNonUsedProjectCount(ctx, s.SafeDB, search, retry)

	if err != nil {
		return 0, utils.CustomErrorf(err)
	}

	return count, nil
}

// func: 현장근태 사용되지 않은 프로젝트(타입별)
// @param
// -
func (s *ServiceProject) GetNonUsedProjectListByType(ctx context.Context, page entity.Page, search entity.NonUsedProject, retry string, typeString string) (*entity.NonUsedProjects, error) {
	pageSql := entity.PageSql{}
	pageSql, err := pageSql.OfPageSql(page)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	list, err := s.Store.GetNonUsedProjectListByType(ctx, s.SafeDB, pageSql, search, retry, typeString)

	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	return list, nil
}

// func: 현장근태 사용되지 않은 프로젝트 수(타입별)
// @param
// -
func (s *ServiceProject) GetNonUsedProjectCountByType(ctx context.Context, search entity.NonUsedProject, retry string, typeString string) (int, error) {
	count, err := s.Store.GetNonUsedProjectCountByType(ctx, s.SafeDB, search, retry, typeString)

	if err != nil {
		return 0, utils.CustomErrorf(err)
	}

	return count, nil
}

// 현장별 프로젝트 조회
func (s *ServiceProject) GetProjectBySite(ctx context.Context, sno int64) (entity.ProjectInfos, error) {
	projectInfos, err := s.Store.GetProjectBySite(ctx, s.SafeDB, sno)
	if err != nil {
		return entity.ProjectInfos{}, utils.CustomErrorf(err)
	}
	return projectInfos, nil
}

// func: 현장 프로젝트 추가
// @param
// -
func (s *ServiceProject) AddProject(ctx context.Context, project entity.ReqProject) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	err = s.Store.AddProject(ctx, tx, project)
	if err != nil {
		return utils.CustomErrorf(err)
	}
	return
}

// func: 현장 기본 프로젝트 변경
// @param
// -
func (s *ServiceProject) ModifyDefaultProject(ctx context.Context, project entity.ReqProject) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	err = s.Store.ModifyDefaultProject(ctx, tx, project)
	if err != nil {
		return utils.CustomErrorf(err)
	}
	return
}

// func: 현장 프로젝트 사용여부 변경
// @param
// -
func (s *ServiceProject) ModifyUseProject(ctx context.Context, project entity.ReqProject) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	err = s.Store.ModifyUseProject(ctx, tx, project)
	if err != nil {
		return utils.CustomErrorf(err)
	}
	return
}

// func: 현장 프로젝트 삭제
// @param
// -
func (s *ServiceProject) RemoveProject(ctx context.Context, sno int64, jno int64) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	err = s.Store.RemoveProject(ctx, tx, sno, jno)
	if err != nil {
		return utils.CustomErrorf(err)
	}
	return
}
