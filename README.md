# 1billion-rows-concurrency

| Attempt Number | Approach | Execution Time | Diff | Commit |
|-----------------|---|---|---|--|
|0| Naive Implementation: Read temperatures into a map of cities. Iterate serially over each key (city) in map to find min, max and average temperatures.| 2:15.320s | |[naiveImplementation](https://github.com/piyu37/1billion-rows-concurrency/pull/1/commits/d312accf1fd55e3090a55b55ad52662f98f05b10)|
|1| Evaluate each city in map concurrently using goroutines and decouple reading and processing of file content. A buffered goroutine is used to communicate between the two processes.|5:3.988s|+167.680s|[simpleConcurrency](https://github.com/piyu37/1billion-rows-concurrency/pull/2/commits/2d2c175eaaa823fbd0be48b03f39997d743a05cb)|
|2| Instead of sending each line to the channel, now sending 100 lines chunked together. Also, to minimise garbage collection, not freeing up memory when resetting a slice.|2:50.906s|+35.586s|[batchProcessing](https://github.com/piyu37/1billion-rows-concurrency/pull/3/commits/4c5fc8a25360ded4518cf2d8fa4f6ae7acc4fa62)|
|3| In the station <> temperatures map, replaced the value for each key (city) to preprocessed min, max, count and sum of all temperatures instead of storing all recorded temperatures for the city.|2:10.078s|-5.242s|[refactorLogic](https://github.com/piyu37/1billion-rows-concurrency/pull/4/commits/225a57fc82171d496d4344ecbebb7b714e289020)|
|4| Use producer consumer pattern to read file in chunks and process the chunks in parallel and reduce memory allocation by processing each read chunk into a map. Result channel now can collate the smaller processed chunk maps.|1:12.530s|-62.790s|[chunkImplementation](https://github.com/piyu37/1billion-rows-concurrency/pull/5/commits/0e9952122ebdf7ddc2fe688410e5c9a3f0c4dafc)|
|5| Convert byte slice to string directly instead of using a strings.Builder.|0:43.865s|-91.455s||