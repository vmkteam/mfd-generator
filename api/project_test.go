package api

import (
	"strconv"
	"testing"

	"github.com/vmkteam/mfd-generator/mfd"
)

func TestProjectServiceUpdateRejectsUnsupportedGoPGVersions(t *testing.T) {
	for _, version := range []int{8, 9, 11} {
		t.Run(strconv.Itoa(version), func(t *testing.T) {
			current := mfd.NewProject("current", mfd.GoPG10)
			service := NewProjectService(&Store{CurrentProject: current})

			err := service.Update(mfd.Project{GoPGVer: version})
			if err == nil {
				t.Fatalf("Update() error = nil for go-pg version %d", version)
			}

			if service.CurrentProject != current {
				t.Fatalf("Update() changed CurrentProject for go-pg version %d", version)
			}
		})
	}
}

func TestProjectServiceUpdateAcceptsSupportedGoPGVersion(t *testing.T) {
	service := NewProjectService(&Store{})
	project := mfd.Project{Name: "updated", GoPGVer: mfd.GoPG10}

	if err := service.Update(project); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if service.CurrentProject == nil || service.CurrentProject.Name != project.Name {
		t.Fatalf("CurrentProject = %#v, want project %q", service.CurrentProject, project.Name)
	}
}
