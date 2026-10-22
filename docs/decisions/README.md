# Architecture decision records

Short records of significant decisions: the context, the options, what was
chosen and why. To add one, copy [0000-template.md](0000-template.md) to the
next number. Do not rewrite old records; supersede them with a new one.

| # | Decision | Status |
|---|---|---|
| [0001](0001-cli-framework.md) | Use Cobra for the command-line interface | Accepted |
| [0002](0002-sqlite-driver.md) | Use modernc.org/sqlite (pure Go) for local history | Accepted |
| [0003](0003-icmp-privilege-strategy.md) | Prefer unprivileged ICMP; never require root or Administrator | Accepted |
