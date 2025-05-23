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

  const [appliances, setAppliances] = useState<Array<SelectableValue<string>>>([]);
  const [loadingAppliances, setLoadingAppliances] = useState(false);
  const [applianceError, setApplianceError] = useState<string | null>(null);

  const [serviceUris, setServiceUris] = useState<Array<SelectableValue<string>>>([]);
  const [loadingServiceUris, setLoadingServiceUris] = useState(false);
  const [serviceUriError, setServiceUriError] = useState<string | null>(null);

  // Fetch endpoints on mount
  useEffect(() => {
    setLoadingEndpoints(true);
    setEndpointError(null);
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

  // Fetch appliances when endpoint_id changes
  useEffect(() => {
    if (!query.endpoint_id) {
      setAppliances([]);
      setApplianceError(null);
      return;
    }
    setLoadingAppliances(true);
    setApplianceError(null);
    datasource
      .getResource('appliance-list', { endpointId: query.endpoint_id })
      .then((result: any) => {
        const opts = (result || []).map((ap: any) => ({
          label: ap.label || ap.id,
          value: ap.id,
        }));
        setAppliances(opts);
        setLoadingAppliances(false);
      })
      .catch((err: any) => {
        setApplianceError('Failed to load appliances');
        setLoadingAppliances(false);
      });
  }, [datasource, query.endpoint_id]);

  // Fetch service URIs when endpoint_id and appliance_id change
  useEffect(() => {
    if (!query.endpoint_id || !query.appliance_id) {
      setServiceUris([]);
      setServiceUriError(null);
      return;
    }
    setLoadingServiceUris(true);
    setServiceUriError(null);
    datasource
      .getResource('service-list', { endpointId: query.endpoint_id, applianceId: query.appliance_id })
      .then((result: any) => {
        const opts = (result || []).map((svc: any) => ({
          label: svc.label || svc.uri,
          value: svc.uri,
        }));
        setServiceUris(opts);
        setLoadingServiceUris(false);
      })
      .catch((err: any) => {
        setServiceUriError('Failed to load service URIs');
        setLoadingServiceUris(false);
      });
  }, [datasource, query.endpoint_id, query.appliance_id]);

  const onFieldChange = (field: keyof MyQuery) => (event: React.ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value;
    onChange({ ...query, [field]: value === '' ? '' : value });
  };
  const onSelectEndpoint = (option: SelectableValue<string>) => {
    // Clear appliance_id if endpoint changes
    onChange({ ...query, endpoint_id: option?.value ?? '', appliance_id: '' });
  };
  const onSelectAppliance = (option: SelectableValue<string>) => {
    onChange({ ...query, appliance_id: option?.value ?? '' });
  };
  const onSelectServiceUri = (option: SelectableValue<string>) => {
    onChange({ ...query, service_uri: option?.value ?? '' });
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
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Select
            options={appliances}
            value={appliances.find((a) => a.value === query.appliance_id) || null}
            onChange={onSelectAppliance}
            isLoading={loadingAppliances}
            width={40}
            placeholder={query.endpoint_id ? 'Select appliance...' : 'Select endpoint first'}
            disabled={!query.endpoint_id}
          />
          {applianceError && <span style={{ color: 'red', marginLeft: 8 }}>{applianceError}</span>}
        </div>
      </InlineField>
      <InlineField label="Service URI">
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Select
            options={serviceUris}
            value={serviceUris.find((s) => s.value === query.service_uri) || null}
            onChange={onSelectServiceUri}
            isLoading={loadingServiceUris}
            width={40}
            placeholder={query.endpoint_id && query.appliance_id ? 'Select service URI...' : 'Select endpoint and appliance first'}
            disabled={!query.endpoint_id || !query.appliance_id}
          />
          {serviceUriError && <span style={{ color: 'red', marginLeft: 8 }}>{serviceUriError}</span>}
        </div>
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
