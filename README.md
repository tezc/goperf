## Performance counters for Go

Just a wrapper around [perf_event_open](https://man7.org/linux/man-pages/man2/perf_event_open.2.html).

Useful when you want to get performance counters for some code snippet.  
Normally, you use 'perf' itself but it's not always possible to extract  
some piece of code from your project and isolate it for performance counters.

## Notes
- CPUs have limited PMU registers. So, performance counters can be activated at the  
same time are limited.  
- Some performance counters can be scheduled on specific PMU only. So, combination of  
some performance counters may not work or they will be multiplexed. 
- Check out "Measurement Time" section in the report to see if it's multiplexed.  
  (less than %100)
- Not all counters are supported by Linux or your CPU.
- If you are surprised that some counters does not work, search scheduling  
  algorithm of performance counters online.
- This tool will measure all threads of the process, including gc threads.
- This tool is direct translation from C version : https://github.com/tezc/sc/tree/master/perf

## Config
To allow recording kernel events, you may need to run :

```
sudo sh -c 'echo 1 >/proc/sys/kernel/perf_event_paranoid'
```

## Usage :


```go
package main

import (
	"github.com/tezc/goperf"
	"hash/crc32"
	"os"
)

func main() {

	var x uint32 = 0
	data := []byte("hello world!")

	goperf.Start()

	// long running code
	for i := 0; i < 10000000; i++ {
		x += crc32.ChecksumIEEE(data)
	}

	goperf.End()

	os.Exit(int(x))
}

```

Output : 

```
| Event                     | Value              | Measurement time  
---------------------------------------------------------------
| time (seconds)            | 0.22               | (100,00%)  
| cpu-clock                 | 219,754,463.00     | (100.00%)  
| task-clock                | 219,754,469.00     | (100.00%)  
| page-faults               | 4.00               | (100.00%)  
| context-switches          | 0.00               | (100.00%)  
| cpu-migrations            | 0.00               | (100.00%)  
| page-fault-minor          | 4.00               | (100.00%)  
| cpu-cycles                | 911,266,695.00     | (100.00%)  
| instructions              | 2,300,365,430.00   | (100.00%)  
| cache-misses              | 14,812.00          | (100.00%)  
| L1D-read-miss             | 11,385.00          | (100.00%)  
| L1I-read-miss             | 48,424.00          | (100.00%)  
 
```


You can add or disable counters, check Counters table for the full list :  


```go
package main

import (
	"github.com/tezc/goperf"
	"hash/crc32"
	"os"
)

func main() {

	var x uint32 = 0
	data := []byte("hello world!")

	goperf.Disable("page-faults")
	goperf.Disable("context-switches")
	goperf.Enable("L1D-read-miss")
	goperf.Start()

	// long running code
	for i := 0; i < 10000000; i++ {
		x += crc32.ChecksumIEEE(data)
	}

	goperf.End()

	os.Exit(int(x))
}
```

Run multiple time : 

```go
package main

import (
	"github.com/tezc/goperf"
	"hash/crc32"
	"os"
)

func main() {

	var x uint32 = 0
	data := []byte("hello world!")
	
	goperf.Start()
	// long running code
	for i := 0; i < 10000000; i++ {
		x += crc32.ChecksumIEEE(data)
	}
	goperf.End()

	
	
	test := []byte("test!")

	goperf.Start()
	// long running code
	for i := 0; i < 10000000; i++ {
		x += crc32.ChecksumIEEE(test)
	}
	goperf.End()
	

	os.Exit(int(x))
}
```
