package cache

import "testing"

func TestIsDistributedHotPathKey(t *testing.T) {
	for _, key := range []string{
		"deployments:tenant:tenant-1",
		"wallet:tenant:tenant-1",
		"tenant:host:acme.localhost",
	} {
		if !isDistributedHotPathKey(key) {
			t.Fatalf("expected %q to be distributed", key)
		}
	}

	for _, key := range []string{
		"rpm:key:test",
		"pricing:model:gpt-4o-mini",
		"exchange_rate:latest",
		"tenant:id:tenant-1",
	} {
		if isDistributedHotPathKey(key) {
			t.Fatalf("expected %q to remain memory-first", key)
		}
	}
}
