package alist_v3

import (
	"encoding/json"
	"testing"
)

func TestMeRespAcceptsScalarRole(t *testing.T) {
	var resp MeResp
	if err := json.Unmarshal([]byte(`{"role":1}`), &resp); err != nil {
		t.Fatalf("unmarshal scalar role: %v", err)
	}
	if len(resp.Role) != 1 || resp.Role[0] != 1 {
		t.Fatalf("role = %#v, want []int{1}", resp.Role)
	}
}

func TestMeRespAcceptsRoleList(t *testing.T) {
	var resp MeResp
	if err := json.Unmarshal([]byte(`{"role":[1,2]}`), &resp); err != nil {
		t.Fatalf("unmarshal role list: %v", err)
	}
	if len(resp.Role) != 2 || resp.Role[0] != 1 || resp.Role[1] != 2 {
		t.Fatalf("role = %#v, want []int{1, 2}", resp.Role)
	}
}
