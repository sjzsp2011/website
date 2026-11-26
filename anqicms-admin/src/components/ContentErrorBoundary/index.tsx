import { useIntl, useLocation } from '@umijs/max';
import { Button, Result } from 'antd';
import React from 'react';

type InnerProps = {
  children?: React.ReactNode;
  // 当前路由路径，用来判断出错后用户是否已经切换到别的页面
  pathname: string;
  title: string;
  retryText: string;
};

type InnerState = {
  error: Error | null;
};

/**
 * 替换 pro-layout 默认的 ErrorBoundary。
 * 默认实现一旦进入错误状态，只有刷新整个页面才能恢复：某个页面渲染出错后，点其它任何菜单/tab
 * 看到的都是同一个错误页。这里改成：
 * 1. 出错后切换到别的页面时自动清除错误，正常渲染新页面；
 * 2. 提供「重试」按钮，原地重新渲染当前页面。
 */
class ContentErrorBoundaryInner extends React.Component<InnerProps, InnerState> {
  state: InnerState = { error: null };

  // 记录是在哪个页面上出的错
  private errorPath = '';

  static getDerivedStateFromError(error: Error): Partial<InnerState> {
    return { error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    this.errorPath = this.props.pathname;
    // 保留完整堆栈，方便在控制台排查
    console.error(error, errorInfo);
  }

  componentDidUpdate() {
    const { error } = this.state;
    // 已经离开出错页面：清掉错误状态
    if (error && this.props.pathname !== this.errorPath) {
      this.setState({ error: null });
    }
  }

  private handleRetry = () => {
    this.setState({ error: null });
  };

  render() {
    const { error } = this.state;
    const hasError = !!error && this.props.pathname === this.errorPath;

    if (!hasError) {
      return this.props.children;
    }

    return (
      <Result
        status="error"
        title={this.props.title}
        subTitle={error?.message}
        extra={
          <Button type="primary" onClick={this.handleRetry}>
            {this.props.retryText}
          </Button>
        }
      />
    );
  }
}

const ContentErrorBoundary: React.FC<{ children?: React.ReactNode }> = ({
  children,
}) => {
  const { pathname } = useLocation();
  const intl = useIntl();

  return (
    <ContentErrorBoundaryInner
      pathname={pathname}
      title={intl.formatMessage({ id: 'component.errorBoundary.title' })}
      retryText={intl.formatMessage({ id: 'component.errorBoundary.retry' })}
    >
      {children}
    </ContentErrorBoundaryInner>
  );
};

export default ContentErrorBoundary;
