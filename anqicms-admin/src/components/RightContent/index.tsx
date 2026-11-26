import { anqiRestart } from '@/services';
import { FormattedMessage, useIntl, useModel } from '@umijs/max';
import { Modal, Space, message } from 'antd';
import React from 'react';
import HeaderSearch from '../HeaderSearch';
import Avatar from './AvatarDropdown';
import './index.less';

export type SiderTheme = 'light' | 'dark';

// 右上角只保留：搜索、头像菜单、重启。站点名链接、帮助图标、后台语言切换已按需求去掉。
const GlobalHeaderRight: React.FC = () => {
  const { initialState } = useModel('@@initialState');
  const intl = useIntl();

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
      <Avatar menu />
      <div className="restart" onClick={confirmRestart}>
        <FormattedMessage id="component.right-content.restart" />
      </div>
    </Space>
  );
};
export default GlobalHeaderRight;
