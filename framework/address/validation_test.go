package address_test

import (
	"strings"
	"testing"

	"github.com/sirrchat/SirrMesh/framework/address"
)

func TestValidMailboxName(t *testing.T) {
	if !address.ValidMailboxName("caddy.bug") {
		t.Error("caddy.bug should be valid mailbox name")
	}
}

func TestValidDomain(t *testing.T) {
	for _, c := range []struct {
		Domain string
		Valid  bool
	}{
		{Domain: "sirrmesh.email", Valid: true},
		{Domain: "", Valid: false},
		{Domain: "sirrmesh.email.", Valid: true},
		{Domain: "..", Valid: false},
		{Domain: strings.Repeat("a", 256), Valid: false},
		{Domain: "äõäoaõoäaõaäõaoäaoaäõoaäooaoaoiuaiauäõiuüõaõäiauõaaa.tld", Valid: true},
		{Domain: "xn--oaoaaaoaoaoaooaoaoiuaiauiuaiauaaa-f1cadccdcmd01eddchqcbe07a.tld", Valid: true},
	} {
		if actual := address.ValidDomain(c.Domain); actual != c.Valid {
			t.Errorf("expected domain %v to be valid=%v, but got %v", c.Domain, c.Valid, actual)
		}
	}
}
