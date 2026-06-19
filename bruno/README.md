# Bruno Collection

Die Collection definiert in `collection.bru` die Variable `baseUrl` mit `http://localhost:3000`. Die Requests funktionieren deshalb auch dann, wenn im Bruno-Client noch kein Environment ausgewaehlt ist.

Das Environment `local` kann diese Werte ueberschreiben und enthaelt:

- `baseUrl`: Referenzwert fuer lokale Tests
- `adminToken`: optionaler Bearer Token, falls der Server mit `ADMIN_TOKEN` gestartet wurde

Wenn `ADMIN_TOKEN` nicht gesetzt ist, koennen `Create`, `Update` und `Delete` ohne Token ausgefuehrt werden. Wenn `ADMIN_TOKEN` gesetzt ist, in Bruno das Environment `local` auswaehlen und `adminToken` passend setzen.
