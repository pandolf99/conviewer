package sse

import (
	"bytes"
	"testing"
)

func TestSerialize(t *testing.T) {
	m := new(SseMsg)
	m.SetEvent([]byte("myEvent"))
	m.SetData([]byte("Hello World"))
	expected := []byte("event:myEvent\ndata:Hello World\n")
	actual := Serialize(m)
	if bytes.Compare(expected, actual) != 0 {
		t.Fail()
	}
	t.Logf("Expected:\n%s\n", expected)
	t.Logf("Actual:\n%s\n", actual)
}

func TestUnserialize(t *testing.T) {
	msg := []byte("event:myEvent\ndata:Hello World\n")
	m := new(SseMsg)
	err := Unserialize(msg, m)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if bytes.Compare(m.Event(), []byte("myEvent")) != 0 {
		t.Fail()
	}
	t.Log("Expected Event: myEvent\n")
	t.Logf("Actual Event: %s\n", m.Event())
	if bytes.Compare(m.Data(), []byte("Hello World")) != 0 {
		t.Fail()
	}
	t.Log("Expected Data: Hello World\n")
	t.Logf("Actual Data: %s\n", m.Data())
}
