package driver

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// ConST 系列公共基类（提取 811A/820/860 重复逻辑）
// ---------------------------------------------------------------------------

// constBaseDriver 封装 ConST 系列打压设备的公共行为。
// 各子型号仅通过 SCPI 命令差异（stableCmd / pressureCmd / unitCmd 等）来区分。
type constBaseDriver struct {
	base *tcpConnectionDriver
}

// constConnect 公共连接逻辑：先建立 TCP，再轮询稳定状态直到设备就绪。
func (d *constBaseDriver) constConnect(ctx context.Context, stableCmd string) error {
	if err := d.base.Connect(ctx); err != nil {
		return err
	}
	for i := 0; i < 10; i++ {
		resp, err := d.base.sendSCPICommand(ctx, stableCmd, 2*time.Second)
		if err == nil && (resp == "0" || resp == "1") {
			break
		}
	}
	return nil
}

// constDisconnect 公共断开逻辑。
func (d *constBaseDriver) constDisconnect(ctx context.Context) error {
	return d.base.Disconnect(ctx)
}

// constReadStability 公共读取稳定状态逻辑。
func (d *constBaseDriver) constReadStability(ctx context.Context, stableCmd string) (bool, error) {
	resp, err := d.base.sendSCPICommand(ctx, stableCmd, 3*time.Second)
	if err != nil {
		return false, fmt.Errorf("read stability: %w", err)
	}
	return strings.TrimSpace(resp) == "1", nil
}

// constReadTargetRange 公共读取目标范围逻辑。
func (d *constBaseDriver) constReadTargetRange(ctx context.Context) (min, max float64, err error) {
	resp, err := d.base.sendSCPICommand(ctx, "PRESsure:TARGet:RANGe?", 3*time.Second)
	if err != nil {
		return 0, 0, fmt.Errorf("read target range: %w", err)
	}
	return parseTargetRange(resp)
}
