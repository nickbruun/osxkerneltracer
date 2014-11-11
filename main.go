package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	callsExpr = regexp.MustCompile("^(.*?)`(.*?) (\\d+)$")
)

// Trace call.
type TraceCall struct {
	// Module.
	Module string

	// Method.
	Method string

	// Calls.
	Calls uint64
}

// Trace.
type Trace []TraceCall

// Longest trace method name length.
func (t Trace) LongestMethodNameLength() (l int) {
	l = 0

	for _, c := range t {
		cl := len(c.Method)
		if cl > l {
			l = cl
		}
	}

	return
}

// Longest trace module name length.
func (t Trace) LongestModuleNameLength() (l int) {
	l = 0

	for _, c := range t {
		cl := len(c.Module)
		if cl > l {
			l = cl
		}
	}

	return
}

// Maximum number of calls for any call in the trace.
func (t Trace) MaximumCalls() (m uint64) {
	m = 0

	for _, c := range t {
		if c.Calls > m {
			m = c.Calls
		}
	}

	return
}

// Total number of calls in the trace.
func (t Trace) TotalCalls() (s uint64) {
	s = 0

	for _, c := range t {
		s += c.Calls
	}

	return
}

func (t Trace) Len() int           { return len(t) }
func (t Trace) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t Trace) Less(i, j int) bool {
	a, b := &t[i], &t[j]

	if a.Calls > b.Calls {
		return true
	} else if a.Calls == b.Calls {
		if a.Module < b.Module {
			return true
		} else if a.Module == b.Module {
			return a.Method < b.Method
		} else {
			return false
		}
	} else {
		return false
	}
}

func main() {
	// Parse arguments.
	duration := flag.Duration("d", 5 * time.Second, "trace duration")
	flag.Parse()

	// Set up to run dtrace.
	cmd := exec.Command("/usr/sbin/dtrace", "-n", `
#pragma D option quiet
profile:::profile-1001hz
/arg0/
{
    @pc[arg0] = count();
}
dtrace:::END
{
    printa("%a %@d\n", @pc);
}`)

	var outputBuffer bytes.Buffer
	cmd.Stdout = &outputBuffer
	cmd.Stderr = os.Stderr

	// Start running dtrace.
	if err := cmd.Start(); err != nil {
		os.Stderr.Write([]byte("Failed to start kernel trace\n"))
		os.Exit(1)
	}

	os.Stderr.Write([]byte(fmt.Sprintf("Running trace for %v...\n", *duration)))

	// Wait for dtrace to either fail or for the trace period to end.
	timerChan := time.After(*duration)
	errChan := make(chan error)

	go func() {
		errChan <- cmd.Wait()
	}()

	select {
	case <-errChan:
		os.Stderr.Write([]byte("Kernel trace failed\n"))
		os.Exit(1)

	case <-timerChan:
		cmd.Process.Signal(os.Interrupt)
	}

	// Wait for the dtrace process to finish running.
	if err := <- errChan; err != nil {
		os.Stderr.Write([]byte("Kernel trace failed\n"))
		os.Exit(1)
	}

	// Parse the kernel trace.
	outputLines := strings.Split(string(outputBuffer.Bytes()), "\n")
	trace := make(Trace, 0, len(outputLines))

	for _, l := range outputLines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}

		m := callsExpr.FindStringSubmatch(l)
		if m == nil {
			continue
		}

		calls, _ := strconv.ParseUint(m[3], 10, 64)

		trace = append(trace, TraceCall{
			Module: m[1],
			Method: m[2],
			Calls: calls,
		})
	}

	// Output the kernel trace.
	sort.Sort(trace)

	moduleNameTitle := "Module"
	moduleNameWidth := trace.LongestModuleNameLength()
	if moduleNameWidth < len(moduleNameTitle) {
		moduleNameWidth = len(moduleNameTitle)
	}

	methodNameTitle := "Method"
	methodNameWidth := trace.LongestMethodNameLength()
	if methodNameWidth < len(methodNameTitle) {
		methodNameWidth = len(methodNameTitle)
	}

	callsTitle := "Calls"
	maximumCalls := trace.MaximumCalls()
	callsWidth := len(fmt.Sprintf("%d", maximumCalls))
	if callsWidth < len(callsTitle) {
		callsWidth = len(callsTitle)
	}

	shareTitle := "Share"
	shareWidth := 10
	totalCalls := trace.TotalCalls()

	lineFormat := fmt.Sprintf("%%-%ds | %%-%ds | %%%ds | %%%ds\n", moduleNameWidth, methodNameWidth, callsWidth, shareWidth)
	fmt.Printf(lineFormat, moduleNameTitle, methodNameTitle, callsTitle, shareTitle)

	fmt.Printf("%s-+-%s-+-%s-+-%s\n", strings.Repeat("-", moduleNameWidth), strings.Repeat("-", methodNameWidth), strings.Repeat("-", callsWidth), strings.Repeat("-", shareWidth))

	for _, c := range trace {
		fmt.Printf(lineFormat, c.Module, c.Method, fmt.Sprintf("%d", c.Calls), fmt.Sprintf("%.4f %%", float32(c.Calls * 100) / float32(totalCalls)))
	}
}
