import { DataSourceInstanceSettings, CoreApp, ScopedVars } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';

import { MyQuery, MyDataSourceOptions, DEFAULT_QUERY } from './types';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<MyQuery> {
    return DEFAULT_QUERY;
  }

  applyTemplateVariables(query: MyQuery, scopedVars: ScopedVars) {
    // No queryText field anymore; just return the query as-is or apply template replacement to relevant fields if needed
    return {
      ...query,
      // Example: endpoint_id: getTemplateSrv().replace(query.endpoint_id, scopedVars),
      // Add similar lines for other fields if you want template variable support
    };
  }

  filterQuery(query: MyQuery): boolean {
    // Only run query if required fields are present
    return !!query.endpoint_id && !!query.appliance_id && !!query.service_uri && !!query.data_point;
  }
}
