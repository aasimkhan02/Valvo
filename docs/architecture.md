
# P.A.R.L-AI — System Architecture

## 1. Overview

P.A.R.L-AI (Planet-Scale Adaptive Rate Limiter with AI Control Plane) is a globally distributed, low-latency rate-limiting system designed to enforce multi-dimensional quotas at planet scale.

The core system prioritizes:

- Microsecond-level local decisions
- High availability under partitions
- Bounded global correctness without consensus
- Clear separation between deterministic control logic and AI advisory logic

**AI is never involved in request admission.**

## 2. Design Goals

### Primary Goals

- Sub-millisecond request admission latency
- Graceful behavior under network partitions
- Predictable, bounded quota overshoot
- Support for multi-key limits (tenant, region, API, user)
- Operational transparency and explainability

### Explicit Non-Goals

- Strong global consistency
- Exactly-once global quota enforcement
- AI-driven real-time admission decisions

## 3. High-Level Architecture

```
Clients
    |
    v
API Gateway (Envoy / Nginx)
    |
    v
P.A.R.L Core Node (Go)
    |
    +-- Local Fast Path (hot path)
    |
    +-- Admission Controller
    |
    +-- Lease Manager
    |
    +-- CRDT Replica Layer
    |
    +-- Metrics & Tracing

Async / Out-of-band:

Prometheus / OTel
    |
    v
AI Control Plane (Python)
```

## 4. Request Flow (Hot Path)

### Step-by-Step

1. **Gateway forwards request metadata**
     - `tenant_id`
     - `region`
     - `resource`
     - `cost`

2. **Local Fast Path executes**
     - Token Bucket
     - Sliding Window
     - Per-key sharded state
     - Lock-free reads
     - EWMA Predictor (top-K keys only)
         - Detects burst acceleration
         - Adjusts local admission aggressiveness

3. **Admission Controller decides**
     - `ALLOW`
     - `SOFT_DENY` (grace mode)
     - `HARD_DENY`

4. **Response returned immediately**
     - No network calls
     - No disk IO
     - No AI involvement
     - 99.99% of requests end here

## 5. Local Fast Path

### Responsibilities

- Enforce local rate limits
- Maintain rolling usage windows
- Serve requests in constant time

### Key Properties

- In-memory only
- Sharded maps to reduce contention
- Deterministic behavior
- Safe under process restarts

## 6. Admission Controller

The admission controller converts local and global signals into a final decision.

### Modes

- **Strict** — hard enforcement
- **Soft** — probabilistic throttling
- **Grace** — partial allowance during anomalies

### Inputs

- Local usage
- EWMA trend
- Active lease capacity
- Policy constraints

### Output

A single deterministic decision with metadata.

## 7. Lease Manager (Bounded Overshoot)

To avoid global coordination on every request, nodes operate under time-bound quota leases.

### Lease Properties

- Pre-allocated quota slices
- Time-limited validity
- Hard upper bounds
- No recursive borrowing

### Guarantees

- Overshoot is bounded
- Global limits are eventually respected
- System remains available during partitions

## 8. CRDT Replica Layer

Global quota state is synchronized using PN-Counters with delta-based gossip.

### Why CRDTs

- No leader election
- Partition tolerant
- Convergent by design

### Tradeoff

- Approximate global view
- Accepts short-term divergence

This is a deliberate choice to favor availability over strict consistency.

## 9. Cluster & Gossip

### Responsibilities

- Membership tracking
- Failure detection
- State dissemination

### Behavior

- Periodic gossip
- Adaptive fanout
- Tolerant to node churn

No quorum or consensus is required for admission.

## 10. Observability Architecture

### Metrics

- Requests allowed / denied
- Lease exhaustion rate
- Overshoot percentage
- Predictor accuracy

### Tracing

- Admission decision path
- Latency attribution

Instrumentation is defined early but exporters are pluggable.

## 11. AI Control Plane (Out-of-Band)

AI components operate only on historical and aggregated data.

### Responsibilities

- Explain incidents
- Recommend policy changes
- Analyze traffic replays
- Provide natural-language operational insights

### Explicit Constraints

- No request-path access
- No direct policy mutation
- Human-in-the-loop required
- Deterministic logic always wins

## 12. Failure Scenarios

| Failure | Behavior |
|---------|----------|
| Node crash | Local state lost, leases rebalanced |
| Network partition | Independent operation with bounded overshoot |
| CRDT lag | Temporary divergence, eventual convergence |
| AI outage | No impact on traffic |

The system fails open in a controlled manner, never catastrophically.

## 13. Tradeoffs Summary

| Choice | Benefit | Cost |
|--------|---------|------|
| No consensus | Low latency | Approximate global limits |
| Local fast path | Availability | Temporary overshoot |
| AI off hot path | Safety | Slower adaptation |

All tradeoffs are explicit and documented.

## 14. Architecture Invariants

These must never change:

- AI is advisory only
- Admission is deterministic
- Hot path has no network calls
- Overshoot is bounded
- Safety limits are enforced locally

## 15. Conclusion

P.A.R.L-AI is designed as a production-grade control system, not an academic toy.

It favors:

- Clear contracts
- Explicit tradeoffs
- Operational safety
- Explainability over magic

This foundation allows aggressive optimization without architectural risk.
