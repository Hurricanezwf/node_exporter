package hurricanezwf_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/prometheus/node_exporter/hurricanezwf"
)

func TestDiskMounts(t *testing.T) {
	mounts, err := hurricanezwf.DiskMounts()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range mounts {
		t.Logf("%s %s\n", m.DeviceName, strings.Join(m.MountPaths, " | "))
	}

	podInfo, err := hurricanezwf.TryDecodePodInfoForDevice(newMockLogger(), mounts, "vda1", newMockKubeReader())
	if err != nil {
		t.Fatal(err)
	}

	b, err := json.MarshalIndent(podInfo, "", "   ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Result:")
	fmt.Println(string(b))
}
