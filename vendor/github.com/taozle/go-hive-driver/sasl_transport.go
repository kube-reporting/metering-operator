package hive

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/emersion/go-sasl"
	"io"
)

const (
	saslStatusStart    int8 = 1
	saslStatusOK            = 2
	saslStatusBad           = 3
	saslStatusError         = 4
	saslStatusComplete      = 5
)

// All implementations are copied from TFramedTransport except `Open`.
type TSaslClientTransport struct {
	transport     thrift.TTransport
	clientFactory func() sasl.Client

	client    sasl.Client
	buf       bytes.Buffer
	reader    *bufio.Reader
	frameSize uint32 //Current remaining size of the frame. if ==0 read next frame header
	buffer    [4]byte
	maxLength uint32
}

func NewTSaslClientTransport(transport thrift.TTransport, factory func() sasl.Client) thrift.TTransport {
	return &TSaslClientTransport{
		transport:     transport,
		clientFactory: factory,

		reader:    bufio.NewReader(transport),
		maxLength: thrift.DEFAULT_MAX_LENGTH,
	}
}

func (c *TSaslClientTransport) Read(p []byte) (n int, err error) {
	if c.frameSize == 0 {
		c.frameSize, err = c.readFrameHeader()
		if err != nil {
			return
		}
	}
	if c.frameSize < uint32(len(p)) {
		frameSize := c.frameSize
		tmp := make([]byte, c.frameSize)
		n, err = c.Read(tmp)
		copy(p, tmp)
		if err == nil {
			err = thrift.NewTTransportExceptionFromError(fmt.Errorf("Not enough frame size %d to read %d bytes", frameSize, len(p)))
			return
		}
	}
	got, err := c.reader.Read(p)
	c.frameSize = c.frameSize - uint32(got)
	//sanity check
	if c.frameSize < 0 {
		return 0, thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION, "Negative frame size")
	}
	return got, thrift.NewTTransportExceptionFromError(err)
}

func (c *TSaslClientTransport) readFrameHeader() (uint32, error) {
	buf := c.buffer[:4]
	if _, err := io.ReadFull(c.reader, buf); err != nil {
		return 0, err
	}
	size := binary.BigEndian.Uint32(buf)
	if size < 0 || size > c.maxLength {
		return 0, thrift.NewTTransportException(thrift.UNKNOWN_TRANSPORT_EXCEPTION, fmt.Sprintf("Incorrect frame size (%d)", size))
	}
	return size, nil
}

func (c *TSaslClientTransport) Write(p []byte) (n int, err error) {
	n, err = c.buf.Write(p)
	return n, thrift.NewTTransportExceptionFromError(err)
}

func (c *TSaslClientTransport) Close() error {
	c.client = nil
	return c.transport.Close()
}

func (c *TSaslClientTransport) Flush() (err error) {
	size := c.buf.Len()
	buf := c.buffer[:4]
	binary.BigEndian.PutUint32(buf, uint32(size))
	_, err = c.transport.Write(buf)
	if err != nil {
		c.buf.Truncate(0)
		return thrift.NewTTransportExceptionFromError(err)
	}
	if size > 0 {
		if n, err := c.buf.WriteTo(c.transport); err != nil {
			print("Error while flushing write buffer of size ", size, " to transport, only wrote ", n, " bytes: ", err.Error(), "\n")
			c.buf.Truncate(0)
			return thrift.NewTTransportExceptionFromError(err)
		}
	}
	err = c.transport.Flush()
	return thrift.NewTTransportExceptionFromError(err)
}

func (c *TSaslClientTransport) RemainingBytes() uint64 {
	return uint64(c.frameSize)
}

func (c *TSaslClientTransport) Open() error {
	if !c.IsOpen() {
		if err := c.transport.Open(); err != nil {
			return err
		}
	}
	if c.client != nil {
		return thrift.NewTTransportException(thrift.NOT_OPEN, "sasl transport is already opened")
	}

	c.client = c.clientFactory()
	mech, ir, err := c.client.Start()
	if err != nil {
		return err
	}

	// Send initial response
	if err := c.sendMessage(saslStatusStart, []byte(mech)); err != nil {
		return err
	}
	if err := c.sendMessage(saslStatusOK, ir); err != nil {
		return err
	}

	// SASL negotiation loop
	for {
		status, payload, err := c.recvMessage()
		if err != nil {
			return err
		}

		if status != saslStatusOK && status != saslStatusComplete {
			return thrift.NewTTransportException(thrift.NOT_OPEN, fmt.Sprintf("Bad status: %d (%s)", status, payload))
		}

		if status == saslStatusComplete {
			break
		}

		response, err := c.client.Next(payload)
		if err != nil {
			return err
		}

		if err := c.sendMessage(saslStatusOK, response); err != nil {
			return err
		}
	}

	return nil
}

func (c *TSaslClientTransport) sendMessage(status int8, body []byte) error {
	header := make([]byte, 5)
	header[0] = byte(status)
	binary.BigEndian.PutUint32(header[1:], uint32(len(body)))

	if _, err := c.transport.Write(append(header, body...)); err != nil {
		return err
	}

	return c.transport.Flush()
}

func (c *TSaslClientTransport) recvMessage() (int8, []byte, error) {
	header := make([]byte, 5)
	if _, err := c.transport.Read(header); err != nil {
		return 0, nil, err
	}

	status := int8(header[0])
	length := binary.BigEndian.Uint32(header[1:])
	if length == 0 {
		return status, nil, nil
	}

	payload := make([]byte, length)
	if _, err := c.transport.Read(payload); err != nil {
		return 0, nil, err
	}

	return status, payload, nil
}

func (c *TSaslClientTransport) IsOpen() bool {
	return c.transport.IsOpen()
}
