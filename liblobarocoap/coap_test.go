package liblobarocoap

import (
	"github.com/Lobaro/coap-go/coapmsg"
	"testing"
)

func TestInit(t *testing.T) {

}

func TestSendPingReceivePong(t *testing.T) {
	socket := NewSocket()

	pingMsg := coapmsg.Message{
		Type: coapmsg.Confirmable,
		Code: coapmsg.COAPCode(0),
	}

	msgBytes, err := pingMsg.MarshalBinary()
	if err != nil {
		t.Error("Failed to marshal CoAP message")
	}

	HandleIncomingUartPacket(socket, 10, msgBytes)

	response := <-PendingResponses

	if response.Socket.Handle != socket.Handle {
		t.Error("Expected socket handle in SendMessageHandler to be", socket.Handle, "but is", response.Socket.Handle)
	}

	pongMsg, err := coapmsg.ParseMessage(response.Data)
	if err != nil {
		t.Error("Failed to parse CoAP message", err)
	}

	if len(pongMsg.Payload) != 0 {
		t.Error("Pong payload must be empty but has len", len(pongMsg.Payload))
	}
	if pongMsg.Code != 0 {
		t.Error("Expected code 0 but got", pongMsg.Code)
	}
	if pongMsg.Type != coapmsg.Reset {
		t.Error("Expected type coapmsg.Reset but got", pongMsg.Type)
	}
}
