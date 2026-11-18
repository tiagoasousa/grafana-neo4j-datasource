import React, { ChangeEvent } from 'react';
import { Field, Input, SecretInput, Stack, Switch, Text } from '@grafana/ui';
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
    <Stack direction="column" gap={3}>
      <Stack direction="column" gap={2}>
        <Text variant="h3">Connection</Text>
        <Field label="URL" description="Neo4j Bolt address, e.g. neo4j://localhost:7687" htmlFor="config-editor-url">
          <Input
            id="config-editor-url"
            onChange={(e: ChangeEvent<HTMLInputElement>) => onJsonDataChange('url', e.target.value)}
            value={jsonData.url ?? ''}
            placeholder="neo4j://localhost:7687"
          />
        </Field>
        <Field
          label="Database"
          description="Leave empty for the server default database"
          htmlFor="config-editor-database"
        >
          <Input
            id="config-editor-database"
            onChange={(e: ChangeEvent<HTMLInputElement>) => onJsonDataChange('database', e.target.value)}
            value={jsonData.database ?? ''}
            placeholder="default"
          />
        </Field>
      </Stack>

      <Stack direction="column" gap={2}>
        <Text variant="h3">Authentication</Text>
        <Field label="Username" description="Leave empty for no authentication" htmlFor="config-editor-username">
          <Input
            id="config-editor-username"
            onChange={(e: ChangeEvent<HTMLInputElement>) => onJsonDataChange('username', e.target.value)}
            value={jsonData.username ?? ''}
            placeholder="neo4j"
          />
        </Field>
        <Field label="Password" description="Leave empty for no authentication" htmlFor="config-editor-password">
          <SecretInput
            id="config-editor-password"
            isConfigured={secureJsonFields.password}
            value={secureJsonData?.password}
            placeholder="password"
            onReset={onResetPassword}
            onChange={onPasswordChange}
          />
        </Field>
      </Stack>

      {secureSocksSupported && (
        <Stack direction="column" gap={2}>
          <Text variant="h3">Private Data Source Connect</Text>
          <Field
            label="Secure SOCKS Proxy"
            description="Route the connection through Grafana's secure SOCKS proxy to reach a private network. Required for Grafana Private Data Source Connect (PDC)."
            htmlFor="config-editor-secure-socks"
          >
            <Switch
              id="config-editor-secure-socks"
              value={jsonData.enableSecureSocksProxy || false}
              onChange={(e: ChangeEvent<HTMLInputElement>) =>
                onJsonDataChange('enableSecureSocksProxy', e.currentTarget.checked)
              }
            />
          </Field>
        </Stack>
      )}
    </Stack>
  );
}
