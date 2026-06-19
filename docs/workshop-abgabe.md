# Programmierworkshop am 19.6.2026

## Namen

TODO

## Link zum Git-Repository

https://github.com/buan1027/workshop

## KI-Werkzeuge

- ChatGPT/Codex als Coding-Agent fuer Analyse, Framework-Auswahl, Implementierung, Tests und Dokumentation.

### Agenten

- Codex im lokalen Workspace `C:\Users\anna\dev\workshop`.

### Chat-URLs

- TODO: Chat-URL aus der verwendeten Sitzung eintragen, falls gefordert.

## Frameworks und Bibliotheken

### REST-Schnittstelle (Lesen und Neuanlegen)

- `net/http` aus der Go-Standardbibliothek als HTTP-Basis.
- `github.com/go-chi/chi/v5` fuer Routing, Route-Gruppen und Middleware.

Begruendung: `chi` ist leichtgewichtig, gut dokumentiert und bleibt nah an idiomatischem Go. Dadurch ist die Loesung im Workshop gut erklaerbar.

### Validierung (nur Neuanlegen)

- Manuelle Validierung in `internal/domain`, aufgerufen aus `internal/service`.

Begruendung: Fuer das kleine Datenmodell sind klare eigene Regeln einfacher zu erklaeren als ein zusaetzliches Validation-Framework. Validiert werden Pflichtfelder, die 17-stellige FIN, Enum-Werte, Datumsfelder, `kilometerstand >= 0` und optionale Relationsdaten. Die Service-Schicht fuehrt diese Validierung aus, bevor das Repository PostgreSQL aufruft.

### OR-Mapping (fuer PostgreSQL)

- Kein klassisches ORM.
- `github.com/jackc/pgx/v5` mit `pgxpool` fuer direkten PostgreSQL-Zugriff.
- Echtes PostgreSQL-Backend ueber Docker Compose (`postgres:17`) mit Schema- und Seed-Skripten.

Begruendung: Das Datenbankschema existiert bereits. Direkte SQL-Queries sind transparent, schnell umzusetzen und vermeiden Mapping-Probleme mit PostgreSQL-Enums. Damit entspricht der Server dem DB-Server-Ansatz aus den vorherigen Abgaben, ohne fuer den Workshop zusaetzliche ORM-Komplexitaet einzufuehren.

### Optional: OIDC mit Keycloak

- Nicht umgesetzt.
- Stattdessen gibt es optional `ADMIN_TOKEN`: Wenn gesetzt, brauchen schreibende Endpunkte `Authorization: Bearer <token>`.

Begruendung: Keycloak war optional und haette in vier Stunden zu viel Zeit vom Kernserver abgezogen. Der einfache Token-Schutz ist kein vollwertiger OIDC-Ersatz, zeigt aber, wo Security-Middleware im Server sitzt.

### Einfacher Integrationstest

- Unit-/Handler-Tests laufen mit `go test ./...`.
- Optionaler PostgreSQL-Integrationstest:

```powershell
$env:INTEGRATION_DATABASE_URL="postgres://gebrauchtwagen:gebrauchtwagen@localhost:5432/gebrauchtwagen?sslmode=disable"
go test ./internal/repository
```

### Demo-Daten

- Beim Serverstart wird der Datenbestand standardmaessig neu befuellt.
- Das entspricht dem Demo-Verhalten aus den vorherigen Abgaben.
- Abschaltbar mit `RESET_DATABASE_ON_START=false`, falls Daten lokal erhalten bleiben sollen.

### Linting und statische Codeanalyse

- `gofmt` fuer einheitliche Formatierung.
- `go vet` als offizielles Go-Analysewerkzeug fuer verdaechtige Konstrukte.
- `staticcheck` fuer zusaetzliche statische Analyse, Bugs, Performance- und Vereinfachungshinweise.
- `govulncheck` fuer bekannte Sicherheitsluecken in tatsaechlich verwendeten Go-Abhaengigkeiten.

Lokal gebuendelt in:

```powershell
.\scripts\check.ps1
```

### Bruno

- Manuelle REST-Requests liegen in `bruno/`.
- Die Umgebung `local` verwendet `baseUrl = http://localhost:3000`.
- Der optionale `adminToken` kann gesetzt werden, wenn der Server mit `ADMIN_TOKEN` gestartet wurde.

## Prompts/Requests an KI-Agent/en

- Aufgabenstellung analysieren und Ressourcen, Endpunkte, Validierungen und Risiken ableiten.
- Geeignete Go-Frameworks fuer REST, Datenbankzugriff, Validierung und Tests vergleichen.
- Empfohlene Loesung mit `chi` und `pgx` umsetzen.
- Projektstruktur, CRUD-Endpunkte, Fehlerbehandlung, Tests, Docker-Setup und README erstellen.
- OpenAPI-Beschreibung und Demo-Ablauf fuer die Abgabe dokumentieren.
