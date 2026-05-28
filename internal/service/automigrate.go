package service

import "context"

type IAutoMigrate interface {
	Up(ctx context.Context) error
	Status(ctx context.Context) (version uint, dirty bool, err error)
}

var localAutoMigrate IAutoMigrate

func AutoMigrate() IAutoMigrate {
	if localAutoMigrate == nil {
		panic("implement not found for interface IAutoMigrate")
	}
	return localAutoMigrate
}

func RegisterAutoMigrate(i IAutoMigrate) {
	localAutoMigrate = i
}
