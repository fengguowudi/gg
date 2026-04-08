# Cross-Layer Thinking Guide

> **Purpose**: Think through data flow across layers before implementing.

---

## The Problem

**Most bugs happen at layer boundaries**, not within layers.

Common cross-layer bugs:
- API returns format A, frontend expects format B
- Database stores X, service transforms to Y, but loses data
- Multiple layers implement the same logic differently

---

## Before Implementing Cross-Layer Features

### Step 1: Map the Data Flow

Draw out how data moves:

```
Source -> Transform -> Store -> Retrieve -> Transform -> Display
```

For each arrow, ask:
- What format is the data in?
- What could go wrong?
- Who is responsible for validation?

### Step 2: Identify Boundaries

| Boundary | Common Issues |
|----------|---------------|
| API -> Service | Type mismatches, missing fields |
| Service -> Database | Format conversions, null handling |
| Backend -> Frontend | Serialization, date formats |
| Component -> Component | Props shape changes |

### Step 3: Define Contracts

For each boundary:
- What is the exact input format?
- What is the exact output format?
- What errors can occur?

---

## Common Cross-Layer Mistakes

### Mistake 1: Implicit Format Assumptions

**Bad**: Assuming date format without checking

**Good**: Explicit format conversion at boundaries

### Mistake 2: Scattered Validation

**Bad**: Validating the same thing in multiple layers

**Good**: Validate once at the entry point

### Mistake 3: Leaky Abstractions

**Bad**: Component knows about database schema

**Good**: Each layer only knows its neighbors

### Mistake 4: Assuming One I/O Direction Starts First

**Bad**: Building a transport handshake that only starts on the first `Write()`

This breaks protocols such as SSH where the client may block on `Read()` waiting for the server banner, or where helper goroutines inside higher-level libraries read before application payload writes happen.

**Good**: Model transport handshakes as explicit state machines and support both `read-first` and `write-first` flows when the protocol boundary requires it.

### Mistake 5: Dropping Buffered Bytes at Protocol Boundaries

**Bad**: Parsing a proxy or tunnel response with a temporary `bufio.Reader`, then switching back to raw `net.Conn` reads

This can silently discard bytes that arrived immediately after the protocol response, such as the first SSH banner bytes after an HTTP `200 Connection Established`.

**Good**: Preserve and reuse the buffered reader, or otherwise drain and replay buffered bytes across the boundary.

---

## Checklist for Cross-Layer Features

Before implementation:
- [ ] Mapped the complete data flow
- [ ] Identified all layer boundaries
- [ ] Defined format at each boundary
- [ ] Decided where validation happens
- [ ] For transport handshakes, checked whether the protocol can be `read-first`, `write-first`, or both
- [ ] For parser-to-stream transitions, checked how buffered bytes are preserved after the boundary

After implementation:
- [ ] Tested with edge cases (null, empty, invalid)
- [ ] Verified error handling at each boundary
- [ ] Checked data survives round-trip
- [ ] Tested both handshake directions when the transport is stateful
- [ ] Verified no bytes are lost immediately after protocol upgrade or tunnel establishment

---

## When to Create Flow Documentation

Create detailed flow docs when:
- Feature spans 3+ layers
- Multiple teams are involved
- Data format is complex
- Feature has caused bugs before
