package mfd

import (
	"encoding/xml"
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
