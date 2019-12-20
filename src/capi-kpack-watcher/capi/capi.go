package capi

import "capi_kpack_watcher/model"

// CAPI represents the actions that a client can make to CAPI.
type CAPI interface {
	PATCHBuild(guid string, status model.BuildStatus) error
}
