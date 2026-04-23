import React from 'react';
import { Card, Tag, Typography, Space, Avatar } from 'antd';
import { UserOutlined, SoundOutlined } from '@ant-design/icons';
import { SongResult } from '../api';

const { Text, Title } = Typography;

const STYLE_COLORS = ['#7c3aed', '#6366f1', '#0891b2', '#0d9488'];
const MOOD_COLORS = ['#b45309', '#15803d', '#be185d', '#6d28d9'];

interface Props {
  song: SongResult;
  rank: number;
  query: string;
}

const SongCard: React.FC<Props> = ({ song, rank, query }) => {
  const artists = song.artists?.map(a => a.name).join('、') ?? '—';
  const albumName = song.album?.name ?? '—';
  const coverUrl = song.album?.picUrl;
  const styles: string[] = Array.isArray(song.style) ? song.style : [];
  const moods: string[] = Array.isArray(song.mood) ? song.mood : [];
  const keywords: string[] = Array.isArray(song.keywords) ? song.keywords : [];
  const normalizedQuery = query.trim().toLowerCase();
  const matchedKeywords = keywords.filter((k) => {
    const normalizedKeyword = k.trim().toLowerCase();
    return normalizedKeyword.length > 0 && (
      normalizedQuery.includes(normalizedKeyword) || normalizedKeyword.includes(normalizedQuery)
    );
  });

  return (
    <Card
      hoverable
      style={{
        background: 'rgba(15, 23, 42, 0.75)',
        border: '1px solid rgba(99,102,241,0.2)',
        borderRadius: 14,
        backdropFilter: 'blur(10px)',
        transition: 'all 0.25s ease',
        overflow: 'hidden',
      }}
      bodyStyle={{ padding: '16px 20px' }}
    >
      <div style={{ display: 'flex', gap: 16, alignItems: 'flex-start' }}>
        {/* Rank + cover */}
        <div style={{ position: 'relative', flexShrink: 0 }}>
          <Avatar
            src={coverUrl}
            size={72}
            shape="square"
            icon={<SoundOutlined />}
            style={{ borderRadius: 10, border: '1px solid rgba(99,102,241,0.3)' }}
          />
          <div style={{
            position: 'absolute', top: -8, left: -8,
            width: 22, height: 22, borderRadius: '50%',
            background: 'linear-gradient(135deg, #6366f1, #7c3aed)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontSize: 11, fontWeight: 700, color: '#fff',
          }}>
            {rank}
          </div>
        </div>

        {/* Info */}
        <div style={{ flex: 1, minWidth: 0 }}>
          <Title level={5} style={{ color: '#e2e8f0', margin: 0, marginBottom: 2 }} ellipsis>
            {song.name}
          </Title>
          <Space size={4} style={{ marginBottom: 10 }}>
            <UserOutlined style={{ color: '#94a3b8', fontSize: 12 }} />
            <Text style={{ color: '#94a3b8', fontSize: 13 }}>{artists}</Text>
            <Text style={{ color: '#4b5563', fontSize: 13 }}>·</Text>
            <Text style={{ color: '#64748b', fontSize: 13 }} ellipsis>{albumName}</Text>
          </Space>

          {/* Style tags */}
          {styles.length > 0 && (
            <div style={{ marginBottom: 6 }}>
              {styles.slice(0, 4).map((s, i) => (
                <Tag key={i} style={{
                  background: `${STYLE_COLORS[i % STYLE_COLORS.length]}22`,
                  borderColor: STYLE_COLORS[i % STYLE_COLORS.length],
                  color: STYLE_COLORS[i % STYLE_COLORS.length],
                  borderRadius: 6, fontSize: 12, marginBottom: 4,
                }}>
                  {s}
                </Tag>
              ))}
            </div>
          )}

          {/* Mood tags */}
          {moods.length > 0 && (
            <div style={{ marginBottom: 6 }}>
              {moods.slice(0, 4).map((m, i) => (
                <Tag key={i} style={{
                  background: `${MOOD_COLORS[i % MOOD_COLORS.length]}22`,
                  borderColor: MOOD_COLORS[i % MOOD_COLORS.length],
                  color: MOOD_COLORS[i % MOOD_COLORS.length],
                  borderRadius: 6, fontSize: 12, marginBottom: 4,
                }}>
                  {m}
                </Tag>
              ))}
            </div>
          )}

          {/* Keywords */}
          {keywords.length > 0 && (
            <div>
              {keywords.slice(0, 5).map((k, i) => (
                <Tag key={i} style={{
                  background: 'rgba(255,255,255,0.04)',
                  borderColor: 'rgba(255,255,255,0.1)',
                  color: '#94a3b8', borderRadius: 6, fontSize: 11, marginBottom: 4,
                }}>
                  #{k}
                </Tag>
              ))}
            </div>
          )}

          {matchedKeywords.length > 0 && (
            <div style={{
              marginTop: 10,
              padding: '8px 10px',
              borderRadius: 8,
              background: 'rgba(99,102,241,0.12)',
              border: '1px solid rgba(99,102,241,0.32)',
            }}>
              <Text style={{ color: '#c4b5fd', fontSize: 12 }}>
                推荐理由：关键词匹配 {matchedKeywords.slice(0, 3).join(' / ')}
              </Text>
            </div>
          )}
        </div>
      </div>
    </Card>
  );
};

export default SongCard;
