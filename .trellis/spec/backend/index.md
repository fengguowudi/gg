# Backend Development Guidelines

> Best practices for backend development in this project.

---

## Overview

This directory contains backend development guidelines derived from the current `gg` codebase.

The project is a Go CLI backend with config-file persistence, protocol-specific dialer packages, Linux-specific tracing code, and no database layer.

---

## Guidelines Index

| Guide | Description | Status |
|-------|-------------|--------|
| [Directory Structure](./directory-structure.md) | Go CLI module organization and file layout | Documented |
| [Database Guidelines](./database-guidelines.md) | Config persistence conventions and explicit non-use of a database layer | Documented |
| [Error Handling](./error-handling.md) | Sentinel errors, wrapping, and CLI exit boundaries | Documented |
| [Quality Guidelines](./quality-guidelines.md) | Code standards, required patterns, and testing expectations | Documented |
| [Logging Guidelines](./logging-guidelines.md) | Logrus usage, verbosity levels, and logging caveats | Documented |

---

## How to Fill These Guidelines

For each guideline file:

1. Document your project's **actual conventions** (not ideals)
2. Include **code examples** from your codebase
3. List **forbidden patterns** and why
4. Add **common mistakes** your team has made

The goal is to help AI assistants and new team members understand how YOUR project works.

---

**Language**: All documentation should be written in **English**.
