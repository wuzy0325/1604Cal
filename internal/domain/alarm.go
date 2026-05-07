package domain

// AlarmConfig 报警配置，控制精度超限的判定与响应行为。
// 标定流程使用 PrecisionThreshold（满量程百分比），计量流程使用 Threshold + IsRelative。
type AlarmConfig struct {
	Enabled            bool    `json:"enabled"`
	PrecisionThreshold float64 `json:"precisionThreshold"` // 标定用：满量程百分比
	Threshold          float64 `json:"threshold"`          // 计量用：绝对/相对阈值
	IsRelative         bool    `json:"isRelative"`          // 计量用：是否相对目标值
	SoundEnabled       bool    `json:"soundEnabled"`
	ConfirmOnAlarm     bool    `json:"confirmOnAlarm"`
	EnabledChannels    []int   `json:"enabledChannels"`
}
