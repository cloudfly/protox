package doc

import (
	"context"
	"testing"
)

type LedgerGeneratedApiServiceHandler interface {
	UpdateLedger(context.Context, *ServiceInfo) (*MethodInfo, error)
}

func TestParser(t *testing.T) {
	info, err := ParseService((*LedgerGeneratedApiServiceHandler)(nil))
	if err != nil {
		panic(err)
	}
	t.Log(info.String())
}
