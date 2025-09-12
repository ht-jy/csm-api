package service

import (
	"context"
	"csm-api/auth"
	"csm-api/entity"
	"csm-api/store"
	"csm-api/txutil"
	"csm-api/utils"
	"strconv"
	"strings"
)

type ServiceCompare struct {
	SafeDB  store.Queryer
	SafeTDB store.Beginner
	Store   store.CompareStore
}

// 일일 근로자 비교 리스트
func (s *ServiceCompare) GetCompareList(ctx context.Context, compare entity.Compare, isRole bool, retry string, order string) ([]entity.Compare, error) {

	uno, _ := auth.GetContext(ctx, auth.Uno{})

	workerlist, err := s.Store.GetDailyWorkerList(ctx, s.SafeDB, compare, isRole, uno, retry, order)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	tbmList, err := s.Store.GetTbmList(ctx, s.SafeDB, compare, retry, order)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	deductionList, err := s.Store.GetDeductionList(ctx, s.SafeDB, compare, retry, order)
	if err != nil {
		return nil, utils.CustomErrorf(err)
	}

	// " ", "-" 제거
	replacer := strings.NewReplacer(" ", "", "-", "")
	// 주민번호 앞자리+뒷자리첫번째만 반환 ex) 000101-1234567 -> 0001011
	cleanRegNo := func(reg string) string {
		cleaned := replacer.Replace(reg)
		var birth string

		// 주민번호 앞 6자리
		if len(cleaned) >= 6 {
			birth = cleaned[:6]
		}

		// 주민번호가 7자리가 안될 경우 그대로 리턴
		if len(cleaned) < 7 {
			return birth
		}

		// 7번째 값 숫자로 변환 숫자로 변환이 안될시 그대로 리턴
		ch := cleaned[6]
		digit, err := strconv.Atoi(string(ch))
		if err != nil {
			return birth
		}

		// 0, 2, 4, 6, 8 = 여자, 1, 3, 5, 7, 9 = 남자
		if digit%2 == 0 {
			return birth + "2"
		}
		return birth + "1"
	}
	// 성별: 주민번호
	getBirthToRegNo := func(reg string) string {
		cleaned := replacer.Replace(reg)

		if len(cleaned) < 7 {
			return ""
		}

		ch := cleaned[6]
		digit, err := strconv.Atoi(string(ch))
		if err != nil {
			return ""
		}

		if digit%2 == 0 {
			return "여"
		}
		return "남"
	}
	// 생년월일 + 성별을 주민번호 앞자리+뒷자리첫번째 형태로 반환 ex) 00-01-01, 남 ->  0001011
	cleanBirthRegNo := func(reg string, gender string) string {
		cleaned := replacer.Replace(reg)

		// 성별에 따라 1 or 2
		if gender == "남" {
			return cleaned + "1"
		} else if gender == "여" {
			return cleaned + "2"
		}
		return cleaned
	}

	// TBM map: 동명이인 고려한 slice map
	tbmMap := make(map[entity.TbmKey][]entity.Tbm)
	for _, tbm := range tbmList {
		key := entity.TbmKey{
			tbm.Sno.Int64,
			0,
			//tbm.Jno.Int64,
			tbm.UserNm.String,
			tbm.Department.String,
			tbm.TbmDate.Time}
		tbmMap[key] = append(tbmMap[key], tbm)
	}

	// 공제 map: 주민번호 기준과 동명이인 기준
	deductionRegMap := make(map[entity.DeductionRegKey]entity.Deduction)
	deductionMap := make(map[entity.DeductionKey][]entity.Deduction)
	for _, d := range deductionList {
		regKey := entity.DeductionRegKey{
			d.Sno.Int64,
			0,
			//d.Jno.Int64,
			replacer.Replace(d.Phone.String),
			cleanBirthRegNo(d.RegNo.String, d.Gender.String),
			d.RecordDate.Time,
		}
		deductionRegMap[regKey] = d
		key := entity.DeductionKey{
			d.Sno.Int64,
			0,
			//d.Jno.Int64,
			d.UserNm.String,
			d.Department.String,
			d.RecordDate.Time,
		}
		deductionMap[key] = append(deductionMap[key], d)
	}
	var compareList []entity.Compare

	// 근태 기준 비교
	for _, worker := range workerlist {
		compareTemp := entity.Compare{
			Jno:           worker.Jno,
			UserKey:       worker.UserKey,
			UserId:        worker.UserId,
			UserNm:        worker.UserNm,
			Department:    worker.Department,
			DiscName:      worker.DiscName,
			Phone:         worker.Phone,
			Gender:        utils.ParseNullString(getBirthToRegNo(worker.RegNo.String)),
			IsTbm:         utils.ParseNullString("N"),
			DeviceNm:      worker.DeviceNm,
			RecordDate:    worker.RecordDate,
			WorkerInTime:  worker.InRecogTime,
			WorkerOutTime: worker.OutRecogTime,
			CompareState:  worker.CompareState,
			IsDeadline:    worker.IsDeadline,
		}

		// TBM 비교
		tbmKey := entity.TbmKey{
			worker.Sno.Int64,
			0,
			worker.UserNm.String,
			worker.Department.String,
			worker.RecordDate.Time,
		}
		//if worker.CompareState.String == "S" || worker.CompareState.String == "X" {
		//	tbmKey.Jno = worker.Jno.Int64
		//}
		if tbms, ok := tbmMap[tbmKey]; ok && len(tbms) > 0 {
			compareTemp.IsTbm = utils.ParseNullString("Y")
			compareTemp.DiscName = tbms[0].DiscName
			tbmMap[tbmKey] = tbms[1:] // 하나만 소비
			if len(tbmMap[tbmKey]) == 0 {
				delete(tbmMap, tbmKey)
			}
		}

		// 공제 비교 (RegNo 기준)
		deductKey := entity.DeductionRegKey{
			worker.Sno.Int64,
			0,
			replacer.Replace(worker.Phone.String),
			cleanRegNo(worker.RegNo.String),
			worker.RecordDate.Time,
		}
		//if worker.CompareState.String == "S" || worker.CompareState.String == "X" {
		//	deductKey.Jno = worker.Jno.Int64
		//}

		if deduction, ok := deductionRegMap[deductKey]; ok {
			compareTemp.DeductionInTime = deduction.InRecogTime
			compareTemp.DeductionOutTime = deduction.OutRecogTime
			compareTemp.DeductionBirth = deduction.RegNo
			delete(deductionRegMap, deductKey)

			dk := entity.DeductionKey{deduction.Sno.Int64, deduction.Jno.Int64, deduction.UserNm.String, deduction.Department.String, deduction.RecordDate.Time}
			if list := deductionMap[dk]; len(list) > 0 {
				deductionMap[dk] = list[1:]
				if len(deductionMap[dk]) == 0 {
					delete(deductionMap, dk)
				}
			}
		}

		compareList = append(compareList, compareTemp)
	}

	// 근태x 남은 TBM-공제 비교
	for _, tbmSlice := range tbmMap {
		for _, tbm := range tbmSlice {
			compareTemp := entity.Compare{
				Jno:          tbm.Jno,
				UserNm:       tbm.UserNm,
				Department:   tbm.Department,
				DiscName:     tbm.DiscName,
				IsTbm:        utils.ParseNullString("Y"),
				RecordDate:   tbm.TbmDate,
				CompareState: utils.ParseNullString("C"),
			}

			deductionKey := entity.DeductionKey{tbm.Sno.Int64, tbm.Jno.Int64, tbm.UserNm.String, tbm.Department.String, tbm.TbmDate.Time}
			if dList, ok := deductionMap[deductionKey]; ok && len(dList) > 0 {
				d := dList[0]
				compareTemp.DeductionInTime = d.InRecogTime
				compareTemp.DeductionOutTime = d.OutRecogTime
				compareTemp.Gender = d.Gender
				compareTemp.DeductionBirth = d.RegNo
				compareTemp.UserId = d.Phone
				deductionMap[deductionKey] = dList[1:]
				if len(deductionMap[deductionKey]) == 0 {
					delete(deductionMap, deductionKey)
				}

				// 삭제 동기화 (regMap도)
				regKey := entity.DeductionRegKey{d.Sno.Int64, d.Jno.Int64, replacer.Replace(d.Phone.String), cleanRegNo(d.RegNo.String), d.RecordDate.Time}
				delete(deductionRegMap, regKey)
			}

			compareList = append(compareList, compareTemp)
		}
	}

	// 근태x TBMx 공제 처리
	for _, dList := range deductionMap {
		for _, d := range dList {
			compareTemp := entity.Compare{
				Jno:              d.Jno,
				UserId:           d.Phone,
				UserNm:           d.UserNm,
				Department:       d.Department,
				Gender:           d.Gender,
				IsTbm:            utils.ParseNullString("N"),
				CompareState:     utils.ParseNullString("C"),
				RecordDate:       d.RecordDate,
				DeductionInTime:  d.InRecogTime,
				DeductionOutTime: d.OutRecogTime,
				DeductionBirth:   d.RegNo,
			}
			compareList = append(compareList, compareTemp)
		}
	}

	return compareList, nil
}

// 근로자 비교 반영
func (s *ServiceCompare) ModifyWorkerCompareApply(ctx context.Context, workers entity.WorkerDailys) (err error) {
	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// 근로자 정보: IRIS_WORKER_SET
	// 선택한 프로젝트로 수정
	if err = s.Store.ModifyWorkerCompareApply(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 근로자 비교 반영 - 근로자 일일 정보: IRIS_WORKER_DAILY_SET
	// 반영상태, 선택한 프로젝트로 수정
	if err = s.Store.ModifyDailyWorkerCompareApply(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 근로자 비교 반영 - TBM 등록 정보: IRIS_TBM_SET
	// 선택한 프로젝트로 수정
	if err = s.Store.ModifyTbmCompareApply(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 근로자 비교 반영 - 퇴직공제 등록 정보: IRIS_DEDUCTION_SET
	// 선택한 프로젝트로 수정
	if err = s.Store.ModifyDeductionCompareApply(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	// 비교 상태 수정 로그 등록
	if err = s.Store.AddCompareLog(ctx, tx, workers); err != nil {
		return utils.CustomErrorf(err)
	}

	return
}
