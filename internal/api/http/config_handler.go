package http

import (
	"net/http"
	"time"
)

type deviceConnectConfigPayload struct {
	ConnectAttemptTimeoutMs    int `json:"connectAttemptTimeoutMs"`
	ConnectMaxAttempts         int `json:"connectMaxAttempts"`
	ConnectInitialBackoffMs    int `json:"connectInitialBackoffMs"`
	ConnectMaxBackoffMs        int `json:"connectMaxBackoffMs"`
	DisconnectAttemptTimeoutMs int `json:"disconnectAttemptTimeoutMs"`
	DisconnectMaxAttempts      int `json:"disconnectMaxAttempts"`
	DisconnectInitialBackoffMs int `json:"disconnectInitialBackoffMs"`
	DisconnectMaxBackoffMs     int `json:"disconnectMaxBackoffMs"`
}

// deviceConnectConfigHandler 返回当前生效的连接可靠性配置。
// 该接口用于前端设备面板可视化 timeout/retry 策略，便于现场排障与参数核对。
func (s *apiServer) deviceConnectConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	payload := deviceConnectConfigPayload{
		ConnectAttemptTimeoutMs:    durationToMilliseconds(s.connectConfig.ConnectAttemptTimeout),
		ConnectMaxAttempts:         s.connectConfig.ConnectMaxAttempts,
		ConnectInitialBackoffMs:    durationToMilliseconds(s.connectConfig.ConnectInitialBackoff),
		ConnectMaxBackoffMs:        durationToMilliseconds(s.connectConfig.ConnectMaxBackoff),
		DisconnectAttemptTimeoutMs: durationToMilliseconds(s.connectConfig.DisconnectAttemptTimeout),
		DisconnectMaxAttempts:      s.connectConfig.DisconnectMaxAttempts,
		DisconnectInitialBackoffMs: durationToMilliseconds(s.connectConfig.DisconnectInitialBackoff),
		DisconnectMaxBackoffMs:     durationToMilliseconds(s.connectConfig.DisconnectMaxBackoff),
	}

	writeSuccess(w, http.StatusOK, payload)
}

func durationToMilliseconds(value time.Duration) int {
	if value <= 0 {
		return 0
	}
	return int(value / time.Millisecond)
}
