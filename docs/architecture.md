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

Der Server verwendet `pgx` statt ORM. Das passt hier gut, weil das PostgreSQL-Schema bereits vorhanden ist und PostgreSQL-Enums direkt genutzt werden. Die Neuanlage eines Fahrzeugs mit optionalem Standort, Hauptuntersuchung und Schaeden erfolgt transaktional.

## Fehlerbehandlung

HTTP-Fehler werden als `application/problem+json` zurueckgegeben. Fachliche Fehler wie `ErrNotFound`, `ErrVersionConflict` und Validierungsfehler werden im Handler auf passende HTTP-Statuscodes abgebildet.

## Komponenten

Siehe PlantUML-Diagramm: `docs/diagramme/src/komponenten.puml`.
