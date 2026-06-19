# Programmierworkshop am 19.6.2026

## Namen

Anna Wiedemann, Gruppe 2

## Link zum Git-Repository

https://github.com/buan1027/workshop

## KI-Werkzeuge

- ChatGPT/Codex als Coding-Agent fuer Analyse, Framework-Auswahl, Implementierung, Tests und Dokumentation.

### Agenten

- Codex im lokalen Workspace `C:\Users\anna\dev\workshop`.
- Arbeitsweise: iterativ, mit kleinen Commits nach dem Prinzip "Commit early, commit often".

### Chat-URLs

- Keine oeffentliche Chat-URL verfuegbar. Die Arbeit erfolgte in einer lokalen Codex-Desktop-Sitzung.

## Frameworks und Bibliotheken

### REST-Schnittstelle (Lesen und Neuanlegen)

- `net/http` aus der Go-Standardbibliothek als HTTP-Basis.
- `github.com/go-chi/chi/v5` fuer Routing, Route-Gruppen und Middleware.

Begruendung: `chi` ist leichtgewichtig, gut dokumentiert und bleibt nah an idiomatischem Go. Dadurch ist die Loesung im Workshop gut erklaerbar.

Umgesetzt wurden Healthchecks sowie CRUD-Endpunkte fuer die Hauptressource `Gebrauchtwagen`:

```text
GET    /health/liveness
GET    /health/readiness
GET    /api/gebrauchtwagen
GET    /api/gebrauchtwagen/{id}
POST   /api/gebrauchtwagen
PUT    /api/gebrauchtwagen/{id}
DELETE /api/gebrauchtwagen/{id}
```

Die Listenabfrage unterstuetzt Filter und Paging mit `page`, `size` und `count-only`.

Schluessel im Modell:

- `id` ist der technische Primaerschluessel in PostgreSQL und wird automatisch vergeben.
- `fin` ist 17-stellig, eindeutig und fachlich der stabile Fahrzeugschluessel fuer Clients.
- Die REST-Pfade verwenden aktuell die technische `id`; die `fin` wird in Requests und Responses mitgefuehrt und serverseitig eindeutig validiert.

### Validierung (nur Neuanlegen)

- Manuelle Validierung in `internal/domain`, aufgerufen aus `internal/service`.

Begruendung: Fuer das kleine Datenmodell sind klare eigene Regeln einfacher zu erklaeren als ein zusaetzliches Validation-Framework. Validiert werden Pflichtfelder, die 17-stellige FIN, Enum-Werte, Datumsfelder, `kilometerstand >= 0` und optionale Relationsdaten. Die Service-Schicht fuehrt diese Validierung aus, bevor das Repository PostgreSQL aufruft.

Hinweis: Die Aufgabenstellung nennt "nur Neuanlegen". Im Server wird dieselbe fachliche Validierung auch fuer `PUT` wiederverwendet, weil Updates sonst inkonsistente Daten erzeugen koennten.

### OR-Mapping (fuer PostgreSQL)

- Kein klassisches ORM.
- `github.com/jackc/pgx/v5` mit `pgxpool` fuer direkten PostgreSQL-Zugriff.
- Echtes PostgreSQL-Backend ueber Docker Compose (`postgres:17`) mit Schema- und Seed-Skripten.

Begruendung: Das Datenbankschema existiert bereits. Direkte SQL-Queries sind transparent, schnell umzusetzen und vermeiden Mapping-Probleme mit PostgreSQL-Enums. Damit entspricht der Server dem DB-Server-Ansatz aus den vorherigen Abgaben, ohne fuer den Workshop zusaetzliche ORM-Komplexitaet einzufuehren.

### Optional: OIDC mit Keycloak

- Optional vorbereitet.
- Standard ist weiterhin der einfache `ADMIN_TOKEN`: Wenn gesetzt, brauchen schreibende Endpunkte `Authorization: Bearer <token>`.
- Mit `AUTH_MODE=keycloak` prueft der Server Bearer Tokens ueber OpenID Connect gegen `KEYCLOAK_ISSUER_URL`.

Begruendung: Keycloak war optional. Deshalb ist die Integration bewusst abschaltbar und faellt bei fehlender Keycloak-Konfiguration auf den stabilen `ADMIN_TOKEN`-Modus zurueck, damit der Kernserver im Workshop weiter lauffaehig bleibt.

### Einfacher Integrationstest

- Unit-/Handler-Tests laufen mit `go test ./...`.
- Optionaler PostgreSQL-Integrationstest gegen eine echte laufende Datenbank:

```powershell
$env:INTEGRATION_DATABASE_URL="postgres://gebrauchtwagen:gebrauchtwagen@localhost:5432/gebrauchtwagen?sslmode=disable"
go test ./internal/repository
```

Getestet werden unter anderem:

- Validierung der fachlichen Eingaben
- Healthcheck-Handler
- REST-Create, Detailantwort, ETag, `If-None-Match` und `If-Match`
- Paging-Metadaten und ungueltige Paging-Parameter
- PostgreSQL-CRUD, relationale Detaildaten und echte Limit-/Offset-Paginierung

### Demo-Daten

- Beim Serverstart wird der Datenbestand standardmaessig neu befuellt.
- Das entspricht dem Demo-Verhalten aus den vorherigen Abgaben.
- Abschaltbar mit `RESET_DATABASE_ON_START=false`, falls Daten lokal erhalten bleiben sollen.

### Optimistische Synchronisation

- Das Datenmodell enthaelt `version`.
- `GET /api/gebrauchtwagen/{id}` liefert die aktuelle Version als `ETag`.
- `PUT` verlangt `If-Match`; bei veralteter Version antwortet der Server mit `412 Precondition Failed`.
- Nach erfolgreichem Update wird `version` in PostgreSQL erhoeht.

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
- Die Umgebung `local` verwendet `baseUrl = http://127.0.0.1:3000`, damit das VS-Code-Bruno-Plugin lokal stabil aufloest.
- Der optionale `adminToken` kann gesetzt werden, wenn der Server mit `ADMIN_TOKEN` gestartet wurde.
- Fuer Keycloak gibt es einen Auth-Request, der ein Access Token fuer `admin` holt und als `adminToken` speichert. Die vorhandenen schreibenden REST-Requests verwenden diese Variable bereits als Bearer Token.

### Weitere Artefakte

- `README.md`: Start-, Test- und Architekturhinweise fuer die schnelle Nutzung.
- `docs/architecture.md`: Kurzbeschreibung der Schichten.
- `docs/projekthandbuch.adoc`: Kompaktes Projekthandbuch in AsciiDoc.
- `docs/demo-guide.md`: Vorfuehrablauf mit konkreten Requests.
- `docs/diagramme/src/*.puml`: PlantUML-Quellen fuer Komponenten, ER-Modell und Update-Sequenz.
- `docs/openapi.yaml`: Schlanke OpenAPI-Beschreibung.
- `.github/workflows/go.yml`: CI mit Formatierung, Analyse, Tests und Vulnerability-Check.

## Prompts/Requests an KI-Agent/en

- Aufgabenstellung analysieren und Ressourcen, Endpunkte, Validierungen und Risiken ableiten.
- Geeignete Go-Frameworks fuer REST, Datenbankzugriff, Validierung und Tests vergleichen.
- Empfohlene Loesung mit `chi` und `pgx` umsetzen.
- Projektstruktur, CRUD-Endpunkte, Fehlerbehandlung, Tests, Docker-Setup und README erstellen.
- OpenAPI-Beschreibung und Demo-Ablauf fuer die Abgabe dokumentieren.
- Reales PostgreSQL-Backend und automatisches Neuladen der Demo-Daten umsetzen.
- Optimistische Synchronisation mit `version`, `ETag` und `If-Match` pruefen.
- Paging explizit fuer REST und PostgreSQL testen.
- Optionale Keycloak/OIDC-Authentifizierung so vorbereiten, dass der Server ohne Keycloak weiter funktioniert.
