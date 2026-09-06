package engine

import (
	"context"
	"errors"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/catnet-io/engine/pkg/coreerr"
	"github.com/catnet-io/engine/pkg/results"
)

func TestScanConcurrency(t *testing.T) {
	ips := []string{"192.0.2.1", "192.0.2.2", "192.0.2.3", "192.0.2.4", "192.0.2.5"}
	cfg := DefaultConfig()
	cfg.PingTimeoutMs = 10 // small timeout for test
	cfg.MaxThreads = 2

	var mu sync.Mutex
	var eventDevices []results.DeviceInfo

	report, err := StartScan(context.Background(), ips, cfg, func(event ScanEvent) {
		if event.Type == EventResult && event.Device != nil {
			mu.Lock()
			eventDevices = append(eventDevices, *event.Device)
			mu.Unlock()
		}
	})

	if err != nil {
		t.Fatalf("StartScan failed: %v", err)
	}

	if report == nil {
		t.Fatalf("Expected report, got nil")
	}

	if len(report.Devices) != len(ips) {
		t.Errorf("Expected %d results in report, got %d", len(ips), len(report.Devices))
	}

	// Verify all ips are present in report
	ipMap := make(map[string]bool)
	for _, r := range report.Devices {
		ipMap[r.IP] = true
	}
	for _, ip := range ips {
		if !ipMap[ip] {
			t.Errorf("Expected IP %s in report", ip)
		}
	}

	if len(eventDevices) != len(ips) {
		t.Errorf("Expected %d results from events, got %d", len(ips), len(eventDevices))
	}
}

func TestScanEventPointerAliasing(t *testing.T) {
	ips := []string{"192.0.2.1", "192.0.2.2", "192.0.2.3"}
	cfg := DefaultConfig()
	cfg.PingTimeoutMs = 1
	cfg.MaxThreads = 1 // Single thread makes loop reuse the same variable if aliasing exists

	var mu sync.Mutex
	var pointers []*results.DeviceInfo

	_, err := StartScan(context.Background(), ips, cfg, func(event ScanEvent) {
		if event.Type == EventResult && event.Device != nil {
			mu.Lock()
			pointers = append(pointers, event.Device)
			mu.Unlock()
		}
	})

	if err != nil {
		t.Fatalf("StartScan failed: %v", err)
	}

	if len(pointers) != len(ips) {
		t.Fatalf("Expected %d results, got %d", len(ips), len(pointers))
	}

	// Verify that each pointer points to a distinct IP, meaning pointers are not aliased
	ipMap := make(map[string]bool)
	for _, ptr := range pointers {
		if ipMap[ptr.IP] {
			t.Errorf("Duplicate IP found, pointer aliasing issue detected: %s", ptr.IP)
		}
		ipMap[ptr.IP] = true
	}
}

func TestScanCancellation(t *testing.T) {
	ips := make([]string, 100)
	for i := 0; i < 100; i++ {
		ips[i] = "192.0.2." + strconv.Itoa(i)
	}
	cfg := DefaultConfig()
	cfg.PingTimeoutMs = 1000
	cfg.MaxThreads = 2

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel almost immediately
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	_, err := StartScan(ctx, ips, cfg, nil)
	if err == nil {
		t.Fatalf("Expected cancellation error, got nil")
	}

	if !errors.Is(err, coreerr.ErrCancelled) {
		t.Errorf("Expected coreerr.ErrCancelled, got %v", err)
	}
}

func TestParallelScans(t *testing.T) {
	ips1 := []string{"192.0.2.1", "192.0.2.2"}
	ips2 := []string{"192.0.2.3", "192.0.2.4"}
	cfg := DefaultConfig()
	cfg.PingTimeoutMs = 10
	cfg.MaxThreads = 2

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, err := StartScan(context.Background(), ips1, cfg, nil)
		if err != nil {
			t.Errorf("Scan 1 failed: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		_, err := StartScan(context.Background(), ips2, cfg, nil)
		if err != nil {
			t.Errorf("Scan 2 failed: %v", err)
		}
	}()

	wg.Wait()
}

func TestScanReportCompleteness(t *testing.T) {
	ips := []string{"192.0.2.1", "192.0.2.2", "192.0.2.3"}
	cfg := DefaultConfig()
	cfg.PingTimeoutMs = 10
	cfg.MaxThreads = 2

	report, err := StartScan(context.Background(), ips, cfg, nil)
	if err != nil {
		t.Fatalf("StartScan failed: %v", err)
	}

	if len(report.Devices) != report.Total {
		t.Errorf("len(Devices)=%d != Total=%d", len(report.Devices), report.Total)
	}
}

func BenchmarkStartScan(b *testing.B) {
	ips := make([]string, 100)
	for i := 0; i < 100; i++ {
		ips[i] = "127.0.0.1" // Localhost to avoid dropping packets
	}
	cfg := DefaultConfig()
	cfg.PingTimeoutMs = 10
	cfg.MaxThreads = 10
	cfg.DefaultPorts = []int{} // No port scanning for baseline

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = StartScan(context.Background(), ips, cfg, nil)
	}
}

func TestNoGoroutineLeakOnCancel(t *testing.T) {
	ips := make([]string, 50)
	for i := range ips {
		ips[i] = "192.0.2." + itoa(i+1)
	}
	cfg := DefaultConfig()
	cfg.PingTimeoutMs = 500
	cfg.MaxThreads = 10
	cfg.DefaultPorts = []int{}

	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	_, _ = StartScan(ctx, ips, cfg, nil)

	for i := 0; i < 50; i++ {
		after := runtime.NumGoroutine()
		if after <= before+3 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	after := runtime.NumGoroutine()
	t.Errorf("Goroutine leak: before=%d after=%d", before, after)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [10]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}

func TestStartScanDefensiveTimeout(t *testing.T) {
	ips := []string{"192.0.2.1"}
	cfg := DefaultConfig()
	cfg.PingTimeoutMs = 100
	cfg.PortTimeoutMs = 200
	// 25 ports with concurrency of 10 means 3 batches.
	// 3 batches * 200ms = 600ms for ports.
	// 100ms ping + 600ms ports = 700ms max per host.
	// 1 host / 1 thread = 700ms total + 1 min buffer.
	cfg.DefaultPorts = make([]int, 25)
	cfg.MaxThreads = 1

	start := time.Now()
	// Pass context without deadline so defensivo timeout applies
	_, err := StartScan(context.Background(), ips, cfg, nil)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("StartScan failed: %v", err)
	}

	// The actual scan will finish almost instantly since 192.0.2.1 doesn't respond and times out.
	// Wait, the defensive timeout calculation shouldn't affect the normal execution time.
	// But we just want to ensure it doesn't crash or create an incorrectly small timeout that fails the scan immediately.
	if duration > 5*time.Second {
		t.Errorf("Scan took too long, might be stuck")
	}
}

func TestSlowCallbackDoesNotStallWorkers(t *testing.T) {
	ips := make([]string, 50)
	for i := range ips {
		ips[i] = "127.0.0.1"
	}

	cfg := DefaultConfig()
	cfg.MaxThreads = 16
	cfg.PingTimeoutMs = 50
	cfg.PortTimeoutMs = 50
	cfg.DefaultPorts = []int{}

	callbackDelay := 20 * time.Millisecond
	start := time.Now()

	report, _ := StartScan(context.Background(), ips, cfg, func(ev ScanEvent) {
		if ev.Type == EventResult {
			time.Sleep(callbackDelay) // simulate slow UI render
		}
	})

	elapsed := report.EndTime.Sub(start)

	// With async dispatch, scan should complete significantly faster
	// than 50 hosts × 20ms callback delay = 1000ms if workers were stalled sequentially.
	// 50 hosts * 20ms = 1000ms.
	maxAllowed := 800 * time.Millisecond
	if elapsed > maxAllowed {
		t.Errorf("slow callback stalled workers: elapsed %v > %v", elapsed, maxAllowed)
	}
}

func TestNoGoroutineLeakOnPrematureCancel(t *testing.T) {
	before := runtime.NumGoroutine()

	ips := make([]string, 100)
	for i := range ips {
		ips[i] = "10.0.0." + itoa(i+1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	var processed int32
	go func() {
		for atomic.LoadInt32(&processed) < 10 {
			time.Sleep(time.Millisecond)
		}
		cancel() // cancel after 10 hosts processed
	}()

	cfg := DefaultConfig()
	cfg.DefaultPorts = []int{}
	cfg.PingTimeoutMs = 50

	_, _ = StartScan(ctx, ips, cfg, func(ev ScanEvent) {
		if ev.Type == EventResult {
			atomic.AddInt32(&processed, 1)
		}
	})

	// Allow goroutines to settle
	time.Sleep(100 * time.Millisecond)
	runtime.GC()

	after := runtime.NumGoroutine()
	leak := after - before
	if leak > 3 { // tolerance for test harness goroutines
		t.Errorf("goroutine leak detected: %d goroutines created and not cleaned up", leak)
	}
}
