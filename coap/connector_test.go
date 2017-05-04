package coap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/Lobaro/coap-go/coapmsg"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

type TestConnector struct {
	//ReceiveBuf *SafeBuffer // Data that is received by the client (connection reader)
	//SendBuf    *SafeBuffer // Data that is send by the client (connection writer)

	In   *PacketBuffer
	Out  *PacketBuffer
	conn *serialConnection
}

func NewTestConnector() *TestConnector {
	return &TestConnector{
		In:  &PacketBuffer{name: "in"},
		Out: &PacketBuffer{name: "out"},
	}
}

func (c *TestConnector) FakeReceiveData(data []byte) {
	c.In.WritePacket(data)
}

func (c *TestConnector) FakeReceiveMessage(msg coapmsg.Message) error {
	p := msg.MarshalBinary()
	c.In.WritePacket(p)
	return nil
}

func (c *TestConnector) WaitForSendMessage(timeout time.Duration) (coapmsg.Message, error) {
	buf := &bytes.Buffer{}

	ctx, _ := context.WithTimeout(context.Background(), timeout)
	for {
		tmp, isPrefix, err := c.Out.ReadPacket()
		if err != nil && err != io.EOF {
			return coapmsg.NewMessage(), err
		}

		buf.Write(tmp)

		if buf.Len() > 0 && !isPrefix {
			break
		}

		select {
		case <-ctx.Done():
			return coapmsg.NewMessage(), errors.New(fmt.Sprintf("WaitForSendMessage Timeout: %d", c.Out.Len()))
		default:
		}
	}
	return coapmsg.ParseMessage(buf.Bytes())
}

func (c *TestConnector) GetSendMessage() (coapmsg.Message, error) {

	p, _, err := c.Out.ReadPacket()
	if err != nil {
		return coapmsg.NewMessage(), err
	}
	return coapmsg.ParseMessage(p)
}

func (c *TestConnector) GetSendData() ([]byte, error) {
	p, _, err := c.Out.ReadPacket()
	if err != nil {
		return nil, err
	}
	return p, nil
}

type PacketBuffer struct {
	name    string
	mu      sync.Mutex
	packets [][]byte
}

var NO_PACKET = errors.New("No Packets availiable")

func (rw *PacketBuffer) ReadPacket() (p []byte, isPrefix bool, err error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if len(rw.packets) > 0 {
		res := rw.packets[0]
		logrus.WithField("raw", res).Info("ReadPacket from " + rw.name)
		rw.packets = rw.packets[1:len(rw.packets)]
		return res, false, nil
	}
	return nil, true, io.EOF
}

func (rw *PacketBuffer) WritePacket(p []byte) (err error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	logrus.WithField("raw", p).Info("WritePacket to " + rw.name)
	rw.packets = append(rw.packets, p)
	return nil
}

func (rw *PacketBuffer) Len() int {
	return len(rw.packets)
}

func (c *TestConnector) Connect(host string) (*serialConnection, error) {

	if c.conn != nil {
		return c.conn, nil
	}

	conn := &serialConnection{
		config: nil,
		port:   nil,
		/*
			reader:   slip.NewReader(c.ReceiveBuf),
			writer:   slip.NewWriter(c.SendBuf),
		*/
		reader:   c.In,
		writer:   c.Out,
		deadline: time.Now().Add(UART_CONNECTION_TIMEOUT),
	}

	conn.Open()

	c.conn = conn

	return conn, nil
}