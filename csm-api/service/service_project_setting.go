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
	"time"
)

type ServiceProjectSetting struct {
	SafeDB        store.Queryer
	SafeTDB       store.Beginner
	Store         store.ProjectSettingStore
	WorkHourStore store.WorkHourStore
}

// func: 프로젝트에 설정된 공수 조회
// @param
// - jno: 프로젝트pk
func (s *ServiceProjectSetting) GetManHourList(ctx context.Context, jno int64) (*entity.ManHours, error) {

	manhours, err := s.Store.GetManHourList(ctx, s.SafeDB, jno)
	if err != nil {
		return &entity.ManHours{}, utils.CustomErrorf(err)
	}

	return manhours, nil

}

// func: 공수 수정 및 추가 (수정 시 기존 공수 삭제 후 새로 넣는 방식)
// @param
// - manHours: 공수 정보 배열
func (s *ServiceProjectSetting) MergeManHours(ctx context.Context, manHours *entity.ManHours) (err error) {

	jno := (*manHours)[0].Jno.Int64
	user := (*manHours)[0].Base

	deleteManhours, err := s.Store.GetManHourList(ctx, s.SafeDB, jno)
	if err != nil {
		return utils.CustomErrorf(err)
	}
	// jno에 해당하는 공수 찾기
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// jno에 해당하는 공수 모두 삭제
	if deleteManhours != nil && len(*deleteManhours) > 0 {
		for _, deleteManhour := range *deleteManhours {
			if deleteManhour == nil {
				continue
			}
			deleteManhour.Message.Valid = true
			deleteManhour.Message.String = fmt.Sprintf(`[DELETE] mhno:[before:%d, after: N/A]|work_hour:[before: %d, after: N/A]|man_hour:[before:%.2f, after: N/A]|jno:[before:%d, after: N/A]|etc:[before:%s, after: N/A]`, deleteManhour.Mhno.Int64, deleteManhour.WorkHour.Int64, deleteManhour.ManHour.Float64, deleteManhour.Jno.Int64, deleteManhour.Etc.String)
			deleteManhour.Base = user

			// 삭제
			//if err = s.DeleteManHour(ctx, deleteManhour.Mhno.Int64, *deleteManhour); err != nil {
			//	return utils.CustomErrorf(err)
			//}
			// 공수 삭제
			if err = s.Store.DeleteManHour(ctx, tx, deleteManhour.Mhno.Int64); err != nil {
				return utils.CustomErrorf(err)
			}

			// 공수 삭제 시 근로자 업데이트
			jno = deleteManhour.Jno.Int64
			user = deleteManhour.Base
			if err = s.WorkHourStore.ModifyWorkHourByJno(ctx, tx, jno, user, nil); err != nil {
				return utils.CustomErrorf(err)
			}

			// 로그 기록
			if err = s.Store.ManHourLog(ctx, tx, *deleteManhour); err != nil {
				return utils.CustomErrorf(err)
			}
		}
	}

	// 받아온 공수 추가
	for _, manHour := range *manHours {
		if !manHour.Message.Valid {
			continue
		}

		// 추가
		err = s.Store.AddManHour(ctx, tx, *manHour)
		if err != nil {
			return utils.CustomErrorf(err)
		}

		// 로그 남기기
		if err = s.Store.ManHourLog(ctx, tx, *manHour); err != nil {
			return utils.CustomErrorf(err)
		}
	}

	if !user.ModUser.Valid || !user.ModUno.Valid {
		user.ModUser = utils.ParseNullString("SYSTEM")
		user.ModUno = utils.ParseNullInt("0")
	}

	// 공수에 맞춰 근로자 업데이트
	if err = s.WorkHourStore.ModifyWorkHourByJno(ctx, tx, jno, user, nil); err != nil {
		return utils.CustomErrorf(err)
	}

	return
}

// func: 프로젝트 설정 정보 추가 및 수정
// @param
// - ProjectSetting
func (s *ServiceProjectSetting) MergeProjectSetting(ctx context.Context, project entity.ProjectSetting) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	if !project.Message.Valid {
		return
	}

	count, err := s.Store.MergeProjectSetting(ctx, tx, project)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	if count <= 0 {
		return
	} else {
		// 프로젝트 재설정 시 근로자 업데이트
		jno := project.Jno.Int64
		user := project.Base
		if err = s.WorkHourStore.ModifyWorkHourByJno(ctx, tx, jno, user, nil); err != nil {
			return utils.CustomErrorf(err)
		}
	}

	if err = s.Store.ProjectSettingLog(ctx, tx, project); err != nil {
		return utils.CustomErrorf(err)
	}

	return
}

// func: 프로젝트 미설정 정보 업데이트(스케줄러)
// @param
// -
func (s *ServiceProjectSetting) CheckProjectSetting(ctx context.Context) (count int, err error) {

	projectManHours := &entity.ProjectSettings{}
	if projectManHours, err = s.Store.GetCheckProjectManHours(ctx, s.SafeDB); err != nil {
		return 0, utils.CustomErrorf(err)
	}
	for _, projectManHour := range *projectManHours {
		// 기본 공수 추가하기
		manHourMore := &entity.ManHour{}

		manHourMore.WorkHour = utils.ParseNullInt("8")
		manHourMore.ManHour = utils.ParseNullFloat("0.5")
		manHourMore.Jno = projectManHour.Jno
		manHourMore.Message = utils.ParseNullString(fmt.Sprintf("[ADD] jno:[before:N/A, after:%d]|work_hour:[before:N/A, after:8]|man_hour:[before:N/A, after:0.5]|etc:[before:N/A, after:]", projectManHour.Jno.Int64))
		manHours := entity.ManHours{manHourMore}

		if err = s.MergeManHours(ctx, &manHours); err != nil {
			return 0, utils.CustomErrorf(err)
		}
	}

	projects := &entity.ProjectSettings{}
	if projects, err = s.Store.GetCheckProjectSetting(ctx, s.SafeDB); err != nil {
		return 0, utils.CustomErrorf(err)
	}

	for _, project := range *projects {

		// 프로젝트 기본값으로 설정하기
		setting := &entity.ProjectSetting{}

		setting.Jno = project.Jno
		loc, _ := time.LoadLocation("Asia/Seoul")
		setting.InTime = null.NewTime(time.Date(2006, 01, 02, 8, 0, 0, 0, loc), true)
		setting.OutTime = null.NewTime(time.Date(2006, 01, 02, 17, 0, 0, 0, loc), true)
		setting.RespiteTime = utils.ParseNullInt("30")
		setting.CancelCode = utils.ParseNullString("NO_DAY")
		setting.Message = utils.ParseNullString(fmt.Sprintf("[ADD] jno:[before:N/A, after:%d]|in_time:[before:N/A, after:2006-01-02T08:00:00+09:00]|out_time:[before:N/A, after:2006-01-02T17:00:00+09:00]|respite_time:[before:N/A, after:30]|cancel_code:[before:N/A, after:NO_DAY]", project.Jno.Int64))
		if err = s.MergeProjectSetting(ctx, *setting); err != nil {
			return 0, utils.CustomErrorf(err)
		}

	}
	count = (len(*projectManHours) + len(*projects)) / 2

	return
}

// func: 프로젝트 설정 정보 가져오기
// @param
// - jno: 프로젝트PK
func (s *ServiceProjectSetting) GetProjectSetting(ctx context.Context, jno int64, isRole bool) (*entity.ProjectSettings, error) {

	uno, _ := auth.GetContext(ctx, auth.Uno{})

	setting, err := s.Store.GetProjectSetting(ctx, s.SafeDB, isRole, uno, jno)
	if err != nil {
		return &entity.ProjectSettings{}, utils.CustomErrorf(err)
	}

	manHours, err := s.GetManHourList(ctx, jno)
	if err != nil {
		return &entity.ProjectSettings{}, utils.CustomErrorf(err)
	}

	if len(*setting) > 0 {
		(*setting)[0].ManHours = manHours
	}

	return setting, nil

}

// func: 공수 삭제
// @param
// - mhno: 공수pk
func (s *ServiceProjectSetting) DeleteManHour(ctx context.Context, mhno int64, manhour entity.ManHour) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// 공수 삭제
	if err = s.Store.DeleteManHour(ctx, tx, mhno); err != nil {
		return utils.CustomErrorf(err)
	}

	// 공수 삭제 시 근로자 업데이트
	jno := manhour.Jno.Int64
	user := manhour.Base
	if err = s.WorkHourStore.ModifyWorkHourByJno(ctx, tx, jno, user, nil); err != nil {
		return utils.CustomErrorf(err)
	}

	// 로그 기록
	if err = s.Store.ManHourLog(ctx, tx, manhour); err != nil {
		return utils.CustomErrorf(err)
	}
	return
}

// 공수 추가(삭제 없이 추가만)
func (s *ServiceProjectSetting) AddManHour(ctx context.Context, manhour entity.ManHour) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	if err = s.Store.AddManHour(ctx, tx, manhour); err != nil {
		return utils.CustomErrorf(err)
	}

	manhour.Message = utils.ParseNullString(fmt.Sprintf("[ADD] jno:[before:N/A, after:%d]|work_hour:[before:N/A, after:%d]|man_hour:[before:N/A, after:%f]|etc:[before:N/A, after:%s]", manhour.Jno.Int64, manhour.WorkHour.Int64, manhour.ManHour.Float64, manhour.Etc.String))

	// 로그 기록
	if manhour.Message.Valid {
		if err = s.Store.ManHourLog(ctx, tx, manhour); err != nil {
			return utils.CustomErrorf(err)
		}
	}

	return
}
