# NF-ToolBox

License server and toolbox CLI for NF Software tools.

## Components

- **Server** (`cmd/server/`) - License API server
- **Toolbox** (`cmd/toolbox/`) - CLI for managing tools
- **License Package** (`pkg/license/`) - Shared verification library

## Quick Start

### 1. Setup Database

```bash
createdb license_db
psql -d license_db < migrations/001_initial.sql
```

### 2. Generate Keys

```bash
go run cmd/keygen/main.go
# Add output to .env
```

### 3. Configure

```bash
cp .env.example .env
# Edit .env with your values
```

### 4. Run Server

```bash
make run
# or
go run cmd/server/main.go
```

### 5. Build Toolbox

```bash
go build -o bin/nf-toolbox cmd/toolbox/main.go
```

## Toolbox Commands

```bash
nf-toolbox login          # Login to your account
nf-toolbox logout         # Logout
nf-toolbox list           # List all tools
nf-toolbox install <tool> # Install a tool
nf-toolbox uninstall <tool>
nf-toolbox update         # Update tools & renew license
nf-toolbox sync           # Sync license after purchase
nf-toolbox devices        # List activated devices
nf-toolbox status         # Show current status
```

## API Endpoints

### Public
- `POST /api/v1/auth/register` - Create account
- `POST /api/v1/auth/login` - Login

### Protected (requires Bearer token)
- `POST /api/v1/activate` - Activate license on device
- `POST /api/v1/renew` - Renew license
- `GET /api/v1/license` - Get current license
- `GET /api/v1/devices` - List devices
- `DELETE /api/v1/devices/{id}` - Deactivate device
- `GET /api/v1/tools` - List tools catalog
- `GET /api/v1/me` - Get user info

### Webhooks
- `POST /webhooks/stripe` - Stripe payment events

## Integrating Tools

Add to your tool's `main.go`:

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

## Deploy to Fly.io

```bash
fly launch
fly postgres create --name nf-toolbox-db
fly secrets set JWT_SECRET=xxx PRIVATE_KEY=yyy
fly deploy
```

## File Structure

```
~/.nf-tools/
├── bin/           # Installed tool binaries
├── config/
│   ├── session.json   # Auth token
│   ├── license.json   # Signed license
│   └── state.json     # Renewal state
└── cache/
    └── versions.json  # Available versions
```

## License

Proprietary - NF Software
