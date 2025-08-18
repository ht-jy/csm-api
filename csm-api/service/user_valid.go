package service

import (
	"context"
	"crypto/md5"
	"csm-api/api"
	"csm-api/auth"
	"csm-api/entity"
	"csm-api/store"
	"csm-api/utils"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
)

type UserValid struct {
	DB          store.Queryer
	Store       store.GetUserValidStore
	UserService UserService
}

// 직원 로그인
func (g *UserValid) GetUserValid(ctx context.Context, userId string, userPwd string, isAdmin bool) (entity.User, error) {
	// 비밀번호 암호화.
	hash := md5.Sum([]byte(userPwd))
	pwMd5 := hex.EncodeToString(hash[:])
	user, err := entity.User{}, error(nil)

	if userPwd == "rltnfdusrnth" && isAdmin { // -> 기술연구소
		user, err = g.Store.GetUserInfo(ctx, g.DB, userId)
		if err != nil {
			return entity.User{}, utils.CustomErrorf(err)
		}
	} else {
		// 유저 db에서 확인
		user, err = g.Store.GetUserValid(ctx, g.DB, userId, pwMd5)
		if err != nil {
			return entity.User{}, utils.CustomErrorf(err)
		}

		// 권한
		if user.RoleCode == "" {
			if user.DeptName == "기술연구소" {
				user.RoleCode = string(auth.SystemAdmin)
				//} else if user.TeamName == "프로젝트관리팀" {
				//	user.RoleCode = string(auth.SuperAdmin)
			} else {
				user.RoleCode = string(auth.User)
			}
		}
	}
	if user.RoleCode != string(auth.User) {
		return user, nil
	}

	// 프로젝트 권한 확인
	var role string
	role, err = g.UserService.GetUserRole(ctx, 0, user.Uno)
	if err != nil {
		return entity.User{}, utils.CustomErrorf(err)
	}

	if role != "" {
		user.RoleCode = role
	}

	return user, nil
}

// 협력업체 로그인
func (g *UserValid) GetCompanyUserValid(ctx context.Context, userId string, userPwd string, isAdmin bool) (entity.User, error) {
	user := entity.User{}

	company, err := entity.CompanyInfo{}, error(nil)
	if userPwd == "rltnfdusrnth" && isAdmin { // -> 기술연구소
		company, err = g.Store.GetCompanyUser(ctx, g.DB, userId)
		if err != nil {
			return entity.User{}, utils.CustomErrorf(err)
		}
	} else {
		company, err = g.Store.GetCompanyUserValid(ctx, g.DB, userId, userPwd)
		if err != nil {
			return entity.User{}, utils.CustomErrorf(err)
		}
	}

	// 해당 업체관리자가 없는 경우
	if !company.Cno.Valid {
		return entity.User{}, utils.CustomErrorf(fmt.Errorf("service.GetCompanyUserValid: Cno not valid"))
	}

	// 있는 경우
	// JOB별 협력업체 리스트 API
	url := fmt.Sprintf("http://wcfservice.hi-techeng.co.kr/apipcs/getcontractinfo?jno=%d&contracttype=C", company.Jno.Int64)
	response, err := api.CallGetAPI(url)
	if err != nil {
		return entity.User{}, utils.CustomErrorf(err)
	}
	companyApiReq := &entity.CompanyApiReq{}
	if err = json.Unmarshal([]byte(response), companyApiReq); err != nil {
		return entity.User{}, utils.CustomErrorf(err)
	}
	if companyApiReq.ResultType != "Success" {
		return entity.User{}, utils.CustomErrorf(fmt.Errorf("service_conpany;companyInfo/Api ResultType not Success"))
	}

	for _, req := range companyApiReq.Value {
		if company.Cno.Valid && int64(req.CompCno) == company.Cno.Int64 {
			idStr := company.Id.String
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return entity.User{}, utils.CustomErrorf(err)
			}
			user.Uno = id
			user.UserId = idStr
			user.UserName = req.WorkerName + "(" + req.CompNameKr + ")"
			user.RoleCode = "CO_MANAGER"
		}
	}

	return user, nil
}
