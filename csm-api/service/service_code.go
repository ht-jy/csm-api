package service

import (
	"context"
	"csm-api/entity"
	"csm-api/store"
	"fmt"
)

type ServiceCode struct {
	SafeDB store.Queryer
	Store  store.CodeStore
}

func (s *ServiceCode) GetCodeList(ctx context.Context, pCode string) (*entity.Codes, error) {
	list, err := s.Store.GetCodeList(ctx, s.SafeDB, pCode)
	if err != nil {
		//TODO: 에러 아카이브
		return nil, fmt.Errorf("service_code/GetCodeList err: %w", err)
	}

	return list, nil
}

// 코드트리 조회
func (s *ServiceCode) GetCodeTree(ctx context.Context) (*entity.CodeTrees, error) {

	// 코드리스트 조회
	codes, err := s.Store.GetCodeTree(ctx, s.SafeDB)
	if err != nil {
		//TODO: 에러 아카이브
		return nil, fmt.Errorf("service_code/GetCodeSetList err: %w", err)
	}

	// 트리구조로 반환
	trees, err := entity.ConvertCodesToCodeTree(*codes, "")
	if err != nil {
		//TODO: 에러 아카이브
		return nil, fmt.Errorf("service_code/ConvertCodesToCodeTree err: %w", err)
	}

	return &trees, nil

}
