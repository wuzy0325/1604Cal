package domain

// AlarmConfig 报警配置，控制校准过程中精度超限的判定与响应行为。
type AlarmConfig struct {
	Enabled            bool    `json:"enabled"`            // 是否启用报警
	PrecisionThreshold float64 `json:"precisionThreshold"` // 精度阈值百分比（如 5.0 表示 5%）
	SoundEnabled       bool    `json:"soundEnabled"`       // 报警时是否播放声音
	ConfirmOnAlarm     bool    `json:"confirmOnAlarm"`     // 报警时是否需要用户确认才能继续
	EnabledChannels    []int   `json:"enabledChannels"`    // 参与报警判定的通道索引列表
}
