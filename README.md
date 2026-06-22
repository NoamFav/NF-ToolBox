# 🧰 NF-ToolBox

<div align="center">

<img src="https://img.shields.io/badge/go-1.21+-00ADD8.svg?style=for-the-badge&logo=go" alt="Go">
<img src="https://img.shields.io/badge/license-Proprietary-orange.svg?style=for-the-badge" alt="License">

**License server and toolbox CLI for NF Software tools**

[Setup](#setup) · [Toolbox Commands](#toolbox-commands) · [API](#api-endpoints) · [Deploy](#deploy)

</div>

---

NF-ToolBox is a license management system consisting of a Go API server, a CLI toolbox for end users, and a shared license verification library that any NF tool can embed.

---

## Components

| Component | Path | Purpose |
|-----------|------|---------|
| **Server** | `cmd/server/` | License API server |
| **Toolbox** | `cmd/toolbox/` | CLI for managing tools |
| **License Package** | `pkg/license/` | Shared verification library |

---

## Setup

```bash
# 1. Create database
createdb license_db
psql -d license_db < migrations/001_initial.sql

# 2. Generate keys
go run cmd/keygen/main.go
# Add output to .env

# 3. Configure
cp .env.example .env

# 4. Run server
make run

# 5. Build toolbox
go build -o bin/nf-toolbox cmd/toolbox/main.go
```

---

## Toolbox Commands

```bash
nf-toolbox login              # Login to your account
nf-toolbox logout             # Logout
nf-toolbox list               # List all tools
nf-toolbox install <tool>     # Install a tool
nf-toolbox uninstall <tool>   # Uninstall a tool
nf-toolbox update             # Update tools and renew license
nf-toolbox sync               # Sync license after purchase
nf-toolbox devices            # List activated devices
nf-toolbox status             # Show current status
```

---

## API Endpoints

### Public
- `POST /api/v1/auth/register` — Create account
- `POST /api/v1/auth/login` — Login

### Protected (Bearer token required)
- `POST /api/v1/activate` — Activate license on device
- `POST /api/v1/renew` — Renew license
- `GET  /api/v1/license` — Get current license
- `GET  /api/v1/devices` — List devices
- `DELETE /api/v1/devices/{id}` — Deactivate device
- `GET  /api/v1/tools` — List tools catalog
- `GET  /api/v1/me` — Get user info

### Webhooks
- `POST /webhooks/stripe` — Stripe payment events

---

## Integrating Tools

```go
import "github.com/nf-software/nf-toolbox/pkg/license"

func main() {
    if err := license.CheckEntitlement("your-tool-name"); err != nil {
        fmt.Fprintf(os.Stderr, "License error: %v\n", err)
        fmt.Fprintln(os.Stderr, "Run: nf-toolbox login && nf-toolbox install your-tool")
        os.Exit(1)
    }
    // Your tool code...
}
```

---

## Deploy

```bash
fly launch
fly postgres create --name nf-toolbox-db
fly secrets set JWT_SECRET=xxx PRIVATE_KEY=yyy
fly deploy
```

---

## File Structure

```
~/.nf-tools/
├── bin/                # Installed tool binaries
├── config/
│   ├── session.json    # Auth token
│   ├── license.json    # Signed license
│   └── state.json      # Renewal state
└── cache/
    └── versions.json   # Available versions
```

---

## License

Proprietary — NF Software

---

<div align="center">
Made with ❤️ by <a href="https://github.com/NoamFav">NoamFav</a>
</div>
