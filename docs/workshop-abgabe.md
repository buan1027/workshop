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

- Manuelle Validierung in `internal/domain`.

Begruendung: Fuer das kleine Datenmodell sind klare eigene Regeln einfacher zu erklaeren als ein zusaetzliches Validation-Framework. Validiert werden Pflichtfelder, Enum-Werte und `kilometerstand >= 0`.

### OR-Mapping (fuer PostgreSQL)

- Kein klassisches ORM.
- `github.com/jackc/pgx/v5` mit `pgxpool` fuer direkten PostgreSQL-Zugriff.

Begruendung: Das Datenbankschema existiert bereits. Direkte SQL-Queries sind transparent, schnell umzusetzen und vermeiden Mapping-Probleme mit PostgreSQL-Enums.

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

## Prompts/Requests an KI-Agent/en

- Aufgabenstellung analysieren und Ressourcen, Endpunkte, Validierungen und Risiken ableiten.
- Geeignete Go-Frameworks fuer REST, Datenbankzugriff, Validierung und Tests vergleichen.
- Empfohlene Loesung mit `chi` und `pgx` umsetzen.
- Projektstruktur, CRUD-Endpunkte, Fehlerbehandlung, Tests, Docker-Setup und README erstellen.
- OpenAPI-Beschreibung und Demo-Ablauf fuer die Abgabe dokumentieren.
