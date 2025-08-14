import { test, expect } from '@grafana/plugin-e2e';
import { MyDataSourceOptions, MySecureJsonData } from '../src/types';

test('smoke: should render config editor', async ({ createDataSourceConfigPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });
  await expect(page.getByText('Client ID')).toBeVisible();
});
test('"Save & test" should be successful when configuration is valid', async () => {
  // Skip this test since it requires valid WEMS API credentials
  test.skip();
});

test('"Save & test" should fail when configuration is invalid', async () => {
  // Skip this test since it requires valid WEMS API credentials to test properly
  test.skip();
});
