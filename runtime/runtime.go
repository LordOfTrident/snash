package runtime

import (
	_ "embed"
	"os"
)

//go:embed rc.snash
var RC string

var (
	RCFile = "rc.snash"
)

func WriteFile(path, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(content)

	return nil
}
