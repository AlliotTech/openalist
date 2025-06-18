package op_test

import (
	"testing"

	_ "github.com/AlliotTech/openalist/drivers"
	"github.com/AlliotTech/openalist/internal/op"
)

func TestDriverItemsMap(t *testing.T) {
	itemsMap := op.GetDriverInfoMap()
	if len(itemsMap) != 0 {
		t.Logf("driverInfoMap: %v", itemsMap)
	} else {
		t.Errorf("expected driverInfoMap not empty, but got empty")
	}
}
