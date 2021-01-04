package goperf

import (
	"fmt"
	"golang.org/x/sys/unix"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"time"
	"unsafe"
)

const (
	L1D  = unix.PERF_COUNT_HW_CACHE_L1D
	L1I  = unix.PERF_COUNT_HW_CACHE_L1I
	LL   = unix.PERF_COUNT_HW_CACHE_LL
	DTLB = unix.PERF_COUNT_HW_CACHE_DTLB
	ITLB = unix.PERF_COUNT_HW_CACHE_ITLB
	BPU  = unix.PERF_COUNT_HW_CACHE_BPU
	NODE = unix.PERF_COUNT_HW_CACHE_NODE
)

const (
	READ     = unix.PERF_COUNT_HW_CACHE_OP_READ
	WRITE    = unix.PERF_COUNT_HW_CACHE_OP_WRITE
	PREFETCH = unix.PERF_COUNT_HW_CACHE_OP_PREFETCH
)

const (
	ACCESS = unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS
	MISS   = unix.PERF_COUNT_HW_CACHE_RESULT_MISS
)

type counter struct {
	Name    string
	Type    uint32
	Config  uint64
	Enabled bool
}

func cache(cache uint32, op uint32, result uint32) uint64 {
	return uint64(cache | (op << 16) | (result << 16))
}

var counters = [...]counter{
	{"cpu-clock", unix.PERF_TYPE_SOFTWARE, unix.PERF_COUNT_SW_CPU_CLOCK, true},
	{"task-clock", unix.PERF_TYPE_SOFTWARE, unix.PERF_COUNT_SW_TASK_CLOCK, true},
	{"page-faults", unix.PERF_TYPE_SOFTWARE, unix.PERF_COUNT_SW_PAGE_FAULTS, true},
	{"context-switches", unix.PERF_TYPE_SOFTWARE, unix.PERF_COUNT_SW_CONTEXT_SWITCHES, true},
	{"cpu-migrations", unix.PERF_TYPE_SOFTWARE, unix.PERF_COUNT_SW_CPU_MIGRATIONS, true},
	{"page-fault-minor", unix.PERF_TYPE_SOFTWARE, unix.PERF_COUNT_SW_PAGE_FAULTS_MIN, true},
	{"page-fault-major", unix.PERF_TYPE_SOFTWARE, unix.PERF_COUNT_SW_PAGE_FAULTS_MAJ, false},
	{"alignment-faults", unix.PERF_TYPE_SOFTWARE, unix.PERF_COUNT_SW_ALIGNMENT_FAULTS, false},
	{"emulation-faults", unix.PERF_TYPE_SOFTWARE, unix.PERF_COUNT_SW_EMULATION_FAULTS, false},
	{"cpu-cycles", unix.PERF_TYPE_HARDWARE, unix.PERF_COUNT_HW_CPU_CYCLES, true},
	{"instructions", unix.PERF_TYPE_HARDWARE, unix.PERF_COUNT_HW_INSTRUCTIONS, true},
	{"cache-references", unix.PERF_TYPE_HARDWARE, unix.PERF_COUNT_HW_CACHE_REFERENCES, false},
	{"cache-misses", unix.PERF_TYPE_HARDWARE, unix.PERF_COUNT_HW_CACHE_MISSES, true},
	{"branch-instructions", unix.PERF_TYPE_HARDWARE, unix.PERF_COUNT_HW_BRANCH_INSTRUCTIONS, false},
	{"branch-misses", unix.PERF_TYPE_HARDWARE, unix.PERF_COUNT_HW_BRANCH_MISSES, false},
	{"bus-cycles", unix.PERF_TYPE_HARDWARE, unix.PERF_COUNT_HW_BUS_CYCLES, false},
	{"stalled-cycles-frontend", unix.PERF_TYPE_HARDWARE, unix.PERF_COUNT_HW_STALLED_CYCLES_FRONTEND, false},
	{"stalled-cycles-backend", unix.PERF_TYPE_HARDWARE, unix.PERF_COUNT_HW_STALLED_CYCLES_BACKEND, false},
	{"ref-cpu-cycles", unix.PERF_TYPE_HARDWARE, unix.PERF_COUNT_HW_REF_CPU_CYCLES, false},
	{"L1D-read-access", unix.PERF_TYPE_HW_CACHE, cache(L1D, READ, ACCESS), false},
	{"L1D-read-miss", unix.PERF_TYPE_HW_CACHE, cache(L1D, READ, MISS), true},
	{"L1D-write-access", unix.PERF_TYPE_HW_CACHE, cache(L1D, WRITE, ACCESS), false},
	{"L1D-write-miss", unix.PERF_TYPE_HW_CACHE, cache(L1D, WRITE, MISS), false},
	{"L1D-prefetch-access", unix.PERF_TYPE_HW_CACHE, cache(L1D, PREFETCH, ACCESS), false},
	{"L1D-prefetch-miss", unix.PERF_TYPE_HW_CACHE, cache(L1D, PREFETCH, MISS), false},
	{"L1I-read-access", unix.PERF_TYPE_HW_CACHE, cache(L1I, READ, ACCESS), false},
	{"L1I-read-miss", unix.PERF_TYPE_HW_CACHE, cache(L1I, READ, MISS), true},
	{"L1I-write-access", unix.PERF_TYPE_HW_CACHE, cache(L1I, WRITE, ACCESS), false},
	{"L1I-write-miss", unix.PERF_TYPE_HW_CACHE, cache(L1I, WRITE, MISS), false},
	{"L1I-prefetch-access", unix.PERF_TYPE_HW_CACHE, cache(L1I, PREFETCH, ACCESS), false},
	{"L1I-prefetch-miss", unix.PERF_TYPE_HW_CACHE, cache(L1I, PREFETCH, MISS), false},
	{"LL-read-access", unix.PERF_TYPE_HW_CACHE, cache(LL, READ, ACCESS), false},
	{"LL-read-miss", unix.PERF_TYPE_HW_CACHE, cache(LL, READ, MISS), false},
	{"LL-write-access", unix.PERF_TYPE_HW_CACHE, cache(LL, WRITE, ACCESS), false},
	{"LL-write-miss", unix.PERF_TYPE_HW_CACHE, cache(LL, WRITE, MISS), false},
	{"LL-prefetch-access", unix.PERF_TYPE_HW_CACHE, cache(LL, PREFETCH, ACCESS), false},
	{"LL-prefetch-miss", unix.PERF_TYPE_HW_CACHE, cache(LL, PREFETCH, MISS), false},
	{"DTLB-read-access", unix.PERF_TYPE_HW_CACHE, cache(DTLB, READ, ACCESS), false},
	{"DTLB-read-miss", unix.PERF_TYPE_HW_CACHE, cache(DTLB, READ, MISS), false},
	{"DTLB-write-access", unix.PERF_TYPE_HW_CACHE, cache(DTLB, WRITE, ACCESS), false},
	{"DTLB-write-miss", unix.PERF_TYPE_HW_CACHE, cache(DTLB, WRITE, MISS), false},
	{"DTLB-prefetch-access", unix.PERF_TYPE_HW_CACHE, cache(DTLB, PREFETCH, ACCESS), false},
	{"DTLB-prefetch-miss", unix.PERF_TYPE_HW_CACHE, cache(DTLB, PREFETCH, MISS), false},
	{"ITLB-read-access", unix.PERF_TYPE_HW_CACHE, cache(ITLB, READ, ACCESS), false},
	{"ITLB-read-miss", unix.PERF_TYPE_HW_CACHE, cache(ITLB, READ, MISS), false},
	{"ITLB-write-access", unix.PERF_TYPE_HW_CACHE, cache(ITLB, WRITE, ACCESS), false},
	{"ITLB-write-miss", unix.PERF_TYPE_HW_CACHE, cache(ITLB, WRITE, MISS), false},
	{"ITLB-prefetch-access", unix.PERF_TYPE_HW_CACHE, cache(ITLB, PREFETCH, ACCESS), false},
	{"ITLB-prefetch-miss", unix.PERF_TYPE_HW_CACHE, cache(ITLB, PREFETCH, MISS), false},
	{"BPU-read-access", unix.PERF_TYPE_HW_CACHE, cache(BPU, READ, ACCESS), false},
	{"BPU-read-miss", unix.PERF_TYPE_HW_CACHE, cache(BPU, READ, MISS), false},
	{"BPU-write-access", unix.PERF_TYPE_HW_CACHE, cache(BPU, WRITE, ACCESS), false},
	{"BPU-write-miss", unix.PERF_TYPE_HW_CACHE, cache(BPU, WRITE, MISS), false},
	{"BPU-prefetch-access", unix.PERF_TYPE_HW_CACHE, cache(BPU, PREFETCH, ACCESS), false},
	{"BPU-prefetch-miss", unix.PERF_TYPE_HW_CACHE, cache(BPU, PREFETCH, MISS), false},
	{"NODE-read-access", unix.PERF_TYPE_HW_CACHE, cache(NODE, READ, ACCESS), false},
	{"NODE-read-miss", unix.PERF_TYPE_HW_CACHE, cache(NODE, READ, MISS), false},
	{"NODE-write-access", unix.PERF_TYPE_HW_CACHE, cache(NODE, WRITE, ACCESS), false},
	{"NODE-write-miss", unix.PERF_TYPE_HW_CACHE, cache(NODE, WRITE, MISS), false},
	{"NODE-prefetch-access", unix.PERF_TYPE_HW_CACHE, cache(NODE, PREFETCH, ACCESS), false},
	{"NODE-prefetch-miss", unix.PERF_TYPE_HW_CACHE, cache(NODE, PREFETCH, MISS), false},
}

var initialized bool
var running bool
var total time.Duration
var start time.Time

type item struct {
	event  counter
	value  float64
	active float64
	fd     int
}

var items [len(counters)]item

func clear() {
	total = 0
	running = false
	initialized = false

	for i := range items {
		items[i].event = counters[i]
		items[i].active = 0
		items[i].value = 0
		items[i].fd = -1
	}
}

type readFormat struct {
	value       uint64
	timeEnabled uint64
	timeRunning uint64
}

func set() {

	flag := unix.PERF_FORMAT_TOTAL_TIME_RUNNING |
		unix.PERF_FORMAT_TOTAL_TIME_ENABLED

	for i := range items {
		if !counters[i].Enabled {
			continue
		}

		t := unix.PerfEventAttr{
			Type:        items[i].event.Type,
			Size:        0,
			Config:      items[i].event.Config,
			Read_format: uint64(flag),
			Bits:        3, // disabled & inherit
		}

		fd, err := unix.PerfEventOpen(&t, 0, -1, -1, unix.PERF_FLAG_FD_CLOEXEC)
		if err != nil {
			str := fmt.Errorf("perfEventOpen : %v", err)
			panic(str)
		}

		items[i].fd = fd
	}

}

func Enable(counter string) {
	for i := range counters {
		if counters[i].Name == counter {
			counters[i].Enabled = true
			return
		}
	}

	str := fmt.Errorf("unknown counter : %s", counter)
	panic(str)
}

func Disable(counter string) {
	for i := range counters {
		if counters[i].Name == counter {
			counters[i].Enabled = false
			return
		}
	}

	str := fmt.Errorf("unknown counter : %s", counter)
	panic(str)
}

func Start() {
	if !initialized {
		clear()
		set()
		initialized = true
	}

	err := unix.Prctl(unix.PR_TASK_PERF_EVENTS_ENABLE, 0, 0, 0, 0)
	if err != nil {
		str := fmt.Errorf("prctl : %v", err)
		panic(str)
	}

	start = time.Now()
	running = true
}

func Pause() {
	if !initialized {
		panic(initialized)
	}

	if !running {
		return
	}

	err := unix.Prctl(unix.PR_TASK_PERF_EVENTS_DISABLE, 0, 0, 0, 0)
	if err != nil {
		str := fmt.Errorf("prctl : %v", err)
		panic(str)
	}

	total = total + time.Now().Sub(start)
	running = false
}

func End() {
	if !initialized {
		panic(initialized)
	}

	Pause()
	readCounters()

	for i := range items {
		if counters[i].Enabled {
			_ = unix.Close(items[i].fd)
		}
	}

	p := message.NewPrinter(language.English)

	_, _ = p.Printf("\n| %-25s | %-18s | %s  \n", "Event", "Value", "Measurement time")
	_, _ = p.Printf("---------------------------------------------------------------\n")
	_, _ = p.Printf("| %-25s | %-18.2f | %s  \n", "time (seconds)", float64(total)/1e9, "(100,00%)")

	for i := range items {
		if counters[i].Enabled {
			_, _ = p.Printf("| %-25s | %-18.2f | (%.2f%%)  \n", items[i].event.Name,
				items[i].value, items[i].active*100)
		}
	}

	clear()
}

func readCounters() {
	var f readFormat

	for i := range items {
		n := 1.0
		p := make([]byte, 64)

		if !counters[i].Enabled {
			continue
		}

		rd, err := unix.Read(items[i].fd, p)
		if err != nil {
			str := fmt.Errorf("failed to read counters : %v", err)
			panic(str)
		}

		if rd != int(unsafe.Sizeof(f)) {
			str := fmt.Errorf("read less than expected")
			panic(str)
		}

		f = *((*readFormat)(unsafe.Pointer(&p[0])))

		if f.timeEnabled > 0 && f.timeRunning > 0 {
			n = float64(f.timeRunning / f.timeEnabled)
			items[i].active = n
		}

		items[i].value = items[i].value + float64(f.value)*n
	}
}
