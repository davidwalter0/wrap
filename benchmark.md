*Benchmark*

Test benchmarks with profiles show improved performance using wrapped
buffer for chained calls.

```
    go test -o wrap.test main_test.go buffer.go wrap.go;
    rm *.profiles;
    for test in $(grep -o 'Benchmark[^(]*' main_test.go); do
        ./wrap.test -test.run XYZ -test.bench ${test} -test.v \
        -test.benchtime 5s -test.benchmem -test.cpuprofile ${test}.profiles;
        echo top5| go tool pprof ${test}.profiles;  done 2>&1| \
            grep -Ev "^ok|^PASS|^ *Entering" | sed -e 's,(pprof),,g' | tee text2 | grep "^ *Benchmark_"
```

- Recreating the call chain has enough overhead to affect a high
  performance application.
- Caching the result of a call chain gives a result similar to pure
  discrete calls.
- Buffered calls with chained or pure calls reduce overhead
- In a single threaded benchmark test the bpool implementation doesn't
  seem to improve performance over per instance creation of a
  bytes.buffer.

```
    Benchmark_Buffer_BP_G-4                     2000000000           0.00 ns/op        0 B/op          0 allocs/op
    Benchmark_Buffer_BP_Make_One_Chain_D-4      2000000000           0.00 ns/op        0 B/op          0 allocs/op
    Benchmark_Buffer_BP_Multi_Chain_Create_A-4  1000000000           0.91 ns/op        1 B/op          0 allocs/op

    Benchmark_Buffer_H-4                        2000000000           0.00 ns/op        0 B/op          0 allocs/op
    Benchmark_Buffer_Make_One_Chain_E-4         2000000000           0.00 ns/op        0 B/op          0 allocs/op
    Benchmark_Buffer_Multi_Chain_Create_B-4     2000000000           0.55 ns/op        0 B/op          0 allocs/op

    Benchmark_UnBuffered_I-4                    1000000000           0.91 ns/op        1 B/op          0 allocs/op
    Benchmark_UnBuffered_Make_One_Chain_F-4     1000000000           0.90 ns/op        1 B/op          0 allocs/op
    Benchmark_UnBuffered_Multi_Chain_Create_C-4 1000000000           0.91 ns/op        1 B/op          0 allocs/op
```


```
    Benchmark_Buffer_BP_G-4   	2000000000	         0.00 ns/op	       0 B/op	       0 allocs/op
     10ms of 10ms total (  100%)
    Showing top 5 nodes out of 16 (cum >= 10ms)
          flat  flat%   sum%        cum   cum%
          10ms   100%   100%       10ms   100%  runtime.memclr
             0     0%   100%       10ms   100%  bytes.(*Buffer).Write
             0     0%   100%       10ms   100%  bytes.(*Buffer).grow
             0     0%   100%       10ms   100%  bytes.makeSlice
             0     0%   100%       10ms   100%  command-line-arguments.Benchmark_Buffer_BP_G
    Benchmark_Buffer_BP_Make_One_Chain_D-4   	2000000000	         0.00 ns/op	       0 B/op	       0 allocs/op
     40ms of 40ms total (  100%)
    Showing top 5 nodes out of 25 (cum >= 10ms)
          flat  flat%   sum%        cum   cum%
          10ms 25.00% 25.00%       10ms 25.00%  fmt.(*pp).fmtPointer
          10ms 25.00% 50.00%       10ms 25.00%  runtime.getfull
          10ms 25.00% 75.00%       10ms 25.00%  runtime.scanblock
          10ms 25.00%   100%       10ms 25.00%  runtime.scanobject
             0     0%   100%       10ms 25.00%  command-line-arguments.Benchmark_Buffer_BP_Make_One_Chain_D

    Benchmark_Buffer_BP_Multi_Chain_Create_A-4   	1000000000	         0.91 ns/op	       1 B/op	       0 allocs/op
     4660ms of 11110ms total (41.94%)
    Dropped 89 nodes (cum <= 55.55ms)
    Showing top 5 nodes out of 143 (cum >= 1070ms)
          flat  flat%   sum%        cum   cum%
        1400ms 12.60% 12.60%     1400ms 12.60%  runtime.memclr
         960ms  8.64% 21.24%      960ms  8.64%  runtime.memmove
         890ms  8.01% 29.25%      990ms  8.91%  fmt.(*fmt).writePadding
         870ms  7.83% 37.08%     3150ms 28.35%  runtime.mallocgc
         540ms  4.86% 41.94%     1070ms  9.63%  runtime.scanobject

    Benchmark_Buffer_H-4   	2000000000	         0.00 ns/op	       0 B/op	       0 allocs/op
     20ms of 20ms total (  100%)
    Showing top 5 nodes out of 23 (cum >= 10ms)
          flat  flat%   sum%        cum   cum%
          10ms 50.00% 50.00%       10ms 50.00%  runtime.gentraceback
          10ms 50.00%   100%       10ms 50.00%  runtime.memclr
             0     0%   100%       10ms 50.00%  bytes.(*Buffer).Write
             0     0%   100%       10ms 50.00%  bytes.(*Buffer).grow
             0     0%   100%       10ms 50.00%  bytes.makeSlice

    Benchmark_Buffer_Make_One_Chain_E-4   	2000000000	         0.00 ns/op	       0 B/op	       0 allocs/op
     40ms of 40ms total (  100%)
    Showing top 5 nodes out of 34 (cum >= 20ms)
          flat  flat%   sum%        cum   cum%
          20ms 50.00% 50.00%       20ms 50.00%  runtime.memclr
          10ms 25.00% 75.00%       10ms 25.00%  runtime.(*gcWork).get
          10ms 25.00%   100%       10ms 25.00%  runtime.scanblock
             0     0%   100%       20ms 50.00%  bytes.(*Buffer).Write
             0     0%   100%       20ms 50.00%  bytes.(*Buffer).grow

    Benchmark_Buffer_Multi_Chain_Create_B-4   	2000000000	         0.55 ns/op	       0 B/op	       0 allocs/op
     7000ms of 14730ms total (47.52%)
    Dropped 107 nodes (cum <= 73.65ms)
    Showing top 5 nodes out of 132 (cum >= 4370ms)
          flat  flat%   sum%        cum   cum%
        2450ms 16.63% 16.63%     2450ms 16.63%  runtime.memclr
        1470ms  9.98% 26.61%     1470ms  9.98%  runtime.memmove
        1200ms  8.15% 34.76%     1820ms 12.36%  runtime.scanobject
         940ms  6.38% 41.14%      950ms  6.45%  fmt.(*fmt).writePadding
         940ms  6.38% 47.52%     4370ms 29.67%  runtime.mallocgc

    Benchmark_UnBuffered_I-4   	1000000000	         0.91 ns/op	       1 B/op	       0 allocs/op
     5440ms of 10980ms total (49.54%)
    Dropped 92 nodes (cum <= 54.90ms)
    Showing top 5 nodes out of 123 (cum >= 1400ms)
          flat  flat%   sum%        cum   cum%
        1980ms 18.03% 18.03%     1980ms 18.03%  runtime.memclr
         990ms  9.02% 27.05%      990ms  9.02%  runtime.memmove
         920ms  8.38% 35.43%      960ms  8.74%  fmt.(*fmt).writePadding
         830ms  7.56% 42.99%     3470ms 31.60%  runtime.mallocgc
         720ms  6.56% 49.54%     1400ms 12.75%  runtime.scanobject

    Benchmark_UnBuffered_Make_One_Chain_F-4   	1000000000	         0.90 ns/op	       1 B/op	       0 allocs/op
     5420ms of 11020ms total (49.18%)
    Dropped 81 nodes (cum <= 55.10ms)
    Showing top 5 nodes out of 133 (cum >= 3510ms)
          flat  flat%   sum%        cum   cum%
        2150ms 19.51% 19.51%     2150ms 19.51%  runtime.memclr
         930ms  8.44% 27.95%      930ms  8.44%  runtime.memmove
         860ms  7.80% 35.75%     1250ms 11.34%  runtime.scanobject
         840ms  7.62% 43.38%      890ms  8.08%  fmt.(*fmt).writePadding
         640ms  5.81% 49.18%     3510ms 31.85%  runtime.mallocgc

    Benchmark_UnBuffered_Multi_Chain_Create_C-4   	1000000000	         0.91 ns/op	       1 B/op	       0 allocs/op
     5430ms of 11200ms total (48.48%)
    Dropped 86 nodes (cum <= 56ms)
    Showing top 5 nodes out of 130 (cum >= 3380ms)
          flat  flat%   sum%        cum   cum%
        1950ms 17.41% 17.41%     1950ms 17.41%  runtime.memclr
        1060ms  9.46% 26.88%     1060ms  9.46%  runtime.memmove
         940ms  8.39% 35.27%     1010ms  9.02%  fmt.(*fmt).writePadding
         880ms  7.86% 43.12%     1380ms 12.32%  runtime.scanobject
         600ms  5.36% 48.48%     3380ms 30.18%  runtime.mallocgc

```
