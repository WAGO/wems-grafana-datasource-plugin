import React from 'react';
import { InlineField, Input } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onFieldChange = (field: keyof MyQuery) => (event: React.ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value;
    onChange({ ...query, [field]: value === '' ? null : value });
  };
  const onBoolFieldChange = (field: keyof MyQuery) => (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, [field]: event.target.checked });
  };

  return (
    <>
      <InlineField label="Endpoint ID">
        <Input value={query.endpoint_id || ''} onChange={onFieldChange('endpoint_id')} width={40} />
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
