package domain

// WorkflowConfig 标定和计量模块的共享工作流配置。
type WorkflowConfig struct {
	Channels          []int     `json:"channels"`
	MinPressure       float64   `json:"minPressure"`
	MaxPressure       float64   `json:"maxPressure"`
	PointCount        int       `json:"pointCount"`
	Precision         int       `json:"precision"`
	AverageCount      int       `json:"averageCount"`
	PrecisionLevel    float64   `json:"precisionLevel"`
	StableWaitMs      int       `json:"stableWaitMs"`
	StabilityTimeoutMs int      `json:"stabilityTimeoutMs"`
	ControlMode       string    `json:"controlMode"`
	PressureMode      string    `json:"pressureMode"`
	CustomPoints      []float64 `json:"customPoints"`
}
