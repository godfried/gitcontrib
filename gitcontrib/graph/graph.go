package graph

import (
	"flag"
	"fmt"
	"math"
	"os"
	"time"
)


var days = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

const (
	delim   = ","
	tick    = "▇"
	sm_tick = "▏"
)

var (
	fileName, title, format, suffix, colour, customTick, customDelim        string
	width                                                                   int
	noLabels, vertical, stacked, differentScale, calendar, verbose, version bool
	startDate                                                               Date
)

type Date time.Time

func (d *Date) Set(val string) error {
	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return err
	}
	*d = Date(t)
	return nil
}

func (d *Date) String() string {
	return time.Time(*d).String()
}

func init() {
	flag.StringVar(&fileName, "file", "", "data file name (comma or space separated). Defaults to stdin.")
	flag.StringVar(&title, "title", "Graph", "Title of graph.")
	flag.IntVar(&width, "width", 50, "Width of graph.")
	flag.StringVar(&format, "format", "{:<5.2f}", "Format specifier to use.")
	flag.StringVar(&suffix, "suffix", "", "string to add as a suffix to all data points.")
	flag.BoolVar(&noLabels, "no-labels", false, "Do not print the label column.")
	flag.StringVar(&colour, "colour", "red", "Graph bar color(s).")
	flag.BoolVar(&vertical, "vertical", false, "Vertical graph.")
	flag.BoolVar(&stacked, "stacked", false, "Stacked bar graph.")
	flag.BoolVar(&differentScale, "different-scale", false, "Categories have different scales.")
	flag.BoolVar(&calendar, "calendar", false, "Calendar Heatmap chart.")
	flag.Var(&startDate, "start-date", "Start date for Calendar chart.")
	flag.StringVar(&customTick, "custom-tick", "", "Custom tick mark, emoji approved.")
	flag.StringVar(&customDelim, "delim", "", "Custom delimiter, default ',' or ' '.")
	flag.BoolVar(&verbose, "verbose", false, "Verbose output, helpful for debugging.")
	flag.BoolVar(&version, "version", false, "Display version and exit.")
	flag.Parse()

}

func main() {
	if version {
		fmt.Println("termgraph v0.1")
		os.Exit(0)
	}
	_, labels, data, colors := readData()
	if calendar {
		calendarHeatmap(data, labels, args)
	} else {
		chart(colors, data, args, labels)
	}
}

func findMaxLabelLength(labels []string) int {
	length := 0
	for _, l := range labels {
		if len(l) > length {
			length = len(l)
		}
	}
	return length
}

func find(list [][]float64, initial float64, comparer func(x, y float64) float64) float64 {
	for _, subList := range list {
		for _, v := range subList {
			initial = comparer(initial, v)
		}
	}
	return initial
}

func findMin(list [][]float64) float64 {
	return find(list, math.MaxFloat64, math.Min)
}

func findMax(list [][]float64) float64 {
	return find(list, -math.MaxFloat64, math.Max)
}

func normalize(dataLists [][]float64, width float64) [][]float64 {
	minData := findMin(dataLists)
	// We offset by the minimum if there's a negative.
	offData := make([][]float64, 0, len(dataLists))
	if minData < 0 {
		minData = math.Abs(minData)
		for i, list := range dataLists {
			offData = append(offData, make([]float64, 0, len(list)))
			for _, d := range list {
				offData[i] = append(offData[i], d+minData)
			}
		}
	} else {
		offData = dataLists
	}
	minData = findMin(offData)
	maxData := findMax(offData)
	if maxData < width {
		// Don't need to normalize if the max value
		// is less than the width we allow.
		return offData
	}
	// max_dat / width is the value for a single tick. norm_factor is the
	// inverse of this value
	// If you divide a number to the value of single tick, you will find how
	// many ticks it does contain basically.
	normFactor := width / maxData
	normalData := make([][]float64, 0, len(offData))
	for i, list := range offData {
		normalData = append(normalData, make([]float64, 0, len(list)))
		for _, d := range list {
			normalData[i] = append(normalData[i], d*normFactor)
		}
	}
	return normalData

}

type Config struct {
	NoLabels bool
}

type Colour int

var (
	Red =     Colour(91)
	Blue =    Colour(94)
	Green=   Colour(92)
	Magenta= Colour(95)
	Yellow=  Colour(93)
	Black=   Colour(90)
	Cyan=    Colour(96)
}


func horizRows(labels []string, data, normalData [][]float64, config Config, colors []Colour){
// Prepare the horizontal graph.
// Each row is printed through the print_row function.
valMin := findMin(data)

for i, l := range labels{
	if config.NoLabels {
		l = ""
	}else {
		l = l //"{:<{x}}: ".format(labels[i], x = find_max_label_length(labels))
	}
	values := data[i]
	numBlocks := normalData[i]

	for _, j := range values{
		// In Multiple series graph 1st category has label at the beginning,
		// whereas the rest categories have only spaces.
		if j > 0{
			lenLabel := len(l)
			l = Z * len_label
tail = ' {}{}'.format(args['format'].format(values[j]),
args['suffix'])
if colors:
color = colors[j]
else:
color = None

if not args['vertical']:
print(label, end="")

yield(values[j], int(num_blocks[j]), val_min, color)

if not args['vertical']:
print(tail)

# Prints a row of the horizontal graph.
def print_row(value, num_blocks, val_min, color):
"""A method to print a row for a horizontal graphs.
i.e:
1: ▇▇ 2
2: ▇▇▇ 3
3: ▇▇▇▇ 4
"""
if color:
sys.stdout.write(f'\033[{color}m') # Start to write colorized.

if num_blocks < 1 and (value > val_min or value > 0):
# Print something if it's not the smallest
# and the normal value is less than one.
sys.stdout.write(SM_TICK)
else:
for _ in range(num_blocks):
sys.stdout.write(TICK)

if color:
sys.stdout.write('\033[0m') # Back to original.

def stacked_graph(labels, data, normal_data, len_categories, args, colors):
"""Prepare the horizontal stacked graph.
Each row is printed through the print_row function."""
val_min = find_min(data)

for i in range(len(labels)):
if args['no_labels']:
# Hide the labels.
label = ''
else:
label = "{:<{x}}: ".format(labels[i],
x=find_max_label_length(labels))

print(label, end="")
values = data[i]
num_blocks = normal_data[i]

for j in range(len(values)):
print_row(values[j], int(num_blocks[j]), val_min, colors[j])

tail = ' {}{}'.format(args['format'].format(sum(values)),
args['suffix'])
print(tail)
