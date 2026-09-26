package registry

import "testing"

func TestPutGetList(t *testing.T) {
	r := New()
	rec := &OrganismRecord{
		ID:      "org-1",
		RootCID: "bafytest",
		Name:    "research-agent-01",
		Status:  "created",
		NodeID:  "node:abc",
	}
	if err := r.PutOrganism(rec); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got := r.GetOrganism("org-1")
	if got == nil || got.Name != "research-agent-01" {
		t.Fatalf("Get failed: %+v", got)
	}
	if r.Count() != 1 {
		t.Fatalf("Count = %d", r.Count())
	}
	list := r.ListOrganisms()
	if len(list) != 1 {
		t.Fatalf("List len = %d", len(list))
	}
	r.DeleteOrganism("org-1")
	if r.Count() != 0 {
		t.Fatal("expected empty after delete")
	}
}

func TestPutRequiresID(t *testing.T) {
	r := New()
	if err := r.PutOrganism(&OrganismRecord{}); err == nil {
		t.Fatal("expected error for empty ID")
	}
}
