import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export interface MyQuery extends DataQuery {
  endpoint_id: string;
  appliance_id: string;
  service_uri: string;
  data_point: string;
  aggregate_function?: string;
  create_empty_values?: boolean;
  unit?: string;
}

export const DEFAULT_QUERY: Partial<MyQuery> = {};

export interface DataPoint {
  Time: number;
  Value: number;
}

export interface DataSourceResponse {
  datapoints: DataPoint[];
}

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  client_id?: string;
  base_url?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  client_secret?: string;
}
