import React, { ChangeEvent } from 'react';
import { Box, Field, Input, SecretInput, Stack, Switch } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps, FeatureToggles } from '@grafana/data';
import { config } from '@grafana/runtime';
import { selectors } from '@grafana/e2e-selectors';
import { gte } from 'semver';
import { MyDataSourceOptions, MySecureJsonData } from '../types';
import { ConfigSection } from './config/ConfigSection';
import { LeftSidebar, SidebarHeader } from './config/LeftSidebar';

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

  // PDC (secure SOCKS proxy) is offered when Grafana exposes the feature toggle
  // and is >= 10 (the official Grafana PDC gating pattern).
  const secureSocksSupported =
    config.featureToggles['secureSocksDSProxyEnabled' as keyof FeatureToggles] && gte(config.buildInfo.version, '10.0.0');

  const headers: SidebarHeader[] = [
    { label: 'Connection', id: 'connection' },
    { label: 'Authentication', id: 'authentication' },
    ...(secureSocksSupported ? [{ label: 'Private data source connect', id: 'pdc', isOptional: true }] : []),
    { label: 'Save & test', id: selectors.pages.DataSource.saveAndTest },
  ];

  return (
    <Stack gap={4} alignItems="flex-start">
      <LeftSidebar headers={headers} />
      <Box flex={1} minWidth="450px">
        <ConfigSection
          id="connection"
          title="Connection"
          description="The Neo4j server to connect to. Use a bolt:// or neo4j:// URL."
        >
          <Field label="URL" description="Neo4j Bolt address, e.g. neo4j://localhost:7687" htmlFor="config-editor-url">
            <Input
              id="config-editor-url"
              onChange={(e: ChangeEvent<HTMLInputElement>) => onJsonDataChange('url', e.target.value)}
              value={jsonData.url ?? ''}
              placeholder="neo4j://localhost:7687"
            />
          </Field>
        </ConfigSection>

        <ConfigSection id="authentication" title="Authentication" description="Credentials used to authenticate with Neo4j. Leave empty for no authentication.">
          <Stack direction="column" gap={2}>
            <Field label="Username" htmlFor="config-editor-username">
              <Input
                id="config-editor-username"
                onChange={(e: ChangeEvent<HTMLInputElement>) => onJsonDataChange('username', e.target.value)}
                value={jsonData.username ?? ''}
                placeholder="neo4j"
              />
            </Field>
            <Field label="Password" htmlFor="config-editor-password">
              <SecretInput
                id="config-editor-password"
                isConfigured={secureJsonFields.password}
                value={secureJsonData?.password}
                placeholder="password"
                onReset={onResetPassword}
                onChange={onPasswordChange}
              />
            </Field>
            <Field label="Default database" description="Leave empty for the server default database" htmlFor="config-editor-database">
              <Input
                id="config-editor-database"
                onChange={(e: ChangeEvent<HTMLInputElement>) => onJsonDataChange('database', e.target.value)}
                value={jsonData.database ?? ''}
                placeholder="default"
              />
            </Field>
          </Stack>
        </ConfigSection>

        {secureSocksSupported && (
          <ConfigSection
            id="pdc"
            title="Private data source connect"
            description="Establish a Private Data Source Connect tunnel that enables Grafana Cloud to connect securely to this data source hosted in a private network."
          >
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
          </ConfigSection>
        )}
      </Box>
    </Stack>
  );
}
