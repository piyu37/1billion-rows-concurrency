# 1billion-rows-concurrency

| Attempt Number | Approach | Execution Time | Diff | Commit |
|-----------------|---|---|---|--|
|0| Naive Implementation: Read temperatures into a map of cities. Iterate serially over each key (city) in map to find min, max and average temperatures.| 2:15.320s | |[naiveImplementation](https://github.com/piyu37/1billion-rows-concurrency/pull/1/commits/d312accf1fd55e3090a55b55ad52662f98f05b10)|
|1| Evaluate each city in map concurrently using goroutines and decouple reading and processing of file content. A buffered goroutine is used to communicate between the two processes.|5:3.988s|+288.668s||