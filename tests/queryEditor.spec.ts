import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render query editor', async ({ panelEditPage, readProvisionedDataSource, selectors }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await expect(panelEditPage.getByGrafanaSelector(selectors.components.CodeEditor.container)).toBeVisible();
});

test('typing a Cypher query should trigger a new query request', async ({
  panelEditPage,
  readProvisionedDataSource,
  selectors,
  page,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await panelEditPage.getByGrafanaSelector(selectors.components.CodeEditor.container).click();
  const queryReq = panelEditPage.waitForQueryDataRequest();
  await page.keyboard.type('RETURN 1');
  await expect(await queryReq).toBeTruthy();
});

test('data query should return the queried value', async ({
  panelEditPage,
  readProvisionedDataSource,
  selectors,
  page,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await panelEditPage.getByGrafanaSelector(selectors.components.CodeEditor.container).click();
  await page.keyboard.type('RETURN 1 AS one');
  await panelEditPage.setVisualization('Table');
  await expect(panelEditPage.refreshPanel()).toBeOK();
  await expect(panelEditPage.panel.data).toContainText(['1']);
});
