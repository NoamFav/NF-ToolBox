<div align="center">

<img src="https://capsule-render.vercel.app/api?type=venom&height=220&color=gradient&customColorList=12&text=NF-TOOLBOX&fontSize=70&fontColor=fff&animation=twinkling&desc=License%20Server%20%2B%20CLI%20for%20the%20NF%20Software%20Ecosystem&descSize=16&descAlignY=65&stroke=FFFFFF&strokeWidth=1" alt="NF-ToolBox Banner" />

<img src="https://readme-typing-svg.herokuapp.com?font=Fira+Code&size=20&pause=1000&color=00D9FF&center=true&vCenter=true&multiline=true&repeat=true&width=900&height=60&lines=Auth+%E2%86%92+License+%E2%86%92+Device+Activation+%E2%86%92+Renewal;Go+API+%C2%B7+Postgres+%C2%B7+Stripe+webhooks+%C2%B7+CLI+toolbox" alt="Typing SVG" />

<br>

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white&labelColor=0D1117)](https://go.dev)
[![Postgres](https://img.shields.io/badge/Postgres-4169E1?style=for-the-badge&logo=postgresql&logoColor=white&labelColor=0D1117)](https://www.postgresql.org)
[![Stripe](https://img.shields.io/badge/Stripe-Webhooks-635BFF?style=for-the-badge&logo=stripe&logoColor=white&labelColor=0D1117)](https://stripe.com)
[![License](https://img.shields.io/badge/Proprietary-FF4444?style=for-the-badge&labelColor=0D1117)](#license)

</div>

<img src="https://user-images.githubusercontent.com/73097560/115834477-dbab4500-a447-11eb-908a-139a6edaec5c.gif" width="100%">

<div align="center">
  <img src="https://readme-typing-svg.herokuapp.com?font=Orbitron&size=26&pause=1000&color=00D9FF&center=true&width=800&lines=%F0%9F%A4%96+WHAT+IS+NF-TOOLBOX+%3F" alt="What is NF-ToolBox" />
</div>
<br>

NF-ToolBox is the license backbone for the NF Software tool line: a Go API server that issues and verifies licenses, a CLI (`nf-toolbox`) end users install tools through, and a shared Go package (`pkg/license`) any NF tool embeds to check entitlement at startup.

<img src="https://user-images.githubusercontent.com/73097560/115834477-dbab4500-a447-11eb-908a-139a6edaec5c.gif" width="100%">

<div align="center">
  <img src="https://readme-typing-svg.herokuapp.com?font=Orbitron&size=26&pause=1000&color=FF69B4&center=true&width=800&lines=%E2%9A%A1+COMPONENTS+%E2%9A%A1" alt="Components" />
</div>
<br>

| Component | Path | Purpose |
|-----------|------|---------|
| **Server** | `cmd/server/` | License API — auth, activation, renewal, Stripe webhooks |
| **Toolbox CLI** | `cmd/toolbox/` | `login · list · install · update · sync · devices · status` |
| **License package** | `pkg/license/` | Fingerprinting, signing, verification — embed in any NF tool |
| **Keygen** | `cmd/keygen/` | Generates signing keys for the server |

<img src="https://user-images.githubusercontent.com/73097560/115834477-dbab4500-a447-11eb-908a-139a6edaec5c.gif" width="100%">

<div align="center">
  <img src="https://readme-typing-svg.herokuapp.com?font=Orbitron&size=26&pause=1000&color=6A5ACD&center=true&width=800&lines=%E2%9C%A8+QUICKSTART+%E2%9C%A8" alt="Quickstart" />
</div>
<br>

```sh
createdb license_db
psql -d license_db < migrations/001_initial.sql

go run cmd/keygen/main.go        # generate keys, add to .env
cp .env.example .env

make run                          # start the server
go build -o bin/nf-toolbox cmd/toolbox/main.go
```

```sh
nf-toolbox login
nf-toolbox install <tool>
nf-toolbox sync        # renew license after purchase
nf-toolbox status
```

<details>
<summary><b>🔌 API Endpoints</b></summary>
<br>

| Endpoint | Auth | Purpose |
|----------|------|---------|
| `POST /api/v1/auth/register` · `/login` | — | Account creation & login |
| `POST /api/v1/activate` · `/renew` | Bearer | License activation / renewal |
| `GET /api/v1/license` · `/devices` · `/tools` · `/me` | Bearer | Read current state |
| `DELETE /api/v1/devices/{id}` | Bearer | Deactivate a device |
| `POST /webhooks/stripe` | Stripe sig | Payment events |

</details>

<details>
<summary><b>🧩 Integrating a tool</b></summary>
<br>

```go
import "github.com/nf-software/nf-toolbox/pkg/license"

func main() {
    if err := license.CheckEntitlement("your-tool-name"); err != nil {
        fmt.Fprintln(os.Stderr, "Run: nf-toolbox login && nf-toolbox install your-tool")
        os.Exit(1)
    }
}
```

</details>

<img src="https://user-images.githubusercontent.com/73097560/115834477-dbab4500-a447-11eb-908a-139a6edaec5c.gif" width="100%">

<div align="center">

<img src="https://readme-typing-svg.herokuapp.com?font=Orbitron&size=20&pause=1000&color=6A5ACD&center=true&width=800&lines=Thanks+for+stopping+by!" alt="Footer typing" />

<br>

Made with ♥ by [NoamFav](https://github.com/NoamFav) · Proprietary — NF Software

<img src="https://capsule-render.vercel.app/api?type=waving&height=100&color=gradient&customColorList=12&section=footer" />

</div>
