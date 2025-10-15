package state

import (
	"fmt"
	"os"
	"path/filepath"
)

const Name = "rph"
var CachePath string

func Setup() {
	var dir, err = os.UserCacheDir()
	if err != nil {
		fmt.Println("Unable to get user cache directory")
		os.Exit(1)
	}
	CachePath = filepath.Join(dir, Name)
}
