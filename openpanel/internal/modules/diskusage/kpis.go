package diskusage

import (
	"strconv"
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/modules/dashboard"
)

// UsageTile is a used/limit KPI, Limit is "" when there's no limit
type UsageTile struct {
	Value, Unit string
	Limit       string
	Pct         int
	FromListing bool // disk total added up from the folder list because no quota report exists
}

// DiskKPIs is the tile row above the folder list, nil tiles are left out
type DiskKPIs struct {
	Disk          *UsageTile
	Inodes        *UsageTile
	LargestFolder *Row
	LargestSize   string
	Trash         *UsageTile
	TrashBytes    int64
	TrashHref     string
}

func splitSize(b int64) (value, unit string) {
	value, unit, _ = strings.Cut(humanBytes(b), " ")
	return value, unit
}

// groupThousands writes 312480 as 312,480
func groupThousands(n int64) string {
	s := strconv.FormatInt(n, 10)
	if n < 0 {
		return "-" + groupThousands(-n)
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}

func pctOf(used, limit int64) int {
	if limit <= 0 {
		return 0
	}
	return int(min(100, used*100/limit))
}

// buildDiskKPIs fills the tiles from the dashboard's quota reading, the current folder list and the trash size
func buildDiskKPIs(quota *dashboard.DiskInodes, rows []Row, atRoot bool, trashBytes int64, trashOK bool) DiskKPIs {
	var k DiskKPIs
	if quota != nil {
		used, limit := quota.DiskUsed*1024, quota.DiskSoft*1024
		t := &UsageTile{Pct: pctOf(used, limit)}
		t.Value, t.Unit = splitSize(used)
		if limit > 0 {
			t.Limit = humanBytes(limit)
		}
		k.Disk = t
		in := &UsageTile{Value: groupThousands(quota.InodesUsed), Pct: pctOf(quota.InodesUsed, quota.InodesSoft)}
		if quota.InodesSoft > 0 {
			in.Limit = groupThousands(quota.InodesSoft)
		}
		k.Inodes = in
	} else if atRoot && len(rows) > 0 {
		var total float64
		for _, r := range rows {
			total += parseHumanSize(r.Count)
		}
		t := &UsageTile{FromListing: true}
		t.Value, t.Unit = splitSize(int64(total))
		k.Disk = t
	}
	// rows come sorted biggest first
	if len(rows) > 0 {
		k.LargestFolder = &rows[0]
		k.LargestSize = humanBytes(int64(parseHumanSize(rows[0].Count)))
	}
	if trashOK {
		t := &UsageTile{}
		t.Value, t.Unit = splitSize(trashBytes)
		k.Trash, k.TrashBytes = t, trashBytes
	}
	return k
}
