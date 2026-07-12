# Shahrah gateway load-test report

Generated: 2026-07-11T23:38:34Z

- Machine: Intel(R) Core(TM) i7-5500U CPU @ 2.40GHz; 2 cores/4 logical; 15.92 GB RAM; Docker CPUs=4, memory=8285044736 bytes
- Logical CPUs: 4
- Endpoint: `http://127.0.0.1:4000/v1/chat/completions`
- Generator on same host: yes

## Capacity estimate

- Maximum passing ramp rate: **9 RPS**
- Break point: not reached within the configured maximum
- Recommended rate with 80% safety factor: **9 RPS**
- Sustained verification passed: **true**

## Stages

| Stage | Target RPS | Achieved RPS | Errors | p95 | Avg host CPU | Result |
|---|---:|---:|---:|---:|---:|---|
| warmup | 9 | 9.0 | 0.00% | 61.4 ms | 71.0% | pass |
| stable | 9 | 9.0 | 0.00% | 63.5 ms | 60.4% | pass |
| burst | 14 | 14.0 | 0.00% | 128.3 ms | 73.4% | pass |

## Notes

- The load generator ran on the same host, so its CPU and memory overhead are included.
- Every request used unique prompt content to avoid response-cache hits.
- A stage passes only when errors, p95 latency, achieved rate, and average host CPU all satisfy thresholds.
