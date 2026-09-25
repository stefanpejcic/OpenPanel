package mysqlmanager

import "testing"

func TestToIntUnsigned(t *testing.T) {
	for _, v := range []any{uint64(6291456), int64(6291456), uint32(6291456), int32(6291456), float64(6291456), []byte("6291456"), "6291456"} {
		if got := ToInt(v); got != 6291456 {
			t.Errorf("ToInt(%T) = %d", v, got)
		}
	}
	if ToInt(nil) != 0 {
		t.Error("nil should be 0")
	}
}
