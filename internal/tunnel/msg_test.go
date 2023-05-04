package tunnel

import (
	"bytes"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ net.Conn = (*mockNetConn)(nil)

type mockNetConn struct {
	bytes.Buffer
}

func (*mockNetConn) Close() error                       { return nil }
func (*mockNetConn) LocalAddr() net.Addr                { return nil }
func (*mockNetConn) RemoteAddr() net.Addr               { return nil }
func (*mockNetConn) SetDeadline(t time.Time) error      { return nil }
func (*mockNetConn) SetReadDeadline(t time.Time) error  { return nil }
func (*mockNetConn) SetWriteDeadline(t time.Time) error { return nil }

var _ = Describe("Msg", func() {
	var (
		conn  *mockNetConn
		codec *msgCodec
	)

	BeforeEach(func() {
		conn = &mockNetConn{}
		codec = newMsgCodec(conn)
	})

	Describe("Write", func() {
		It("should succeed", func() {
			Expect(codec.writeMsg("foo", "bar")).Should(Succeed())

			jsonStr := `["foo","bar"]`
			l := byte(len(jsonStr))
			b := append([]byte{l}, []byte(jsonStr)...)

			Expect(conn.Bytes()).Should(Equal(b))
		})
	})

	Describe("Read", func() {
		It("can read regular message", func() {
			Expect(codec.writeMsg("foo", "bar")).Should(Succeed())

			msg, err := codec.readMsg()
			Expect(err).Should(BeNil())
			Expect(msg).Should(Equal([]string{"foo", "bar"}))
		})

		It("can read empty message", func() {
			Expect(codec.writeMsg()).Should(Succeed())

			msg, err := codec.readMsg()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(msg)).Should(Equal(0))
		})
	})
})
