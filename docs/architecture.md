# Architektur

Der Server ist bewusst als kleine Schichtenarchitektur aufgebaut. Ziel ist eine Loesung, die in einem Workshop schnell lauffaehig ist, aber trotzdem die Verantwortlichkeiten sauber trennt.

## Schichten

```text
cmd/server
  startet Konfiguration, Datenbankpool, Repository, Service und HTTP-Server

internal/http
  REST-Routing, Request-/Response-Mapping, Statuscodes, Problem-Details, CORS

internal/service
  Use Cases, fachliche Validierung, Delegation an Repository

internal/domain
  Datenmodell, DTOs, Enum-Werte, Validierungsregeln, fachliche Fehler

internal/repository
  PostgreSQL-Zugriff mit pgx, SQL-Queries, Transaktionen
```

## Validierung

Die Validierungsregeln sind in `internal/domain` implementiert, weil sie fachlich zum Datenmodell gehoeren. Aufgerufen werden sie aus der Service-Schicht. Dadurch bleibt der HTTP-Handler frei von Fachlogik:

- Handler prueft HTTP-spezifische Dinge wie JSON, Pfadparameter, Header und Query-Parameter.
- Service prueft fachliche Regeln fuer `Create` und `Update`.
- Repository speichert nur bereits validierte Daten.

Das ist fuer den Workshop ein guter Kompromiss: keine zusaetzliche Validation-Library, aber klare Verantwortlichkeiten.

## Datenbankzugriff

Der produktive Server verwendet einen echten PostgreSQL-Server. Lokal wird dieser ueber Docker Compose als `postgres:17` gestartet; Schema und Seed-Daten liegen in `extras/compose/postgres/init/`.

Im Code erzeugt `cmd/server` aus `DATABASE_URL` einen `pgxpool.Pool` und uebergibt ihn an `PostgresGebrauchtwagenRepository`. Es gibt deshalb im laufenden Server keinen In-Memory-Speicher und kein Fake-Backend. Fakes werden nur in Handler-Tests eingesetzt.

Der Server verwendet `pgx` statt ORM. Das passt hier gut, weil das PostgreSQL-Schema bereits vorhanden ist und PostgreSQL-Enums direkt genutzt werden. Die Neuanlage eines Fahrzeugs mit optionalem Standort, Hauptuntersuchung und Schaeden erfolgt transaktional.

## Demo-Daten

Beim Serverstart setzt die Anwendung den Demo-Datenbestand standardmaessig zurueck. Das ist fuer den Workshop praktisch, weil Tests und manuelle Requests nach jedem Neustart wieder denselben Ausgangszustand sehen. Technisch passiert das transaktional ueber `internal/database/demo_seed.sql`.

Das Verhalten ist ueber `RESET_DATABASE_ON_START=false` abschaltbar. In einem echten Produktivbetrieb sollte dieser Wert deaktiviert sein.

## Fehlerbehandlung

HTTP-Fehler werden als `application/problem+json` zurueckgegeben. Fachliche Fehler wie `ErrNotFound`, `ErrVersionConflict` und Validierungsfehler werden im Handler auf passende HTTP-Statuscodes abgebildet.

## Komponenten

Siehe PlantUML-Diagramm: `docs/diagramme/src/komponenten.puml`.
