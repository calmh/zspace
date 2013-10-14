package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
)

type ZFS struct {
	Name          string
	Type          string
	Avail         int64
	Used          int64
	UsedSnap      int64
	UsedDS        int64
	UsedRefReserv int64
}

func List() []ZFS {
	var res []ZFS
	cmd := exec.Command("zfs", "list", "-pHo", "name,type,avail,used,usedsnap,usedds,usedrefreserv,usedchild")
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	br := bytes.NewBuffer(out)
	line, err := br.ReadString('\n')
	for err == nil {
		var z ZFS
		fmt.Sscanf(line, "%s %s %d %d %d %d %d", &z.Name, &z.Type, &z.Avail, &z.Used, &z.UsedSnap, &z.UsedDS, &z.UsedRefReserv)
		res = append(res, z)
		line, err = br.ReadString('\n')
	}
	return res
}

const (
	_ = 1 << (10 * iota)
	KB
	MB
	GB
	TB
)

func byteFmt(b int64) string {
	switch {
	case b >= TB:
		return fmt.Sprintf("%5.02f TB", float64(b)/TB)
	case b >= GB:
		return fmt.Sprintf("%5.01f GB", float64(b)/GB)
	case b >= MB:
		return fmt.Sprintf("%5.01f MB", float64(b)/MB)
	}
	return fmt.Sprintf("%5d KB", b/KB)
}

type Sum struct {
	UsedDS        int64
	UsedSnap      int64
	UsedRefReserv int64
	Total         int64
}

var sums = make(map[string]Sum)

func add(cat string, z ZFS) {
	s := sums[cat]
	s.UsedDS += z.UsedDS
	s.UsedSnap += z.UsedSnap
	s.UsedRefReserv += z.UsedRefReserv
	s.Total += z.UsedDS + z.UsedSnap + z.UsedRefReserv
	sums[cat] = s
}

func loadFsClasses(file string) map[string]*regexp.Regexp {
	fd, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	classes := make(map[string]*regexp.Regexp)
	br := bufio.NewReader(fd)
	line, err := br.ReadString('\n')
	for err == nil {
		line = strings.TrimSpace(line)
		fs := strings.SplitN(line, " ", 2)
		classes[fs[0]] = regexp.MustCompile(fs[1])
		line, err = br.ReadString('\n')
	}
	return classes
}

func main() {
	classes := loadFsClasses("/opt/local/etc/zspace-classes.txt")

	tw := tabwriter.NewWriter(os.Stdout, 4, 4, 2, ' ', 0)
	tw.Write([]byte("CATEGORY\tDATASET\tSNAPSHOT\tRESERVED\tTOTAL\n"))

	l := List()
loop:
	for _, z := range l {
		add("total", z)
		add("type/"+z.Type, z)
		for c, r := range classes {
			if r.MatchString(z.Name) {
				add("class/"+c, z)
				continue loop
			}
		}
		add("class/other", z)
	}

	var keys []string
	for k := range sums {
		if strings.Contains(k, "/") {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)

	for _, k := range keys {
		v := sums[k]
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", k, byteFmt(v.UsedDS), byteFmt(v.UsedSnap), byteFmt(v.UsedRefReserv), byteFmt(v.Total))
	}

	v := sums["total"]
	fmt.Fprintf(tw, "TOTAL\t%s\t%s\t%s\t%s\n", byteFmt(v.UsedDS), byteFmt(v.UsedSnap), byteFmt(v.UsedRefReserv), byteFmt(v.Total))

	tw.Flush()
}
