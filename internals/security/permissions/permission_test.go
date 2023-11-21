package permissions

import (
	"testing"
)

func TestHasPermission(t *testing.T) {

	testsTrue := [][]Permission{
		{
			New("*", "*", "*"),
		},
		{
			New("situation", "*", "*"),
		},
		{
			New("situation", "1", "*"),
		},
		{
			New("situation", "1", "get"),
		},
		{
			New("situation", "*", "get"),
		},
		{
			New("*", "*", "get"),
		},
		{
			New("situation", "1", "create"),
			New("situation", "1", "update"),
			New("situation", "1", "get"),
		},
	}

	testsFalse := [][]Permission{
		{
			New("fact", "*", "*"),
		},
		{
			New("*", "*", "create"),
		},
		{
			New("situation", "2", "*"),
		},
		{
			New("situation", "1", "list"),
		},
	}

	required := New("situation", "1", "get")

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

func TestListMatchingPermissions1(t *testing.T) {
	permissions := []Permission{
		New("situation", "1", "*"),
		New("situation", "2", "*"),
		New("situation", "3", "*"),
	}

	filteredPermissions := ListMatchingPermissions(permissions, New("situation", "2", "get"))
	if len(filteredPermissions) != 1 {
		t.Error("invalid filtered permissions")
	}
	if filteredPermissions[0] != New("situation", "2", "*") {
		t.Error("invalid filtered permissions")
	}
}

func TestListMatchingPermissions2(t *testing.T) {
	permissions := []Permission{
		New("situation", "1", "*"),
		New("situation", "2", "*"),
		New("situation", "3", "*"),
		New("situation", "4", "get"),
		New("fact", "5", "*"),
	}

	filteredPermissions := ListMatchingPermissions(permissions, New("situation", "*", "create"))
	if len(filteredPermissions) != 3 {
		t.Error("invalid filtered permissions")
	}
}

func TestListMatchingPermissions3(t *testing.T) {
	permissions := []Permission{
		New("situation", "1", "*"),
		New("situation", "2", "*"),
		New("situation", "*", "*"),
		New("fact", "5", "*"),
	}

	filteredPermissions := ListMatchingPermissions(permissions, New("situation", "*", "create"))
	if len(filteredPermissions) != 3 {
		t.Error("invalid filtered permissions")
	}
}

func TestGetResourceIDs(t *testing.T) {
	permissions := []Permission{
		New("situation", "1", "*"),
		New("situation", "2", "*"),
		New("situation", "3", "*"),
		New("situation", "4", "get"),
		New("fact", "5", "*"),
	}

	resourceIDs := GetResourceIDs(ListMatchingPermissions(permissions, New("situation", "*", "create")))
	if len(resourceIDs) != 3 {
		t.Error("invalid resourceIDs")
	}
}
