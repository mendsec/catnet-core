## 2024-06-03 - [Concurrent Port Scanning]
**Learning:** Sequential network I/O operations (like `net.DialTimeout`) cause severe performance bottlenecks when waiting for timeouts. Cumulative timeouts block the scan entirely.
**Action:** Always introduce bounded concurrency (e.g., using a channel semaphore) and thread-safe data aggregation (`sync.Mutex`) when handling multiple network operations that may drop packets or time out. Ensure deterministic output by sorting the results before returning.

## 2026-06-04 - Optimize IP Range Allocations
**Learning:** Expanding large IP ranges (like CIDR /16 or full class B dash ranges) causes massive memory allocation pressure in Go because the slice backing array constantly resizes and reallocates, and creating a new `net.IP` in every iteration creates garbage collection overhead.
**Action:** When parsing target ranges, always pre-calculate the total number of IPs (using CIDR mask size or end-start+1) and initialize the slice with `make([]string, 0, capacity)`. Additionally, reuse a single `net.IP` buffer across loop iterations when generating sequential IPs.

## 2024-05-24
*   **Optimization:** Removed the overhead of spawning an external shell process (`exec.Command`) for ICMP pings on Windows by leveraging the native Win32 `IcmpSendEcho` API from `iphlpapi.dll`.
*   **Bottleneck Addressed:** High process-creation overhead on Windows. Spawning `ping.exe` for every IP scan is significantly slower than direct DLL calls, especially when scanning large subnets concurrently.
*   **Edge Case / Learning:** The ICMP reply structure on Windows returns the actual status code at offset 4 of the payload byte array. The lazy loading of `iphlpapi.dll` is efficient enough for this context as the network latency dominates, though global loading could save a few nanoseconds. The change eliminates command injection risks while providing a massive performance boost.

2023-10-27: Found infinite loop and OOM vulnerability when parsing CIDRs containing 0.0.0.0 and exceeding limits. Prevented infinite looping by making increment function return overflow indication, and prevented OOMs by capping CIDR generation limit to 65536 IPs.

## 2026-06-05 - [MAC Address Resolution Fast Path]
**Learning:** Spawning an external process (`fork`/`exec` via `exec.Command`) to read the local ARP table (`arp -an`) creates massive performance penalties and OS scheduling contention when scaled across many concurrent scanning threads. For example, spawning a process for 100 IPs can take ~300ms, whereas reading a file takes ~2.5ms.
**Action:** When gathering MAC addresses on POSIX systems (specifically Linux), implement a fast path that reads directly from `/proc/net/arp`. Only fallback to `exec.Command` if the file doesn't exist or isn't mounted (e.g. macOS/BSD).
## 2023-10-27 - Atomic Index Loop Over Buffered Channel
**Learning:** For distributing pre-known arrays of work (like IP slices up to 65536 items) across worker threads, creating a buffered channel and pushing all items into it introduces massive setup overhead and memory allocation (O(N)).
**Action:** Use a pre-allocated array and a shared `int32` atomic index (`atomic.AddInt32(&index, 1)`) inside the worker thread loop to read the slice dynamically without synchronization channels. This gives ~3x speedup.

## 2024-06-07 - [Fixed Worker Pools vs Semaphore Goroutines]
**Learning:** For distributing many tiny tasks (like port scans per IP, up to 65535 ports), launching a new goroutine for every single task and limiting concurrency via a channel semaphore introduces significant memory allocation and scheduling overhead. A benchmark showed that a fixed pool of workers iterating over an array using `atomic.AddInt32` is over 35x faster in raw scheduling overhead.
**Action:** When performing high-volume concurrent tasks over pre-known arrays, prefer spawning a fixed pool of `N` worker goroutines that pull tasks via an atomic index counter (`atomic.AddInt32`), rather than spinning up `M` goroutines constrained by an `N`-capacity semaphore channel.

## 2023-10-25 - Zero-allocation parsing of `/proc/net/arp`
**Learning:** In environments with heavily populated ARP tables, reading and converting `/proc/net/arp` into a string array using `strings.Split` and `strings.Fields` creates a significant O(N) memory allocation and garbage collection bottleneck. This degrades throughput heavily during highly concurrent per-IP discovery scans.
**Action:** Always prefer zero-allocation byte indexing functions (`bytes.Index`, `bytes.Fields`) when parsing structured Linux system files (`/proc/*`) sequentially, particularly inside highly concurrent routines or high-frequency loops. Avoid unnecessary `string(data)` type conversions that force complete memory duplication.

## 2026-06-09 - Avoid string allocations in tight loops
**Learning:** Using string manipulation functions like `strings.Split` and `strings.Join` inside O(n) loops (e.g., building graphs from large reports) causes massive and unnecessary allocation overhead (10-20 memory allocations per host due to slice/string buffers).
**Action:** When extracting a specific part of a string (like a subnet from an IP address), prefer zero-allocation byte indexing functions (`strings.LastIndexByte`) and simple string slicing (`string[:index]`) instead of `Split` + `Join` to prevent GC bottlenecks.

## 2026-06-20 - [Zero-Allocation Struct Keys for O(N^2) Hashmaps]
**Learning:** Using string concatenation (`src + "-" + dst`) to create unique keys for maps inside O(N^2) loops (like graph edge generation) causes massive memory allocations (millions of objects) and garbage collection bottlenecks because each concatenation creates a new string in memory.
**Action:** Always use an anonymous or dedicated struct with string fields (`type edgeKey struct { src, dst string }`) as the map key. Go handles struct keys efficiently without additional heap allocations, resulting in a 2x performance increase and almost zero allocations per loop iteration.
## 2025-03-10 - [Avoid strings.Join for large arrays]
**Learning:** Using `strings.Join` along with `strconv.Itoa` to format an array of integers (like open ports) inside a tight loop causes multiple memory allocations per item, creating significant overhead in operations like CSV exporting.
**Action:** Preallocate a byte slice buffer (`var portsBuf []byte`) and use `strconv.AppendInt` to construct the delimited string. Reuse the buffer via `portsBuf = portsBuf[:0]` on each loop iteration to reduce allocations from O(N) to O(1) for the entire list.

## 2024-06-23 - [Zero-Allocation Fixed-Length String Formatting]
**Learning:** Chaining string operations like `strings.ToUpper(strings.ReplaceAll(...))` for fixed-length formats (like 17-character MAC addresses) causes multiple heap allocations per call. In high-volume sequential parsing tasks (like reading `/proc/net/arp`), this creates a measurable garbage collection bottleneck.
**Action:** When transforming fixed-length strings in hot paths, use a stack-allocated byte array (`var out [17]byte`) and manual byte iteration (e.g. `c - 'a' + 'A'`) to perform conversions without allocating new strings until the final cast. This reduces allocations from multiple to exactly 1 (for the final string) and significantly increases throughput.
## 2026-07-28 - [Validating Context Cancellation for Blocking OS Executions]
**Learning:** Functions that wrap long-running system operations (such as `exec.CommandContext` or `net.DialContext`) generally handle context cancellation automatically, but early aborts for network calls need explicit testing. Merely relying on lower-level Go standard library propagation does not automatically guarantee zero side-effects if the cancellation happens before execution logic completes in consumer boundaries.
**Action:** Always include fine-grained tests that explicitly pass a pre-cancelled context to network discovery and connection routines (e.g. Ping, LookupAddr) to definitively prove zero leakage of OS processes or goroutines on early aborts.
