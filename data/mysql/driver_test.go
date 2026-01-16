package mysql

import (
	"testing"
)

func TestDriverName(t *testing.T) {
	d := &driver{}
	if got := d.Name(); got != "mysql" {
		t.Errorf("Name() = %q, want %q", got, "mysql")
	}
}
