package service

import (
	"context"
	"csm-api/entity"
	"csm-api/store"
	"csm-api/txutil"
	"csm-api/utils"
	"fmt"
	"github.com/guregu/null"
	"github.com/xuri/excelize/v2"
	"strconv"
	"strings"
	"time"
)

type ServiceExcel struct {
	SafeDB      store.Queryer
	SafeTDB     store.Beginner
	Store       store.ExcelStore
	WorkerStore store.WorkerStore
	FileStore   store.UploadFileStore
}

func mustGet(f *excelize.File, sheet, cell string) string {
	val, _ := f.GetCellValue(sheet, cell)
	return strings.TrimSpace(val)
}

// TBM excel import
func (s *ServiceExcel) ImportTbm(ctx context.Context, path string, tbm entity.Tbm, file entity.UploadFile) (err error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	// tbm 차수
	order, err := s.Store.GetTbmOrder(ctx, s.SafeDB, tbm)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	// 파일 차수
	uploadRound, err := s.FileStore.GetUploadRound(ctx, s.SafeDB, file)
	if err != nil {
		return utils.CustomErrorf(err)
	}
	file.UploadRound = utils.ParseNullInt(strconv.Itoa(uploadRound))

	var tbmList []entity.Tbm

	type cellPair struct {
		nameCol string
		signCol string
	}

	pairs := []cellPair{
		{"C", "D"},
		{"G", "H"},
		{"K", "L"},
	}

	// 시트 전체 순회
	for _, sheetName := range f.GetSheetList() {
		for row := 25; row <= 34; row++ {
			for _, pair := range pairs {
				nameCell := pair.nameCol + fmt.Sprint(row)
				signCell := pair.signCol + fmt.Sprint(row)

				name, _ := f.GetCellValue(sheetName, nameCell)
				sign, _ := f.GetCellValue(sheetName, signCell)

				if name != "" && sign != "" {
					newTbm := entity.Tbm{
						Sno:        tbm.Sno,
						Jno:        tbm.Jno,
						Department: tbm.Department,
						DiscName:   tbm.DiscName,
						TbmDate:    tbm.TbmDate,
						TbmOrder:   utils.ParseNullInt(order),
						UserNm:     utils.ParseNullString(name),
						Base: entity.Base{
							RegUser: tbm.RegUser,
							RegUno:  tbm.RegUno,
						},
					}
					tbmList = append(tbmList, newTbm)
				}
			}
		}
	}

	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}
	defer txutil.DeferTx(tx, &err)

	// tbm 저장
	if err = s.Store.AddTbmExcel(ctx, tx, tbmList); err != nil {
		return utils.CustomErrorf(err)
	}

	// file 정보 저장
	if err = s.FileStore.AddUploadFile(ctx, tx, file); err != nil {
		return utils.CustomErrorf(err)
	}

	return
}

// 퇴직공제 excel import
func (s *ServiceExcel) ImportDeduction(ctx context.Context, path string, deduction entity.Deduction, file entity.UploadFile) (err error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	order, err := s.Store.GetDeductionOrder(ctx, s.SafeDB, deduction)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	// 파일 차수
	uploadRound, err := s.FileStore.GetUploadRound(ctx, s.SafeDB, file)
	if err != nil {
		return utils.CustomErrorf(err)
	}
	file.UploadRound = utils.ParseNullInt(strconv.Itoa(uploadRound))

	siteNm, err := s.Store.GetDeductionSiteNameBySno(ctx, s.SafeDB, deduction.Sno.Int64)
	if err != nil {
		return utils.CustomErrorf(err)
	}

	sheetName := f.GetSheetName(0)

	var deductionList []entity.Deduction

	// 3번째 행부터 (1-based index, 엑셀은 B3부터 시작)
	for rowIdx := 3; ; rowIdx++ {
		// B열(근무날짜)
		cellAddr := fmt.Sprintf("B%d", rowIdx)
		dateStr, err := f.GetCellValue(sheetName, cellAddr)
		if err != nil || strings.TrimSpace(dateStr) == "" {
			break
		}
		if dateStr == "" || utils.ConvertMMDDYYToYYMMDD(dateStr) != deduction.RecordDate.Time.Format("06-01-02") {
			continue
		}

		// C열(현장명)
		siteName, _ := f.GetCellValue(sheetName, fmt.Sprintf("C%d", rowIdx))
		if utils.NormalizeForEqual(siteName) != utils.NormalizeForEqual(siteNm) {
			continue
		}

		// G열(이름)
		userNm := mustGet(f, sheetName, fmt.Sprintf("G%d", rowIdx))
		// F열(회사명)
		department := mustGet(f, sheetName, fmt.Sprintf("F%d", rowIdx))
		// I열(성별)
		gender := mustGet(f, sheetName, fmt.Sprintf("N%d", rowIdx))
		// H열(생년월일)
		regNo := mustGet(f, sheetName, fmt.Sprintf("H%d", rowIdx))
		// I열(전화번호)
		phone := mustGet(f, sheetName, fmt.Sprintf("I%d", rowIdx))
		normalizedPhone := strings.ReplaceAll(strings.ReplaceAll(phone, "-", ""), " ", "")
		if len(normalizedPhone) == 10 && strings.HasPrefix(normalizedPhone, "1") {
			normalizedPhone = "0" + normalizedPhone
		}
		// O열(출근시간)
		inTime := mustGet(f, sheetName, fmt.Sprintf("O%d", rowIdx))
		// P열(퇴근시간)
		outTime := mustGet(f, sheetName, fmt.Sprintf("P%d", rowIdx))

		newDeduction := entity.Deduction{
			Sno:          deduction.Sno,
			UserNm:       utils.ParseNullString(userNm),
			Department:   utils.ParseNullString(department),
			Gender:       utils.ParseNullString(gender),
			RegNo:        utils.ParseNullString(utils.ConvertMMDDYYToYYMMDD(regNo)),
			Phone:        utils.ParseNullString(normalizedPhone),
			InRecogTime:  utils.ParseNullDateTime(deduction.RecordDate.Time.Format("2006-01-02"), inTime),
			OutRecogTime: utils.ParseNullDateTime(deduction.RecordDate.Time.Format("2006-01-02"), outTime),
			RecordDate:   deduction.RecordDate,
			DeductOrder:  utils.ParseNullString(order),
			Base: entity.Base{
				RegUser: deduction.RegUser,
				RegUno:  deduction.RegUno,
			},
		}

		if newDeduction.UserNm.Valid && newDeduction.Department.Valid && newDeduction.Gender.Valid && newDeduction.Phone.Valid {
			deductionList = append(deductionList, newDeduction)
		}
	}

	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return utils.CustomErrorf(err)
	}
	defer txutil.DeferTx(tx, &err)

	// 퇴직공제 저장
	if err = s.Store.AddDeductionExcel(ctx, tx, deductionList); err != nil {
		return utils.CustomErrorf(err)
	}

	// file 정보 저장
	if err = s.FileStore.AddUploadFile(ctx, tx, file); err != nil {
		return utils.CustomErrorf(err)
	}

	return
}

// 현장근로자 업로드
func (s *ServiceExcel) ImportAddDailyWorker(ctx context.Context, path string, worker entity.WorkerDaily) (list entity.WorkerDailys, err error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return list, utils.CustomErrorf(err)
	}

	sheet := f.GetSheetName(0)
	var excels []entity.WorkerDailyExcel

	row := 2
	for {
		// B: (이름) 기준으로 값이 없으면 종료
		userNm, err := f.GetCellValue(sheet, fmt.Sprintf("B%d", row))
		if err != nil || strings.TrimSpace(userNm) == "" {
			break
		}
		// C: 생년월일
		birthRaw, _ := f.GetCellValue(sheet, fmt.Sprintf("C%d", row))
		regNo := strings.ReplaceAll(birthRaw, "-", "")
		if len(regNo) == 8 {
			regNo = regNo[2:] // 앞 2자리 제거
		}
		// D: 핸드폰번호
		rawPhone, _ := f.GetCellValue(sheet, fmt.Sprintf("D%d", row))
		normalizedPhone := strings.ReplaceAll(strings.ReplaceAll(rawPhone, "-", ""), " ", "")
		if strings.HasPrefix(normalizedPhone, "1") {
			normalizedPhone = "0" + normalizedPhone
		}
		// E: 근로날짜
		workDate, _ := f.GetCellValue(sheet, fmt.Sprintf("E%d", row))
		if !utils.IsYYYYMMDD(workDate) {
			workDate = utils.NormalizeYYMMDD(utils.ConvertMMDDYYToYYMMDD(workDate))
		}
		// F: 출근시간 → 시간 서식으로 저장됨
		inTimeRaw, err := f.GetCellValue(sheet, fmt.Sprintf("F%d", row))
		if err != nil {
			return list, utils.CustomErrorf(err)
		}
		inTime := inTimeRaw
		if timeVal, err := f.GetCellValue(sheet, fmt.Sprintf("F%d", row), excelize.Options{RawCellValue: false}); err == nil {
			inTime = timeVal
		}
		// G: 퇴근시간 → 시간 서식으로 저장됨
		outTimeRaw, err := f.GetCellValue(sheet, fmt.Sprintf("G%d", row))
		if err != nil {
			return list, utils.CustomErrorf(err)
		}
		outTime := outTimeRaw
		if timeVal, err := f.GetCellValue(sheet, fmt.Sprintf("G%d", row), excelize.Options{RawCellValue: false}); err == nil {
			outTime = timeVal
		}
		// H: 공수
		workHour, _ := f.GetCellValue(sheet, fmt.Sprintf("H%d", row))

		excels = append(excels, entity.WorkerDailyExcel{
			RegNo:    regNo,
			UserNm:   userNm,
			Phone:    normalizedPhone,
			WorkDate: workDate,
			InTime:   inTime,
			OutTime:  outTime,
			WorkHour: workHour,
		})

		row++
	}

	var workers entity.WorkerDailys
	regDate := null.NewTime(time.Now(), true)
	for _, excel := range excels {
		temp := entity.WorkerDaily{
			Sno:          worker.Sno,
			Jno:          worker.Jno,
			UserNm:       utils.ParseNullString(excel.UserNm),
			UserId:       utils.ParseNullString(excel.Phone),
			RegNo:        utils.ParseNullString(excel.RegNo),
			RecordDate:   utils.ParseNullDate(excel.WorkDate),
			Phone:        utils.ParseNullString(excel.Phone),
			InRecogTime:  utils.ParseNullDateTime(excel.WorkDate, utils.NormalizeHHMM(excel.InTime)),
			OutRecogTime: utils.ParseNullDateTime(excel.WorkDate, utils.NormalizeHHMM(excel.OutTime)),
			WorkHour:     utils.ParseNullFloat(excel.WorkHour),
			CompareState: utils.ParseNullString("X"),
			WorkState:    utils.ParseNullString("02"),
			Base: entity.Base{
				RegDate: regDate,
				RegUser: worker.RegUser,
				RegUno:  worker.RegUno,
			},
			WorkerReason: entity.WorkerReason{
				Reason:     worker.Reason,
				ReasonType: worker.ReasonType,
				HisStatus:  utils.ParseNullString("AFTER"),
			},
		}

		var userKey string
		if userKey, err = s.WorkerStore.GetDailyWorkerUserKey(ctx, s.SafeDB, temp); err != nil {
			continue
		}
		temp.UserKey = utils.ParseNullString(userKey)
		workers = append(workers, &temp)
	}

	// 업로드 전 데이터 조회
	beforeList, err := s.WorkerStore.GetDailyWorkerBeforeList(ctx, s.SafeDB, workers)
	if err != nil {
		return list, utils.CustomErrorf(err)
	}
	for i := range beforeList {
		beforeList[i].HisStatus = utils.ParseNullString("BEFORE")
		beforeList[i].RegDate = regDate
	}

	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return list, utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	// 업로드 데이터 추가/수정
	if list, err = s.WorkerStore.AddDailyWorkers(ctx, s.SafeDB, tx, workers); err != nil {
		return list, utils.CustomErrorf(err)
	}

	// 변경사항 로그 저장
	if err = s.WorkerStore.MergeSiteBaseWorkerLog(ctx, tx, list); err != nil {
		return list, utils.CustomErrorf(err)
	}

	// 업로드 전 데이터 저장
	if err = s.WorkerStore.AddHistoryDailyWorkers(ctx, tx, workers); err != nil {
		return list, utils.CustomErrorf(err)
	}
	// 업로드 후 데이터 저장
	if err = s.WorkerStore.AddHistoryDailyWorkers(ctx, tx, beforeList); err != nil {
		return list, utils.CustomErrorf(err)
	}

	return
}

// 전체근로자 업로드
func (s *ServiceExcel) ImportAddWorker(ctx context.Context, path string, worker entity.Worker) (successList entity.Workers, err error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return successList, utils.CustomErrorf(err)
	}

	sheet := f.GetSheetName(0)
	var excels entity.Workers

	row := 2
	for {
		// B: (이름) 기준으로 값이 없으면 종료
		userNm, err := f.GetCellValue(sheet, fmt.Sprintf("B%d", row))
		if err != nil || strings.TrimSpace(userNm) == "" {
			break
		}
		// C: 주민등록번호
		identityNumberRaw, _ := f.GetCellValue(sheet, fmt.Sprintf("C%d", row))
		regNo := strings.ReplaceAll(identityNumberRaw, "-", "")
		if len(regNo) == 8 {
			regNo = regNo[2:] // 앞 2자리 제거
		}
		// D: 아이디
		rawId, _ := f.GetCellValue(sheet, fmt.Sprintf("D%d", row))
		normalizedId := strings.ReplaceAll(strings.ReplaceAll(rawId, "-", ""), " ", "")
		if strings.HasPrefix(normalizedId, "1") {
			normalizedId = "0" + normalizedId
		}
		// E: 부서/조직명
		department, _ := f.GetCellValue(sheet, fmt.Sprintf("E%d", row))

		// F: 핸드폰 번호
		rawPhone, _ := f.GetCellValue(sheet, fmt.Sprintf("D%d", row))
		normalizedPhone := strings.ReplaceAll(strings.ReplaceAll(rawPhone, "-", ""), " ", "")
		if strings.HasPrefix(normalizedPhone, "1") {
			normalizedPhone = "0" + normalizedPhone
		}

		// G: 공종
		discName, _ := f.GetCellValue(sheet, fmt.Sprintf("G%d", row))

		// H: 퇴직여부
		retire, _ := f.GetCellValue(sheet, fmt.Sprintf("H%d", row))
		var isRetire string
		if utils.NormalizeBool(retire) {
			isRetire = "Y"
		} else {
			isRetire = "N"
		}

		// 근로자 구분
		codeNm, _ := f.GetCellValue(sheet, fmt.Sprintf("I%d", row))

		newWorker := &entity.Worker{
			Sno:        worker.Sno,
			Jno:        worker.Jno,
			UserNm:     utils.ParseNullString(userNm),
			RegNo:      utils.ParseNullString(identityNumberRaw),
			UserId:     utils.ParseNullString(normalizedId),
			Department: utils.ParseNullString(department),
			Phone:      utils.ParseNullString(normalizedPhone),
			DiscName:   utils.ParseNullString(discName),
			IsRetire:   utils.ParseNullString(isRetire),
			CodeNm:     utils.ParseNullString(codeNm),
			Base: entity.Base{
				RegUser: worker.RegUser,
				RegUno:  worker.RegUno,
			},
			WorkerReason: entity.WorkerReason{
				Reason:     worker.Reason,
				ReasonType: worker.ReasonType,
				HisStatus:  utils.ParseNullString("AFTER"),
			},
		}

		excels = append(excels, newWorker)

		row++
	}

	tx, err := txutil.BeginTxWithMode(ctx, s.SafeTDB, false)
	if err != nil {
		return successList, utils.CustomErrorf(err)
	}

	defer txutil.DeferTx(tx, &err)

	var count int64
	var failList entity.Workers
	// 업로드 데이터 추가/수정
	for _, excelWorker := range excels {

		if count, err = s.WorkerStore.MergeWorker(ctx, tx, *excelWorker); err != nil {
			return successList, utils.CustomErrorf(err)
		} else if count != 0 {
			successList = append(successList, excelWorker)
		} else {
			failList = append(failList, excelWorker)
		}
	}

	return
}
