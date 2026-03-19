package publish

import (
	"fmt"
	"path"
	"time"
)

type RemotePaths struct {
	Staging string
	Dated   string
	Latest  string
}

func BuildRemotePaths(base string, date time.Time) RemotePaths {
	y, m, d := date.Date()
	stamp := date.Format("20060102")
	return RemotePaths{
		Staging: path.Join(base, "_staging", stamp),
		Dated:   path.Join(base, fmt.Sprintf("%04d", y), fmt.Sprintf("%02d", int(m)), fmt.Sprintf("%02d", d)),
		Latest:  path.Join(base, "latest"),
	}
}
