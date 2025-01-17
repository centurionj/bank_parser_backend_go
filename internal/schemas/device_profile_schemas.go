package schemas

type DeviceProfile struct {
	UserAgent string
	Platform  string
	Screen    struct {
		Width  int
		Height int
	}
}
