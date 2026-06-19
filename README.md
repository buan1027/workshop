# Workshop Server

Go-REST-Server fuer das Gebrauchtwagen-Datenmodell aus dem Softwareengineering-Workshop.

## Ziel

Der Server stellt eine einfache REST-API fuer die Hauptressource `Gebrauchtwagen` bereit. Er ist bewusst klein gehalten: klare Schichten, direkte PostgreSQL-Anbindung und gut erklaerbare Validierung.
Der HTTP-Server verwendet konservative Timeouts und beendet sich bei `Ctrl+C` geordnet.

## Tech Stack

- Go mit `net/http` als Basis
- `github.com/go-chi/chi/v5` fuer Routing und Middleware
- `github.com/jackc/pgx/v5` fuer PostgreSQL-Zugriff
- Go-Standardbibliothek fuer Tests (`testing`, `httptest`)

`chi` wurde gewaehlt, weil es leichtgewichtig und kompatibel mit `net/http` ist. `pgx` passt gut, weil das Datenbankschema bereits in PostgreSQL existiert und SQL dadurch transparent bleibt.

## Projektstruktur

```text
cmd/server/              Einstiegspunkt des Servers
internal/config/         Konfiguration ueber Umgebungsvariablen
internal/domain/         Datenmodell, Validierung, fachliche Fehler
internal/http/           Router, Handler, Problem-Details
internal/repository/     Repository-Interface und PostgreSQL-Implementierung
```

## Konfiguration

Siehe `.env.example`.

Wichtige Variablen:

- `APP_ADDR`: HTTP-Adresse, Standard `:3000`
- `DATABASE_URL`: PostgreSQL-Verbindung
- `ADMIN_TOKEN`: optionaler Bearer Token fuer schreibende Endpunkte

Beispiel:

```powershell
$env:APP_ADDR=":3000"
$env:DATABASE_URL="postgres://gebrauchtwagen:gebrauchtwagen@localhost:5432/gebrauchtwagen?sslmode=disable"
go run ./cmd/server
```

## Starten

Voraussetzung: Go ist installiert und eine passende PostgreSQL-Datenbank mit dem Schema `gebrauchtwagen` laeuft.

```powershell
go mod tidy
go run ./cmd/server
```

Lokale PostgreSQL-Datenbank starten:

```powershell
docker compose -f extras/compose/postgres/compose.yml up -d
```

Healthcheck:

```powershell
Invoke-RestMethod http://localhost:3000/health/liveness
Invoke-RestMethod http://localhost:3000/health/readiness
```

## REST-Endpunkte

```text
GET    /health/liveness
GET    /health/readiness
GET    /api/gebrauchtwagen
GET    /api/gebrauchtwagen/{id}
POST   /api/gebrauchtwagen
PUT    /api/gebrauchtwagen/{id}
DELETE /api/gebrauchtwagen/{id}
```

Filter fuer die Liste:

```text
marke, modell, fahrzeugklasse, kraftstoffart, schadenfrei, page, size, count-only
```

Beispiel fuer Neuanlage:

```powershell
Invoke-RestMethod http://localhost:3000/api/gebrauchtwagen `
  -Method Post `
  -ContentType "application/json" `
  -Body '{"marke":"VW","modell":"Golf","fahrzeugklasse":"KOMPAKTKLASSE","kraftstoffart":"BENZIN","schadenfrei":true,"kilometerstand":12000}'
```

Weitere Beispielrequests liegen in `docs/requests.http`.
Ein kompakter Vorfuehrablauf liegt in `docs/demo-guide.md`.
Eine schlanke OpenAPI-Beschreibung liegt in `docs/openapi.yaml`.

## Validierung und Fehler

Beim Schreiben werden Pflichtfelder und Enum-Werte validiert:

- `marke` und `modell` duerfen nicht leer sein
- `fahrzeugklasse` muss einem bekannten Enum entsprechen
- `kraftstoffart` muss einem bekannten Enum entsprechen
- `kilometerstand` muss mindestens `0` sein

Fehler werden als `application/problem+json` zurueckgegeben. Beispiele:

- `400 Bad Request`: ungueltiges JSON
- `404 Not Found`: ID nicht gefunden
- `422 Unprocessable Entity`: Validierungsfehler
- `428 Precondition Required`: `If-Match` fehlt bei `PUT`
- `412 Precondition Failed`: Versionskonflikt bei `PUT`

## Tests

```powershell
go test ./...
```

Die vorhandenen Tests pruefen Validierung, Healthcheck, Create, Detail mit ETag und Fehlerfaelle. Die HTTP-Tests nutzen ein Fake-Repository und brauchen keine echte Datenbank.
Auf GitHub fuehrt der Workflow `.github/workflows/go.yml` Formatpruefung und Tests automatisch aus.

Optionaler Integrationstest gegen PostgreSQL:

```powershell
$env:INTEGRATION_DATABASE_URL="postgres://gebrauchtwagen:gebrauchtwagen@localhost:5432/gebrauchtwagen?sslmode=disable"
go test ./internal/repository
```

## Demo-Ablauf

```powershell
docker compose -f extras/compose/postgres/compose.yml up -d
go run ./cmd/server
```

In einem zweiten Terminal:

```powershell
Invoke-RestMethod http://localhost:3000/health/readiness
Invoke-RestMethod http://localhost:3000/api/gebrauchtwagen
```

Wenn `ADMIN_TOKEN` gesetzt ist, brauchen `POST`, `PUT` und `DELETE` den Header `Authorization: Bearer <token>`.

## Docker-Image

```powershell
docker build -t workshop-server .
docker run --rm -p 3000:3000 `
  -e DATABASE_URL="postgres://gebrauchtwagen:gebrauchtwagen@host.docker.internal:5432/gebrauchtwagen?sslmode=disable" `
  workshop-server
```

## Naechste sinnvolle Erweiterungen

- Integrationstest gegen echte PostgreSQL-Datenbank
- Relationen `Standort`, `Schaden` und `Hauptuntersuchung` in Detailantworten aufnehmen
- Keycloak/OIDC nur als optionaler Zusatz, wenn der Kern stabil laeuft
