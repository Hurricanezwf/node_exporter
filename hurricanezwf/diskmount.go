package hurricanezwf

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-kit/log"
	"github.com/tidwall/gjson"
)

type DiskMount struct {
	DeviceName string
	MountPaths []string
}

// DiskMounts 读取并解析 /proc/mounts 文件内容;
func DiskMounts() (map[string]DiskMount, error) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/mounts, %w", err)
	}
	defer f.Close()

	diskMounts := make(map[string]DiskMount)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, " ")
		if len(fields) < 2 {
			continue
		}
		device := strings.TrimPrefix(fields[0], "/dev/")
		mountPath := fields[1]

		mounts, ok := diskMounts[device]
		if !ok {
			mounts = DiskMount{
				DeviceName: device,
				MountPaths: []string{},
			}
		}
		mounts.MountPaths = append(mounts.MountPaths, mountPath)
		diskMounts[device] = mounts
	}

	return diskMounts, nil
}

// TryDecodePodInfoForDevice 尝试从设备挂载的路径中解析出 pod 信息;
// 如果 deviceName 在 diskMounts 中不存在, 将返回 ErrNotFound;
func TryDecodePodInfoForDevice(logger log.Logger, diskMounts map[string]DiskMount, deviceName string, kubeReader KubeReader) ([]PodInfo, error) {
	if diskMounts == nil {
		return nil, ErrNotFound
	}

	mounts, ok := diskMounts[deviceName]
	if !ok {
		mounts, ok = diskMounts[fmt.Sprintf("/dev/%s", deviceName)]
		if !ok {
			return nil, ErrNotFound
		}
	}

	podContainerToVols := make(map[string]map[string]map[string]any)
	for _, path := range mounts.MountPaths {
		// For example1 PVC:      `/dev/vdg  /var/lib/kubelet/pods/82e31898-1e29-4c93-b63b-db2f8393c6e6/volume-subpaths/pvc-3431657b-94fa-425e-a719-ee4acb1720c9/gateserver/2 ext4 rw,relatime 0 0`
		// For example2 EmptyDir: `/dev/vda1 /var/lib/kubelet/pods/374d4e5c-a31d-45e1-9e81-9669e2a2c0f0/volume-subpaths/observ-router-trademarketserver-tx-sea-rometa-prod-rometa-10001/router-trademarketserver-tx-sea-rometa-prod-rometa-10001/2`
		if !strings.Contains(path, "volume-subpaths") {
			continue
		}
		fields := strings.Split(path, "/")
		if len(fields) < 8 || fields[1] != "var" || fields[2] != "lib" || fields[3] != "kubelet" || fields[4] != "pods" {
			continue
		}
		podUID := fields[5]
		vol := fields[7]
		container := fields[8]
		fmt.Printf("%s / %s/ %s\n", podUID, vol, container)

		if v := podContainerToVols[podUID]; v == nil {
			podContainerToVols[podUID] = make(map[string]map[string]any)
		}
		if v := podContainerToVols[podUID][container]; v == nil {
			podContainerToVols[podUID][container] = make(map[string]any)
		}
		podContainerToVols[podUID][container][vol] = nil
	}

	// Note: 这里统一错误降级;
	infoList := []PodInfo{}
	for podUID, containerToVols := range podContainerToVols {
		if podUID == "" {
			continue
		}
		info := PodInfo{}
		// 读取 pod 名字;
		podJSONManifests, err := kubeReader.PodReader().GetByUID(context.Background(), podUID)
		if err != nil {
			logger.Log("error:", fmt.Sprintf("could not get pod JSON manifests with uid %s, %v", podUID, err))
			continue
		}
		j := gjson.Parse(podJSONManifests)
		info.PodName = j.Get("metadata").Get("name").String()
		if info.PodName == "" {
			logger.Log("error:", fmt.Sprintf("empty .meta.name from pod manifests with uid %s", podUID))
			continue
		}
		info.PodNamespace = j.Get("metadata").Get("namespace").String()
		info.PodContainers = make(map[string]ContainerInfo)

		for container, volset := range containerToVols {
			info.PodContainers[container] = ContainerInfo{
				PVCNames:      make(map[string]any),
				EmptyDirNames: make(map[string]any),
			}
			for vol, _ := range volset {
				if strings.HasPrefix(vol, "pvc-") {
					// pvc;
					pvcUID := strings.TrimPrefix(vol, "pvc-")
					pvcJSONManifests, err := kubeReader.PersistentVolumeClaimReader().GetByUID(context.Background(), pvcUID)
					if err != nil {
						logger.Log("error:", fmt.Sprintf("could not find pvc with uid %s, %v", pvcUID, err))
						continue
					}
					pvcName := gjson.Parse(pvcJSONManifests).Get("metadata").Get("name").String()
					info.PodContainers[container].PVCNames[pvcName] = nil
					continue
				}
				// emptydir;
				info.PodContainers[container].EmptyDirNames[vol] = nil
				continue
			}
		}

		infoList = append(infoList, info)
	}

	return infoList, nil
}
