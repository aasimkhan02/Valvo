
# Core Algorithms & Mathematical Foundations

## 1. Design Philosophy

P.A.R.L-AI is a control system, not a rules engine.

**Key principles:**

- Deterministic admission
- Bounded error instead of global coordination
- Predictive dampening instead of reactive throttling
- Separation of mechanism (Go core) and analysis (AI plane)

## 2. Rate Limiting Primitives

### 2.1 Token Bucket (Primary Enforcer)

Each `(tenant, region, resource)` key maintains:

- **Capacity** `C`
- **Refill rate** `r` (tokens/sec)
- **Current tokens** `T(t)`

**Update rule:**

```
T(t) = min(C, T(t₀) + r · (t − t₀))
```

A request of cost `k` is allowed if:

```
T(t) ≥ k
```

Otherwise denied or degraded.

**Why token bucket:**
- Burst-friendly
- Constant-time
- Well-understood failure behavior

### 2.2 Sliding Window (Fairness Guard)

Used to prevent micro-burst abuse that token buckets may allow.

For window size `W`:

```
usage = Σ requests in (t − W, t]
```

**Admission constraint:**

```
usage ≤ window_limit
```

**Implemented via:**
- Ring buffers
- Fixed time buckets
- No per-request allocations

## 3. EWMA Predictor (Burst Detection)

Applied only to top-K hot keys to control CPU cost.

### 3.1 EWMA Formula

For observed request rate `xₜ`:

```
Sₜ = α·xₜ + (1−α)·Sₜ₋₁
```

Where:
- `α ∈ (0,1)` controls responsiveness
- Higher `α` → faster reaction, noisier

### 3.2 First Derivative (Acceleration)

To detect rate of change:

```
Δₜ = Sₜ − Sₜ₋₁
```

**Burst condition:**

```
Δₜ > burst_threshold
```

**Triggers:**
- Early lease tightening
- Probabilistic throttling
- Grace mode entry

## 4. Admission Controller Logic

The admission controller evaluates three signals:

- Local capacity
- Lease headroom
- Predictor trend

### 4.1 Decision Function

```
decision = f(local_ok, lease_ok, trend)
```

| Condition | Result |
|-----------|--------|
| All OK | `ALLOW` |
| Local OK + lease tight | `SOFT_DENY` |
| Predictor spike | `PROBABILISTIC` |
| Safety violated | `HARD_DENY` |

No randomness except in explicit probabilistic mode.

## 5. Probabilistic Throttling

Used to smooth transitions instead of cliff failures.

### 5.1 Probability Function

```
p_deny = clamp(
    (usage − soft_limit) / (hard_limit − soft_limit),
    0, 1
)
```

Request denied with probability `p_deny`.

**Benefits:**
- Prevents retry storms
- Preserves partial throughput
- Reduces oscillations

## 6. Lease-Based Global Coordination

### 6.1 Motivation

Global counters on every request are:

- Too slow
- Partition fragile
- Operationally expensive

Instead: time-bounded quota leasing

### 6.2 Lease Model

Each node holds:

```
L = { quota, expiry_time }
```

**Constraints:**

```
Σ L_node ≤ global_limit + ε
```

Where `ε` is bounded overshoot.

### 6.3 Lease Consumption

Local admission decrements lease quota. If exhausted:

- Enter grace mode
- Trigger async lease refresh
- Never block hot path

### 6.4 Overshoot Bound

Worst-case overshoot:

```
overshoot ≤ N · max_lease
```

Where:
- `N` = number of nodes
- `max_lease` is strictly capped

This bound is configurable and enforced.

## 7. CRDT-Based Global State

### 7.1 PN-Counter

Each node maintains:

- `P` = increments
- `N` = decrements
- `value = ΣP − ΣN`

Merged via component-wise max.

**Properties:**
- Associative
- Commutative
- Idempotent
- Guarantees convergence without coordination

### 7.2 Delta-Gossip Optimization

Only deltas since last sync are exchanged:

```
Δ = current_state − last_sent_state
```

Reduces bandwidth and convergence time.

## 8. Hot Key Management

Top-K hot keys tracked via:

- Count-Min Sketch (optional)
- Periodic heap selection

Only these keys:

- Run predictors
- Emit detailed metrics
- Influence AI analysis

Cold keys stay cheap.

## 9. Determinism Guarantees

Admission decision must be:

- Pure function of local state
- Free of external calls
- Time-bounded

This ensures:

- Reproducibility
- Debuggability
- Replay correctness

## 10. Summary of Mathematical Guarantees

| Property | Guarantee |
|----------|-----------|
| Latency | O(1) per request |
| Overshoot | Strictly bounded |
| Convergence | Eventual |
| Availability | Partition tolerant |
| Stability | Damped, not oscillatory |
