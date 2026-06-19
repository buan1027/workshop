# Workshop Server

Go-REST-Server fuer das Gebrauchtwagen-Datenmodell aus dem Softwareengineering-Workshop.

## Ziel

Der Server stellt eine einfache REST-API fuer die Hauptressource `Gebrauchtwagen` bereit. Er ist bewusst klein gehalten: klare Schichten, direkte PostgreSQL-Anbindung und gut erklaerbare Validierung.
Der HTTP-Server verwendet konservative Timeouts und beendet sich bei `Ctrl+C` geordnet.

Dieses Repository ist das Ergebnis des Programmierworkshops am 19.6.2026. Die abgabeorientierte Zusammenfassung nach Vorgabe liegt in `docs/workshop-abgabe.md`.

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
internal/service/        Use Cases und Aufruf der fachlichen Validierung
internal/repository/     Repository-Interface und PostgreSQL-Implementierung
internal/database/       Eingebettetes SQL zum Zuruecksetzen der Demo-Daten
internal/auth/           Optionaler Schreibschutz mit ADMIN_TOKEN oder Keycloak/OIDC
```

Eine ausfuehrlichere Architekturbeschreibung liegt in `docs/architecture.md`.

## Konfiguration

Siehe `.env.example`.

Wichtige Variablen:

- `APP_ADDR`: HTTP-Adresse, Standard `:3000`
- `DATABASE_URL`: PostgreSQL-Verbindung
- `RESET_DATABASE_ON_START`: setzt Demo-Daten beim Serverstart neu, Standard `true`
- `AUTH_MODE`: Authentifizierungsmodus fuer schreibende Endpunkte, Standard `admin-token`
- `ADMIN_TOKEN`: optionaler Bearer Token fuer schreibende Endpunkte
- `KEYCLOAK_ISSUER_URL`: optionaler Keycloak-Issuer fuer `AUTH_MODE=keycloak`
- `KEYCLOAK_CLIENT_ID`: optionale Audience-Pruefung fuer `AUTH_MODE=keycloak`

Beispiel:

```powershell
$env:APP_ADDR=":3000"
$env:DATABASE_URL="postgres://gebrauchtwagen:gebrauchtwagen@localhost:5432/gebrauchtwagen?sslmode=disable"
$env:RESET_DATABASE_ON_START="true"
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

Das ist kein Mock und keine In-Memory-Datenbank: Der Compose-Stack startet einen echten `postgres:17`-Container. Das Schema und die Beispieldaten werden beim ersten Start aus `extras/compose/postgres/init/` geladen.
Zusaetzlich setzt der Server beim Start standardmaessig den Demo-Datenbestand neu. Dadurch ist der Zustand nach jedem Serverstart wieder definiert, wie in den vorherigen Abgaben. Fuer laengerlebige lokale Daten kann das Verhalten mit `RESET_DATABASE_ON_START=false` deaktiviert werden.

Kompletten Docker-Stack mit App starten:

```powershell
docker compose -f extras/compose/postgres/compose.yml --profile app up -d --build
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

Schluessel:

- `id` ist der automatisch vergebene technische Primaerschluessel der Datenbank und wird fuer REST-Pfade wie `/api/gebrauchtwagen/{id}` verwendet.
- `fin` ist 17-stellig und eindeutig. Sie ist der fachliche Fahrzeugschluessel, den ein Client stabil anzeigen und wiedererkennen kann.

Filter fuer die Liste:

```text
marke, modell, fahrzeugklasse, kraftstoffart, schadenfrei, page, size, count-only
```

Paging-Beispiel:

```powershell
Invoke-RestMethod "http://localhost:3000/api/gebrauchtwagen?page=1&size=2"
```

Die Antwort enthaelt `data`, `total`, `page` und `size`.

Beispiel fuer Neuanlage:

```powershell
Invoke-RestMethod http://localhost:3000/api/gebrauchtwagen `
  -Method Post `
  -ContentType "application/json" `
  -Body '{"fin":"WVWZZZ1JZXW000001","marke":"VW","modell":"Golf","fahrzeugklasse":"KOMPAKTKLASSE","kraftstoffart":"BENZIN","schadenfrei":true,"kilometerstand":12000}'
```

Weitere Beispielrequests liegen in `docs/requests.http`.
Ein kompakter Vorfuehrablauf liegt in `docs/demo-guide.md`.
Eine schlanke OpenAPI-Beschreibung liegt in `docs/openapi.yaml`.
Eine Bruno-Collection fuer manuelle API-Tests liegt in `bruno/`.

## Validierung und Fehler

Beim Schreiben werden Pflichtfelder und Enum-Werte validiert:

- `marke` und `modell` duerfen nicht leer sein
- `fin` muss genau 17 Zeichen lang sein
- `fahrzeugklasse` muss einem bekannten Enum entsprechen
- `kraftstoffart` muss einem bekannten Enum entsprechen
- `kilometerstand` muss mindestens `0` sein

Fehler werden als `application/problem+json` zurueckgegeben. Beispiele:

- `400 Bad Request`: ungueltiges JSON
- `404 Not Found`: ID nicht gefunden
- `422 Unprocessable Entity`: Validierungsfehler
- `428 Precondition Required`: `If-Match` fehlt bei `PUT`
- `412 Precondition Failed`: Versionskonflikt bei `PUT`

Die optimistische Synchronisation erfolgt ueber die Spalte `version` und ETags wie `"1"`. `GET /api/gebrauchtwagen/{id}` liefert den aktuellen `ETag`; `PUT` erwartet diesen Wert im Header `If-Match` und erhoeht die Version bei erfolgreicher Aktualisierung. Aus Kompatibilitaetsgruenden akzeptiert der Server auch schwache ETags wie `W/"1"`.

## Tests

```powershell
go test ./...
```

Die vorhandenen Tests pruefen Validierung, Healthcheck, Create, Detail mit ETag, optimistische Versionierung, Paging und Fehlerfaelle. Die HTTP-Tests nutzen bewusst ein Fake-Repository, damit Handler-Fehler schnell und isoliert getestet werden koennen. Der produktive Serverpfad nutzt dagegen immer `pgx` und PostgreSQL.
Auf GitHub fuehrt der Workflow `.github/workflows/go.yml` Formatpruefung und Tests automatisch aus.

## Linting und statische Analyse

Lokale Standardpruefung:

```powershell
.\scripts\check.ps1
```

Enthalten sind:

- `gofmt -l .` fuer Formatierung
- `go vet ./...` fuer offizielle Go-Analyse
- `go test ./...` fuer Tests
- `staticcheck ./...` ueber `go run` fuer zusaetzliche statische Analyse
- `govulncheck ./...` ueber `go run` fuer bekannte Sicherheitsluecken

Falls gerade kein Netzwerk verfuegbar ist:

```powershell
.\scripts\check.ps1 -SkipOnlineTools
```

Abhaengigkeiten aktualisieren:

```powershell
go get -u ./...
go mod tidy
.\scripts\check.ps1
```

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

## Authentifizierung

Standardmaessig bleibt der Server workshop-freundlich lauffaehig:

- Ohne `ADMIN_TOKEN` sind schreibende Endpunkte offen.
- Mit `ADMIN_TOKEN` brauchen `POST`, `PUT` und `DELETE` den passenden Bearer Token.
- Mit `AUTH_MODE=keycloak` prueft der Server Bearer Tokens gegen `KEYCLOAK_ISSUER_URL`.

Wenn `AUTH_MODE=keycloak` gesetzt ist, aber Keycloak nicht erreichbar oder nicht konfiguriert ist, startet der Server weiter und faellt sichtbar geloggt auf den `ADMIN_TOKEN`-Modus zurueck. Fuer Produktivbetrieb waere ein harter Fehler sinnvoller; fuer den Workshop bleibt so der Kernserver benutzbar.

Beispiel fuer Keycloak:

```powershell
$env:AUTH_MODE="keycloak"
$env:KEYCLOAK_ISSUER_URL="http://localhost:8080/realms/gebrauchtwagen"
$env:KEYCLOAK_CLIENT_ID="workshop-server"
go run ./cmd/server
```

## Docker-Image

```powershell
docker build -t workshop-server .
docker run --rm -p 3000:3000 `
  -e DATABASE_URL="postgres://gebrauchtwagen:gebrauchtwagen@host.docker.internal:5432/gebrauchtwagen?sslmode=disable" `
  workshop-server
```

## Umgesetzter Stand und optionale Erweiterungen

Umgesetzt:

- REST-Endpunkte fuer Healthchecks und Gebrauchtwagen-CRUD
- Suche mit Filtern, Paging und `count-only`
- Detailantwort mit `Standort`, `Schaden` und `Hauptuntersuchung`
- Neuanlage mit optionalen relationalen Daten in einer Transaktion
- Automatisches Zuruecksetzen der Demo-Daten beim Serverstart
- Optimistische Synchronisation mit `version`, `ETag`, `If-None-Match` und `If-Match`
- Optionaler Schreibschutz mit `ADMIN_TOKEN` oder Keycloak/OIDC
- Unit-, Handler- und optionale PostgreSQL-Integrationstests
- Dockerfile, Docker Compose, Bruno-Collection, OpenAPI-Beschreibung und GitHub Actions
- Linting und statische Analyse mit `gofmt`, `go vet`, `staticcheck` und `govulncheck`

Optional, falls noch Zeit bleibt:

- Vollstaendiger Keycloak-Test inklusive Realm-/Client-/User-Export fuer reproduzierbare Imports
- Ausfuehrlicher Image-Scan, z.B. Trivy oder OWASP Dependency Check im CI
