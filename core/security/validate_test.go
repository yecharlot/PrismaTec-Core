package security

import (
	"os"
	"testing"
)

func TestValidateName(t *testing.T) {
	if err := ValidateName("research-agent-01"); err != nil {
		t.Fatal(err)
	}
	if ValidateName("") == nil || ValidateName("bad name!") == nil {
		t.Fatal("expected reject")
	}
}

func TestAIPToken(t *testing.T) {
	os.Unsetenv("PRISMATEC_AIP_TOKEN")
	if err := CheckAIPToken("", ""); err != nil {
		t.Fatal("open mode should allow")
	}
	t.Setenv("PRISMATEC_AIP_TOKEN", "secret")
	if err := CheckAIPToken("", ""); err == nil {
		t.Fatal("should require token")
	}
	if err := CheckAIPToken("Bearer secret", ""); err != nil {
		t.Fatal(err)
	}
	if err := CheckAIPToken("", "secret"); err != nil {
		t.Fatal(err)
	}
	if err := CheckAIPToken("Bearer wrong", ""); err == nil {
		t.Fatal("wrong token")
	}
}
