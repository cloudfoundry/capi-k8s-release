package watcher

import (
	"capi_kpack_watcher/capi_model"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . BuildUpdater
type BuildUpdater interface {
	UpdateBuild(guid string, capi_model capi_model.Build) error
}

