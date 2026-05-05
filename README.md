# relayctl

Lightweight reverse proxy manager with hot-reload and rule-based routing.

---

## Installation

```bash
go install github.com/yourname/relayctl@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/relayctl.git && cd relayctl && go build -o relayctl .
```

---

## Usage

Define your routing rules in a `relayctl.yaml` config file:

```yaml
listen: ":8080"
routes:
  - match: "/api/"
    upstream: "http://localhost:3000"
  - match: "/static/"
    upstream: "http://localhost:4000"
```

Start the proxy:

```bash
relayctl start --config relayctl.yaml
```

Reload rules without downtime:

```bash
relayctl reload
```

Check proxy status:

```bash
relayctl status
```

---

## Features

- **Hot-reload** — update routing rules with zero downtime
- **Rule-based routing** — path prefix and header matching
- **Lightweight** — single binary, no runtime dependencies
- **Simple config** — plain YAML, no boilerplate

---

## License

MIT © 2024 yourname