package exec

import (
	"fmt"
	"os/exec"
	"runtime"
)

var ErrUnsupportedPlatform = fmt.Errorf("exec: unsupported platform")

func Browser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	}
	return ErrUnsupportedPlatform
}
