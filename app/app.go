package app

import (
	"github.com/b3log/wide/log"
	"os"
)

// 定义应用端操作结构.
type app struct{}

// 声明应用端操作接口.
var App = app{}

var logger = log.NewLogger(os.Stdout)
