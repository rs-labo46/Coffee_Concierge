package controller

import (
	"encoding/json"
	"testing"

	"coffee-spa/usecase"
)

func TestWsCtlOnPatchPrefInvalidDiff(t *testing.T) {
	ctl := NewWsCtl(&searchFlowUCMock{patchPrefFn: func(usecase.PatchPrefIn) (usecase.PatchPrefOut, error) {
		t.Fatal("should not be called")
		return usecase.PatchPrefOut{}, nil
	}}, &sessionUCMock{})
	ev := wsClientEvent{Type: "pref.patch", SessionID: 1, Diff: json.RawMessage(`{"flavor":`)}
	if err := ctl.onPatchPref(nil, nil, "", ev); err != ErrInvalidRequest {
		t.Fatalf("err=%v", err)
	}
}
