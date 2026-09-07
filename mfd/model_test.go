package mfd

import (
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func TestTableMapping_Packages(t *testing.T) {
	tests := []struct {
		name    string
		Entries []Entry
		want    map[string]string
	}{
		{
			name: "base test",
			Entries: []Entry{
				{
					XMLName: xml.Name{Local: "common"},
					Value:   "users",
				},
				{
					XMLName: xml.Name{Local: "vfs"},
					Value:   "vfsFiles,vfsFolders",
				},
				{
					XMLName: xml.Name{Local: "news"},
					Value:   "news,categories,tags",
				},
			},
			want: map[string]string{
				"users":      "common",
				"vfsFiles":   "vfs",
				"vfsFolders": "vfs",
				"news":       "news",
				"categories": "news",
				"tags":       "news",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := &TableMapping{
				Entries: tt.Entries,
			}
			if got := tm.Packages(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Packages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectIsConsistentRejectsUnsupportedGoPGVersions(t *testing.T) {
	for _, version := range []int{8, 9, 11} {
		t.Run(strconv.Itoa(version), func(t *testing.T) {
			project := &Project{GoPGVer: version}
			if err := project.IsConsistent(); err == nil {
				t.Fatalf("IsConsistent() error = nil for go-pg version %d", version)
			}
		})
	}
}

func TestSaveMFDRejectsUnsupportedGoPGVersions(t *testing.T) {
	for _, version := range []int{8, 9, 11} {
		t.Run(strconv.Itoa(version), func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "project.mfd")

			err := SaveMFD(filename, &Project{GoPGVer: version})
			if err == nil {
				t.Fatalf("SaveMFD() error = nil for go-pg version %d", version)
			}

			if _, statErr := os.Stat(filename); !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("SaveMFD() created output file for go-pg version %d: %v", version, statErr)
			}
		})
	}
}

func TestSaveMFDPersistsSupportedGoPGVersionWithoutLoadedNamespaces(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "project.mfd")
	project := &Project{
		GoPGVer:        GoPG10,
		NamespaceNames: []string{"portal"},
	}

	if err := SaveMFD(filename, project); err != nil {
		t.Fatalf("SaveMFD() error = %v", err)
	}

	saved := &Project{}
	if err := UnmarshalFile(filename, saved); err != nil {
		t.Fatalf("UnmarshalFile() error = %v", err)
	}

	if !reflect.DeepEqual(saved.NamespaceNames, project.NamespaceNames) {
		t.Fatalf("saved NamespaceNames = %v, want %v", saved.NamespaceNames, project.NamespaceNames)
	}
}
