import { Component, type ReactNode } from 'react';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, error: null };

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    console.error('ErrorBoundary caught:', error, info.componentStack);
  }

  handleReload = () => {
    window.location.reload();
  };

  handleGoHome = () => {
    window.location.href = '/';
  };

  render() {
    if (!this.state.hasError) {
      return this.props.children;
    }

    return (
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        minHeight: '100vh', padding: 24, fontFamily: 'system-ui, sans-serif',
      }}>
        <div style={{ textAlign: 'center', maxWidth: 480 }}>
          <h1 style={{ fontSize: 24, fontWeight: 600, marginBottom: 8 }}>
            Something went wrong
          </h1>
          <p style={{ color: '#64748b', marginBottom: 24, lineHeight: 1.5 }}>
            An unexpected error occurred. You can try reloading the page or going back to the dashboard.
          </p>
          {this.state.error && (
            <pre style={{
              background: '#f8fafc', border: '1px solid #e2e8f0', borderRadius: 8,
              padding: 16, marginBottom: 24, fontSize: 13, textAlign: 'left',
              overflow: 'auto', maxHeight: 120, color: '#dc2626',
            }}>
              {this.state.error.message}
            </pre>
          )}
          <div style={{ display: 'flex', gap: 12, justifyContent: 'center' }}>
            <button onClick={this.handleGoHome} style={{
              padding: '10px 20px', background: '#f1f5f9', border: '1px solid #e2e8f0',
              borderRadius: 8, cursor: 'pointer', fontSize: 14,
            }}>
              Go to Dashboard
            </button>
            <button onClick={this.handleReload} style={{
              padding: '10px 20px', background: '#2563eb', color: '#fff',
              border: 'none', borderRadius: 8, cursor: 'pointer', fontSize: 14,
            }}>
              Reload Page
            </button>
          </div>
        </div>
      </div>
    );
  }
}
