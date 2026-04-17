package driver

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// tcpConnectionDriver 负责维护 TCP 级别连接与命令交互。
type tcpConnectionDriver struct {
	model   string
	address string
	mu      sync.Mutex
	conn    net.Conn
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
	_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	drainBuf := make([]byte, 4096)
	_, _ = conn.Read(drainBuf)
	_ = conn.SetReadDeadline(time.Time{})
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

// sendCommand 发送命令并读取响应（带超时），用于 WTN1604 协议。
func (d *tcpConnectionDriver) sendCommand(ctx context.Context, cmd string, readTimeout time.Duration) (string, error) {
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
	buf := make([]byte, 4096)
	n, err := d.conn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("%s read response: %w", d.model, err)
	}
	return strings.TrimSpace(string(buf[:n])), nil
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

// ---------------------------------------------------------------------------
// WTN1604 计量设备驱动
// ---------------------------------------------------------------------------

type WTN1604Driver struct {
	base *tcpConnectionDriver
}

func newWTN1604Driver(host string, port int) *WTN1604Driver {
	return &WTN1604Driver{base: newTCPConnectionDriver("WTN1604", host, port)}
}

func (d *WTN1604Driver) Connect(ctx context.Context) error    { return d.base.Connect(ctx) }
func (d *WTN1604Driver) Disconnect(ctx context.Context) error { return d.base.Disconnect(ctx) }

func (d *WTN1604Driver) ReadValveStatus(ctx context.Context) (string, error) {
	resp, err := d.base.sendWTN1604Command(ctx, "@01  0", 3*time.Second)
	if err != nil {
		return "", fmt.Errorf("read valve status: %w", err)
	}

	raw := strings.TrimSpace(resp)
	val := strings.TrimSpace(strings.TrimPrefix(raw, "A"))
	if val == "" {
		val = raw
	}

	if num, parseErr := strconv.Atoi(strings.TrimSpace(val)); parseErr == nil {
		switch num {
		case 1:
			return "calibration", nil
		case 0, 2, 3:
			// 现场兼容：部分 1604 固件在 RUN/测量态返回 3，按 measurement 处理。
			return "measurement", nil
		}
	}

	switch strings.ToLower(strings.TrimSpace(val)) {
	case "calibration", "calibrate", "open", "opened", "on":
		return "calibration", nil
	case "measurement", "measure", "close", "closed", "off":
		return "measurement", nil
	default:
		return raw, nil
	}
}

func (d *WTN1604Driver) SetValveStatus(ctx context.Context, status string) error {
	var cmd string
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "calibration", "calibrate", "1":
		cmd = "w0C01"
	case "measurement", "measure", "2":
		cmd = "w0C00"
	default:
		return fmt.Errorf("invalid valve status: %s", status)
	}
	resp, err := d.base.sendWTN1604Command(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set valve status: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("set valve status failed: response %q", resp)
	}
	return nil
}

func (d *WTN1604Driver) ReadUnit(ctx context.Context) (string, error) {
	resp, err := d.base.sendWTN1604Command(ctx, "u01101", 3*time.Second)
	if err != nil {
		return "", fmt.Errorf("read unit: %w", err)
	}
	unit, ok := parseWTN1604Unit(resp)
	if !ok {
		return coefficientToUnit(strings.TrimSpace(resp)), nil
	}
	return unit, nil
}

func (d *WTN1604Driver) SetUnit(ctx context.Context, unit string) error {
	coef, ok := unitToCoefficient(unit)
	if !ok {
		return fmt.Errorf("unsupported unit: %s", unit)
	}
	cmd := fmt.Sprintf("v01101 %s", coef)
	resp, err := d.base.sendWTN1604Command(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set unit: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("set unit failed: response %q", resp)
	}
	return nil
}

func (d *WTN1604Driver) CollectData(ctx context.Context, channels []int) ([]float64, error) {
	bitmap := channelsToBitmap(channels)
	cmd := fmt.Sprintf("r%s0", bitmap)
	resp, err := d.base.sendWTN1604Command(ctx, cmd, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("collect data: %w", err)
	}
	if strings.HasPrefix(resp, "N") {
		return nil, fmt.Errorf("device error: %s", resp)
	}
	parts := strings.Fields(resp)
	values := make([]float64, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			continue
		}
		values = append(values, v)
	}
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
	return values, nil
}

func (d *WTN1604Driver) ReadDeviceInfo(ctx context.Context) (map[string]string, error) {
	info := make(map[string]string)
	resp, err := d.base.sendWTN1604Command(ctx, "A", 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("device communication test: %w", err)
	}
	info["commTest"] = resp
	if resp, err = d.base.sendWTN1604Command(ctx, "q00", 3*time.Second); err == nil {
		info["model"] = resp
	}
	if resp, err = d.base.sendWTN1604Command(ctx, "q01", 3*time.Second); err == nil {
		info["version"] = resp
	}
	return info, nil
}

func (d *WTN1604Driver) Reset(ctx context.Context) error {
	resp, err := d.base.sendWTN1604Command(ctx, "B", 3*time.Second)
	if err != nil {
		return fmt.Errorf("reset: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("reset failed: response %q", resp)
	}
	return nil
}

func (d *WTN1604Driver) StartCalibration(ctx context.Context, channels []int, pressurePoints int, avgPoints int) error {
	bitmap := channelsToBitmap(channels)
	cmd := fmt.Sprintf("C 00 %s %d 1 %d", bitmap, pressurePoints, avgPoints)
	resp, err := d.base.sendWTN1604Command(ctx, cmd, 5*time.Second)
	if err != nil {
		return fmt.Errorf("start calibration: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("start calibration failed: response %q", resp)
	}
	return nil
}

func (d *WTN1604Driver) CollectCalibrationPoint(ctx context.Context, pointIndex int, targetPressure float64) ([]float64, error) {
	cmd := fmt.Sprintf("C 01 %d %.2f", pointIndex, targetPressure)
	resp, err := d.base.sendWTN1604Command(ctx, cmd, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("collect calibration point: %w", err)
	}
	if strings.HasPrefix(resp, "N") {
		return nil, fmt.Errorf("device error: %s", resp)
	}
	parts := strings.Fields(resp)
	values := make([]float64, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			continue
		}
		values = append(values, v)
	}
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
	return values, nil
}

func (d *WTN1604Driver) PerformFitting(ctx context.Context) error {
	resp, err := d.base.sendWTN1604Command(ctx, "C 02", 10*time.Second)
	if err != nil {
		return fmt.Errorf("perform fitting: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("perform fitting failed: response %q", resp)
	}
	return nil
}

func (d *WTN1604Driver) EndCalibration(ctx context.Context) error {
	resp, err := d.base.sendWTN1604Command(ctx, "C 03", 5*time.Second)
	if err != nil {
		return fmt.Errorf("end calibration: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("end calibration failed: response %q", resp)
	}
	return nil
}

func (d *WTN1604Driver) SaveCoefficients(ctx context.Context) error {
	resp, err := d.base.sendWTN1604Command(ctx, "w08", 3*time.Second)
	if err != nil {
		return fmt.Errorf("save zero coefficient: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("save zero coefficient failed: response %q", resp)
	}
	resp, err = d.base.sendWTN1604Command(ctx, "w09", 3*time.Second)
	if err != nil {
		return fmt.Errorf("save gain coefficient: %w", err)
	}
	if resp != "A" {
		return fmt.Errorf("save gain coefficient failed: response %q", resp)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ConST 811A 打压设备驱动 (标准SCPI)
// ---------------------------------------------------------------------------

type ConST811ADriver struct {
	base *tcpConnectionDriver
}

func newConST811ADriver(host string, port int) *ConST811ADriver {
	return &ConST811ADriver{base: newTCPConnectionDriver("ConST 811A", host, port)}
}

func (d *ConST811ADriver) Connect(ctx context.Context) error {
	if err := d.base.Connect(ctx); err != nil {
		return err
	}
	for i := 0; i < 10; i++ {
		resp, err := d.base.sendSCPICommand(ctx, "PRESsure:MODule1:STABle?", 2*time.Second)
		if err == nil && (resp == "0" || resp == "1") {
			break
		}
	}
	return nil
}

func (d *ConST811ADriver) Disconnect(ctx context.Context) error { return d.base.Disconnect(ctx) }

func (d *ConST811ADriver) SetTargetPressure(ctx context.Context, target float64) error {
	cmd := fmt.Sprintf("PRESsure:TARGet %.4f", target)
	_, err := d.base.sendSCPICommand(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set target pressure: %w", err)
	}
	return nil
}

func (d *ConST811ADriver) Stop(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "PRESsure:MODE VENT", 3*time.Second)
	if err != nil {
		return fmt.Errorf("stop pressure: %w", err)
	}
	return nil
}

func (d *ConST811ADriver) Exhaust(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "PRESsure:MODE VENT", 3*time.Second)
	if err != nil {
		return fmt.Errorf("exhaust: %w", err)
	}
	return nil
}

func (d *ConST811ADriver) ReadCurrentPressure(ctx context.Context) (float64, error) {
	resp, err := d.base.sendSCPICommand(ctx, "PRESsure0?", 3*time.Second)
	if err != nil {
		return 0, fmt.Errorf("read current pressure: %w", err)
	}
	return parseSCPIPressure(resp)
}

func (d *ConST811ADriver) ReadUnit(ctx context.Context) (string, error) {
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := d.base.sendSCPICommand(ctx, "PRESsure0?", 3*time.Second)
		if err != nil {
			return "", fmt.Errorf("read unit: %w", err)
		}
		parts := strings.SplitN(resp, ",", 2)
		if len(parts) >= 2 {
			unit := strings.TrimSpace(parts[1])
			if isValidPressureUnit(unit) {
				return unit, nil
			}
		}
	}
	return "", nil
}

func (d *ConST811ADriver) SetUnit(ctx context.Context, unit string) error {
	unitCode, ok := pressureUnitToCode(unit)
	if !ok {
		return fmt.Errorf("unsupported pressure unit: %s", unit)
	}
	cmd := fmt.Sprintf("PRESsure:MODule1:UNIT %s", unitCode)
	_, err := d.base.sendSCPICommand(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set unit: %w", err)
	}
	return nil
}

func (d *ConST811ADriver) ReadStability(ctx context.Context) (bool, error) {
	resp, err := d.base.sendSCPICommand(ctx, "PRESsure:MODule1:STABle?", 3*time.Second)
	if err != nil {
		return false, fmt.Errorf("read stability: %w", err)
	}
	return strings.TrimSpace(resp) == "1", nil
}

func (d *ConST811ADriver) StartControl(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "PRESsure:MODE CONTROL", 3*time.Second)
	if err != nil {
		return fmt.Errorf("start pressure control: %w", err)
	}
	return nil
}

func (d *ConST811ADriver) ReadTargetRange(ctx context.Context) (min, max float64, err error) {
	resp, err := d.base.sendSCPICommand(ctx, "PRESsure:TARGet:RANGe?", 3*time.Second)
	if err != nil {
		return 0, 0, fmt.Errorf("read target range: %w", err)
	}
	return parseTargetRange(resp)
}

// ---------------------------------------------------------------------------
// ConST 820 打压设备驱动 (简化SCPI)
// ---------------------------------------------------------------------------

type ConST820Driver struct {
	base *tcpConnectionDriver
}

func newConST820Driver(host string, port int) *ConST820Driver {
	return &ConST820Driver{base: newTCPConnectionDriver("ConST 820", host, port)}
}

func (d *ConST820Driver) Connect(ctx context.Context) error {
	if err := d.base.Connect(ctx); err != nil {
		return err
	}
	for i := 0; i < 10; i++ {
		resp, err := d.base.sendSCPICommand(ctx, "OUTPut:PRESsure:STABle?", 2*time.Second)
		if err == nil && (resp == "0" || resp == "1") {
			break
		}
	}
	return nil
}

func (d *ConST820Driver) Disconnect(ctx context.Context) error { return d.base.Disconnect(ctx) }

func (d *ConST820Driver) SetTargetPressure(ctx context.Context, target float64) error {
	cmd := fmt.Sprintf("SOURce:PRESsure %.4f", target)
	_, err := d.base.sendSCPICommand(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set target pressure: %w", err)
	}
	return nil
}

func (d *ConST820Driver) Stop(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "OUTPut:PRESsure:MODE VENT", 3*time.Second)
	if err != nil {
		return fmt.Errorf("stop pressure: %w", err)
	}
	return nil
}

func (d *ConST820Driver) Exhaust(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "OUTPut:PRESsure:MODE VENT", 3*time.Second)
	if err != nil {
		return fmt.Errorf("exhaust: %w", err)
	}
	return nil
}

func (d *ConST820Driver) ReadCurrentPressure(ctx context.Context) (float64, error) {
	resp, err := d.base.sendSCPICommand(ctx, "MEASure:SCALar:PRESsure1?", 3*time.Second)
	if err != nil {
		return 0, fmt.Errorf("read current pressure: %w", err)
	}
	return parseSCPIPressure(resp)
}

func (d *ConST820Driver) ReadUnit(ctx context.Context) (string, error) {
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := d.base.sendSCPICommand(ctx, "MEASure:SCALar:PRESsure1?", 3*time.Second)
		if err != nil {
			return "", fmt.Errorf("read unit: %w", err)
		}
		parts := strings.SplitN(resp, ",", 2)
		if len(parts) >= 2 {
			unit := strings.TrimSpace(parts[1])
			if isValidPressureUnit(unit) {
				return unit, nil
			}
		}
	}
	return "", nil
}

func (d *ConST820Driver) SetUnit(ctx context.Context, unit string) error {
	unitCode, ok := pressureUnitToCode820(unit)
	if !ok {
		return fmt.Errorf("unsupported pressure unit: %s", unit)
	}
	cmd := fmt.Sprintf("UNIT:PRESsure %s", unitCode)
	_, err := d.base.sendSCPICommand(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set unit: %w", err)
	}
	return nil
}

func (d *ConST820Driver) ReadStability(ctx context.Context) (bool, error) {
	resp, err := d.base.sendSCPICommand(ctx, "OUTPut:PRESsure:STABle?", 3*time.Second)
	if err != nil {
		return false, fmt.Errorf("read stability: %w", err)
	}
	return strings.TrimSpace(resp) == "1", nil
}

func (d *ConST820Driver) StartControl(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "OUTPut:PRESsure:MODE CONTROL", 3*time.Second)
	if err != nil {
		return fmt.Errorf("start pressure control: %w", err)
	}
	return nil
}

func (d *ConST820Driver) ReadTargetRange(ctx context.Context) (min, max float64, err error) {
	resp, err := d.base.sendSCPICommand(ctx, "PRESsure:TARGet:RANGe?", 3*time.Second)
	if err != nil {
		return 0, 0, fmt.Errorf("read target range: %w", err)
	}
	return parseTargetRange(resp)
}

// ---------------------------------------------------------------------------
// ConST 860 打压设备驱动 (标准SCPI，部分简化)
// ---------------------------------------------------------------------------

type ConST860Driver struct {
	base *tcpConnectionDriver
}

func newConST860Driver(host string, port int) *ConST860Driver {
	return &ConST860Driver{base: newTCPConnectionDriver("ConST 860", host, port)}
}

func (d *ConST860Driver) Connect(ctx context.Context) error {
	if err := d.base.Connect(ctx); err != nil {
		return err
	}
	for i := 0; i < 10; i++ {
		resp, err := d.base.sendSCPICommand(ctx, "PRESsure:STABle?", 2*time.Second)
		if err == nil && (resp == "0" || resp == "1") {
			break
		}
	}
	return nil
}

func (d *ConST860Driver) Disconnect(ctx context.Context) error { return d.base.Disconnect(ctx) }

func (d *ConST860Driver) SetTargetPressure(ctx context.Context, target float64) error {
	cmd := fmt.Sprintf("PRESsure:TARGet %.4f", target)
	_, err := d.base.sendSCPICommand(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set target pressure: %w", err)
	}
	return nil
}

func (d *ConST860Driver) Stop(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "PRESsure:MODule:CONTrol VENT", 3*time.Second)
	if err != nil {
		return fmt.Errorf("stop pressure: %w", err)
	}
	return nil
}

func (d *ConST860Driver) Exhaust(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "PRESsure:MODule:CONTrol VENT", 3*time.Second)
	if err != nil {
		return fmt.Errorf("exhaust: %w", err)
	}
	return nil
}

func (d *ConST860Driver) ReadCurrentPressure(ctx context.Context) (float64, error) {
	resp, err := d.base.sendSCPICommand(ctx, "PRESsure?", 3*time.Second)
	if err != nil {
		return 0, fmt.Errorf("read current pressure: %w", err)
	}
	return parseSCPIPressure(resp)
}

func (d *ConST860Driver) ReadUnit(ctx context.Context) (string, error) {
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := d.base.sendSCPICommand(ctx, "PRESsure?", 3*time.Second)
		if err != nil {
			return "", fmt.Errorf("read unit: %w", err)
		}
		parts := strings.SplitN(resp, ",", 2)
		if len(parts) >= 2 {
			unit := strings.TrimSpace(parts[1])
			if isValidPressureUnit(unit) {
				return unit, nil
			}
		}
	}
	return "", nil
}

func (d *ConST860Driver) SetUnit(ctx context.Context, unit string) error {
	unitCode, ok := pressureUnitToCode(unit)
	if !ok {
		return fmt.Errorf("unsupported pressure unit: %s", unit)
	}
	cmd := fmt.Sprintf("PRESsure:MODule:UNIT 1,%s", unitCode)
	_, err := d.base.sendSCPICommand(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set unit: %w", err)
	}
	return nil
}

func (d *ConST860Driver) ReadStability(ctx context.Context) (bool, error) {
	resp, err := d.base.sendSCPICommand(ctx, "PRESsure:STABle?", 3*time.Second)
	if err != nil {
		return false, fmt.Errorf("read stability: %w", err)
	}
	return strings.TrimSpace(resp) == "1", nil
}

func (d *ConST860Driver) StartControl(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "PRESsure:MODule:CONTrol CONTROL", 3*time.Second)
	if err != nil {
		return fmt.Errorf("start pressure control: %w", err)
	}
	return nil
}

func (d *ConST860Driver) ReadTargetRange(ctx context.Context) (min, max float64, err error) {
	resp, err := d.base.sendSCPICommand(ctx, "PRESsure:TARGet:RANGe?", 3*time.Second)
	if err != nil {
		return 0, 0, fmt.Errorf("read target range: %w", err)
	}
	return parseTargetRange(resp)
}

// ---------------------------------------------------------------------------
// SPC4000 打压设备驱动 (Mensor兼容命令集)
// ---------------------------------------------------------------------------

type SPC4000Driver struct {
	base *tcpConnectionDriver
}

func newSPC4000Driver(host string, port int) *SPC4000Driver {
	return &SPC4000Driver{base: newTCPConnectionDriver("SPC4000", host, port)}
}

func (d *SPC4000Driver) Connect(ctx context.Context) error    { return d.base.Connect(ctx) }
func (d *SPC4000Driver) Disconnect(ctx context.Context) error { return d.base.Disconnect(ctx) }

func (d *SPC4000Driver) SetTargetPressure(ctx context.Context, target float64) error {
	var cmd string
	if target >= 0 {
		cmd = fmt.Sprintf("GP %.4f", target)
	} else {
		cmd = fmt.Sprintf("GN %.4f", math.Abs(target))
	}
	_, err := d.base.sendSCPICommand(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set target pressure: %w", err)
	}
	return nil
}

func (d *SPC4000Driver) Stop(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "Measure", 3*time.Second)
	if err != nil {
		return fmt.Errorf("stop pressure: %w", err)
	}
	return nil
}

func (d *SPC4000Driver) Exhaust(ctx context.Context) error {
	_, err := d.base.sendSCPICommand(ctx, "Vent", 3*time.Second)
	if err != nil {
		return fmt.Errorf("exhaust: %w", err)
	}
	return nil
}

func (d *SPC4000Driver) ReadCurrentPressure(ctx context.Context) (float64, error) {
	resp, err := d.base.sendSCPICommand(ctx, "RP", 3*time.Second)
	if err != nil {
		return 0, fmt.Errorf("read current pressure: %w", err)
	}
	return parseSCPIPressure(resp)
}

func (d *SPC4000Driver) ReadUnit(ctx context.Context) (string, error) {
	resp, err := d.base.sendSCPICommand(ctx, "Units?", 3*time.Second)
	if err != nil {
		return "", fmt.Errorf("read unit: %w", err)
	}
	return parseSPC4000UnitCode(resp), nil
}

func (d *SPC4000Driver) SetUnit(ctx context.Context, unit string) error {
	code, ok := pressureUnitToCodeSPC4000(unit)
	if !ok {
		return fmt.Errorf("unsupported pressure unit: %s", unit)
	}
	cmd := fmt.Sprintf("Units %s", code)
	_, err := d.base.sendSCPICommand(ctx, cmd, 3*time.Second)
	if err != nil {
		return fmt.Errorf("set unit: %w", err)
	}
	return nil
}

func (d *SPC4000Driver) ReadStability(ctx context.Context) (bool, error) { return true, nil }

func (d *SPC4000Driver) StartControl(ctx context.Context) error { return nil }

func (d *SPC4000Driver) ReadTargetRange(ctx context.Context) (min, max float64, err error) {
	return 0, 0, fmt.Errorf("SPC4000 does not support target range query")
}

// ---------------------------------------------------------------------------
// 辅助函数
// ---------------------------------------------------------------------------

func isValidPressureUnit(unit string) bool {
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "mpa", "kpa", "pa", "bar", "mbar", "psi", "kgf/cm2", "mmhg", "atm", "inhg":
		return true
	default:
		return false
	}
}

func pressureUnitToCode(unit string) (string, bool) {
	m := map[string]string{
		"pa": "1130", "kpa": "1133", "mpa": "1132", "bar": "1105", "mbar": "1104",
		"psi": "1141", "kgf/cm2": "1145", "mmhg": "1134", "atm": "1135",
	}
	code, ok := m[strings.ToLower(strings.TrimSpace(unit))]
	return code, ok
}

func pressureUnitToCode820(unit string) (string, bool) {
	m := map[string]string{"pa": "0", "kpa": "1", "mpa": "2", "psi": "3", "kgf/cm2": "4"}
	code, ok := m[strings.ToLower(strings.TrimSpace(unit))]
	return code, ok
}

func pressureUnitToCodeSPC4000(unit string) (string, bool) {
	m := map[string]string{
		"psi": "1", "atm": "13", "bar": "14", "mbar": "15", "mmhg": "19",
		"kpa": "22", "pa": "23", "kgf/cm2": "26", "mpa": "36",
	}
	code, ok := m[strings.ToLower(strings.TrimSpace(unit))]
	return code, ok
}

func parseSPC4000UnitCode(code string) string {
	m := map[string]string{
		"1": "psi", "13": "atm", "14": "bar", "15": "mbar", "19": "mmHg",
		"22": "kPa", "23": "Pa", "26": "kgf/cm2", "36": "MPa",
	}
	if unit, ok := m[strings.TrimSpace(code)]; ok {
		return unit
	}
	return "MPa"
}

func channelsToBitmap(channels []int) string {
	var bitmap uint16
	for _, ch := range channels {
		if ch >= 1 && ch <= 16 {
			bitmap |= 1 << (ch - 1)
		}
	}
	return fmt.Sprintf("%04X", bitmap)
}

func coefficientToUnit(coef string) string {
	v, err := strconv.ParseFloat(coef, 64)
	if err != nil {
		return coef
	}
	switch {
	case v == 1.0:
		return "psi"
	case approxEqual(v, 0.07031):
		return "kgf/cm2"
	case approxEqual(v, 0.0689476):
		return "bar"
	case approxEqual(v, 68.9476):
		return "mbar"
	case approxEqual(v, 6.89476):
		return "kPa"
	case approxEqual(v, 0.00689476):
		return "MPa"
	case approxEqual(v, 51.7149):
		return "mmHg"
	case approxEqual(v, 0.068046):
		return "atm"
	case approxEqual(v, 6894.76):
		return "Pa"
	default:
		return coef
	}
}

func unitToCoefficient(unit string) (string, bool) {
	coefficients := map[string]float64{
		"psi": 1.0, "kgf/cm2": 0.07031, "bar": 0.0689476, "mbar": 68.9476,
		"kpa": 6.89476, "mpa": 0.00689476, "mmhg": 51.7149, "atm": 0.068046, "pa": 6894.76,
	}
	v, ok := coefficients[strings.ToLower(strings.TrimSpace(unit))]
	if !ok {
		return "", false
	}
	return fmt.Sprintf("%g", v), true
}

func parseSCPIPressure(resp string) (float64, error) {
	resp = strings.TrimSpace(resp)
	parts := strings.SplitN(resp, ",", 2)
	v, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, fmt.Errorf("parse pressure value %q: %w", resp, err)
	}
	return v, nil
}

func parseTargetRange(resp string) (min, max float64, err error) {
	parts := strings.Split(resp, ",")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("invalid target range response: %q", resp)
	}
	min, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse min: %w", err)
	}
	max, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse max: %w", err)
	}
	return min, max, nil
}

func parseWTN1604Unit(response string) (string, bool) {
	val := strings.TrimSpace(response)
	if strings.HasPrefix(val, "A") {
		val = strings.TrimSpace(strings.TrimPrefix(val, "A"))
	}
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return "", false
	}
	unitInt := int(math.Round(v))
	switch unitInt {
	case 0:
		return "kgf/cm2", true
	case 1:
		return "psi", true
	case 6:
		return "kPa", true
	case 6894:
		return "Pa", true
	default:
		return "", false
	}
}

func approxEqual(a, b float64) bool {
	if a == 0 && b == 0 {
		return true
	}
	diff := a - b
	avg := (a + b) / 2
	if avg == 0 {
		return diff == 0
	}
	return (diff/avg) < 0.01 && (diff/avg) > -0.01
}
