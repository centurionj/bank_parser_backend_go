package schemas

type AccountIDRequest struct {
	AccountID int `json:"account_id"`
}

type AccountProperties struct {
	HardwareConcurrency int
	DeviceMemory        int
	DeviceProfile       DeviceProfile
	RandomFrequency     int
	RandomStart         string
	RandomStop          string
	BufferSize          int
	InputChannels       int
	OutputChannels      int
	GPU                 string
	CPU                 string
	BatteryVolume       float64
	IsCharging          bool
	PublicIP            string
}
