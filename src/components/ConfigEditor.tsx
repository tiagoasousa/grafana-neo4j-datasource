import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput, Switch } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps, FeatureToggles } from '@grafana/data';
import { config } from '@grafana/runtime';
import { gte } from 'semver';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

type Props = DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData>;

export function ConfigEditor(props: Props) {
  const { options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onJsonDataChange = (field: keyof MyDataSourceOptions, value: string | boolean) => {
    props.onOptionsChange({ ...options, jsonData: { ...jsonData, [field]: value } });
  };

  const onPasswordChange = (event: ChangeEvent<HTMLInputElement>) => {
    props.onOptionsChange({ ...options, secureJsonData: { ...secureJsonData, password: event.target.value } });
  };

  const onResetPassword = () => {
    props.onOptionsChange({
      ...options,
      secureJsonFields: { ...options.secureJsonFields, password: false },
      secureJsonData: { ...options.secureJsonData, password: '' },
    });
  };

  const secureSocksSupported =
    config.featureToggles['secureSocksDSProxyEnabled' as keyof FeatureToggles] && gte(config.buildInfo.version, '10.0.0');

  return (
    <>
      <InlineField label="URL" labelWidth={14} tooltip="e.g. neo4j://localhost:7687">
        <Input
          id="config-editor-url"
          onChange={(e: ChangeEvent<HTMLInputElement>) => onJsonDataChange('url', e.target.value)}
          value={jsonData.url}
          placeholder="neo4j://localhost:7687"
          width={40}
        />
      </InlineField>
      <InlineField label="Database" labelWidth={14} tooltip="Leave empty for the server default database">
        <Input
          id="config-editor-database"
          onChange={(e: ChangeEvent<HTMLInputElement>) => onJsonDataChange('database', e.target.value)}
          value={jsonData.database}
          placeholder="default"
          width={40}
        />
      </InlineField>
      <InlineField label="Username" labelWidth={14}>
        <Input
          id="config-editor-username"
          onChange={(e: ChangeEvent<HTMLInputElement>) => onJsonDataChange('username', e.target.value)}
          value={jsonData.username}
          placeholder="leave empty for no authentication"
          width={40}
        />
      </InlineField>
      <InlineField label="Password" labelWidth={14}>
        <SecretInput
          required={false}
          id="config-editor-password"
          isConfigured={secureJsonFields.password}
          value={secureJsonData?.password}
          placeholder="leave empty for no authentication"
          width={40}
          onReset={onResetPassword}
          onChange={onPasswordChange}
        />
      </InlineField>
      {secureSocksSupported && (
        <InlineField
          label="Secure SOCKS Proxy"
          labelWidth={14}
          tooltip="Enable proxying the data source connection through the secure SOCKS proxy to a different network. Required for Grafana Private Data Source Connect (PDC)."
        >
          <Switch
            id="config-editor-secure-socks"
            value={jsonData.enableSecureSocksProxy || false}
            onChange={(e: ChangeEvent<HTMLInputElement>) =>
              onJsonDataChange('enableSecureSocksProxy', e.currentTarget.checked)
            }
          />
        </InlineField>
      )}
    </>
  );
}
