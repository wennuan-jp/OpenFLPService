package service

import (
	"openflp.com/model"
)

type FLPService interface {
	Resolve(filePath string) (*model.ResolutionResult, error)
}
