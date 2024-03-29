// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package imports

import (
	"fmt"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestDirectoryPackageInfoReachedStatus(t *testing.T) {
	tests := []struct {
		info       directoryPackageInfo
		target     directoryPackageStatus
		wantStatus bool
		wantError  bool
	}{
		{
			info: directoryPackageInfo{
				status: directoryScanned,
				err:    nil,
			},
			target:     directoryScanned,
			wantStatus: true,
		},
		{
			info: directoryPackageInfo{
				status: directoryScanned,
				err:    fmt.Errorf("error getting to directory scanned"),
			},
			target:     directoryScanned,
			wantStatus: true,
			wantError:  true,
		},
		{
			info:       directoryPackageInfo{},
			target:     directoryScanned,
			wantStatus: false,
		},
	}

	for _, tt := range tests {
		gotStatus, gotErr := tt.info.reachedStatus(tt.target)
		if gotErr != nil {
			if !tt.wantError {
				t.Errorf("unexpected error: %s", gotErr)
			}
			continue
		}

		if tt.wantStatus != gotStatus {
			t.Errorf("reached status expected: %v, got: %v", tt.wantStatus, gotStatus)
		}
	}
}

func TestModCacheInfo(t *testing.T) {
	m := NewDirInfoCache()

	dirInfo := []struct {
		dir  string
		info directoryPackageInfo
	}{
		{
			dir: "mypackage",
			info: directoryPackageInfo{
				status:                 directoryScanned,
				dir:                    "mypackage",
				nonCanonicalImportPath: "example.com/mypackage",
			},
		},
		{
			dir: "bad package",
			info: directoryPackageInfo{
				status: directoryScanned,
				err:    fmt.Errorf("bad package"),
			},
		},
		{
			dir: "mypackage/other",
			info: directoryPackageInfo{
				dir:                    "mypackage/other",
				nonCanonicalImportPath: "example.com/mypackage/other",
			},
		},
	}

	for _, d := range dirInfo {
		m.Store(d.dir, d.info)
	}

	for _, d := range dirInfo {
		val, ok := m.Load(d.dir)
		if !ok {
			t.Errorf("directory not loaded: %s", d.dir)
		}

		if !reflect.DeepEqual(d.info, val) {
			t.Errorf("expected: %v, got: %v", d.info, val)
		}
	}

	var wantKeys []string
	for _, d := range dirInfo {
		wantKeys = append(wantKeys, d.dir)
	}
	sort.Strings(wantKeys)

	gotKeys := m.Keys()
	sort.Strings(gotKeys)

	if len(gotKeys) != len(wantKeys) {
		t.Errorf("different length of keys. expected: %d, got: %d", len(wantKeys), len(gotKeys))
	}

	for i, want := range wantKeys {
		if want != gotKeys[i] {
			t.Errorf("%d: expected %s, got %s", i, want, gotKeys[i])
		}
	}
}

func BenchmarkScanModuleCache(b *testing.B) {
	output, err := exec.Command("go", "env", "GOMODCACHE").Output()
	if err != nil {
		b.Fatal(err)
	}
	gomodcache := strings.TrimSpace(string(output))
	cache := NewDirInfoCache()
	start := time.Now()
	ScanModuleCache(gomodcache, cache, nil)
	b.Logf("initial scan took %v", time.Since(start))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ScanModuleCache(gomodcache, cache, nil)
	}
}
