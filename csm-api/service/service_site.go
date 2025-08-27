package service

import (
	"context"
	"csm-api/auth"
	"csm-api/entity"
	"csm-api/store"
	"csm-api/txutil"
	"csm-api/utils"
	"fmt"
	"github.com/guregu/null"
	"strconv"
	"strings"
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

// struct: 현장 데이터 서비스 구조체
type ServiceSite struct {
	SafeDB                  store.Queryer
	SafeTDB                 store.Beginner
	Store                   store.SiteStore
	ProjectStore            store.ProjectStore
	ProjectDailyStore       store.ProjectDailyStore
	SitePosStore            store.SitePosStore
	SiteDateStore           store.SiteDateStore
	UserService             UserService
	ProjectService          ProjectService
	WeatherApiService       WeatherApiService
	AddressSearchAPIService AddressSearchAPIService
	RestDateApiService      RestDateApiService
}

// func: 현장 관리 리스트 조회
// @param
// - targetDate: 현재시간
func (s *ServiceSite) GetSiteList(ctx context.Context, targetDate time.Time, isRole bool, isNonUse bool) (*entity.Sites, error) {

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

	var status string
	if isNonUse { // 현장 종료된 경우
		status = "S"
	} else {
		status = "Y"
	}

	//현장관리 테이블 조회
	sites, err := s.Store.GetSiteList(ctx, s.SafeDB, targetDate, roleInt, uno, status)
	if err != nil {
		return &entity.Sites{}, utils.CustomErrorf(err)
	}

	// 공휴일 조회
	year := strconv.Itoa(targetDate.Year())
	restDates, err := s.RestDateApiService.GetRestDelDates(year, "")
	if err != nil {
		return &entity.Sites{}, utils.CustomErrorf(err)
	}

	// 주말, 휴무일인지 체크
	isRest := targetDate.Weekday() == time.Saturday || targetDate.Weekday() == time.Sunday
	if !isRest {
		for _, rest := range restDates {
			rd, err := time.ParseInLocation("20060102", strconv.FormatInt(rest.RestDate, 10), targetDate.Location())
			if err != nil {
				continue
			}
			if rd.Year() == targetDate.Year() &&
				rd.Month() == targetDate.Month() &&
				rd.Day() == targetDate.Day() {
				isRest = true
				break
			}
		}
	}

	for _, site := range *sites {
		// 주말,공휴일 이라면 미운영을 휴무일 상태로 변경
		if isRest {
			if site.CurrentSiteStats.String == "C" {
				site.CurrentSiteStats.String = "H"
			}
		}

		sno := site.Sno.Int64
		var projectCnt int64 = 0
		var sumWorkRate int64 = 0

		// 프로젝트 리스트 조회
		projectInfos, err := s.ProjectService.GetProjectList(ctx, sno, targetDate)
		if err != nil {
			return &entity.Sites{}, utils.CustomErrorf(err)
		}
		site.ProjectList = projectInfos

		for _, projectInfo := range *site.ProjectList {
			// 공정률 더하기
			projectCnt++
			sumWorkRate += projectInfo.WorkRate.Int64

			if &projectInfo.Jno != nil {
				// 작업내용
				projectDailyList, err := s.ProjectDailyStore.GetProjectDailyContentList(ctx, s.SafeDB, projectInfo.Jno.Int64, targetDate)
				if err != nil {
					return &entity.Sites{}, utils.CustomErrorf(err)
				}
				projectInfo.DailyContentList = projectDailyList
			}
		}
		if projectCnt == 0 {
			projectCnt++
		}

		// 공정률
		site.WorkRate = null.NewFloat(float64(sumWorkRate)/float64(projectCnt), true)

		// 현장 위치 조회
		sitePos, err := s.SitePosStore.GetSitePosData(ctx, s.SafeDB, sno)
		if err != nil {
			return &entity.Sites{}, utils.CustomErrorf(err)
		}
		if sitePos.RoadAddress.String == "" {
			depthArray := []string{sitePos.RoadAddressNameDepth1.String, sitePos.RoadAddressNameDepth2.String, sitePos.RoadAddressNameDepth3.String, sitePos.RoadAddressNameDepth4.String, sitePos.RoadAddressNameDepth5.String}
			roadAddress := ""
			for _, depth := range depthArray {
				if depth != "" {
					roadAddress = roadAddress + " " + depth
				}
			}
			sitePos.RoadAddress = utils.ParseNullString(strings.Trim(roadAddress, " "))
		}
		site.SitePos = sitePos

		// 현장 날짜 조회
		siteDateData, err := s.SiteDateStore.GetSiteDateData(ctx, s.SafeDB, sno)
		if err != nil {
			return &entity.Sites{}, utils.CustomErrorf(err)
		}
		site.SiteDate = siteDateData
	}

	return sites, nil
}

// func: 현장 데이터 리스트 조회
// @param
// -
func (s *ServiceSite) GetSiteNmList(ctx context.Context, page entity.Page, search entity.Site, nonSite int) (*entity.Sites, error) {
	pageSql := entity.PageSql{}
	pageSql, err := pageSql.OfPageSql(page)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	sites, err := s.Store.GetSiteNmList(ctx, s.SafeDB, pageSql, search, nonSite)
	if err != nil {
		return &entity.Sites{}, utils.CustomErrorf(err)
	}

	return sites, nil
}

// func: 현장 데이터 리스트 개수 조회
// @param
// -
func (s *ServiceSite) GetSiteNmCount(ctx context.Context, search entity.Site, nonSite int) (int, error) {

	count, err := s.Store.GetSiteNmCount(ctx, s.SafeDB, search, nonSite)
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}

	return count, nil
}

// func: 현장 상태 조회
// @param
// -
func (s *ServiceSite) GetSiteStatsList(ctx context.Context, targetDate time.Time) (*entity.Sites, error) {
	// 프로젝트별 휴무일 조회
	sites, err := s.Store.GetSiteStatsList(ctx, s.SafeDB, targetDate)
	if err != nil {
		return &entity.Sites{}, utils.CustomErrorf(err)
	}

	// 공휴일 조회
	year := strconv.Itoa(targetDate.Year())
	restDates, err := s.RestDateApiService.GetRestDelDates(year, "")
	if err != nil {
		return &entity.Sites{}, utils.CustomErrorf(err)
	}

	// 주말, 휴무일인지 체크
	isRest := targetDate.Weekday() == time.Saturday || targetDate.Weekday() == time.Sunday
	if !isRest {
		for _, rest := range restDates {
			rd, err := time.ParseInLocation("20060102", strconv.FormatInt(rest.RestDate, 10), targetDate.Location())
			if err != nil {
				continue
			}
			if rd.Year() == targetDate.Year() &&
				rd.Month() == targetDate.Month() &&
				rd.Day() == targetDate.Day() {
				isRest = true
				break
			}
		}
	}

	// 주말,공휴일 이라면 미운영을 휴무일 상태로 변경
	if isRest {
		for _, site := range *sites {
			if site.CurrentSiteStats.String == "C" {
				site.CurrentSiteStats.String = "H"
			}
		}
	}

	return sites, nil
}

// func: 현장 수정
// @param
// -
func (s *ServiceSite) ModifySite(ctx context.Context, site entity.Site) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	if site.Sno.Int64 == 0 {
		return utils.CustomErrorf(fmt.Errorf("sno parameter is missing"))
	}
	// 비고 정보 수정
	if err = s.Store.ModifySite(ctx, tx, site); err != nil {
		return utils.CustomErrorf(err)
	}

	// 기본 프로젝트 변경
	project := entity.ReqProject{
		Jno: site.DefaultJno,
		Sno: site.Sno,
		Base: entity.Base{
			ModUno:  site.ModUno,
			ModUser: site.ModUser,
		},
	}
	if err = s.ProjectStore.ModifyDefaultProject(ctx, tx, project); err != nil {
		return utils.CustomErrorf(err)
	}

	// 프로젝트 정보 수정
	for _, prj := range *site.ProjectList {
		// 공정률 수정
		workRate := entity.SiteWorkRate{
			Sno:        prj.Sno,
			Jno:        prj.Jno,
			WorkRate:   prj.WorkRate,
			SearchDate: site.SelectDate,
			Base: entity.Base{
				ModUno:  site.ModUno,
				ModUser: site.ModUser,
			},
		}
		if err = s.Store.ModifyWorkRate(ctx, tx, workRate); err != nil {
			return utils.CustomErrorf(err)
		}
	}

	// 날짜 수정 정보가 있는 경우만 실행
	if site.SiteDate != nil {
		siteDate := *site.SiteDate
		if err = s.SiteDateStore.ModifySiteDate(ctx, tx, site.Sno.Int64, siteDate); err != nil {
			return utils.CustomErrorf(err)
		}
	}

	// 장소 수정 할 정보가 있는 경우만 실행
	if site.SitePos != nil && site.SitePos.RoadAddress.String != "" {
		sitePos := *site.SitePos
		point, err := s.AddressSearchAPIService.GetAPILatitudeLongtitude(site.SitePos.RoadAddress.String)
		if err != nil {
			return utils.CustomErrorf(err)
		}
		if point.Latitude != 0.0 {
			sitePos.Latitude.Float64 = point.Latitude
			sitePos.Latitude.Valid = true
		}
		if point.Longitude != 0.0 {
			sitePos.Longitude.Float64 = point.Longitude
			sitePos.Longitude.Valid = true
		}

		if err = s.SitePosStore.ModifySitePosData(ctx, tx, site.Sno.Int64, sitePos); err != nil {
			return utils.CustomErrorf(err)
		}
	}

	return

}

// func: 현장 생성
// @param
// -
func (s *ServiceSite) AddSite(ctx context.Context, jno int64, user entity.User) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	err = s.Store.AddSite(ctx, s.SafeDB, tx, jno, user)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	return nil
}

// func: 현장 사용안함 변경
// @param
// -
func (s *ServiceSite) ModifySiteIsNonUse(ctx context.Context, site entity.ReqSite) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// 현장
	if err = s.Store.ModifySiteIsNonUse(ctx, tx, site); err != nil {
		return utils.CustomErrorf(err)
	}

	// 프로젝트
	if err = s.ProjectStore.ModifyProjectIsNonUse(ctx, tx, site); err != nil {
		return utils.CustomErrorf(err)
	}

	/*
		* 현장의 status만으로 관리(정보 불러올 때, site의 sno로 불러오기 때문에)
		// 위치
		if err = s.SitePosStore.ModifySitePosIsNonUse(ctx, tx, site); err != nil {
			return utils.CustomErrorf(err)
		}

		// 날짜
		if err = s.SiteDateStore.ModifySiteDateIsNonUse(ctx, tx, site); err != nil {
			return utils.CustomErrorf(err)
		}
	*/
	return
}

// func: 현장 사용함 변경
// @param
// -
func (s *ServiceSite) ModifySiteIsUse(ctx context.Context, site entity.ReqSite) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// 현장
	if err = s.Store.ModifySiteIsUse(ctx, tx, site); err != nil {
		return utils.CustomErrorf(err)
	}

	// 프로젝트
	if err = s.ProjectStore.ModifyProjectIsUse(ctx, tx, site); err != nil {
		return utils.CustomErrorf(err)
	}

	/*
		** 현장의 status만으로 관리(정보 불러올 때, site의 sno로 불러오기 때문에)
		// 위치
		if err = s.SitePosStore.ModifySitePosIsUse(ctx, tx, site); err != nil {
			return utils.CustomErrorf(err)
		}

		// 날짜
		if err = s.SiteDateStore.ModifySiteDateIsUse(ctx, tx, site); err != nil {
			return utils.CustomErrorf(err)
		}
	*/
	return
}

// func: 현장 프로젝트 사용안함 변경
// @param
// -
func (s *ServiceSite) ModifySiteJobNonUse(ctx context.Context, site entity.ReqSite) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// 프로젝트
	if err = s.ProjectStore.ModifyProjectIsNonUse(ctx, tx, site); err != nil {
		return utils.CustomErrorf(err)
	}
	return
}

// func: 현장 프로젝트 사용안함 변경
// @param
// -
func (s *ServiceSite) ModifySiteJobUse(ctx context.Context, site entity.ReqSite) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// 프로젝트
	if err = s.ProjectStore.ModifyProjectIsUse(ctx, tx, site); err != nil {
		return utils.CustomErrorf(err)
	}
	return
}

// func: 공정률 전날 수치로 세팅
// @param
// -
func (s *ServiceSite) SettingWorkRate(ctx context.Context, targetDate time.Time) (count int64, err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	count, err = s.Store.SettingWorkRate(ctx, tx, targetDate)
	if err != nil {
		return 0, utils.CustomErrorf(err)
	}

	return
}

// func: 공정률 수정
// @param
// -
func (s *ServiceSite) ModifyWorkRate(ctx context.Context, workRate entity.SiteWorkRate) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	err = s.Store.ModifyWorkRate(ctx, tx, workRate)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	return
}

// func: 날짜별 공정률 조회
// @param
// -
func (s *ServiceSite) GetSiteWorkRateByDate(ctx context.Context, jno int64, month string) (entity.SiteWorkRate, error) {
	data, err := s.Store.GetSiteWorkRateByDate(ctx, s.SafeDB, jno, month)
	if err != nil {
		return data, utils.CustomErrorf(err)
	}
	return data, nil
}

// func: 월별 공정률 조회
// @param
// -
func (s *ServiceSite) GetSiteWorkRateListByMonth(ctx context.Context, jno int64, month string) (entity.SiteWorkRates, error) {
	workRates, err := s.Store.GetSiteWorkRateListByMonth(ctx, s.SafeDB, jno, month)
	if err != nil {
		return workRates, utils.CustomErrorf(err)
	}
	return workRates, nil
}

// func: 공정률 추가
// @param
// -
func (s *ServiceSite) AddWorkRate(ctx context.Context, workRate entity.SiteWorkRate) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	err = s.Store.AddWorkRate(ctx, tx, workRate)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	return
}
