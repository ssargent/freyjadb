# Project: Bitcask-Style Key-Value Store with Optional Partition/Sort Keys

Our objective is to build a **small, embeddable, log-structured key-value database** that:

* **Supports point-lookups with SSD-friendly speed** (append-only writes, one random read per `GET`).
* **Survives crashes** (CRC-guarded records + truncation recovery).
* **Runs highly-concurrent workloads** (many readers, single writer; hot-swappable index).
* **Offers optional DynamoDB-like **Partition Key / Sort Key** semantics** via a thin layering.  
* **Stays educational and test-driven**—each step ships behind CI and unit tests so you can learn storage internals incrementally.

The roadmap below breaks the work into 21 bite-sized items, each with:

* *Deliverable*: what you will code.
* *Core idea & test surface*: how to prove it works.
* *Complexity*: rough 1 (low) – 5 (high) score.

---

| #  | Work item (deliverable)                           | Core idea & test surface                                                                                     | Complexity |
|----|---------------------------------------------------|--------------------------------------------------------------------------------------------------------------|------------|
| 0  | **Repo & CI scaffold**                            | Green build with a single “hello world” test.                                                                | 1 |
| 1  | **Record codec**                                  | Serialize/deserialize `{crc32, key_size, val_size, timestamp, key, value}`. Unit test: encode ➜ decode round-trip, CRC mismatch rejection. | 2 |
| 2  | **Append-only log writer**                        | `put(key, value)` appends to *active.data* and fsyncs every N ms. Test: file grows, tail bytes match encoded record. | 2 |
| 3  | **Sequential log reader**                         | Iterate records from offset 0, validate CRC. Test: file with 3 records yields 3 objects & stops.            | 2 |
| 4  | **In-memory hash index builder**                  | Scan log, keep last offset of each key in `HashMap`. Test: duplicate keys → index points to newest.         | 2 |
| 5  | **Basic KV API (single-threaded)**                | `get`, `put`, `delete` wired to writer & index. Property tests: get(after put) == val; get(after delete) == None. | 2 |
| 6  | **Crash-safe reopen**                             | On startup, read *active.data* until CRC fails → truncate tail, rebuild index. Test: corrupt tail, reopen OK. | 3 |
| 7  | **Hint file format**                              | After *merge* emit *N.hint* = `{key_hash, file_id, offset, size}`.                                           | 1 |
| 8  | **Hint-driven fast bootstrap**                    | mmap hints to build index. Benchmark: boot time ≈ O(#keys).                                                  | 2 |
| 9  | **Log rotation**                                  | When *active.data* ≥ size_limit, close as *N.data*, start *N+1.data*. Test: puts cross boundary.             | 2 |
| 10 | **Compaction / merge utility**                    | Offline: read all `.data`, keep latest per key, drop tombstones, emit fresh file + hint. Test: space shrinks, data intact. | 4 |
| 11 | **Hot-swap compaction**                           | Run compactor concurrently, atomically rename, reload index. Test: readers never miss keys.                  | 4 |
| 12 | **Directory bootstrap**                           | At open, glob `*.data`, order by id, rebuild/load hints. Integration test: survives restarts & merges.       | 2 |
| 13 | **Read/Write concurrency (lock-light)**           | Single writer mutex; readers on immutable index; swap via `ArcSwap`/RCU. Stress test with multi-threads.     | 4 |
| 14 | **Memory packing / slab allocator**               | Store key bytes in contiguous “key log”; index keeps `u64` ptrs + lens. Benchmark: RAM/entry drops ≥ 30 %.   | 3 |
| 15 | **Segment-level Bloom filters**                   | Build Bloom per closed segment; negative `get` short-circuits. Test: random misses touch ≤ 1 segment.        | 3 |
| 16 | **Metrics & health endpoints**                    | Expose Prometheus counters: bytes_written, live_keys, compaction_sec.                                        | 1 |
| 17 | **Partition layer (PK)**                          | Sub-directory per PK; open/close Bitcask instance on demand. Test: keys in different PKs never collide.      | 3 |
| 18 | **Minimal Sort-Key range support**                | In each PK keep in-memory B-tree/vec sorted by SK. `query(pk, range)` streams ordered values. Test: range sorted. | 4 |
| 19 | **Background compaction scheduler**               | Prioritize segments by dead-bytes %, throttle I/O.                                                           | 3 |
| 20 | **CLI / library polish & docs**                   | `bitcask bench`, `bitcask dump`, API docs, examples.                                                         | 1 |
