package templates

import (
	"fmt"
	"testing"
	"time"
)

func TestDecodeManifestStrOk(t *testing.T) {
	data := `
	[project]
	name = 'Qore App'
	version = '0.1.0'
	package = 'eg'
	module-dir = 'internal'
	description = 'Qore app project test'
	repository = 'qoinly.id/qore-test'
	created-at = '2025-08-04T07:26:50Z'

	[dependencies]
	foo = { version = 'latest', path = 'qoinly.id/foo' }
	bar = { version = '0.1.1', path = 'qoinly.id/bar' }
	`

	manifest, _ := decodeManifest([]byte(data))
	// log.Println(err)
	fmt.Println(*manifest)
}

func TestEncodeManifestOk(t *testing.T) {
	data, _ := encodeManifest(&Manifest{
		Project: Project{
			Name:        "Qore App",
			Version:     "0.1.0",
			Package:     "eg",
			ModuleDir:   "internal",
			Description: "Qore app project test",
			Repository:  "qoinly.id/qore-test",
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
		Dependencies: Dependencies{
			"foo": {Version: "latest", Path: "qoinly.id/foo"},
			"bar": {Version: "0.1.1", Path: "qoinly.id/bar"},
		}},
	)
	// fmt.Println(err)
	fmt.Println(string(data))
}
