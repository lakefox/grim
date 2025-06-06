package a

import (
	"grim"
	"os/exec"
	"runtime"
)

func Init() grim.Script {
	return grim.Script{
		Call: func(document *grim.Node) {
			// links := document.QuerySelectorAll("a")

			// for i := range *links {
			// 	v := *links
			// 	v[i].AddEventListener("click", func(e Event) {
			// 		open(v[i].Href)
			// 	})
			// }
		},
	}
}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
