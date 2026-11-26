import { PageContainer } from '@ant-design/pro-components';
import { FormattedMessage, history, useModel } from '@umijs/max';
import { Card, Col, Row } from 'antd';
import React, { useState } from 'react';
import routes from '../../../config/routes';
import './index.less';

import icon_ai from '@/images/icon_ai.png';
import icon_akismet from '@/images/icon_akismet.png';
import icon_anchor from '@/images/icon_anchor.png';
import icon_backup from '@/images/icon_backup.png';
import icon_collector from '@/images/icon_collector.png';
import icon_comment from '@/images/icon_comment.png';
import icon_fileupload from '@/images/icon_fileupload.png';
import icon_finance from '@/images/icon_finance.png';
import icon_friendlink from '@/images/icon_friendlink.png';
import icon_fulltext from '@/images/icon_fulltext.png';
import icon_google from '@/images/icon_google.png';
import icon_group from '@/images/icon_group.png';
import icon_guestbook from '@/images/icon_guestbook.png';
import icon_htmlcache from '@/images/icon_htmlcache.png';
import icon_importapi from '@/images/icon_importapi.png';
import icon_interference from '@/images/icon_interference.png';
import icon_jsonld from '@/images/icon_jsonld.png';
import icon_keyword from '@/images/icon_keyword.png';
import icon_limiter from '@/images/icon_limiter.png';
import icon_llms from '@/images/icon_llms.png';
import icon_material from '@/images/icon_material.png';
import icon_multilang from '@/images/icon_multilang.png';
import icon_order from '@/images/icon_order.png';
import icon_pay from '@/images/icon_pay.png';
import icon_push from '@/images/icon_push.png';
import icon_redirect from '@/images/icon_redirect.png';
import icon_replace from '@/images/icon_replace.png';
import icon_retailer from '@/images/icon_retailer.png';
import icon_rewrite from '@/images/icon_rewrite.png';
import icon_robots from '@/images/icon_robots.png';
import icon_sendmail from '@/images/icon_sendmail.png';
import icon_sitemap from '@/images/icon_sitemap.png';
import icon_storage from '@/images/icon_storage.png';
import icon_timefactor from '@/images/icon_timefactor.png';
import icon_titleimage from '@/images/icon_titleimage.png';
import icon_transfer from '@/images/icon_transfer.png';
import icon_translate from '@/images/icon_translate.png';
import icon_user from '@/images/icon_user.png';
import icon_watermark from '@/images/icon_watermark.png';
import icon_weapp from '@/images/icon_weapp.png';
import icon_wechat from '@/images/icon_wechat.png';

const PluginIndex: React.FC = () => {
  const { initialState } = useModel('@@initialState');
  const [type, setType] = useState<string>('');
  const jumpToPlugin = (item: any) => {
    history.push(item.path);
  };

  const onChangeType = (val: string) => {
    setType(val);
  };

  // 把 /plugin 下的路由整理成卡片列表。
  // 分组（如 seo、system）取其子路由；叶子路由（如 google、llms，自身就是一个功能页，没有 routes）
  // 直接作为一张卡片，避免对其取 .routes.map 时报 undefined.map 而导致整个页面崩溃。
  const pluginCards: any[] = [];
  for (const item of routes as any[]) {
    if (item.path !== '/plugin' || !Array.isArray(item.routes)) {
      continue;
    }
    for (const inner of item.routes) {
      const visible =
        !inner.hideInMenu &&
        inner.name &&
        !inner.unaccessible &&
        (!type || type === inner.type || type === inner.name);
      if (!visible) {
        continue;
      }
      const leaves = Array.isArray(inner.routes) ? inner.routes : [inner];
      for (const leaf of leaves) {
        if (!leaf.name || leaf.unaccessible) {
          continue;
        }
        const isGroupItself = leaf !== inner;
        pluginCards.push({
          key: leaf.path,
          path: leaf.path,
          icon: leaf.icon,
          titleId: isGroupItself
            ? 'menu.plugin.' + inner.name + '.' + leaf.name
            : 'menu.plugin.' + inner.name,
        });
      }
    }
  }

  let permissions =
    initialState?.currentUser?.group?.setting?.permissions || [];
  if (
    initialState?.currentUser?.id !== 1 &&
    initialState?.currentUser?.group_id !== 1
  ) {
    for (let i in routes) {
      if (routes[i].path === '/plugin') {
        // 需要处理
        if (!routes[i].routes) {
          continue;
        }
        for (let j = 0; j < routes[i].routes.length; j++) {
          let item = routes[i].routes[j] as any;
          if (!item.routes) {
            continue;
          }
          for (let k in item.routes) {
            if (permissions.indexOf(item.routes[k].path) === -1) {
              item.routes[k].unaccessible = true;
            }
          }
        }
      }
    }
  }

  const getIcon = (icon: string) => {
    switch (icon) {
      case 'anchor':
        return icon_anchor;
      case 'backup':
        return icon_backup;
      case 'collector':
        return icon_collector;
      case 'comment':
        return icon_comment;
      case 'fileupload':
        return icon_fileupload;
      case 'finance':
        return icon_finance;
      case 'friendlink':
        return icon_friendlink;
      case 'fulltext':
        return icon_fulltext;
      case 'group':
        return icon_group;
      case 'guestbook':
        return icon_guestbook;
      case 'importapi':
        return icon_importapi;
      case 'keyword':
        return icon_keyword;
      case 'material':
        return icon_material;
      case 'order':
        return icon_order;
      case 'pay':
        return icon_pay;
      case 'push':
        return icon_push;
      case 'redirect':
        return icon_redirect;
      case 'replace':
        return icon_replace;
      case 'retailer':
        return icon_retailer;
      case 'rewrite':
        return icon_rewrite;
      case 'robots':
        return icon_robots;
      case 'sendmail':
        return icon_sendmail;
      case 'sitemap':
        return icon_sitemap;
      case 'storage':
        return icon_storage;
      case 'transfer':
        return icon_transfer;
      case 'user':
        return icon_user;
      case 'weapp':
        return icon_weapp;
      case 'wechat':
        return icon_wechat;
      case 'titleimage':
        return icon_titleimage;
      case 'htmlcache':
        return icon_htmlcache;
      case 'ai':
        return icon_ai;
      case 'timefactor':
        return icon_timefactor;
      case 'interference':
        return icon_interference;
      case 'watermark':
        return icon_watermark;
      case 'limiter':
        return icon_limiter;
      case 'multilang':
        return icon_multilang;
      case 'translate':
        return icon_translate;
      case 'jsonld':
        return icon_jsonld;
      case 'akismet':
        return icon_akismet;
      case 'google':
        return icon_google;
      case 'llms':
        return icon_llms;
      default:
        return icon_backup;
    }
  };

  return (
    <PageContainer>
      <Card
        title={
          <div className="module-tags">
            <div
              className={'module-tag ' + (!type ? 'active' : '')}
              onClick={() => {
                onChangeType('');
              }}
            >
              <FormattedMessage id="plugin.type.all" />
            </div>
            {routes.map((item: any) => {
              if (item.path === '/plugin') {
                return item.routes.map((inner: any) => {
                  if (!inner.hideInMenu && inner.name && !inner.unaccessible) {
                    return (
                      <div
                        key={inner.name}
                        className={
                          'module-tag ' + (type === inner.name ? 'active' : '')
                        }
                        onClick={() => {
                          onChangeType(inner.name);
                        }}
                      >
                        <FormattedMessage id={'menu.plugin.' + inner.name} />
                      </div>
                    );
                  } else {
                    return null;
                  }
                });
              } else {
                return null;
              }
            })}
          </div>
        }
      >
        <Row gutter={[20, 20]}>
          {pluginCards.map((card) => (
            <Col key={card.key} sm={6} xs={12}>
              <div
                className="plugin-item"
                onClick={() => {
                  jumpToPlugin(card);
                }}
              >
                <img className="avatar" src={getIcon(card.icon)} />
                <div className="info">
                  <div className="title">
                    <FormattedMessage id={card.titleId} />
                  </div>
                </div>
              </div>
            </Col>
          ))}
        </Row>
      </Card>
    </PageContainer>
  );
};

export default PluginIndex;
