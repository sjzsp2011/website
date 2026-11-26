import NewContainer from '@/components/NewContainer';
import {
  getSettingSensitiveWords,
  saveSettingSensitiveWords,
} from '@/services/setting';
import {
  ProForm,
  ProFormInstance,
  ProFormTextArea,
} from '@ant-design/pro-components';
import { useIntl } from '@umijs/max';
import { Card, message } from 'antd';
import React, { useEffect, useState } from 'react';

const SettingSensitiveFrom: React.FC<any> = () => {
  const formRef = React.createRef<ProFormInstance>();
  const [setting, setSetting] = useState<any>([]);
  const [fetched, setFetched] = useState<boolean>(false);
  const [newKey, setNewKey] = useState<string>('');
  const intl = useIntl();

  const getSetting = async () => {
    const res = await getSettingSensitiveWords();
    let setting = res.data || [];
    setSetting(setting);
    setFetched(true);
  };

  const onTabChange = (key: string) => {
    getSetting().then(() => {
      setNewKey(key);
    });
  };

  useEffect(() => {
    getSetting();
  }, []);

  const onSubmit = async (values: any) => {
    const hide = message.loading(
      intl.formatMessage({ id: 'setting.system.submitting' }),
      0,
    );
    const text = values.words.split('\n');
    saveSettingSensitiveWords(text)
      .then((res) => {
        message.success(res.msg);
      })
      .catch((err) => {
        console.log(err);
      })
      .finally(() => {
        hide();
      });
  };

  return (
    <NewContainer onTabChange={(key) => onTabChange(key)}>
      <Card
        key={newKey}
        title={intl.formatMessage({ id: 'menu.setting.sensitive' })}
      >
        {fetched && (
          <ProForm
            formRef={formRef}
            initialValues={{ words: setting.join('\n') }}
            onFinish={onSubmit}
          >
            <ProFormTextArea
              fieldProps={{ rows: 20 }}
              name="words"
              label={intl.formatMessage({ id: 'setting.sensitive.sync.list' })}
              placeholder={intl.formatMessage({
                id: 'setting.sensitive.sync.placeholder',
              })}
              extra={intl.formatMessage({
                id: 'setting.sensitive.sync.description',
              })}
            />
          </ProForm>
        )}
      </Card>
    </NewContainer>
  );
};

export default SettingSensitiveFrom;
