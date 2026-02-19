package sde

import "testing"

func TestIsRigGroupName(t *testing.T) {
	tests := []struct {
		name       string
		categoryID int32
		groupName  string
		want       bool
	}{
		{name: "rig armor", categoryID: 7, groupName: "Rig Armor", want: true},
		{name: "rig launcher", categoryID: 7, groupName: "Rig Launcher", want: true},
		{name: "rig navigation lowercase", categoryID: 7, groupName: "rig navigation", want: true},
		{name: "non-rig module", categoryID: 7, groupName: "Energy Weapon", want: false},
		{name: "rig blueprint", categoryID: 9, groupName: "Rig Blueprint", want: false},
		{name: "ship group", categoryID: 6, groupName: "Tactical Destroyer", want: false},
		{name: "empty", categoryID: 7, groupName: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRigGroupName(tt.categoryID, tt.groupName); got != tt.want {
				t.Fatalf("isRigGroupName(%d, %q) = %v, want %v", tt.categoryID, tt.groupName, got, tt.want)
			}
		})
	}
}
