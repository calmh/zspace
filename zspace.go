package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

type ZFS struct {
	Name              string
	Type              string
	Avail             uint64
	Used              uint64
	UsedSnap          uint64
	UsedDS            uint64
	UsedRefReserv     uint64
	LogicalReferenced uint64
}

type Sum struct {
	UsedDS            uint64
	UsedSnap          uint64
	UsedRefReserv     uint64
	Total             uint64
	LogicalReferenced uint64
}

var (
	zfsOptions = []string{"list", "-pHo", "name,type,avail,used,usedsnap,usedds,usedrefreserv,logicalreferenced"}
	lineFmt    = "%-16s  %10s  %10s  %10s  %10s  %8s\n"
	sums       = make(map[string]Sum)
)

func list(host string) []ZFS {
	var res []ZFS
	var cmd *exec.Cmd
	if host == "" {
		cmd = exec.Command("zfs", zfsOptions...)
	} else {
		sshOptions := append([]string{host, "zfs"}, zfsOptions...)
		cmd = exec.Command("ssh", sshOptions...)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	br := bytes.NewBuffer(out)
	line, err := br.ReadString('\n')
	for err == nil {
		var z ZFS
		fmt.Sscanf(line, "%s %s %d %d %d %d %d %d", &z.Name, &z.Type, &z.Avail, &z.Used, &z.UsedSnap, &z.UsedDS, &z.UsedRefReserv, &z.LogicalReferenced)
		res = append(res, z)
		line, err = br.ReadString('\n')
	}
	return res
}

func gb(b uint64) string {
	const GB = 1 << 30
	return fmt.Sprintf("%7.01f GB", float64(b)/GB)
}

func comp(s Sum) string {
	r := float64(s.LogicalReferenced) / float64(s.UsedDS)
	return fmt.Sprintf("%.02fx", r)
}

func add(cat string, z ZFS) {
	s := sums[cat]
	s.UsedDS += z.UsedDS
	s.UsedSnap += z.UsedSnap
	s.UsedRefReserv += z.UsedRefReserv
	s.Total += z.UsedDS + z.UsedSnap + z.UsedRefReserv
	s.LogicalReferenced += z.LogicalReferenced
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
	host := flag.String("h", "", "Host name (default localhost)")
	cf := flag.String("c", "/opt/local/etc/zspace-classes.txt", "Classes file")
	flag.Parse()

	classes := loadFsClasses(*cf)

	l := list(*host)
	fmt.Printf(lineFmt, "CATEGORY", "DATASET", "SNAPSHOT", "REFRES", "TOTAL", "COMPRESS")

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
		fmt.Printf(lineFmt, k, gb(v.UsedDS), gb(v.UsedSnap), gb(v.UsedRefReserv), gb(v.Total), comp(v))
	}

	v := sums["total"]
	fmt.Printf(lineFmt, "TOTAL", gb(v.UsedDS), gb(v.UsedSnap), gb(v.UsedRefReserv), gb(v.Total), comp(v))
}
