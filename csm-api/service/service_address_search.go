package service

import (
	"csm-api/api"
	"csm-api/config"
	"csm-api/entity"
	"csm-api/utils"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

/**
 * @author 작성자: 정지영
 * @created 작성일: 2025-03-13
 * @modified 최종 수정일:
 * @modifiedBy 최종 수정자:
 * @modified description
 * -
 */

// struct: 브이월드 주소 api 조회
type ServiceAddressSearch struct {
	ApiKey *config.ApiConfig
}

// func: 도로명주소로 위도, 경도 조회
// @param
// - roadAddress string: 도로명주소
func (s *ServiceAddressSearch) GetAPILatitudeLongtitude(roadAddress string) (*entity.Point, error) {

	if roadAddress == "" {
		return nil, utils.CustomErrorf(fmt.Errorf("roadAddress parameter is missing"))
	}
	// apiUrl주소
	apiUrl := fmt.Sprintf(`https://api.vworld.kr/req/search?key=%s&version=%s&service=%s&request=%s&type=%s&crs=%s&category=%s&query=%s&domain=%s&format=%s&errorFormat=%s`,
		url.QueryEscape(s.ApiKey.VworldApiKey), // key
		"2.0",                                  // version
		"search",                               // service
		"search",                               //request
		"address",                              // type
		"EPSG:4326",                            // crs
		"ROAD",                                 // category
		url.QueryEscape(roadAddress),           // query
		"csm.htenc.co.kr",                      // domain
		"json",                                 // format
		"json",                                 // errFormat
	)

	// api 호출
	body, err := api.CallGetAPI(apiUrl)
	if err != nil {
		return nil, utils.CustomMessageErrorf("call GetWhetherSrtNcst API", err)
	}

	// api response struct
	type AddressItem struct {
		Id      string   `json:"id"`
		Address struct{} `json:"address"`
		Point   struct {
			Longitude string `json:"x"`
			Latitude  string `json:"y"`
		} `json:"point"`
	}
	type AddressItems []AddressItem

	type AddressSearching struct {
		Response struct {
			Status string `json:"status"`
			Result struct {
				Crs   string       `json:"crs"`
				Type  string       `json:"type"`
				Items AddressItems `json:"items"`
			} `json:"result"`
			Error struct {
				Level string `json:"level"`
				Code  int    `json:"code"`
				Text  string `json:"text"`
			}
		} `json:"response"`
	}

	// 반환 경/위도 객체
	point := &entity.Point{}

	// api response parse
	var res AddressSearching
	if err = json.Unmarshal([]byte(body), &res); err != nil {
		return nil, utils.CustomMessageErrorf("AddressSearch api JSON parse", err)
	}

	if res.Response.Status != "OK" {
		return nil, utils.CustomErrorf(fmt.Errorf("AddressSearch api response err: %s", res.Response.Error.Text))

	} else {
		items := res.Response.Result.Items

		// 반환 값 추가
		point.Latitude, _ = strconv.ParseFloat(items[0].Point.Latitude, 64)
		point.Longitude, _ = strconv.ParseFloat(items[0].Point.Longitude, 64)
	}

	return point, nil
}

// 지도 x, y좌표 조회
// @params
//   - roadAddress : 도로명 주소
func (s *ServiceAddressSearch) GetAPISiteMapPoint(roadAddress string) (*entity.MapPoint, error) {
	if roadAddress == "" {
		return nil, utils.CustomErrorf(fmt.Errorf("roadAddress parameter is missing"))
	}
	if roadAddress == "undefined" {
		return &entity.MapPoint{
			X: "14209677.672145272",
			Y: "4141263.42632809",
		}, nil
	}

	// apiUrl주소
	apiUrl := fmt.Sprintf(`https://api.vworld.kr/req/address?key=%s&version=%s&service=%s&request=%s&type=%s&crs=%s&address=%s&domain=%s&format=%s&errorFormat=%s&simple=%s`,
		url.QueryEscape(s.ApiKey.VworldApiKey), // key
		"2.0",                                  // version
		"address",                              // service
		"getcoord",                             // request
		"road",                                 // type
		"EPSG:900913",                          // crs
		url.QueryEscape(roadAddress),           // address
		"csm.htenc.co.kr",                      // domain
		"json",                                 // format
		"json",                                 // errformat
		"true",                                 // simple
	)
	type SiteMapPoint struct {
		Response struct {
			Status string `json:"status"`
			Result struct {
				Crs   string          `json:"crs"`
				Point entity.MapPoint `json:"point"`
			} `json:"result"`
			Error struct {
				Level string `json:"level"`
				Code  string `json:"code"`
				Text  string `json:"text"`
			}
		} `json:"response"`
	}
	// api 호출
	body, err := api.CallGetAPI(apiUrl)
	if err != nil {
		return nil, utils.CustomMessageErrorf("call GetWhetherSrtNcst API", err)
	}

	// 응답 json으로 변환
	var res SiteMapPoint
	if err = json.Unmarshal([]byte(body), &res); err != nil {
		return nil, utils.CustomMessageErrorf("SiteMapPoint api JSON parse", err)
	}

	if res.Response.Status != "OK" {
		return nil, utils.CustomErrorf(fmt.Errorf("SiteMapPoint api response", res.Response.Error.Text))
	}

	return &res.Response.Result.Point, nil
}
