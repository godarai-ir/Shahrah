//go:build windows

package main

import (
	"errors"
	"syscall"
	"unsafe"
)

type filetime struct {
	LowDateTime  uint32
	HighDateTime uint32
}

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

type windowsHostSampler struct {
	getSystemTimes       *syscall.LazyProc
	globalMemoryStatusEx *syscall.LazyProc
	previousIdle         uint64
	previousKernel       uint64
	previousUser         uint64
	initialized          bool
}

func newHostSampler() hostSampler {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	return &windowsHostSampler{
		getSystemTimes:       kernel32.NewProc("GetSystemTimes"),
		globalMemoryStatusEx: kernel32.NewProc("GlobalMemoryStatusEx"),
	}
}

func (s *windowsHostSampler) Sample() (hostSample, error) {
	var idle, kernel, user filetime
	ok, _, callErr := s.getSystemTimes.Call(
		uintptr(unsafe.Pointer(&idle)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	)
	if ok == 0 {
		return hostSample{}, callErr
	}

	memory := memoryStatusEx{Length: uint32(unsafe.Sizeof(memoryStatusEx{}))}
	ok, _, callErr = s.globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memory)))
	if ok == 0 {
		return hostSample{}, callErr
	}

	idleTicks := filetimeTicks(idle)
	kernelTicks := filetimeTicks(kernel)
	userTicks := filetimeTicks(user)
	result := hostSample{AvailableMemoryBytes: memory.AvailPhys}

	if s.initialized {
		idleDelta := idleTicks - s.previousIdle
		kernelDelta := kernelTicks - s.previousKernel
		userDelta := userTicks - s.previousUser
		total := kernelDelta + userDelta
		if total > 0 && total >= idleDelta {
			result.CPUPercent = float64(total-idleDelta) / float64(total) * 100
			result.Valid = true
		}
	}

	s.previousIdle = idleTicks
	s.previousKernel = kernelTicks
	s.previousUser = userTicks
	s.initialized = true
	if !result.Valid && memory.TotalPhys == 0 {
		return hostSample{}, errors.New("host resource counters are unavailable")
	}
	return result, nil
}

func filetimeTicks(value filetime) uint64 {
	return uint64(value.HighDateTime)<<32 | uint64(value.LowDateTime)
}
