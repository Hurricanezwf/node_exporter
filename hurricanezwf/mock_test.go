package hurricanezwf_test

import (
	"context"

	"github.com/prometheus/node_exporter/hurricanezwf"
)

type MockLogger struct{}

func newMockLogger() *MockLogger {
	return &MockLogger{}
}

func (l *MockLogger) Log(keyvals ...interface{}) error {
	return nil
}

type MockKubeReader struct{}

func newMockKubeReader() *MockKubeReader {
	return &MockKubeReader{}
}

func (r *MockKubeReader) PodReader() hurricanezwf.KubePodReader {
	return &MockKubePodReader{}
}

func (r *MockKubeReader) PersistentVolumeClaimReader() hurricanezwf.KubePersistentVolumeClaimReader {
	return &MockKubePVCReader{}
}

type MockKubePodReader struct{}

func (r *MockKubePodReader) GetByUID(ctx context.Context, uid string) (string, error) {
	return `{"kind": "Pod", "metadata":{"name":"podname", "namespace":"default"}}`, nil
}

type MockKubePVCReader struct{}

func (r *MockKubePVCReader) GetByUID(ctx context.Context, uid string) (string, error) {
	return `{"kind": "PersistentVolumeClaim", "metadata":{"name":"pvcname", "namespace":"default"}}`, nil
}
