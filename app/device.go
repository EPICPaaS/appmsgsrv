package app

// 定义移动端操作结构.
type device struct{}

// 声明移动端操作接口.
var Device = device{}

const (
	DEVICE_TYPE_IOS     = "iOS"
	DEVICE_TYPE_ANDROID = "android"
)
