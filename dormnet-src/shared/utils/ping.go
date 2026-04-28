package utils

import (
	"net"
	"os"
	"time"

	"github.com/openwrt-dormnet/dormnet/shared/errx"
	"go.uber.org/zap"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type Pinger interface {
	Ping(target string, count int) (int, errx.Exception)
	PingUntilFailed(target string, count int) errx.Exception
	PingUntilSuccess(target string, count int) errx.Exception
	PingOnce(target string) (*icmp.MessageBody, errx.Exception)
}

type pinger struct {
	dialer *net.Dialer
	id     int
}

func NewPinger(iface string) Pinger {
	return &pinger{
		dialer: NewIfaceController(iface).CreateDialer(5 * time.Second),
		id:     os.Getegid() & 0xffff,
	}
}

func NewPingerDefault() Pinger {
	return &pinger{
		id: os.Getegid() & 0xffff,
	}
}

func (p *pinger) createConnection(target string) (conn net.Conn, err errx.Exception) {
	var err2 error
	if p.dialer != nil {
		conn, err2 = p.dialer.Dial("ip4:icmp", target)
	} else {
		conn, err2 = net.Dial("ip4:icmp", target)
	}
	if err2 != nil {
		err = errx.NewExceptionWithError(err2, "failed to create icmp connection")
	}
	return
}

func (p *pinger) runInConnection(target string, block func(net.Conn) errx.Exception) errx.Exception {
	conn, err := p.createConnection(target)
	if err != nil {
		return errx.NewExceptionWithCause(err, "failed to start ping connection")
	}
	defer conn.Close()
	return block(conn)
}

func (p *pinger) Ping(target string, count int) (success int, err errx.Exception) {
	err = p.runInConnection(target, func(conn net.Conn) errx.Exception {
		success = 0
		DelayJoinTasks(count, time.Second, func(index int) bool {
			if _, err = p.sendPing(conn, index); err != nil {
				err = errx.NewExceptionWithCause(err, "multi ping failed", zap.Int("index", index))
			} else {
				success += 1
			}
			return false
		})
		return err
	})
	return
}

func (p *pinger) PingUntilFailed(target string, count int) errx.Exception {
	return p.runInConnection(target, func(conn net.Conn) errx.Exception {
		DelayJoinTasks(count, time.Second, func(index int) bool {
			if _, err := p.sendPing(conn, index); err != nil {
				err = errx.NewExceptionWithCause(err, "multi ping failed", zap.Int("index", index))
			}
			return false
		})
		return nil
	})
}

func (p *pinger) PingUntilSuccess(target string, count int) errx.Exception {
	return p.runInConnection(target, func(conn net.Conn) errx.Exception {
		var err errx.Exception
		DelayJoinTasks(count, time.Second, func(index int) bool {
			_, err = p.sendPing(conn, index)
			if err == nil {
				return true
			}
			err = errx.NewExceptionWithCause(err, "multi ping failed")
			return false
		})
		return err
	})
}

func (p *pinger) PingOnce(target string) (body *icmp.MessageBody, err errx.Exception) {
	err = p.runInConnection(target, func(conn net.Conn) errx.Exception {
		var err errx.Exception
		if body, err = p.sendPing(conn, 0); err != nil {
			return errx.NewExceptionWithCause(err, "ping failed")
		}
		return nil
	})
	return
}

var pingPayload []byte

func (p *pinger) sendPing(conn net.Conn, seq int) (*icmp.MessageBody, errx.Exception) {
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho, // type 8
		Code: 0,
		Body: &icmp.Echo{
			ID:   p.id,
			Seq:  seq & 0xffff,
			Data: pingPayload,
		},
	}
	b, err := msg.Marshal(nil)
	if err != nil {
		return nil, errx.NewExceptionWithError(err, "failed to marshal ICMP message")
	}
	if _, err = conn.Write(b); err != nil {
		return nil, errx.NewExceptionWithError(err, "failed to send ping")
	}
	if err = conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return nil, errx.NewExceptionWithError(err, "failed to set read deadline")
	}
	reply := make([]byte, 1500)
	count, err := conn.Read(reply)
	if err != nil {
		return nil, errx.NewExceptionWithError(err, "failed to read ping")
	}
	if count <= 0 {
		return nil, errx.NewException("ping reply nothing")
	}
	replyMsg, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), reply[:count])
	if err != nil {
		return nil, errx.NewExceptionWithError(err, "failed to parse ping reply")
	}
	if echoReply, ok := replyMsg.Body.(*icmp.Echo); ok && echoReply.ID == p.id && echoReply.Seq == seq {
		return &replyMsg.Body, nil
	}
	if _, ok := replyMsg.Body.(*icmp.RawBody); ok {
		return &replyMsg.Body, nil
	}
	return nil, errx.NewException("ping reply mismatched")
}

func init() {
	pingPayload = make([]byte, 56)
	for i := 0; i < len(pingPayload); i++ {
		pingPayload[i] = byte('a' + (i % 26))
	}
}
