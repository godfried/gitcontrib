package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
)

type statistics struct {
	commits int
	filesChanged int
	insertions int
	deletions int
	name string
	email string
}

func (s *statistics) String() string  {
	return fmt.Sprintf("Name: %s\nEmail: %s\nCommits: %d\nFiles Changed: %d\nInsertions: %d\nDeletions: %d\nDelta: %d\n", s.name, s.email, s.commits, s.filesChanged, s.insertions, s.deletions, s.insertions-s.deletions)
}

type statType string

const (
	statFilesChanged = statType("fileschanged")
	statInsertions = statType("insertions")
	statDeletions= statType("deletions")
	statCommits= statType("commits")
)

type StatType struct {
	statType
}

func (s *StatType) Set(val string) error {
	st := statType(strings.ToLower(val))
	switch st {
	case statFilesChanged,statInsertions, statDeletions, statCommits:
		s.statType = st
	default:
		return fmt.Errorf("unknown stat type '%s'", val)
	}
	return nil
}

func (s *StatType) String() string {
	return string(s.statType)
}

func main() {
	var repoPath string
	var stat StatType
	flag.StringVar(&repoPath, "repo", "", "path of repo to pull statistics for.")
	flag.Var(&stat, "stat", "stat to plot.")
	flag.Parse()
	c := exec.Command("git", "-C", repoPath, "log", "--no-merges", "--shortstat", "--format=name:<%an>,email:<%ae>")
	fmt.Println(strings.Join(c.Args, " "))
	so, err := c.StdoutPipe()
	if err != nil {
		panic(err)
	}
	if err := c.Start(); err != nil {
		panic(err)
	}
	r := bufio.NewReader(so)
	scanner := bufio.NewScanner(r)
	results := make(map[string]*statistics, 1000)
	var currentStat *statistics
	var ok bool
	for scanner.Scan() {
		data := bytes.TrimSpace(scanner.Bytes())
		if len(data) == 0 {
			continue
		}
		if bytes.HasPrefix(data, []byte("name")) {
			split := bytes.Split(data, []byte(","))
			authorName := string(bytes.Split(split[0], []byte(":"))[1])
			authorEmail := string(bytes.Split(split[1], []byte(":"))[1])
			currentStat, ok = results[authorEmail]
			if !ok {
				currentStat = &statistics{
					name:authorName,
					email:authorEmail,
				}
				results[authorEmail] = currentStat
			}
			currentStat.commits ++
		} else if bytes.Contains(data, []byte("files changed")){
			err = currentStat.read(data)
			if err != nil {
				panic(err)
			}
		}
	}
	err = c.Wait()
	if err != nil {
		panic(err)
	}
	err = plot(results, stat)
	if err != nil {
		panic(err)
	}
}

func (s *statistics) read( data []byte) error  {
	const (
		filesChanged = "files changed"
		insertions = "insertions(+)"
		deletions = "deletions(-)"
	)
	split := bytes.Split(data, []byte(","))
	for _, sp := range split {
		vals := bytes.SplitN(bytes.TrimSpace(sp), []byte(" "), 2)
		tipe := string(bytes.TrimSpace(vals[1]))
		conv, err := strconv.Atoi(string(vals[0]))
		if err != nil {
			return err
		}
		switch tipe {
		case filesChanged:
			s.filesChanged += conv
		case insertions:
			s.insertions += conv
		case deletions:
			s.deletions += conv
		}
	}
	return nil
}

func (s *statistics) getStat(st StatType) int {
	switch st.statType {
	case statFilesChanged:
		return s.filesChanged
	case statInsertions:
		return s.insertions
	case statDeletions:
		return s.deletions
	case statCommits:
		return s.commits
	default:
		panic("invalid stat type: "+ st.String())
	}
}

func plot(stats map[string]*statistics, stat StatType) error {
	dataFile, err := ioutil.TempFile("", "gitcontrib-data-*.csv")
	if err != nil {
		return err
	}
	defer dataFile.Close()
	for _, s := range stats {
		_, err := dataFile.WriteString(fmt.Sprintf("%s,%d\n", s.name, s.getStat(stat)))
		if err != nil {
			return err
		}
	}
	if err = dataFile.Sync(); err != nil {
		return err
	}
	if err = dataFile.Close(); err != nil {
		return err
	}
	plotFile, err := ioutil.TempFile("", "gitcontrib-pie-*.gnuplot")
	if err != nil {
		return err
	}
	defer plotFile.Close()
	err = template.Must(template.New("gnuplot").Parse(gnuplotScript)).Execute(plotFile, dataFile.Name())
	if err != nil {
		return err
	}
	if err = plotFile.Sync(); err != nil {
		return err
	}
	if err = plotFile.Close(); err != nil {
		return err
	}
	c := exec.Command("gnuplot", "-d", plotFile.Name())
	return c.Run()
}

var gnuplotScript = `
filename = '{{.}}'
set terminal dumb
rowi = 1
rowf = 7

# obtain sum(column(2)) from rows 'rowi' to 'rowf'
set datafile separator ','
stats filename u 2 every ::rowi::rowf noout prefix "A"

# rowf should not be greater than length of file
rowf = (rowf-rowi > A_records - 1 ? A_records + rowi - 1 : rowf)

angle(x)=x*360/A_sum
percentage(x)=x*100/A_sum

# circumference dimensions for pie-chart
centerX=0
centerY=0
radius=1

# label positions
yposmin = 0.0
yposmax = 0.95*radius
xpos = 1.5*radius
ypos(i) = yposmax - i*(yposmax-yposmin)/(1.0*rowf-rowi)

#-------------------------------------------------------------------
# now we can configure the canvas
set style fill solid 1     # filled pie-chart
unset key                  # no automatic labels
unset tics                 # remove tics
unset border               # remove borders; if some label is missing, comment to see what is happening

set size ratio -1              # equal scale length
set xrange [-radius:2*radius]  # [-1:2] leaves space for labels
set yrange [-radius:radius]    # [-1:1]

#-------------------------------------------------------------------
pos = 0             # init angle
colour = 0          # init colour

# 1st line: plot pie-chart
# 2nd line: draw colored boxes at (xpos):(ypos)
# 3rd line: place labels at (xpos+offset):(ypos)
plot filename u (centerX):(centerY):(radius):(pos):(pos=pos+angle($2)):(colour=colour+1) every ::rowi::rowf w circle lc var,\
     for [i=0:rowf-rowi] '+' u (xpos):(ypos(i)) w p pt 5 ps 4 lc i+1,\
     for [i=0:rowf-rowi] filename u (xpos):(ypos(i)):(sprintf('%05.2f%% %s', percentage($2), stringcolumn(1))) every ::i+rowi::i+rowi w labels left offset 3,0
`