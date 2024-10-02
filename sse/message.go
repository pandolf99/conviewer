package sse

import (
	"bytes"
	"errors"
)

var (
	ErrNonSerializable = errors.New("Unserialize: Data can't be SSE serialized")
)

//TODO make SSE messages conform to the HTML5 standard

//For a message to be SseSerializable
//it should have event and data fields
//this is an interface in case the user 
//wants to pass around its own struct
//through the sse server
type SseSerializable interface {
	Event() []byte
	Data() []byte
	SetEvent([]byte)
	SetData([]byte)
}

func Serialize(s SseSerializable) []byte {
	var buff bytes.Buffer 
	buff.WriteString("event:")
	buff.Write(bytes.TrimSpace(s.Event()))
	buff.WriteRune('\n')
	buff.WriteString("data:")
	buff.Write(bytes.TrimSpace(s.Data()))
	buff.Write([]byte{'\n'})
	return buff.Bytes()
}


//Unserialize data into s
//throws an error if data does not follow
//sse message format
func Unserialize(msg []byte, s SseSerializable) error {
	msgSlice := bytes.SplitN(msg, []byte{'\n'}, 2)
	if len(msgSlice) != 2 {
		return ErrNonSerializable
	}
	event := msgSlice[0]
	data := msgSlice[1]
	if ind := bytes.Index(event, []byte("event:")); ind != 0 {
		return ErrNonSerializable
	}
	if ind := bytes.Index(data, []byte("data:")); ind != 0 {
		return ErrNonSerializable
	}
	event = bytes.SplitN(event, []byte{':'}, 2)[1] 
	data = bytes.SplitN(data, []byte{':'}, 2)[1] 
	s.SetData(bytes.TrimSpace(data))
	s.SetEvent(bytes.TrimSpace(event))
	return nil
}

//default implenentation of a SseSeriablizable
//provided by the package
type SseMsg struct {
	event []byte
	data []byte
}

func (m *SseMsg) Event() []byte {
	return m.event
}

func (m *SseMsg) SetEvent(event []byte) {
	m.event = event
}

func (m *SseMsg) Data() []byte {
	return m.data
}

func (m *SseMsg) SetData(data []byte) {
	 m.data = data
}


