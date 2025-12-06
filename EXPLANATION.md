# Explanation & Strategy

Here is a deeper dive into how I built the order matching engine and why I made certain choices.

## My Approach

### Core Data Structures
I used a standard order book structure. I have a map of `PriceLevel`s for both Bids and Asks.
- **Bids**: Sorted descending (highest price first).
- **Asks**: Sorted ascending (lowest price first).

### Optimization: O(log N) Insertion
At first, I was just doing a linear scan to find the right price level to insert into. It was O(N) which is kinda slow if the book gets deep. So I switched it to use `sort.Search` (binary search) to find the spot in O(log N). That helped a lot with performance stability.

### Matching Logic
For the matching part, it's a loop. When a new order comes in, I check the opposite side of the book. If the prices cross (Bid >= Ask), I make a trade. I keep doing that until the order is filled or I run out of matching orders.

## Concurrency Model
I decided to use a **single-threaded event loop** for the engine.
- All commands (submit, cancel, snapshot) go into a buffered channel.
- A single goroutine pulls them out and processes them one by one.

**Why?**
I know some people use locks (mutexes) for everything, but I think channels are cleaner here.
1.  **No Race Conditions**: Only one thing touches the book at a time.
2.  **Performance**: It's actually really fast because there's no lock contention. The CPU cache stays hot.

## Metrics
I added a `/metrics` endpoint that gives you the current stats.
- `throughput_orders_per_sec`: How many orders we're crunching per second.
- `orders_in_book`: How many active orders are sitting in memory.
- `p50`, `p99` latencies are calculated in the load test script rather than live because calculating p99 live on every request is heavy and I didn't want to slow down the engine.

## Trade-offs & Decisions
- **In-Memory**: Everything is in RAM. If the server crashes, data is gone. For this assignment, as the doc suggests, a DB would have caused slowdowns, so I stuck to memory.
- **UUIDs**: I used `google/uuid` for IDs. Random ints are faster but collisions are scary, so UUID is safer.
- **Float vs Int**: I used `int64` for prices and quantities to avoid floating point weirdness. 100.00 is stored as 10000 (assuming 2 decimal places).
