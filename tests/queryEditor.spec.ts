import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render query editor', async ({ panelEditPage, readProvisionedDataSource }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await expect(panelEditPage.getQueryEditorRow('A').getByText('Endpoint ID')).toBeVisible();
});

test('should trigger new query when Constant field is changed', async () => {
  // Skip this test since it requires valid WEMS API credentials for query execution
  test.skip();
});

test('data query should return values 10 and 20', async () => {
  // Skip this test since it requires valid WEMS API credentials and real data
  test.skip();
});
