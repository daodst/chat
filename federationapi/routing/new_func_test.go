package routing

import (
	"encoding/json"
	"testing"
)

func TestSendOwnerInvite(t *testing.T) {
	marshal, err := json.Marshal(struct {
	}{})
	if err != nil {
		t.Error(err)
	}
	t.Log(string(marshal))
}
