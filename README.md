<div align="center">

```
                    +-----------------------+
                    |    machine-agent       |
                    +-----------------------+
                    | cpu    ||||||||| 24%   |
                    | mem    ||||||    16GB  |
                    | disk   ||||||||  460GB |
                    | load   ||        0.42  |
                    | docker ||| 3 running   |
                    +-----------+-----------+
                                |
                    +-----------+-----------+
                    |    :7891/metadata      |
                    |   { "host": "pi-4",   |
                    |     "os": "linux" }    |
                    +-----------------------+
```

**A lightweight Go agent that exposes system metadata via a JSON HTTP endpoint.**

Deploy it on devices across your local network to inspect and monitor them from a central tool.

```
  [raspberry-pi]        [nas-server]        [desktop]         [your-tool]
   :7891                 :7891               :7891                 |
     |                     |                   |                   |
     +---------------------+-------------------+------- GET /metadata
```

</div>

## Install

```bash
go build -o machine-agent .
```

## Usage

```bash
./machine-agent                    # default port 7891
./machine-agent -port 9999         # custom port
TOKEN=mysecret ./machine-agent     # with auth
```

## Endpoints

### `GET /metadata`

By default returns a slim response with: `host`, `cpu`, `memory`, `load`, `network`.

Use `?include=` to request specific sections or `?include=all` for everything.

```bash
curl http://localhost:7891/metadata                          # default sections
curl http://localhost:7891/metadata?include=all              # everything
curl http://localhost:7891/metadata?include=docker,disks     # specific sections
```

#### Available sections

| Section | Default | Description |
|---------|---------|-------------|
| `host` | yes | hostname, OS, platform, kernel, uptime, boot time, process count |
| `cpu` | yes | model, cores, per-core and total usage % |
| `memory` | yes | total, used, available (bytes + human-readable), usage % |
| `load` | yes | 1, 5, 15 minute averages |
| `network` | yes | interfaces (name, MAC, IPs, flags, MTU) + IO counters |
| `swap` | no | total, used, free (bytes + human-readable), usage % |
| `disks` | no | mounted partitions with device, mountpoint, fstype, sizes |
| `processes` | no | top 50 by CPU (PID, name, status, CPU%, mem%, user, cmdline) |
| `users` | no | logged-in users with terminal and host |
| `docker` | no | server version + all containers (name, image, state, ports, labels) |

### `GET /health`

Returns `{"status":"ok"}`.

## Authentication

Set the `TOKEN` env var to require authentication on all requests. Without it, endpoints are open.

The `Authorization: Bearer` header is preferred, but a `?token=` query param is also supported for convenience (e.g., quick browser access).

```bash
TOKEN=mysecret ./machine-agent

# Header (preferred)
curl -H "Authorization: Bearer mysecret" http://192.168.1.10:7891/metadata

# Query param (convenient for browsers)
curl "http://192.168.1.10:7891/metadata?token=mysecret"
```

## Example

```bash
curl -s http://localhost:7891/metadata?include=memory | jq .
```

```json
{
  "memory": {
    "total_bytes": 17179869184,
    "total": "16.0 GB",
    "used_bytes": 13659144192,
    "used": "12.7 GB",
    "available_bytes": 3520724992,
    "available": "3.3 GB",
    "used_percent": 79.5
  },
  "collected_at": "2026-04-18T12:12:22Z"
}
```

## License

MIT
