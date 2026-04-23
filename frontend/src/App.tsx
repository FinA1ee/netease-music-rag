import React, { useState } from 'react';
import {
  ConfigProvider, theme, Layout, Typography, Input,
  Button, Select, Space, Spin, Empty,
  notification, Tooltip
} from 'antd';
import {
  SearchOutlined, ThunderboltOutlined, RobotOutlined,
  LogoutOutlined, SoundOutlined, DatabaseOutlined
} from '@ant-design/icons';
import LoginPanel from './components/LoginPanel';
import SongCard from './components/SongCard';
import { useJobEvents, JobEvent } from './hooks/useJobEvents';
import { searchSongs, triggerDailyJob, SongResult } from './api';
import './index.css';

const { Header, Content } = Layout;
const { Title, Text } = Typography;
const { TextArea } = Input;

const App: React.FC = () => {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [query, setQuery] = useState('');
  const [limit, setLimit] = useState(5);
  const [results, setResults] = useState<SongResult[]>([]);
  const [searching, setSearching] = useState(false);
  const [jobLoading, setJobLoading] = useState(false);
  const [api, contextHolder] = notification.useNotification();

  // ── Real-time job event notifications via SSE ────────────────────────────
  useJobEvents((event: JobEvent) => {
    const p = event.payload ?? {};
    switch (event.type) {
      case 'job_started':
        api.info({ message: '🚀 Job Started', description: p.message, duration: 4 });
        break;
      case 'playlist_processing':
        api.open({
          type: 'info',
          message: `📋 Playlist ${p.index}/${p.total}`,
          description: `${p.name}  (${p.trackCount} tracks)`,
          duration: 5,
          icon: p.coverImgUrl
            ? <img src={p.coverImgUrl} alt="" style={{ width: 32, height: 32, borderRadius: 6, objectFit: 'cover' }} />
            : undefined,
        });
        break;
      case 'song_analysed':
        api.success({
          message: `✅ ${p.name}`,
          description: `${(p.artists as string[]).join('、')} — ${(p.style as string[] ?? []).slice(0, 2).join(' / ')}`,
          duration: 6,
        });
        break;
      case 'song_skipped':
        api.warning({
          message: `⏭ Skipped: ${p.name}`,
          description: p.reason,
          duration: 4,
        });
        break;
      case 'job_completed':
        api.success({ message: '🎉 Daily Job Done', description: p.message, duration: 8 });
        break;
      case 'embedding_started':
        api.info({ message: '🔢 Embedding Job Started', description: p.message, duration: 4 });
        break;
      case 'embedding_done':
        api.success({ message: '🔢 Embedding Done', description: p.message, duration: 8 });
        break;
    }
  });

  const handleSearch = async () => {
    if (!query.trim()) return;
    setSearching(true);
    setResults([]);
    try {
      const data = await searchSongs(query, limit);
      setResults(data.results ?? []);
      if ((data.results ?? []).length === 0) {
        api.info({ message: 'No results', description: 'Try a different description.' });
      }
    } catch (err: any) {
      api.error({ message: 'Search failed', description: err.message });
    } finally {
      setSearching(false);
    }
  };

  const handleTriggerJob = async () => {
    setJobLoading(true);
    try {
      await triggerDailyJob();
      // Embedding job is now chained automatically on the backend
      // after the daily recommendation job completes.
    } catch (err: any) {
      api.error({ message: 'Failed to trigger job', description: err.message });
    } finally {
      setJobLoading(false);
    }
  };

  // ── Login screen ────────────────────────────────────────────────────────────
  if (!isLoggedIn) {
    return (
      <ConfigProvider theme={{ algorithm: theme.darkAlgorithm, token: { colorPrimary: '#7c3aed' } }}>
        {contextHolder}
        <div className="app-bg">
          <div className="login-center">
            <LoginPanel onLoginSuccess={() => setIsLoggedIn(true)} />
          </div>
        </div>
      </ConfigProvider>
    );
  }

  // ── Main app ────────────────────────────────────────────────────────────────
  return (
    <ConfigProvider theme={{ algorithm: theme.darkAlgorithm, token: { colorPrimary: '#7c3aed' } }}>
      {contextHolder}
      <div className="app-bg">
        <Layout style={{ background: 'transparent', minHeight: '100vh' }}>

          {/* Header */}
          <Header style={{
            background: 'rgba(10, 10, 30, 0.85)',
            backdropFilter: 'blur(12px)',
            borderBottom: '1px solid rgba(99,102,241,0.2)',
            display: 'flex', alignItems: 'center', justifyContent: 'space-between',
            padding: '0 32px', position: 'sticky', top: 0, zIndex: 10,
          }}>
            <Space>
              <SoundOutlined style={{ fontSize: 22, color: '#7c3aed' }} />
              <Title level={4} style={{ color: '#e2e8f0', margin: 0 }}>
                NetEase Music AI
              </Title>
            </Space>
            <Tooltip title="Sign out">
              <Button
                type="text" icon={<LogoutOutlined />}
                style={{ color: '#64748b' }}
                onClick={() => setIsLoggedIn(false)}
              />
            </Tooltip>
          </Header>

          <Content style={{ maxWidth: 780, margin: '0 auto', width: '100%', padding: '48px 24px' }}>

            {/* Hero */}
            <div style={{ textAlign: 'center', marginBottom: 48 }}>
              <Title level={1} style={{
                color: '#e2e8f0', margin: 0, marginBottom: 12,
                background: 'linear-gradient(135deg, #c4b5fd, #818cf8)',
                WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent',
              }}>
                Find Your Perfect Song
              </Title>
              <Text style={{ color: '#94a3b8', fontSize: 16 }}>
                Describe the music you're in the mood for — our AI will find it.
              </Text>
            </div>

            {/* Search Card */}
            <div className="glass-card" style={{ padding: 28, marginBottom: 28 }}>
              <TextArea
                value={query}
                onChange={e => setQuery(e.target.value)}
                placeholder="e.g. 我想听一首快节奏的男声歌曲，充满力量感的摇滚..."
                autoSize={{ minRows: 3, maxRows: 6 }}
                style={{
                  background: 'rgba(255,255,255,0.04)',
                  border: '1px solid rgba(99,102,241,0.3)',
                  borderRadius: 10, color: '#e2e8f0', fontSize: 15, marginBottom: 16,
                  resize: 'none',
                }}
                onPressEnter={e => { if (!e.shiftKey) { e.preventDefault(); handleSearch(); } }}
              />

              <div style={{ display: 'flex', gap: 12, alignItems: 'center', flexWrap: 'wrap' }}>
                <div style={{ flex: 1, minWidth: 120 }}>
                  <Text style={{ color: '#64748b', fontSize: 13, display: 'block', marginBottom: 6 }}>
                    Results
                  </Text>
                  <Select
                    value={limit}
                    onChange={setLimit}
                    style={{ width: '100%' }}
                    options={[1, 2, 3, 5, 8, 10].map(n => ({ value: n, label: `${n} songs` }))}
                  />
                </div>

                <Button
                  type="primary"
                  icon={<SearchOutlined />}
                  size="large"
                  loading={searching}
                  onClick={handleSearch}
                  disabled={!query.trim()}
                  style={{
                    marginTop: 22,
                    background: 'linear-gradient(135deg, #6366f1, #7c3aed)',
                    border: 'none', height: 42, borderRadius: 10, fontWeight: 600,
                    paddingInline: 28,
                  }}
                >
                  Search
                </Button>

                <Button
                  icon={<ThunderboltOutlined />}
                  size="large"
                  loading={jobLoading}
                  onClick={handleTriggerJob}
                  style={{
                    marginTop: 22, height: 42, borderRadius: 10,
                    background: 'rgba(245,158,11,0.1)',
                    border: '1px solid rgba(245,158,11,0.35)',
                    color: '#f59e0b', fontWeight: 600,
                  }}
                >
                  Daily Recommendation
                </Button>
              </div>
            </div>

            {/* Results */}
            {searching && (
              <div style={{ textAlign: 'center', padding: 48 }}>
                <Spin size="large" />
                <div style={{ marginTop: 16, color: '#94a3b8' }}>
                  Searching the vector space...
                </div>
              </div>
            )}

            {!searching && results.length > 0 && (
              <>
                <div style={{ display: 'flex', alignItems: 'center', marginBottom: 20, gap: 10 }}>
                  <RobotOutlined style={{ color: '#7c3aed', fontSize: 16 }} />
                  <Text style={{ color: '#94a3b8' }}>
                    Found <strong style={{ color: '#c4b5fd' }}>{results.length}</strong> matches for "{query}"
                  </Text>
                </div>
                <Space direction="vertical" style={{ width: '100%' }} size={12}>
                  {results.map((song, idx) => (
                    <SongCard key={song.song_id} song={song} rank={idx + 1} />
                  ))}
                </Space>
              </>
            )}

            {!searching && results.length === 0 && query && (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={<Text style={{ color: '#64748b' }}>No songs found. Try a different description.</Text>}
              />
            )}

            {!searching && results.length === 0 && !query && (
              <div className="glass-card" style={{ padding: 32, textAlign: 'center' }}>
                <DatabaseOutlined style={{ fontSize: 36, color: '#4b5563', marginBottom: 12 }} />
                <div style={{ color: '#4b5563', lineHeight: 2 }}>
                  <div>Try: <Text style={{ color: '#818cf8' }}>"一首带有钢琴的日语歌，安静治愈"</Text></div>
                  <div>Or: <Text style={{ color: '#818cf8' }}>"充满力量的粤语摇滚，男声"</Text></div>
                  <div>Or: <Text style={{ color: '#818cf8' }}>"适合睡前听的温柔女声"</Text></div>
                </div>
              </div>
            )}

          </Content>
        </Layout>
      </div>
    </ConfigProvider>
  );
};

export default App;
