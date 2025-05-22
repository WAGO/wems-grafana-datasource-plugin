import React, { useEffect, useState } from 'react';
import { InlineField, Input, Select } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const [endpoints, setEndpoints] = useState<Array<SelectableValue<string>>>([]);
  const [loadingEndpoints, setLoadingEndpoints] = useState(false);
  const [endpointError, setEndpointError] = useState<string | null>(null);

  useEffect(() => {
    setLoadingEndpoints(true);
    setEndpointError(null);
    // Call backend proxy to fetch endpoints
    datasource
      .getResource('endpoint-list')
      .then((result: any) => {
        const opts = (result || []).map((ep: any) => ({
          label: ep.friendlyName ? `${ep.friendlyName} (${ep.endpointId})` : ep.endpointId,
          value: ep.endpointId,
        }));
        setEndpoints(opts);
        setLoadingEndpoints(false);
      })
      .catch((err: any) => {
        setEndpointError('Failed to load endpoints');
        setLoadingEndpoints(false);
      });
  }, [datasource]);

  const onFieldChange = (field: keyof MyQuery) => (event: React.ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value;
    onChange({ ...query, [field]: value === '' ? '' : value });
  };
  const onSelectEndpoint = (option: SelectableValue<string>) => {
    onChange({ ...query, endpoint_id: option?.value ?? '' });
  };
  const onBoolFieldChange = (field: keyof MyQuery) => (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, [field]: event.target.checked });
  };

  return (
    <>
      <InlineField label="Endpoint ID">
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Select
            options={endpoints}
            value={endpoints.find((e) => e.value === query.endpoint_id) || null}
            onChange={onSelectEndpoint}
            isLoading={loadingEndpoints}
            width={40}
            placeholder="Select endpoint..."
          />
          {endpointError && <span style={{ color: 'red', marginLeft: 8 }}>{endpointError}</span>}
        </div>
      </InlineField>
      <InlineField label="Appliance ID">
        <Input value={query.appliance_id || ''} onChange={onFieldChange('appliance_id')} width={40} />
      </InlineField>
      <InlineField label="Service URI">
        <Input value={query.service_uri || ''} onChange={onFieldChange('service_uri')} width={40} />
      </InlineField>
      <InlineField label="Data Point">
        <Input value={query.data_point || ''} onChange={onFieldChange('data_point')} width={40} />
      </InlineField>
      <InlineField label="Aggregate Function (optional)">
        <Input value={query.aggregate_function || ''} onChange={onFieldChange('aggregate_function')} width={20} placeholder="mean" />
      </InlineField>
      <InlineField label="Create Empty Values (optional)">
        <input type="checkbox" checked={!!query.create_empty_values} onChange={onBoolFieldChange('create_empty_values')} />
      </InlineField>
    </>
  );
}
