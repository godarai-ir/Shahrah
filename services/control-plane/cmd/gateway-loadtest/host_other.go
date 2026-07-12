//go:build !windows

package main

type unsupportedHostSampler struct{}

func newHostSampler() hostSampler { return unsupportedHostSampler{} }

func (unsupportedHostSampler) Sample() (hostSample, error) {
	return hostSample{}, nil
}
