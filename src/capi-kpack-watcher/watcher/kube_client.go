package watcher

// TOOD: maybe consider a different package/naming
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . KubeClient
type KubeClient interface {
	GetContainerLogs(podName, containerName string) ([]byte, error)
}
