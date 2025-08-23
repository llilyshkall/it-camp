package usecase

import (
	rep "remarks/internal/repository"
)

type UsecaseInterface interface {
}

type Usecase struct {
	store rep.StoreInterface
}

func NewUsecase(s rep.StoreInterface) UsecaseInterface {
	return &Usecase{
		store: s,
	}
}

// func (u *Usecase) GetRemarks(ctx context.Context, projectID int) ([]*model.Remark, error) {
// 	return u.store.GetRemarksByProjectID(projectID)
// }

// func (u *Usecase) CreateRemark(ctx context.Context, remark *model.Remark) error {
// 	return u.store.AddRemark(remark)
// }
