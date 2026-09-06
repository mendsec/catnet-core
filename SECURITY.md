# Security Policy

## Scope

This policy covers every repository in the [CatNet](https://github.com/catnet-io) ecosystem:

| Repository | Role |
| :--- | :--- |
| [`engine`](https://github.com/catnet-io/engine) | Shared Go scanning engine |
| [`catnet`](https://github.com/catnet-io/catnet) | CLI |
| [`app`](https://github.com/catnet-io/app) | Desktop GUI (Wails + React) |
| [`tui`](https://github.com/catnet-io/tui) | Terminal UI (Bubble Tea) |

## Supported versions

Only the latest release of each repository receives security updates.

## Reporting a vulnerability

**Do not open a public issue for a security vulnerability.**

Report it privately through either channel:

- **GitHub Security Advisories** — [open a private report](https://github.com/catnet-io/engine/security/advisories/new) *(preferred)*
- **Email** — <fabiomendes@mailfence.com>

Please include the affected repository and version, reproduction steps, and the impact you
observed. Expect an acknowledgement within 5 business days.

## Disclosure

Fixes are developed privately and released before public disclosure. Reporters are credited in
the advisory unless they ask otherwise.

## Responsible use

CatNet is network scanning software. Use it only against hosts and networks for which you have
explicit authorization. Unauthorized scanning may be illegal in your jurisdiction.
