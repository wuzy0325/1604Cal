package driver

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

// tcpConnectionDriver 负责维护 TCP 级别连接。
// 当前仅落地连接链路，协议命令交互将在后续迭代补齐。
type tcpConnectionDriver struct {
	model   string
	address string

	mu   sync.Mutex
	conn net.Conn
}

func newTCPConnectionDriver(model string, host string, port int) *tcpConnectionDriver {
	return &tcpConnectionDriver{
		model:   model,
		address: net.JoinHostPort(host, strconv.Itoa(port)),
	}
}

func (d *tcpConnectionDriver) Connect(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.conn != nil {
		_ = d.conn.Close()
		d.conn = nil
	}

	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", d.address)
	if err != nil {
		return fmt.Errorf("%s dial %s: %w", d.model, d.address, err)
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		_ = tcpConn.SetKeepAlive(true)
		_ = tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	d.conn = conn
	return nil
}

func (d *tcpConnectionDriver) Disconnect(_ context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.conn == nil {
		return nil
	}

	if err := d.conn.Close(); err != nil {
		return fmt.Errorf("%s close connection: %w", d.model, err)
	}

	d.conn = nil
	return nil
}

type wtn1604Driver struct {
	base *tcpConnectionDriver
}

func newWTN1604Driver(host string, port int) *wtn1604Driver {
	return &wtn1604Driver{base: newTCPConnectionDriver("WTN1604", host, port)}
}

func (d *wtn1604Driver) Connect(ctx context.Context) error {
	return d.base.Connect(ctx)
}

func (d *wtn1604Driver) Disconnect(ctx context.Context) error {
	return d.base.Disconnect(ctx)
}

type const811ADriver struct {
	base *tcpConnectionDriver
}

func newConST811ADriver(host string, port int) *const811ADriver {
	return &const811ADriver{base: newTCPConnectionDriver("ConST 811A", host, port)}
}

func (d *const811ADriver) Connect(ctx context.Context) error {
	return d.base.Connect(ctx)
}

func (d *const811ADriver) Disconnect(ctx context.Context) error {
	return d.base.Disconnect(ctx)
}

type const820Driver struct {
	base *tcpConnectionDriver
}

func newConST820Driver(host string, port int) *const820Driver {
	return &const820Driver{base: newTCPConnectionDriver("ConST 820", host, port)}
}

func (d *const820Driver) Connect(ctx context.Context) error {
	return d.base.Connect(ctx)
}

func (d *const820Driver) Disconnect(ctx context.Context) error {
	return d.base.Disconnect(ctx)
}
