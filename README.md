# Double-Entry Central Limit Order Book

A concurrent, correctness-first trading engine written in Go.

This project implements a **Central Limit Order Book (CLOB)** and a **double-entry settlement ledger** with the goal of exploring the systems problems behind exchange infrastructure: deterministic order matching, concurrent state mutation, transactional settlement, and financial invariants.

The central design principle is:

> **Matching determines what should happen. The ledger determines what actually happened.**

A trade is therefore not a balance update. It is an atomic transfer of value between two parties.

---

## The Core Invariant

For every executed trade:

```text
                 TRADE
                   │
          ┌────────┴────────┐
          │                 │
       BUYER             SELLER
          │                 │
     Cash decreases    Cash increases
     Asset increases   Asset decreases
```

The system must never create or destroy value as a side effect of matching.

Every movement has a corresponding entry in the ledger:

```text
             ┌─────────────────┐
             │ Double-Entry     │
             │ Ledger           │
             └────────┬────────┘
                      │
          ┌───────────┴───────────┐
          │                       │
      Debit Entry            Credit Entry
          │                       │
          └───────────┬───────────┘
                      │
                 Balanced
```

This gives us a fundamental accounting invariant:

```text
Total Debits == Total Credits
```

The ledger is the durable record of financial truth.

---

## Order Book

The matching engine maintains BUY and SELL interest ordered by **price-time priority**.

```text
             SELL SIDE
       ┌─────────────────┐
       │ 101.00           │
       │ 100.50           │
       │ 100.00 ← Ask     │
       ├─────────────────┤
       │  99.90 ← Bid     │
       │  99.50           │
       │  99.00           │
       └─────────────────┘
             BUY SIDE
```

For each market:

* Higher bids have priority.
* Lower asks have priority.
* Orders at the same price are ordered by arrival time.
* A match occurs when the best bid crosses the best ask.

The matching engine produces **trades**, rather than directly performing arbitrary balance mutations.

---

## Execution Pipeline

The high-level execution path is:

```text
Client
  │
  ▼
API
  │
  ▼
Order Validation
  │
  ▼
Order Book
  │
  ▼
Matching Engine
  │
  ├───────────────┐
  ▼               ▼
Trade         Order State
  │
  ▼
Settlement
  │
  ▼
Double-Entry Ledger
  │
  ▼
PostgreSQL
```

This separation is intentional.

**The order book answers:**

> Which orders can trade?

**The settlement layer answers:**

> What value must move as a consequence?

**The ledger answers:**

> What financial state is authoritative?

---

## Concurrency Model

A trading engine is fundamentally a shared-state concurrency problem.

Multiple execution paths may attempt to observe or mutate the same order simultaneously.

For example:

```text
Order Quantity = 10

Goroutine A → Fill(6)
Goroutine B → Fill(6)
```

Without synchronization, both operations could observe the same previous state and produce an invalid result.

The order model therefore protects state transitions using synchronization primitives such as `sync.RWMutex`.

```text
                 Shared Order
                      │
              ┌───────┴───────┐
              │               │
           Readers          Writers
              │               │
            RLock            Lock
              │               │
          concurrent       exclusive
```

The objective is not simply "thread safety."

The objective is maintaining **state invariants under concurrency**.

For example:

```text
0 <= FilledQty <= Quantity
```

and:

```text
FILLED orders cannot be filled again
CANCELLED orders cannot be filled
```

---

## Transactional Settlement

Matching and settlement are distinct concerns.

Once a trade is produced, financial state must be updated atomically.

The intended transaction boundary is:

```text
BEGIN
  │
  ├── Lock relevant account rows
  │
  ├── Validate balances
  │
  ├── Apply asset transfer
  │
  ├── Apply cash transfer
  │
  ├── Write ledger entries
  │
  └── COMMIT
```

If any step fails:

```text
ROLLBACK
```

No partially applied trade should become durable financial state.

PostgreSQL therefore provides the durability and transactional guarantees required for settlement, while in-memory structures are optimized for the latency-sensitive matching path.

---

## Consistency Over Convenience

The system deliberately treats correctness as a first-class constraint.

Examples of invariants:

```text
FilledQty <= Order.Quantity

Account balance >= 0

Every debit has a corresponding credit

Cancelled orders cannot execute

Filled orders cannot execute again

A successful settlement is atomic

A failed settlement leaves no partial financial state
```

These invariants are more important than any individual implementation detail.

The architecture can change.

The invariants cannot.

---

## Architecture

```text
                         ┌─────────────┐
                         │   Client    │
                         └──────┬──────┘
                                │
                                ▼
                         ┌─────────────┐
                         │  Go API     │
                         └──────┬──────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │ Trading Service │
                       └────────┬────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │ Matching Engine │
                       └────────┬────────┘
                                │
                         ┌──────┴──────┐
                         │             │
                         ▼             ▼
                    Order State      Trades
                                       │
                                       ▼
                              ┌─────────────────┐
                              │   Settlement    │
                              └────────┬────────┘
                                       │
                         ┌─────────────┴─────────────┐
                         │                           │
                         ▼                           ▼
                  Double-Entry Ledger            Events
                         │                           │
                         ▼                           ▼
                    PostgreSQL                   RabbitMQ
                         │
                         ▼
                       Redis
```

The exact topology will evolve as the engine moves toward higher throughput and stronger failure semantics.

---

## Technology

| Component      | Role                               |
| -------------- | ---------------------------------- |
| **Go**         | Core engine and concurrency        |
| **PostgreSQL** | Durable financial state and ledger |
| **Redis**      | Low-latency auxiliary state        |
| **RabbitMQ**   | Asynchronous event delivery        |
| **Docker**     | Reproducible infrastructure        |
| **Chi**        | HTTP layer                         |
| **JWT**        | Authentication                     |

---

## Engineering Focus

This project is primarily an exploration of:

* Central Limit Order Book design
* Price-time priority
* Concurrent shared-state management
* Double-entry accounting
* ACID transactions
* Row-level locking
* State-machine design
* Idempotency
* Failure recovery
* Event-driven architecture
* Durable vs in-memory state
* Race detection
* Load and stress testing
* Observability
* Correctness under contention

Performance is treated as an engineering constraint, **not a marketing claim**. Benchmarks and profiling will determine where optimization is actually justified.

---

## Why I am Building This?

A trading engine is a useful systems problem because it forces several difficult concerns to interact:

```text
Concurrency
     +
State
     +
Ordering
     +
Transactions
     +
Failure
     +
Accounting
     ↓
Correctness
```

The interesting problem is not matching two numbers.

The interesting problem is maintaining a coherent financial state while thousands of operations compete to change it.

That is the problem this project is designed to explore.

---

## Status

**Active development.**

The implementation is being built incrementally from the primitives upward:

```text
Order Model
    ↓
State Transitions
    ↓
Concurrency
    ↓
Order Book
    ↓
Matching
    ↓
Trade Generation
    ↓
Double-Entry Settlement
    ↓
Persistence
    ↓
Events
    ↓
Load / Failure Testing
```

The goal is not to reproduce an exchange by assembling frameworks.

The goal is to understand and engineer the **mechanisms and invariants that make exchange-like systems correct under concurrency and failure**.
