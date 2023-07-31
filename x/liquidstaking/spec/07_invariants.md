<!-- order: 7 -->

# Invariants

This document describes the invariants of liquid staking module.

all of check logics are treated as **OR** conditions, not **AND** condition

**NetAmount invariant check broken when**

- if ls token total supply > 0 but NetAmount  ≤ 0
- if ls token total supply ≤ 0 but (total unbonding balance > 0 or total liquid tokens > 0)

**Chunks invariant check broken when**

- for any Pairing chunk
  - there is a paired insurance
  - balance of chunk is smaller than ChunkSize tokens
- for any Paired chunk
  - there is no paired insurance
  - cannot find paired insurance obj
  - cannot find delegation obj
  - delegation shares < ChunkSize tokens
- for any Unpairing and UnpairingForUnstaking chunk
  - there is no unpairing insurance
  - cannot find unpairing insurance obj
  - **if it is epoch then**
    - cannot find unbonding delegation obj
    - unbonding entries ≠ 1
- for any chunk status == Unspecified

**Insurances invariant check broken when**

- for any Pairing insurance
  - there is a chunk to serve
- for any Paired insurance
  - there is no chunk to serve
  - cannot find serving chunk obj
  - serving chunk status is not Paired
- for any Unpairing insurance
  - there is no chunk to serve
  - cannot find serving chunk obj
- for any Unpaired insurance
  - there is a chunk to serve
- for any UnpairingForWithdrawal insurance
  - there is no chunk to serve
  - cannot find serving chunk obj

**UnpairingForUnstakingChunkInfos invariant check broken when**

- for any info
  - cannot find related chunk obj

**WithdrawInsuranceRequests Invariant check broken when**

- for any req
  - cannot find related insurance obj

**RedelgationInfos Invariant check broken when**

- for any info
  - cannot find related chunk obj