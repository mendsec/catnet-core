//go:build windows

package discovery

import (
	"context"
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

var (
	iphlpapi        = syscall.NewLazyDLL("iphlpapi.dll")
	icmpCreateFile  = iphlpapi.NewProc("IcmpCreateFile")
	icmpSendEcho    = iphlpapi.NewProc("IcmpSendEcho")
	icmpCloseHandle = iphlpapi.NewProc("IcmpCloseHandle")
	sendARP         = iphlpapi.NewProc("SendARP")
)

// osPing faz ping no Windows
// ⚡ Bolt Optimization: Use native IcmpSendEcho from iphlpapi.dll instead of spawning ping.exe.
// This avoids process-creation overhead on Windows for massive concurrent scans.
func osPing(ctx context.Context, ip string, timeoutMs int) bool {
	if net.ParseIP(ip) == nil {
		return false
	}
	if ctx.Err() != nil {
		return false
	}
	if timeoutMs <= 0 {
		timeoutMs = 1000 // safe default
	}

	destIP := net.ParseIP(ip).To4()
	if destIP == nil {
		return false
	}
	var destIPUint32 uint32
	destIPUint32 = uint32(destIP[0]) | uint32(destIP[1])<<8 | uint32(destIP[2])<<16 | uint32(destIP[3])<<24

	// Result channel
	resChan := make(chan bool, 1)

	go func() {
		handle, _, _ := icmpCreateFile.Call()
		if handle == ^uintptr(0) {
			resChan <- false
			return
		}
		defer icmpCloseHandle.Call(handle)

		// The ICMP_ECHO_REPLY structure is size 28 bytes on 32-bit and 32 bytes on 64-bit systems.
		// We'll allocate a 64 byte buffer which is more than enough for the struct + no data payload.
		replySize := 64
		replyBuf := make([]byte, replySize)

		// Send echo
		ret, _, _ := icmpSendEcho.Call(
			handle,
			uintptr(destIPUint32),
			0,
			0,
			0,
			uintptr(unsafe.Pointer(&replyBuf[0])),
			uintptr(replySize),
			uintptr(timeoutMs),
		)

		if ret == 0 {
			resChan <- false
			return
		}

		// Read IP_STATUS from the ICMP_ECHO_REPLY structure.
		// The offset of the Status field is 4 on both 32-bit and 64-bit Windows.
		// We check if it is 0 (IP_SUCCESS).
		status := *(*uint32)(unsafe.Pointer(&replyBuf[4]))
		resChan <- (status == 0)
	}()

	select {
	case <-ctx.Done():
		return false
	case res := <-resChan:
		return res
	}
}

// osGetMAC obtém o MAC usando SendARP no Windows
func osGetMAC(ctx context.Context, ip string) string {
	if ctx.Err() != nil {
		return ""
	}

	destIP := net.ParseIP(ip).To4()
	if destIP == nil {
		return ""
	}
	var destIPUint32 uint32
	destIPUint32 = uint32(destIP[0]) | uint32(destIP[1])<<8 | uint32(destIP[2])<<16 | uint32(destIP[3])<<24

	resChan := make(chan string, 1)

	go func() {
		var mac [6]byte
		macLen := uint32(len(mac))
		// Segurança: mac é um array de tamanho fixo [6]byte alocado na stack.
		// macLen é inicializado com len(mac) == 6 antes da chamada.
		// O acesso via unsafe.Pointer é seguro porque o array não escapa do
		// escopo e seu tamanho é conhecido em tempo de compilação.
		// A validação `macLen == 6` após o retorno garante dados não corrompidos.
		ret, _, _ := sendARP.Call(
			uintptr(destIPUint32),
			0,
			uintptr(unsafe.Pointer(&mac[0])),
			uintptr(unsafe.Pointer(&macLen)),
		)
		if ret == 0 && macLen == 6 {
			resChan <- fmt.Sprintf("%02X-%02X-%02X-%02X-%02X-%02X",
				mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
		} else {
			resChan <- ""
		}
	}()

	select {
	case <-ctx.Done():
		return ""
	case res := <-resChan:
		return res
	}
}
