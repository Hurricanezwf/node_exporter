package hurricanezwf

import "context"

type PodInfo struct {
	PodNamespace  string
	PodName       string
	PodContainers map[string]ContainerInfo
}

type ContainerInfo struct {
	PVCNames      map[string]any
	EmptyDirNames map[string]any
}

type KubeReader interface {
	PodReader() KubePodReader
	PersistentVolumeClaimReader() KubePersistentVolumeClaimReader
}

type KubePodReader interface {
	GetByUID(ctx context.Context, uid string) (string, error)
}

type KubePersistentVolumeClaimReader interface {
	GetByUID(ctx context.Context, uid string) (string, error)
}
