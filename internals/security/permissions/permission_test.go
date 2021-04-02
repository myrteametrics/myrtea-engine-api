package permissions

import "testing"

func TestHasPermission(t *testing.T) {

	testsTrue := [][]Permission{
		{
			{ResourceType: "*", ResourceID: "*", Action: "*"},
		},
		{
			{ResourceType: "situation", ResourceID: "*", Action: "*"},
		},
		{
			{ResourceType: "situation", ResourceID: "1", Action: "*"},
		},
		{
			{ResourceType: "situation", ResourceID: "1", Action: "get"},
		},
		{
			{ResourceType: "situation", ResourceID: "*", Action: "get"},
		},
		{
			{ResourceType: "*", ResourceID: "*", Action: "get"},
		},
		{
			{ResourceType: "situation", ResourceID: "1", Action: "create"},
			{ResourceType: "situation", ResourceID: "1", Action: "update"},
			{ResourceType: "situation", ResourceID: "1", Action: "get"},
		},
	}

	testsFalse := [][]Permission{
		{
			{ResourceType: "fact", ResourceID: "*", Action: "*"},
		},
		{
			{ResourceType: "*", ResourceID: "*", Action: "create"},
		},
		{
			{ResourceType: "situation", ResourceID: "2", Action: "*"},
		},
		{
			{ResourceType: "situation", ResourceID: "1", Action: "list"},
		},
	}

	required := Permission{ResourceType: "situation", ResourceID: "1", Action: "get"}

	for _, permissions := range testsTrue {
		hasPermission := HasPermission(permissions, required)
		if hasPermission != true {
			t.Error("invalid HasPermission")
		}
	}
	for _, permissions := range testsFalse {
		hasPermission := HasPermission(permissions, required)
		if hasPermission != false {
			t.Error("invalid HasPermission")
		}
	}
}

func TestListPermissions1(t *testing.T) {
	permissions := []Permission{
		{ResourceType: "situation", ResourceID: "1", Action: "*"},
		{ResourceType: "situation", ResourceID: "2", Action: "*"},
		{ResourceType: "situation", ResourceID: "3", Action: "*"},
	}

	filteredPermissions := ListPermissions(permissions, "situation", "2", "get")
	if len(filteredPermissions) != 1 {
		t.Error("invalid filtered permissions")
	}
	if (filteredPermissions[0] != Permission{ResourceType: "situation", ResourceID: "2", Action: "*"}) {
		t.Error("invalid filtered permissions")
	}
}

func TestListPermissions2(t *testing.T) {
	permissions := []Permission{
		{ResourceType: "situation", ResourceID: "1", Action: "*"},
		{ResourceType: "situation", ResourceID: "2", Action: "*"},
		{ResourceType: "situation", ResourceID: "3", Action: "*"},
		{ResourceType: "situation", ResourceID: "4", Action: "get"},
		{ResourceType: "fact", ResourceID: "5", Action: "*"},
	}

	filteredPermissions := ListPermissions(permissions, "situation", "*", "create")
	if len(filteredPermissions) != 3 {
		t.Error("invalid filtered permissions")
	}
}
