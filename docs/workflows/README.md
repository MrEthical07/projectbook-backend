# Workflow Documentation

This folder contains route-accurate workflow documentation per routed module.

Each module file describes:
- route inventory
- middleware/policy chain order
- handler -> service -> repository -> store flow
- transaction boundaries
- side effects (cache/rate-limit/auth/document-store)
- expected failure modes and troubleshooting checks

## Module Files

- [auth.md](auth.md)
- [home.md](home.md)
- [project.md](project.md)
- [artifacts.md](artifacts.md)
- [resources.md](resources.md)
- [pages.md](pages.md)
- [calendar.md](calendar.md)
- [sidebar.md](sidebar.md)
- [activity.md](activity.md)
- [team.md](team.md)
- [health.md](health.md)
- [system.md](system.md)

## Reading Order For New Contributors

1. [../architecture.md](../architecture.md)
2. [../auth.md](../auth.md)
3. [../policies.md](../policies.md)
4. module workflow file you are changing
5. [../routeDetails.md](../routeDetails.md) for generated route metadata snapshot

## Conventions Used

- Route notation: `METHOD /path`
- Policy chain order reflects runtime order as registered in `routes.go`
- Transaction notes indicate service-owned `store.WithTx(...)` boundaries
- Cache notes include read tags + invalidation scope where applicable

## Verification After Workflow Changes

```bash
go test ./...
go build ./...
go run ./cmd/superapi-verify ./...
```

If route stacks changed, regenerate route details:

```bash
go run ./cmd/routedocgen
```
