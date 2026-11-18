import { DataSourceJsonData, DataQuery } from '@grafana/schema';

export interface MyQuery extends DataQuery {
  cypherQuery?: string;
  format?: Format;
  database?: string;
}

export enum Format {
  Table = 'table',
  NodeGraph = 'nodegraph',
}

export const DEFAULT_QUERY: Partial<MyQuery> = {
  cypherQuery: 'MATCH (n) RETURN n LIMIT 10',
  format: Format.Table,
};

export interface MyDataSourceOptions extends DataSourceJsonData {
  url?: string;
  database?: string;
  username?: string;
  enableSecureSocksProxy?: boolean;
  // Set by Grafana Cloud when it injects the PDC connection/tunnel selector UI.
  pdcInjected?: boolean;
}

export interface MySecureJsonData {
  password?: string;
}
