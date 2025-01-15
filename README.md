# command-queue-executor

# Plan 

1. [x] Initial structure design
2. [x] Naive ordered map
3. [x] Naive stub for message queue
4. [x] Message queue implementation using RabbitMQ
5. [ ] Setup CI via Github Actions
6. [ ] Models/Commands types
7. [ ] Client-side message generator
8. [ ] Server side consumer
9. [ ] O(1) ordered map
10. [ ] Reconnections with exp tries
11. [ ] Handling multiple clients with gentle shutdown & concurency


# Devlog

1. Starting with naive ordered map impl using plain map + order array (not even a doubly linked list), benchmark:
```
    Size: 100, Store Duration: 8µs, Get Duration: 4.291µs, Delete Duration: 11.25µs
    Size: 1000, Store Duration: 44.292µs, Get Duration: 12.333µs, Delete Duration: 161.375µs
    Size: 10000, Store Duration: 745.833µs, Get Duration: 447.917µs, Delete Duration: 11.9505ms
    Size: 100000, Store Duration: 6.843667ms, Get Duration: 3.400125ms, Delete Duration: 772.933042ms
```
2. Added a simple message queue pub/sub interface with stub implementation based on channels with test
3. Added message queue using RabbitMQ and unified test cases for both implementations
4. Changed message queue pattern from pub/sub to req/rep to get rpc-like api