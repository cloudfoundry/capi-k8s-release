package watcher

// Watcher abstracts the concept of watching resources in Kubernetes.
// Kubernetes controllers are made of components, one of which is an Informer.
// We wrap this Informer object and call it a Watcher.
type Watcher interface {
	// Run starts the Informer in Kubernetes. At this point, new resources can be detected.
	Run()

	// AddFunc is the handler for detecting when new resources have been created.
	AddFunc(obj interface{})

	// UpdateFunc is the handler for detecting when resources have been updated.
	UpdateFunc(oldobj, newobj interface{})
}
