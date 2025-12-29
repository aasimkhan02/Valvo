# Failure & Chaos Analysis

## 1. Failure Philosophy

P.A.R.L-AI is designed to:

- Fail locally
- Degrade gracefully
- Never amplify failures

All failure handling is explicit, not emergent.

## 2. Node-Level Failures

### 2.1 Node Crash

**Impact**
- Local in-memory state lost
- Active leases dropped

**Recovery**
- Node rejoins cluster
- Receives fresh leases
- CRDT state converges

**Guarantee**
- No global outage
- Temporary underutilization only

### 2.2 Process Restart Loop

**Risk**
- Repeated lease acquisition

**Mitigation**
- Lease issuance rate-limited
- Cooldown enforced
- Membership dampening

## 3. Network Partitions

### 3.1 Split-Brain Scenario

**Behavior**
- Each partition enforces limits independently
- Leases cap maximum damage

**Worst Case**
```
global_overshoot ≤ partitions · max_lease
```
This is known, bounded, and acceptable.

### 3.2 Partial Packet Loss

**Effect**
- Gossip lag
- Stale CRDT views

**Mitigation**
- Delta retransmission
- Predictor dampening
- No admission blocking

## 4. Traffic Pathologies

### 4.1 Retry Storms

**Cause**
- Downstream failures
- Aggressive client retries

**Detection**
- High retry ratio
- Burst derivative spike

**Response**
- Probabilistic throttling
- Grace mode
- AI postmortem explanation

### 4.2 Coordinated Bursts

**Cause**
- Cron jobs
- Cache invalidation
- Regional failover

**Handling**
- EWMA detects acceleration
- Leases throttle growth
- No cascading denial

## 5. Predictor Failures

### 5.1 False Positives

**Effect**
- Early throttling

**Mitigation**
- Conservative thresholds
- Hysteresis
- Human-reviewed AI tuning

### 5.2 False Negatives

**Effect**
- Higher overshoot

**Mitigation**
- Lease caps
- Sliding window guard
- Hard safety limits

## 6. CRDT Anomalies

### 6.1 Message Reordering

Safe by design due to:
- Idempotent merges
- Monotonic state

### 6.2 Duplicate Messages

No effect.

## 7. Observability Failures

### 7.1 Metrics Backend Down

**Impact**
- Loss of visibility only

**Guarantee**
- Admission unaffected
- AI degraded gracefully

### 7.2 Tracing Overhead

**Mitigation**
- Sampling
- Trace off hot path
- Hard latency budget

## 8. AI Control Plane Failures

### 8.1 AI Service Down

**Impact**
- No recommendations
- No explanations

**Guarantee**
- Zero traffic impact
- No policy mutation

### 8.2 Incorrect AI Recommendation

**Mitigation**
- Human review
- Diff-based suggestions
- Hard bounds enforced in core

## 9. Configuration Failures

### 9.1 Bad Policy Push

**Protection**
- Schema validation
- Versioned rollout
- Atomic swap

Rollback is instant.

## 10. Chaos Matrix (Summary)

| Failure | System State | User Impact |
|---------|--------------|-------------|
| Node loss | Degraded | None |
| Partition | Independent | Bounded throttling |
| Retry storm | Damped | Partial |
| AI outage | Ignored | None |
| Config bug | Rolled back | None |

## 11. Final Invariant

No single failure can cause global denial or unbounded quota violation.

This invariant is enforced by:
- Local admission
- Lease caps
- Deterministic logic
- Explicit tradeoffs