package log

import (
	"fmt"
	"github.com/cjp2600/assr/config"
	"os"
)

func Printf(format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf(format, a))
}

func Error(err error) {
	fmt.Println(fmt.Sprintf("[ERROR] %s", err.Error()))
}

func Info(text string) {
	fmt.Println(fmt.Sprintf("[INFO] %s", text))
}

func Fatal(text string) {
	fmt.Println(fmt.Sprintf("[FATAL] %s", text))
	os.Exit(0)
}

func Debug(text string) {
	if config.IsDebug() {
		fmt.Println(fmt.Sprintf("[DEBUG] %s", text))
	}
}
