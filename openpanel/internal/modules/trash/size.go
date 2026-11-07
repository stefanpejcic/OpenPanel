package trash

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

// SizeBytes is how much the account's trash takes up, ok is false when du failed or ran out of time
func SizeBytes(ctx context.Context, userContext string) (size int64, ok bool) {
	dir := rawTrashBase(userContext)
	if _, err := tenantfs.Stat(dir); os.IsNotExist(err) {
		return 0, true
	}
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	out, err := tenantfs.OutputIn(cctx, "/home/"+userContext, dir, "du", "-sk", ".")
	if err != nil {
		return 0, false
	}
	return parseDuKB(string(out))
}

// parseDuKB reads the total from `du -sk .` output
func parseDuKB(out string) (int64, bool) {
	fields := strings.Fields(out)
	if len(fields) == 0 {
		return 0, false
	}
	kb, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0, false
	}
	return kb * 1024, true
}
