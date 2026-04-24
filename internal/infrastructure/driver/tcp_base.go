package driver

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// tcpConnectionDriver 负责维护 TCP 级别连接与命令交互。
type tcpConnectionDriver struct {
	model    string
	address  string
	mu       sync.Mutex
	conn     net.Conn
	breaker  *CircuitBreaker
	retryCfg RetryConfig
}

func newTCPConnectionDriver(model string, host string, port int) *tcpConnectionDriver {
	return &tcpConnectionDriver{
		model:    model,
		address:  net.JoinHostPort(host, strconv.Itoa(port)),
		breaker:  NewCircuitBreaker(DefaultCircuitBreakerConfig()),
		retryCfg: DefaultRetryConfig(),
	}
}

func (d *tcpConnectionDriver) Connect(ctx context.Context) error {
	d.mu.Lock()
	if d.conn != nil {
		_ = d.conn.Close()
		d.conn = nil
	}
	d.mu.Unlock()

	if !d.breaker.AllowRequest() {
		return fmt.Errorf("%s: circuit breaker is open", d.model)
	}

	var lastErr error
	rs := NewRetryStrategy(d.retryCfg)
	for attempt := 0; rs.ShouldRetry(attempt); attempt++ {
		if attempt > 0 {
			delay := time.Duration(rs.NextDelay(attempt)) * time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		d.mu.Lock()
		var dialer net.Dialer
		conn, err := dialer.DialContext(ctx, "tcp", d.address)
		if err != nil {
			lastErr = fmt.Errorf("%s dial %s: %w", d.model, d.address, err)
			d.mu.Unlock()
			continue
		}
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			_ = tcpConn.SetKeepAlive(true)
			_ = tcpConn.SetKeepAlivePeriod(30 * time.Second)
		}
		_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
		drainBuf := make([]byte, 4096)
		_, _ = conn.Read(drainBuf)
		_ = conn.SetReadDeadline(time.Time{})
		d.conn = conn
		d.mu.Unlock()

		d.breaker.RecordSuccess()
		return nil
	}

	d.breaker.RecordFailure()
	return fmt.Errorf("%s connect failed after %d attempts: %w", d.model, d.retryCfg.MaxAttempts, lastErr)
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

// sendSCPICommand 发送 SCPI 命令并读取响应（带超时）。
func (d *tcpConnectionDriver) sendSCPICommand(ctx context.Context, cmd string, readTimeout time.Duration) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.conn == nil {
		return "", fmt.Errorf("%s: not connected", d.model)
	}
	_ = d.conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	drainBuf := make([]byte, 4096)
	for {
		if _, err := d.conn.Read(drainBuf); err != nil {
			break
		}
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(5 * time.Second)
	}
	if err := d.conn.SetWriteDeadline(deadline); err != nil {
		return "", fmt.Errorf("%s set write deadline: %w", d.model, err)
	}
	if _, err := fmt.Fprintf(d.conn, "%s\r\n", cmd); err != nil {
		return "", fmt.Errorf("%s write SCPI command %q: %w", d.model, cmd, err)
	}
	if err := d.conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		return "", fmt.Errorf("%s set read deadline: %w", d.model, err)
	}
	var resp strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := d.conn.Read(buf)
		if err != nil {
			break
		}
		resp.Write(buf[:n])
		if strings.Contains(resp.String(), "\n") {
			break
		}
	}
	return strings.TrimSpace(resp.String()), nil
}

// sendWTN1604Command 发送 WTN1604 命令并读取长度前缀响应。
func (d *tcpConnectionDriver) sendWTN1604Command(ctx context.Context, cmd string, readTimeout time.Duration) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.conn == nil {
		return "", fmt.Errorf("%s: not connected", d.model)
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(5 * time.Second)
	}
	if err := d.conn.SetWriteDeadline(deadline); err != nil {
		return "", fmt.Errorf("%s set write deadline: %w", d.model, err)
	}
	if _, err := d.conn.Write([]byte(cmd)); err != nil {
		return "", fmt.Errorf("%s write command %q: %w", d.model, cmd, err)
	}
	if err := d.conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		return "", fmt.Errorf("%s set read deadline: %w", d.model, err)
	}
	lenBuf := make([]byte, 2)
	if _, err := io.ReadFull(d.conn, lenBuf); err != nil {
		return "", fmt.Errorf("%s read length prefix: %w", d.model, err)
	}
	totalLen := int(binary.BigEndian.Uint16(lenBuf))
	if totalLen < 2 {
		return "", fmt.Errorf("%s invalid response length: %d", d.model, totalLen)
	}
	dataLen := totalLen - 2
	data := make([]byte, dataLen)
	if dataLen > 0 {
		if _, err := io.ReadFull(d.conn, data); err != nil {
			return "", fmt.Errorf("%s read response data: %w", d.model, err)
		}
	}
	response := strings.TrimSpace(strings.ReplaceAll(string(data), "\x00", ""))
	return response, nil
}
