import {
  pluginGetTranslateConfig,
  pluginSaveTranslateConfig,
  pluginTranslateLogs,
} from '@/services';
import {
  ActionType,
  ProColumns,
  ProForm,
  ProFormDependency,
  ProFormInstance,
  ProFormRadio,
  ProFormText,
  ProTable,
} from '@ant-design/pro-components';
import { FormattedMessage, useIntl } from '@umijs/max';
import { Alert, Button, Card, Col, Modal, Row, Space, message } from 'antd';
import dayjs from 'dayjs';
import React, { useEffect, useRef, useState } from 'react';

type TranslateSettingsProps = {
  onOpenWritingSettings: () => void;
};

const TranslateSettings: React.FC<TranslateSettingsProps> = ({
  onOpenWritingSettings,
}) => {
  const actionRef = useRef<ActionType>();
  const formRef = useRef<ProFormInstance>();
  const [setting, setSetting] = useState<any>({});
  const [fetched, setFetched] = useState<boolean>(false);
  const [logVisible, setLogVisible] = useState<boolean>(false);
  const intl = useIntl();

  const getSetting = async () => {
    const res = await pluginGetTranslateConfig();
    let setting = res.data || {};
    setSetting(setting);
    formRef.current?.setFieldsValue(setting);
    setFetched(true);
  };

  useEffect(() => {
    getSetting();
  }, []);

  const onSubmit = async (data: any) => {
    const hide = message.loading(
      intl.formatMessage({ id: 'setting.system.submitting' }),
      0,
    );
    pluginSaveTranslateConfig(data)
      .then(async (res) => {
        message.success(res.msg);
        await getSetting();
      })
      .catch((err) => {
        console.log(err);
      })
      .finally(() => {
        hide();
      });
  };

  const handleShowPushLog = () => {
    setLogVisible(true);
  };

  const columns: ProColumns<any>[] = [
    {
      title: 'ID',
      dataIndex: 'id',
      hideInSearch: true,
    },
    {
      title: intl.formatMessage({ id: 'plugin.aigenerate.time' }),
      width: 160,
      dataIndex: 'created_time',
      render: (text, record) =>
        dayjs(record.created_time * 1000).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: intl.formatMessage({ id: 'plugin.translate.lang' }),
      width: 160,
      dataIndex: 'from',
      render: (_, record) => (
        <span>
          {record.from} -&gt; {record.to}
        </span>
      ),
    },
    {
      title: intl.formatMessage({ id: 'content.translate.origin-content' }),
      width: 160,
      dataIndex: 'origin_title',
    },
    {
      title: intl.formatMessage({ id: 'plugin.translate.result' }),
      dataIndex: 'title',
    },
  ];

  return (
    <>
      <Alert
        message={
          <Space wrap>
            <span>
              <FormattedMessage id="plugin.translate.tips" />
            </span>
            <Button size="small" onClick={handleShowPushLog}>
              <FormattedMessage id="plugin.translate.view-log" />
            </Button>
          </Space>
        }
      />
      <div className="mt-normal">
        {fetched && (
          <ProForm
            formRef={formRef}
            onFinish={onSubmit}
            initialValues={setting}
          >
            <ProFormRadio.Group
              name="engine"
              rules={[{ required: true }]}
              label={intl.formatMessage({
                id: 'plugin.translate.engine',
              })}
              options={[
                ...(Object.prototype.hasOwnProperty.call(
                  setting,
                  'openai_base_url',
                )
                  ? [
                      {
                        label: intl.formatMessage({
                          id: 'plugin.translate.engine.openai',
                        }),
                        value: 'openai',
                      },
                    ]
                  : []),
                {
                  label: intl.formatMessage({
                    id: 'plugin.translate.engine.ai',
                  }),
                  value: 'ai',
                },
                {
                  label: intl.formatMessage({
                    id: 'plugin.translate.engine.baidu',
                  }),
                  value: 'baidu',
                },
                {
                  label: intl.formatMessage({
                    id: 'plugin.translate.engine.youdao',
                  }),
                  value: 'youdao',
                },
                {
                  label: intl.formatMessage({
                    id: 'plugin.translate.engine.deepl',
                  }),
                  value: 'deepl',
                },
              ]}
            />
            <ProFormDependency name={['engine']}>
              {({ engine }) => (
                <>
                  {engine === 'openai' && (
                    <>
                      <ProFormText
                        name="openai_base_url"
                        label={intl.formatMessage({
                          id: 'plugin.translate.openai.base-url',
                        })}
                        placeholder="https://api.apiyi.com/v1"
                        rules={[{ required: true, whitespace: true }]}
                      />
                      <ProFormText.Password
                        name="openai_api_key"
                        label={intl.formatMessage({
                          id: 'plugin.translate.openai.api-key',
                        })}
                        extra={intl.formatMessage({
                          id: 'plugin.translate.openai.api-key.help',
                        })}
                        rules={[{ required: true, whitespace: true }]}
                      />
                      <ProFormText
                        name="openai_model"
                        label={intl.formatMessage({
                          id: 'plugin.translate.openai.model',
                        })}
                        rules={[{ required: true, whitespace: true }]}
                      />
                    </>
                  )}
                  {engine === 'ai' && (
                    <Button type="link" onClick={onOpenWritingSettings}>
                      <FormattedMessage id="plugin.translate.ai-config-tips" />
                    </Button>
                  )}
                  {engine === 'baidu' && (
                    <Card
                      size="small"
                      title={intl.formatMessage({
                        id: 'plugin.translate.engine.baidu',
                      })}
                      bordered={false}
                    >
                      <ProFormText
                        name="baidu_app_id"
                        label={intl.formatMessage({
                          id: 'plugin.translate.engine.baidu.app-id',
                        })}
                      />
                      <ProFormText
                        name="baidu_app_secret"
                        label={intl.formatMessage({
                          id: 'plugin.translate.engine.baidu.app-secret',
                        })}
                      />
                    </Card>
                  )}
                  {engine === 'youdao' && (
                    <Card
                      size="small"
                      title={intl.formatMessage({
                        id: 'plugin.translate.engine.youdao',
                      })}
                      bordered={false}
                    >
                      <ProFormText
                        name="youdao_app_key"
                        label={intl.formatMessage({
                          id: 'plugin.translate.engine.youdao.app-id',
                        })}
                      />
                      <ProFormText
                        name="youdao_app_secret"
                        label={intl.formatMessage({
                          id: 'plugin.translate.engine.youdao.app-secret',
                        })}
                      />
                    </Card>
                  )}
                  {engine === 'deepl' && (
                    <Card
                      size="small"
                      title={intl.formatMessage({
                        id: 'plugin.translate.engine.deepl',
                      })}
                      bordered={false}
                    >
                      <ProFormText
                        name="deepl_auth_key"
                        label={intl.formatMessage({
                          id: 'plugin.translate.engine.deepl.auth-key',
                        })}
                      />
                    </Card>
                  )}
                </>
              )}
            </ProFormDependency>
          </ProForm>
        )}
      </div>
      <Modal
        title={intl.formatMessage({ id: 'plugin.translate.logs' })}
        width={900}
        open={logVisible}
        onCancel={() => {
          setLogVisible(false);
        }}
        onOk={() => {
          setLogVisible(false);
        }}
      >
        <ProTable<any>
          actionRef={actionRef}
          rowKey="name"
          search={false}
          tableAlertOptionRender={false}
          request={async (params) => {
            return pluginTranslateLogs(params);
          }}
          columnsState={{
            persistenceKey: 'translate-log-table',
            persistenceType: 'localStorage',
          }}
          columns={columns}
          pagination={false}
          expandable={{
            expandedRowRender: (record) => (
              <Row gutter={16}>
                <Col span={12}>
                  <div className="mt-normal">
                    <div className="font-bold">
                      {intl.formatMessage({
                        id: 'content.translate.origin-content',
                      })}
                    </div>
                    <div className="mt-normal">{record.origin_content}</div>
                  </div>
                </Col>
                <Col span={12}>
                  <div className="mt-normal">
                    <div className="font-bold">
                      {intl.formatMessage({
                        id: 'plugin.translate.result',
                      })}
                    </div>
                    <div className="mt-normal">{record.content}</div>
                  </div>
                </Col>
              </Row>
            ),
            rowExpandable: (record) => record.origin_content !== '',
          }}
        />
      </Modal>
    </>
  );
};

export default TranslateSettings;
