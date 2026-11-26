import { PageContainer } from '@ant-design/pro-components';
import { history } from '@umijs/max';
import { Card } from 'antd';
import TranslateSettings from './components/TranslateSettings';

const PluginTranslate = () => (
  <PageContainer>
    <Card>
      <TranslateSettings
        onOpenWritingSettings={() => history.push('/setting/ai?tab=write')}
      />
    </Card>
  </PageContainer>
);

export default PluginTranslate;
