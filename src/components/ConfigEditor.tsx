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
        clientId: event.target.value,
      },
    });
  };

  const onBaseUrlChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        baseUrl: event.target.value,
      },
    });
  };

  const onClientSecretChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        clientSecret: event.target.value,
      },
    });
  };

  const onResetClientSecret = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        clientSecret: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        clientSecret: '',
      },
    });
  };

  return (
    <>
      <InlineField label="Client-ID" labelWidth={16} interactive tooltip={'WEMS API Client ID'}>
        <Input
          id="config-editor-client-id"
          onChange={onClientIdChange}
          value={jsonData.clientId || ''}
          placeholder="Enter your client ID"
          width={40}
        />
      </InlineField>
      <InlineField label="Client Secret" labelWidth={16} interactive tooltip={'WEMS API Client Secret'}>
        <SecretInput
          required
          id="config-editor-client-secret"
          isConfigured={secureJsonFields.clientSecret}
          value={secureJsonData?.clientSecret || ''}
          placeholder="Enter your client secret"
          width={40}
          onReset={onResetClientSecret}
          onChange={onClientSecretChange}
        />
      </InlineField>
      <InlineField label="Base URL" labelWidth={16} interactive tooltip={'WEMS API Base URL'}>
        <Input
          id="config-editor-base-url"
          onChange={onBaseUrlChange}
          value={jsonData.baseUrl || ''}
          placeholder="https://c1.api.wago.com/wems"
          width={40}
        />
      </InlineField>
    </>
  );
}
