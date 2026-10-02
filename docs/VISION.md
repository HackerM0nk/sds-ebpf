# Vision

## Project Statement

SDS eBPF aims to become an open runtime systems intelligence artifact: a platform that continuously derives a living model of software systems from observed behavior rather than declared intent.

The core idea is simple:

- modern systems are too dynamic to be understood through diagrams, manifests, and tribal knowledge alone,
- runtime behavior is the closest thing we have to operational truth,
- and that runtime truth can be turned into a durable, queryable representation of system structure.

This project exists to build that representation.

## The Problem

Infrastructure understanding is currently fragmented across:

- deployment manifests,
- service registries,
- logs,
- metrics,
- traces,
- dashboards,
- architecture documents,
- and the memory of the people who operate the system.

Each of these captures a slice of reality. None of them reliably maintain a complete, current, machine-readable model of what is actually running, how it is connected, what changed, and what depends on what.

That gap has direct consequences:

- architecture drifts faster than documentation,
- dependencies remain partially implicit,
- incident response is slower than it should be,
- security teams struggle to reason about exposure and blast radius,
- and platform teams lack a trustworthy substrate for higher-level automation.

## What We Are Building

We are building a runtime-native substrate for system understanding.

At maturity, SDS eBPF should do five things well:

1. Observe
   Capture low-level runtime evidence across hosts, containers, and orchestrated environments with low overhead.

2. Resolve
   Convert raw events into stable identities for processes, workloads, services, endpoints, privileges, and infrastructure context.

3. Reconstruct
   Build a temporal graph that represents how the system is actually executing and communicating.

4. Explain
   Expose that graph through diagrams, queries, reports, and operator-facing workflows.

5. Enable
   Serve as a substrate for architecture analysis, security reasoning, drift detection, documentation, and grounded assistant interfaces.

## What Makes The Project Interesting

The project is not interesting because it uses eBPF.

It is interesting because it treats runtime evidence as the foundation for a continuously maintained systems model.

That model is more important than the sensors themselves.

The long-term value of the project depends on three properties:

- the evidence is trustworthy,
- the reconstructed model is useful,
- and the artifact is reproducible enough that others can evaluate, extend, and apply it.

## Core Beliefs

### Runtime Over Declared Intent

Configuration is useful context, not ground truth.

The system should privilege observed execution and communication patterns over intended topology.

### Systems Need A Native Representation

A modern distributed system should have a machine-readable representation of itself that can be updated from evidence, queried by tools, and inspected by humans.

### Evidence And Inference Must Be Separated

Raw observation, identity resolution, graph inference, and explanation are not the same layer.

The project should preserve that separation so confidence, provenance, and error can be reasoned about explicitly.

### Ambiguity Must Be Modeled, Not Hidden

Many relationships in a live system are only probabilistically knowable.

The artifact should track uncertainty, confidence, and provenance instead of pretending every mapping is exact.

### Low Overhead Matters

If the system is too expensive to run, it stops being a broadly useful substrate.

Efficiency is not a nice-to-have; it is a core design constraint.

## Design Commitments

These are the commitments the system should honor as it evolves.

- Evidence-first architecture
- Explicit schema and versioning
- Temporal data model
- Replayable and reproducible workflows
- Incremental graph updates
- Grounded outputs over speculative summaries
- Honest limitations and measurable tradeoffs

## Intended Users

This artifact should serve multiple serious users without collapsing into generic platform marketing.

### Platform And Infrastructure Engineers

They need to understand topology, dependencies, runtime drift, and operational coupling.

### Security Engineers

They need to understand exposure, privilege boundaries, external communication, and blast radius.

### SRE And Incident Responders

They need to reconstruct what changed, what was talking to what, and what downstream impact a failure may have had.

### Researchers And Tool Builders

They need a reusable substrate for evaluating runtime reconstruction, graph-building methods, and downstream systems reasoning.

## System Boundaries

The platform should be broad, but not vague.

It is reasonable for SDS eBPF to own:

- low-level runtime sensing,
- canonical event modeling,
- identity resolution and enrichment,
- temporal graph construction,
- architecture and dependency analysis,
- and grounded higher-level interfaces on top of that graph.

It should not try to become every adjacent category.

## Non-Goals

At least in its core form, this project is not:

- a generic SIEM replacement,
- a full packet capture system,
- a clone of application-level APM tooling,
- an LLM-first observability assistant,
- or a universal solution for perfect ground truth.

Those may overlap with the project, but they should not define it.

## The Artifact We Want On GitHub

The repository should eventually feel like a serious systems project, not a vision deck.

That means the public artifact must support:

- reproducible builds,
- documented schemas,
- sample workloads,
- replayable traces,
- graph construction,
- evaluation harnesses,
- benchmark results,
- and clear artifact-review instructions.

The repo should stand on its own for:

- engineers who want to build on it,
- reviewers who want to reproduce it,
- and readers who want to understand the system design without prior context.

## The Paper We Want To Earn

The paper should emerge from the artifact, not the other way around.

A broad but principled paper from this project should argue that:

- runtime evidence can be turned into a durable systems model,
- that model can be maintained with acceptable overhead,
- and that the resulting representation enables useful classes of analysis that static descriptions do not support well on their own.

The paper can be ambitious. It just has to be anchored in a real system and a measured evaluation.

## North Star Questions

If this project succeeds, it should help answer questions like:

- What is actually running in this environment right now?
- What depends on this service or endpoint?
- What changed in the architecture over the last day?
- Which workloads are talking to infrastructure they should not be talking to?
- What is the likely blast radius if this component fails?
- Which parts of the documented architecture no longer match runtime reality?

These are the kinds of questions the system should eventually make routine.

## Build Philosophy

The right way to build this is:

- define the vision clearly,
- build the evidence path first,
- establish the data model second,
- make graph reconstruction real third,
- and only then expand into richer analysis and interfaces.

Anything that skips those steps risks producing impressive demos without durable systems value.

## Guiding Standard

Every major feature in this repository should pass a simple test:

Does it improve the system's ability to recover, maintain, or use a trustworthy runtime model of real infrastructure?

If not, it is probably not core.

## Relationship To The Rest Of The Repo

This document is the project charter.

- [Research Blueprint](./RESEARCH_BLUEPRINT.md) defines the research shape and evaluation logic.
- [Roadmap](../ROADMAP.md) defines the build sequence and artifact milestones.

The vision should stay stable. The implementation and milestones can evolve underneath it.
