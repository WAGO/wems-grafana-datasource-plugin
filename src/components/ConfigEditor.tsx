import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onClientIdChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        client_id: event.target.value,
      },
    });
  };

  const onBaseUrlChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        base_url: event.target.value,
      },
    });
  };

  const onClientSecretChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        client_secret: event.target.value,
      },
    });
  };

  const onResetClientSecret = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        client_secret: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        client_secret: '',
      },
    });
  };

  return (
    <>
      <InlineField label="Client ID" labelWidth={14} interactive tooltip={'WEMS API Client ID'}>
        <Input
          id="config-editor-client-id"
          onChange={onClientIdChange}
          value={jsonData.client_id || ''}
          placeholder="Enter your client ID"
          width={40}
        />
      </InlineField>
      <InlineField label="Client Secret" labelWidth={14} interactive tooltip={'WEMS API Client Secret'}>
        <SecretInput
          required
          id="config-editor-client-secret"
          isConfigured={secureJsonFields.client_secret}
          value={secureJsonData?.client_secret || ''}
          placeholder="Enter your client secret"
          width={40}
          onReset={onResetClientSecret}
          onChange={onClientSecretChange}
        />
      </InlineField>
      <InlineField label="Base URL" labelWidth={14} interactive tooltip={'WEMS API Base URL (optional, defaults to https://c1.api.wago.com/wems)'}>
        <Input
          id="config-editor-base-url"
          onChange={onBaseUrlChange}
          value={jsonData.base_url || ''}
          placeholder="https://c1.api.wago.com/wems (default)"
          width={40}
        />
      </InlineField>
    </>
  );
}
