import { anqiRestart, getSiteInfo } from '@/services';
import { GlobalOutlined, QuestionCircleOutlined } from '@ant-design/icons';
import {
  FormattedMessage,
  SelectLang as UmiSelectLang,
  useIntl,
  useModel,
} from '@umijs/max';
import { Modal, Space, message } from 'antd';
import React, { useEffect, useState } from 'react';
import HeaderSearch from '../HeaderSearch';
import Avatar from './AvatarDropdown';
import './index.less';

export type SiderTheme = 'light' | 'dark';

const GlobalHeaderRight: React.FC = () => {
  const { initialState } = useModel('@@initialState');
  const [siteInfo, setSiteInfo] = useState<any>({});
  const intl = useIntl();

  const initSiteInfo = async () => {
    getSiteInfo({}).then((res) => {
      setSiteInfo(res?.data || {});
    });
  };

  useEffect(() => {
    initSiteInfo();
  }, []);

  if (!initialState || !initialState.settings) {
    return null;
  }

  const { navTheme, layout } = initialState.settings;
  let className = 'right';

  if ((navTheme === 'realDark' && layout === 'top') || layout === 'mix') {
    className = `right  dark`;
  }

  const confirmRestart = () => {
    Modal.confirm({
      title: intl.formatMessage({
        id: 'component.right-content.restart.confirm',
      }),
      content: intl.formatMessage({
        id: 'component.right-content.restart.confirm.content',
      }),
      onOk: () => {
        const hide2 = message.loading(
          intl.formatMessage({ id: 'component.right-content.restart.doing' }),
          0,
        );
        anqiRestart({})
          .then(() => {})
          .catch(() => {})
          .finally(() => {
            setTimeout(() => {
              hide2();
              window.location.reload();
            }, 3000);
          });
      },
    });
  };

  return (
    <Space className={className}>
      <HeaderSearch
        placeholder={intl.formatMessage({
          id: 'component.right-content.search.placeholder',
        })}
      />
      <a
        href={siteInfo.base_url}
        target={'_blank'}
        className="site-info-item action"
        rel="noreferrer"
      >
        <GlobalOutlined style={{ marginRight: 5 }} />
        {siteInfo.name || siteInfo.base_url}
      </a>
      <span
        style={{ fontSize: 18 }}
        className="action"
        onClick={() => {
          window.open('https://www.anqicms.com/');
        }}
      >
        <QuestionCircleOutlined />
      </span>
      <UmiSelectLang />
      <Avatar menu />
      <div className="restart" onClick={confirmRestart}>
        <FormattedMessage id="component.right-content.restart" />
      </div>
    </Space>
  );
};
export default GlobalHeaderRight;
