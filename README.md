# Neo4j data source for Grafana

Run [Cypher](https://neo4j.com/docs/cypher-manual/current/) queries against a [Neo4j](https://neo4j.com/) database from Grafana, over the Bolt protocol.

## Overview

The data source connects to Neo4j over Bolt and returns query results as Grafana data frames.

**Features**

- Cypher queries over Bolt (`neo4j://` or `bolt://` URLs)
- Two result formats: **Table** and **Node Graph**
- Per-query database selection
- Grafana Private Data Source Connect (PDC) support via the secure SOCKS proxy
- Template-variable interpolation in queries, including the `${__from}` and `${__to}` time macros

## Configuration

| Field | Description |
| ----- | ----------- |
| **URL** | Bolt address, e.g. `neo4j://localhost:7687` |
| **Database** | Optional Neo4j database name. Leave empty to use the server's default database |
| **Username** | Neo4j user for basic authentication. Leave empty for no authentication |
| **Password** | Password for the user above. Leave empty for no authentication |
| **Secure SOCKS Proxy** | Enable when connecting through Grafana Private Data Source Connect (see below) |

## Private Data Source Connect (PDC)

To reach a Neo4j server on a private network from Grafana Cloud, enable **Secure SOCKS Proxy** and configure a [Private Data Source Connect](https://grafana.com/docs/grafana-cloud/connect-externally-hosted/private-data-source-connect/) agent for that network.

- The Grafana instance must have the `secureSocksDSProxyEnabled` feature toggle enabled for the toggle to appear.
- In PDC mode the Bolt connection is tunneled through the SOCKS proxy, which is already encrypted end-to-end. Driver-level TLS Bolt URLs (`neo4j+s://`, `neo4j+ssc://`, `bolt+s://`, `bolt+ssc://`) are **not supported** in this mode — use a plain `neo4j://` or `bolt://` URL.

## Node Graph format

Set the query **Format** to **Node Graph** and return nodes and/or relationships. Each row of the result is inspected: returned nodes become graph nodes, and returned relationships become edges between the returned nodes.

Node Graph styling is driven by node and relationship **properties**. A node property named after one of the Node Graph optional fields styles the node directly; all other properties are shown in the node's details panel.

Node optional fields (set as node properties):

- `color`
- `mainstat`
- `secondarystat`
- `subtitle`
- `icon`
- `nodeRadius`
- `highlighted`
- `arc__*` — arc segments, e.g. `arc__cpu` (node-only)

Edge optional fields (set as relationship properties): `secondarystat`, `thickness`, `color`, `strokeDasharray`, `highlighted`.

A node's `title` property overrides its label as the displayed name.

Example:

```cypher
MATCH (p:Person)-[r:ACTED_IN]->(m:Movie)
RETURN p, r, m
```

## Time filtering

`${__from}` and `${__to}` interpolate to the dashboard time range as epoch milliseconds. Use them in `WHERE` clauses against millisecond-epoch properties:

```cypher
MATCH (e:Event)
WHERE e.ts >= ${__from} AND e.ts <= ${__to}
RETURN e
```

For formatted dates, use Grafana's date formatting, e.g. `${__from:date:YYYY-MM-DD}`.

Dashboard template variables are interpolated into the query text as well.

## Per-query database

The **Database** field in the query editor overrides the data source's configured database for that query only. Leave it empty to use the data source setting.

## Development

The repository ships a Docker Compose development environment with Neo4j and Grafana.

```bash
# start Neo4j and load the development seed data
docker compose up -d neo4j
docker compose up neo4j-provisioning

# run the Go test suite (expects Neo4j on localhost:7687, user neo4j / Password123)
export GOTOOLCHAIN=auto
go test ./... -count=1

# build the frontend into dist/
npm install
npm run build

# run Grafana with the plugin loaded (after npm run build)
npm run server
```
