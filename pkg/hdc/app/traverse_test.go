package app

import (
	"os"
	"testing"
	"testing/fstest"

	"github.com/Dieterbe/sandbox/homedirclean/pkg/hdc"
	"github.com/google/go-cmp/cmp"
)

// TODO run same tests on "regular directory"? these are not specific to zip
// TODO do we have a test anywhere that also checks for adding the "intermediate" dirprints?
// similar test that has a full path AND a zip file?
func TestTraverse(t *testing.T) {

	var tests = []struct {
		name string
		data fstest.MapFS
		want hdc.DirPrint
		err  error
	}{
		{"main", hdc.DataMain, hdc.DataMainPrint, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			dirPrint, _, err := WalkZip(tt.data, "test/in-memory/"+tt.name+".zip", hdc.Sha256FingerPrint, os.Stderr)
			if err != tt.err {
				t.Errorf("WalkZip() error = %v, wantErr %v", err, tt.err)
			}
			if err != nil {
				return
			}

			if diff := cmp.Diff(tt.want, dirPrint); diff != "" {
				t.Errorf("WalkZip() mismatch (-want +got):\n%s", diff)
			}
		})

	}
}
