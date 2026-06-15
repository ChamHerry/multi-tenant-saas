package automigrate

import (
	"io/fs"
	"regexp"
	"strconv"

	"github.com/gogf/gf/v2/errors/gerror"
)

var migrationFilePattern = regexp.MustCompile(`^([0-9]+)_.+\.up\.sql$`)

func LatestMigrationVersion(fsys fs.FS, dir string) (uint64, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return 0, gerror.Wrapf(err, "read migrations dir %s", dir)
	}
	var latest uint64
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := migrationFilePattern.FindStringSubmatch(entry.Name())
		if len(matches) != 2 {
			continue
		}
		version, err := strconv.ParseUint(matches[1], 10, 64)
		if err != nil {
			return 0, gerror.Wrapf(err, "parse migration version %s", entry.Name())
		}
		if version > latest {
			latest = version
		}
	}
	if latest == 0 {
		return 0, gerror.Newf("no up migrations found in %s", dir)
	}
	return latest, nil
}
