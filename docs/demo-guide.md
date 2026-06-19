# Demo-Guide

Dieser Ablauf zeigt den Server mit echter PostgreSQL-Datenbank.

## 1. Datenbank starten

```powershell
docker compose -f extras/compose/postgres/compose.yml up -d
docker compose -f extras/compose/postgres/compose.yml ps
```

Erwartung: Der Container `workshop-gebrauchtwagen-db` ist `healthy`.

## 2. Tests ausfuehren

```powershell
go test ./...
```

Optional mit echter Datenbank:

```powershell
$env:INTEGRATION_DATABASE_URL="postgres://gebrauchtwagen:gebrauchtwagen@localhost:5432/gebrauchtwagen?sslmode=disable"
go test ./internal/repository
```

## 3. Server starten

```powershell
$env:DATABASE_URL="postgres://gebrauchtwagen:gebrauchtwagen@localhost:5432/gebrauchtwagen?sslmode=disable"
go run ./cmd/server
```

Erwartung: Die Konsole zeigt `server listening`.

## 4. Healthchecks pruefen

```powershell
Invoke-RestMethod http://localhost:3000/health/liveness
Invoke-RestMethod http://localhost:3000/health/readiness
```

Erwartung: Beide Antworten enthalten `status = UP`.

## 5. Liste lesen

```powershell
Invoke-RestMethod http://localhost:3000/api/gebrauchtwagen
```

Erwartung: Die drei Seed-Fahrzeuge `VW Golf`, `BMW 320d` und `Audi A4 Avant` werden geliefert.

## 6. Detail mit Relationen lesen

```powershell
Invoke-RestMethod http://localhost:3000/api/gebrauchtwagen/1 | ConvertTo-Json -Depth 8
```

Erwartung: Die Antwort enthaelt neben dem Fahrzeug auch `standort`, `hauptuntersuchung` und `schaeden`.

## 7. Validierungsfehler zeigen

```powershell
Invoke-WebRequest http://localhost:3000/api/gebrauchtwagen `
  -Method Post `
  -ContentType "application/json" `
  -Body '{"marke":"","modell":"","fahrzeugklasse":"FALSCH","kraftstoffart":"BENZIN","schadenfrei":true,"kilometerstand":-1}'
```

Erwartung: Status `422` mit `application/problem+json`.

## 8. Neues Fahrzeug anlegen

```powershell
Invoke-WebRequest http://localhost:3000/api/gebrauchtwagen `
  -Method Post `
  -ContentType "application/json" `
  -Body '{"marke":"Mercedes","modell":"C 200","fahrzeugklasse":"MITTELKLASSE","kraftstoffart":"BENZIN","schadenfrei":true,"kilometerstand":18000}'
```

Erwartung: Status `201 Created` und ein `Location`-Header, z.B. `/api/gebrauchtwagen/4`.

## 9. Update mit Version zeigen

Erst Detail abrufen und `ETag` merken:

```powershell
Invoke-WebRequest http://localhost:3000/api/gebrauchtwagen/1
```

Dann mit passendem `If-Match` aktualisieren:

```powershell
Invoke-WebRequest http://localhost:3000/api/gebrauchtwagen/1 `
  -Method Put `
  -ContentType "application/json" `
  -Headers @{ "If-Match" = '"1"' } `
  -Body '{"marke":"VW","modell":"Golf Variant","fahrzeugklasse":"KOMBI","kraftstoffart":"BENZIN","schadenfrei":true,"kilometerstand":43000}'
```

Erwartung: Status `204 No Content` und ein neuer `ETag`.
