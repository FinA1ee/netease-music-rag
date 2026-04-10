import React, { useState, useEffect, useRef } from 'react';
import {
  Button, Spin, Typography, Space, Card, Tag, Avatar
} from 'antd';
import {
  QrcodeOutlined, ReloadOutlined, CheckCircleOutlined,
  ClockCircleOutlined, LoadingOutlined
} from '@ant-design/icons';
import { generateQR, pollLoginStatus, QRResponse } from '../api';

const { Text, Title } = Typography;

interface Props {
  onLoginSuccess: () => void;
}

type LoginState = 'idle' | 'loading' | 'waiting' | 'scanned' | 'success' | 'expired' | 'error';

const STATUS_MSG: Record<number, { state: LoginState; label: string }> = {
  800: { state: 'expired', label: 'QR code expired. Please refresh.' },
  801: { state: 'waiting', label: 'Waiting for scan...' },
  802: { state: 'scanned', label: 'Scanned! Please confirm in app.' },
  803: { state: 'success', label: 'Login successful!' },
};

const LoginPanel: React.FC<Props> = ({ onLoginSuccess }) => {
  const [qr, setQr] = useState<QRResponse | null>(null);
  const [loginState, setLoginState] = useState<LoginState>('idle');
  const [statusLabel, setStatusLabel] = useState('');
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const stopPolling = () => {
    if (pollRef.current) {
      clearInterval(pollRef.current);
      pollRef.current = null;
    }
  };

  const startPolling = (key: string) => {
    stopPolling();
    pollRef.current = setInterval(async () => {
      try {
        const { code, message } = await pollLoginStatus(key);
        const info = STATUS_MSG[code] ?? { state: 'waiting' as LoginState, label: message };
        setLoginState(info.state);
        setStatusLabel(info.state === 'waiting' ? 'Waiting for scan...' : info.label);

        if (code === 803) {
          stopPolling();
          setTimeout(onLoginSuccess, 800);
        }
        if (code === 800) {
          stopPolling();
        }
      } catch {
        // network glitch — keep polling
      }
    }, 3000);
  };

  const MAX_QR_RETRIES = 3;
  const QR_RETRY_DELAY_MS = 2000;

  const sleep = (ms: number) => new Promise(res => setTimeout(res, ms));

  const handleGenerateQR = async () => {
    stopPolling();
    setLoginState('loading');
    setStatusLabel('');

    let lastErr: Error | null = null;
    for (let attempt = 1; attempt <= MAX_QR_RETRIES; attempt++) {
      try {
        const data = await generateQR();
        setQr(data);
        setLoginState('waiting');
        setStatusLabel('Waiting for scan...');
        startPolling(data.key);
        return; // success — exit
      } catch (err: any) {
        lastErr = err;
        if (attempt < MAX_QR_RETRIES) {
          setStatusLabel(`QR generation failed (attempt ${attempt}/${MAX_QR_RETRIES}), retrying...`);
          await sleep(QR_RETRY_DELAY_MS);
        }
      }
    }

    // All retries exhausted
    setLoginState('error');
    setStatusLabel(`Failed to generate QR after ${MAX_QR_RETRIES} attempts: ${lastErr?.message ?? 'Unknown error'}`);
  };

  useEffect(() => () => stopPolling(), []);

  const statusColor: Record<LoginState, string> = {
    idle: '#666', loading: '#7c3aed', waiting: '#60a5fa',
    scanned: '#f59e0b', success: '#34d399', expired: '#f87171', error: '#f87171',
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 28 }}>
      <div style={{ textAlign: 'center' }}>
        <Title level={2} style={{ color: '#e2e8f0', marginBottom: 8 }}>
          🎵 NetEase Music AI
        </Title>
        <Text style={{ color: '#94a3b8', fontSize: 15 }}>
          Sign in with your NetEase Cloud Music account to get personalised recommendations
        </Text>
      </div>

      <Card
        style={{
          width: 320,
          background: 'rgba(15, 23, 42, 0.8)',
          border: '1px solid rgba(99,102,241,0.3)',
          borderRadius: 16,
          backdropFilter: 'blur(12px)',
        }}
        bodyStyle={{ padding: 28, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 20 }}
      >
        {/* QR area */}
        <div style={{
          width: 200, height: 200,
          background: 'rgba(255,255,255,0.04)',
          borderRadius: 12,
          border: '1px dashed rgba(99,102,241,0.4)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          overflow: 'hidden',
          position: 'relative',
        }}>
          {loginState === 'loading' && (
            <Spin indicator={<LoadingOutlined style={{ fontSize: 40, color: '#7c3aed' }} spin />} />
          )}
          {(loginState === 'idle' || loginState === 'error') && (
            <div style={{ textAlign: 'center', color: '#4b5563' }}>
              <QrcodeOutlined style={{ fontSize: 48 }} />
              <div style={{ marginTop: 8, fontSize: 13 }}>Click below to generate</div>
            </div>
          )}
          {qr && loginState !== 'loading' && (
            <>
              <img
                src={qr.qr_img}
                alt="QR Code"
                style={{
                  width: '100%', height: '100%', objectFit: 'contain',
                  filter: loginState === 'expired' ? 'blur(4px) brightness(0.4)' : 'none',
                  transition: 'filter 0.3s',
                }}
              />
              {loginState === 'success' && (
                <div style={{
                  position: 'absolute', inset: 0,
                  background: 'rgba(6,78,59,0.75)',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  flexDirection: 'column', gap: 8,
                }}>
                  <CheckCircleOutlined style={{ fontSize: 48, color: '#34d399' }} />
                  <Text style={{ color: '#34d399', fontWeight: 600 }}>Logged In</Text>
                </div>
              )}
            </>
          )}
        </div>

        {/* Status */}
        {statusLabel && (
          <Space>
            {loginState === 'waiting' && <ClockCircleOutlined style={{ color: '#60a5fa' }} />}
            {loginState === 'scanned' && <LoadingOutlined style={{ color: '#f59e0b' }} spin />}
            {loginState === 'success' && <CheckCircleOutlined style={{ color: '#34d399' }} />}
            <Text style={{ color: statusColor[loginState], fontSize: 13 }}>{statusLabel}</Text>
          </Space>
        )}

        {/* Buttons */}
        <Space direction="vertical" style={{ width: '100%' }}>
          <Button
            type="primary"
            block
            icon={loginState === 'expired' ? <ReloadOutlined /> : <QrcodeOutlined />}
            loading={loginState === 'loading'}
            onClick={handleGenerateQR}
            disabled={loginState === 'success'}
            style={{
              background: 'linear-gradient(135deg, #6366f1, #7c3aed)',
              border: 'none', height: 40, borderRadius: 8, fontWeight: 600,
            }}
          >
            {qr ? 'Refresh QR Code' : 'Generate QR Code'}
          </Button>
        </Space>

        <Text style={{ color: '#4b5563', fontSize: 12, textAlign: 'center' }}>
          Open NetEase Cloud Music app → Me → Scan QR code
        </Text>
      </Card>
    </div>
  );
};

export default LoginPanel;
