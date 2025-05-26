import React, { useEffect, useState } from 'react';
import { InlineField, Select } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';
import { getBackendSrv } from '@grafana/runtime';

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

  const [dataPoints, setDataPoints] = useState<Array<SelectableValue<string>>>([]);
  const [loadingDataPoints, setLoadingDataPoints] = useState(false);
  const [dataPointError, setDataPointError] = useState<string | null>(null);

  const [unit, setUnit] = useState<string | undefined>(undefined);
  const [validValues, setValidValues] = useState<string[] | undefined>(undefined);

  const aggregationOptions: Array<SelectableValue<string>> = [
    { label: 'Mean', value: 'mean' },
    { label: 'Median', value: 'median' },
    { label: 'Min', value: 'min' },
    { label: 'Max', value: 'max' },
    { label: 'Sum', value: 'sum' },
    { label: 'Count', value: 'count' },
    { label: 'First', value: 'first' },
    { label: 'Last', value: 'last' },
    { label: 'Derivative', value: 'derivative' },
  ];

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
        opts.sort((a: { label: string }, b: { label: string }) => a.label.localeCompare(b.label));
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
        opts.sort((a: { label: string }, b: { label: string }) => a.label.localeCompare(b.label));
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
        opts.sort((a: { label: string }, b: { label: string }) => a.label.localeCompare(b.label));
        setServiceUris(opts);
        setLoadingServiceUris(false);
      })
      .catch((err: any) => {
        setServiceUriError('Failed to load service URIs');
        setLoadingServiceUris(false);
      });
  }, [datasource, query.endpoint_id, query.appliance_id]);

  // Fetch data points when endpoint_id, appliance_id, and service_uri change
  useEffect(() => {
    if (!query.endpoint_id || !query.appliance_id || !query.service_uri) {
      setDataPoints([]);
      setDataPointError(null);
      return;
    }
    setLoadingDataPoints(true);
    setDataPointError(null);
    datasource
      .getResource('datapoint-list', {
        endpointId: query.endpoint_id,
        applianceId: query.appliance_id,
        serviceUri: query.service_uri,
      })
      .then((result: any) => {
        const opts = (result?.dataPoints ? Object.keys(result.dataPoints) : []).map((dp) => ({
          label: dp,
          value: dp,
        }));
        opts.sort((a: { label: string }, b: { label: string }) => a.label.localeCompare(b.label));
        setDataPoints(opts);
        setLoadingDataPoints(false);
      })
      .catch((err: any) => {
        setDataPointError('Failed to load data points');
        setLoadingDataPoints(false);
      });
  }, [datasource, query.endpoint_id, query.appliance_id, query.service_uri]);

  // Fetch unit and validValues when a datapoint is selected
  useEffect(() => {
    if (!query.endpoint_id || !query.appliance_id || !query.service_uri || !query.data_point) {
      setUnit(undefined);
      setValidValues(undefined);
      return;
    }
    // Call backend resource to get the unit and validValues
    getBackendSrv()
      .get(`/api/datasources/${datasource.id}/resources/datapoint-unit`, {
        endpointId: query.endpoint_id,
        applianceId: query.appliance_id,
        serviceUri: query.service_uri,
        datapoint: query.data_point,
      })
      .then((result: any) => {
        setUnit(result.unit || undefined);
        setValidValues(result.validValues || undefined);
      })
      .catch(() => {
        setUnit(undefined);
        setValidValues(undefined);
      });
  }, [query.endpoint_id, query.appliance_id, query.service_uri, query.data_point, datasource.id]);

  // Pass the unit and validValues to the query object so it can be used in the panel
  useEffect(() => {
    let changed = false;
    const newQuery: any = { ...query };
    if (unit !== undefined && query.unit !== unit) {
      newQuery.unit = unit;
      changed = true;
    }
    if (unit === undefined && query.unit) {
      newQuery.unit = undefined;
      changed = true;
    }
    if (validValues !== undefined && JSON.stringify(query.validValues) !== JSON.stringify(validValues)) {
      newQuery.validValues = validValues;
      changed = true;
    }
    if (validValues === undefined && query.validValues) {
      newQuery.validValues = undefined;
      changed = true;
    }
    if (changed) {
      onChange(newQuery);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [unit, validValues]);

  const onSelectEndpoint = (option: SelectableValue<string>) => {
    // Clear appliance_id, service_uri, and data_point if endpoint changes
    onChange({ ...query, endpoint_id: option?.value ?? '', appliance_id: '', service_uri: '', data_point: '' });
  };
  const onSelectAppliance = (option: SelectableValue<string>) => {
    // Clear service_uri and data_point if appliance changes
    onChange({ ...query, appliance_id: option?.value ?? '', service_uri: '', data_point: '' });
  };
  const onSelectServiceUri = (option: SelectableValue<string>) => {
    // Clear data_point if service_uri changes
    onChange({ ...query, service_uri: option?.value ?? '', data_point: '' });
  };
  const onSelectDataPoint = (option: SelectableValue<string>) => {
    onChange({ ...query, data_point: option?.value ?? '' });
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
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Select
            options={dataPoints}
            value={dataPoints.find((d) => d.value === query.data_point) || null}
            onChange={onSelectDataPoint}
            isLoading={loadingDataPoints}
            width={40}
            placeholder={query.endpoint_id && query.appliance_id && query.service_uri ? 'Select data point...' : 'Select endpoint, appliance, and service URI first'}
            disabled={!query.endpoint_id || !query.appliance_id || !query.service_uri}
          />
          {dataPointError && <span style={{ color: 'red', marginLeft: 8 }}>{dataPointError}</span>}
        </div>
      </InlineField>
      <InlineField label="Aggregate Function">
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Select
            options={aggregationOptions}
            value={aggregationOptions.find(opt => opt.value === (query.aggregate_function || 'mean'))}
            onChange={opt => onChange({ ...query, aggregate_function: opt?.value || 'mean' })}
            width={20}
            placeholder="Select aggregation..."
          />
        </div>
      </InlineField>
      <InlineField label="Create Empty Values (optional)">
        <input type="checkbox" checked={!!query.create_empty_values} onChange={onBoolFieldChange('create_empty_values')} />
      </InlineField>
    </>
  );
}
