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
	pL1D  = unix.PERF_COUNT_HW_CACHE_L1D
	pL1I  = unix.PERF_COUNT_HW_CACHE_L1I
	pLL   = unix.PERF_COUNT_HW_CACHE_LL
	pDTLB = unix.PERF_COUNT_HW_CACHE_DTLB
	pITLB = unix.PERF_COUNT_HW_CACHE_ITLB
	pBPU  = unix.PERF_COUNT_HW_CACHE_BPU
	pNODE = unix.PERF_COUNT_HW_CACHE_NODE
)

const (
	pREAD     = unix.PERF_COUNT_HW_CACHE_OP_READ
	pWRITE    = unix.PERF_COUNT_HW_CACHE_OP_WRITE
	pPREFETCH = unix.PERF_COUNT_HW_CACHE_OP_PREFETCH
)

const (
	pACCESS = unix.PERF_COUNT_HW_CACHE_RESULT_ACCESS
	pMISS   = unix.PERF_COUNT_HW_CACHE_RESULT_MISS
)

type counter struct {
	Name    string
	Type    uint32
	Config  uint64
	Enabled bool
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

var items [len(Counters)]item

func cache(cache uint32, op uint32, result uint32) uint64 {
	return uint64(cache | (op << 16) | (result << 16))
}

var Counters = [...]counter{
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
	{"L1D-read-access", unix.PERF_TYPE_HW_CACHE, cache(pL1D, pREAD, pACCESS), false},
	{"L1D-read-miss", unix.PERF_TYPE_HW_CACHE, cache(pL1D, pREAD, pMISS), true},
	{"L1D-write-access", unix.PERF_TYPE_HW_CACHE, cache(pL1D, pWRITE, pACCESS), false},
	{"L1D-write-miss", unix.PERF_TYPE_HW_CACHE, cache(pL1D, pWRITE, pMISS), false},
	{"L1D-prefetch-access", unix.PERF_TYPE_HW_CACHE, cache(pL1D, pPREFETCH, pACCESS), false},
	{"L1D-prefetch-miss", unix.PERF_TYPE_HW_CACHE, cache(pL1D, pPREFETCH, pMISS), false},
	{"L1I-read-access", unix.PERF_TYPE_HW_CACHE, cache(pL1I, pREAD, pACCESS), false},
	{"L1I-read-miss", unix.PERF_TYPE_HW_CACHE, cache(pL1I, pREAD, pMISS), true},
	{"L1I-write-access", unix.PERF_TYPE_HW_CACHE, cache(pL1I, pWRITE, pACCESS), false},
	{"L1I-write-miss", unix.PERF_TYPE_HW_CACHE, cache(pL1I, pWRITE, pMISS), false},
	{"L1I-prefetch-access", unix.PERF_TYPE_HW_CACHE, cache(pL1I, pPREFETCH, pACCESS), false},
	{"L1I-prefetch-miss", unix.PERF_TYPE_HW_CACHE, cache(pL1I, pPREFETCH, pMISS), false},
	{"LL-read-access", unix.PERF_TYPE_HW_CACHE, cache(pLL, pREAD, pACCESS), false},
	{"LL-read-miss", unix.PERF_TYPE_HW_CACHE, cache(pLL, pREAD, pMISS), false},
	{"LL-write-access", unix.PERF_TYPE_HW_CACHE, cache(pLL, pWRITE, pACCESS), false},
	{"LL-write-miss", unix.PERF_TYPE_HW_CACHE, cache(pLL, pWRITE, pMISS), false},
	{"LL-prefetch-access", unix.PERF_TYPE_HW_CACHE, cache(pLL, pPREFETCH, pACCESS), false},
	{"LL-prefetch-miss", unix.PERF_TYPE_HW_CACHE, cache(pLL, pPREFETCH, pMISS), false},
	{"DTLB-read-access", unix.PERF_TYPE_HW_CACHE, cache(pDTLB, pREAD, pACCESS), false},
	{"DTLB-read-miss", unix.PERF_TYPE_HW_CACHE, cache(pDTLB, pREAD, pMISS), false},
	{"DTLB-write-access", unix.PERF_TYPE_HW_CACHE, cache(pDTLB, pWRITE, pACCESS), false},
	{"DTLB-write-miss", unix.PERF_TYPE_HW_CACHE, cache(pDTLB, pWRITE, pMISS), false},
	{"DTLB-prefetch-access", unix.PERF_TYPE_HW_CACHE, cache(pDTLB, pPREFETCH, pACCESS), false},
	{"DTLB-prefetch-miss", unix.PERF_TYPE_HW_CACHE, cache(pDTLB, pPREFETCH, pMISS), false},
	{"ITLB-read-access", unix.PERF_TYPE_HW_CACHE, cache(pITLB, pREAD, pACCESS), false},
	{"ITLB-read-miss", unix.PERF_TYPE_HW_CACHE, cache(pITLB, pREAD, pMISS), false},
	{"ITLB-write-access", unix.PERF_TYPE_HW_CACHE, cache(pITLB, pWRITE, pACCESS), false},
	{"ITLB-write-miss", unix.PERF_TYPE_HW_CACHE, cache(pITLB, pWRITE, pMISS), false},
	{"ITLB-prefetch-access", unix.PERF_TYPE_HW_CACHE, cache(pITLB, pPREFETCH, pACCESS), false},
	{"ITLB-prefetch-miss", unix.PERF_TYPE_HW_CACHE, cache(pITLB, pPREFETCH, pMISS), false},
	{"BPU-read-access", unix.PERF_TYPE_HW_CACHE, cache(pBPU, pREAD, pACCESS), false},
	{"BPU-read-miss", unix.PERF_TYPE_HW_CACHE, cache(pBPU, pREAD, pMISS), false},
	{"BPU-write-access", unix.PERF_TYPE_HW_CACHE, cache(pBPU, pWRITE, pACCESS), false},
	{"BPU-write-miss", unix.PERF_TYPE_HW_CACHE, cache(pBPU, pWRITE, pMISS), false},
	{"BPU-prefetch-access", unix.PERF_TYPE_HW_CACHE, cache(pBPU, pPREFETCH, pACCESS), false},
	{"BPU-prefetch-miss", unix.PERF_TYPE_HW_CACHE, cache(pBPU, pPREFETCH, pMISS), false},
	{"NODE-read-access", unix.PERF_TYPE_HW_CACHE, cache(pNODE, pREAD, pACCESS), false},
	{"NODE-read-miss", unix.PERF_TYPE_HW_CACHE, cache(pNODE, pREAD, pMISS), false},
	{"NODE-write-access", unix.PERF_TYPE_HW_CACHE, cache(pNODE, pWRITE, pACCESS), false},
	{"NODE-write-miss", unix.PERF_TYPE_HW_CACHE, cache(pNODE, pWRITE, pMISS), false},
	{"NODE-prefetch-access", unix.PERF_TYPE_HW_CACHE, cache(pNODE, pPREFETCH, pACCESS), false},
	{"NODE-prefetch-miss", unix.PERF_TYPE_HW_CACHE, cache(pNODE, pPREFETCH, pMISS), false},
}

func clear() {
	total = 0
	running = false
	initialized = false

	for i := range items {
		items[i].event = Counters[i]
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
		if !Counters[i].Enabled {
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

// Enable a counter by its name, check Counters array for the list
// Hardware counters are limited on your CPU (~7 these days).
// Some performance counters cannot be enabled at the same time.
// Unsupported counters (either by OS or your CPU) will fail.
// Some counters can be scheduled at the same PMU on the CPU, so
// they will be multiplexed. You can check measurement time in the output
// to see this is the case.
func Enable(counter string) {
	for i := range Counters {
		if Counters[i].Name == counter {
			Counters[i].Enabled = true
			return
		}
	}

	panic(fmt.Errorf("unknown counter : %s", counter))
}

// Disable a counter by its name, check Counters array for the list
func Disable(counter string) {
	for i := range Counters {
		if Counters[i].Name == counter {
			Counters[i].Enabled = false
			return
		}
	}

	panic(fmt.Errorf("unknown counter : %s", counter))
}

func Start() {
	if running {
		panic("Already started")
	}

	if !initialized {
		clear()
		set()
		initialized = true
	}

	err := unix.Prctl(unix.PR_TASK_PERF_EVENTS_ENABLE, 0, 0, 0, 0)
	if err != nil {
		panic(fmt.Errorf("prctl : %v", err))
	}

	start = time.Now()
	running = true
}

func Pause() {
	if !initialized {
		panic("call Start() first")
	}

	if !running {
		return
	}

	err := unix.Prctl(unix.PR_TASK_PERF_EVENTS_DISABLE, 0, 0, 0, 0)
	if err != nil {
		panic(fmt.Errorf("prctl : %v", err))
	}

	total = total + time.Now().Sub(start)
	running = false
}

func End() {
	if !initialized {
		panic("call Start() first")
	}

	Pause()
	readCounters()

	for i := range items {
		if Counters[i].Enabled {
			_ = unix.Close(items[i].fd)
		}
	}

	p := message.NewPrinter(language.English)

	p.Printf("\n| %-25s | %-18s | %s  \n", "Event", "Value", "Measurement time")
	p.Printf("---------------------------------------------------------------\n")
	p.Printf("| %-25s | %-18.2f | %s  \n", "time (seconds)", float64(total)/1e9, "(100,00%)")

	for i := range items {
		if Counters[i].Enabled {
			p.Printf("| %-25s | %-18.2f | (%.2f%%)  \n", items[i].event.Name,
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

		if !Counters[i].Enabled {
			continue
		}

		rd, err := unix.Read(items[i].fd, p)
		if err != nil {
			panic(fmt.Errorf("failed to read counters : %v", err))
		}

		if rd != int(unsafe.Sizeof(f)) {
			panic(fmt.Errorf("read less than expected"))
		}

		f = *((*readFormat)(unsafe.Pointer(&p[0])))

		if f.timeEnabled > 0 && f.timeRunning > 0 {
			n = float64(f.timeRunning / f.timeEnabled)
			items[i].active = n
		}

		items[i].value = items[i].value + float64(f.value)*n
	}
}
