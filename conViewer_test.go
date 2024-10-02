package conviewer 

import (
	"testing"
	"encoding/json"
)

func TestMarshaling(t *testing.T) {
	msg := &ConMsg{
		numCons: 5,
		activeCons: map[string]struct{}{
			"1.2.3.4": {},
			"1.2.3.5": {},
		},
		idleCons: map[string]struct{}{
			"1.2.3.4": {},
			"1.2.3.5": {},
		},
	}
	s, err := json.Marshal(msg) 
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", s)
}
