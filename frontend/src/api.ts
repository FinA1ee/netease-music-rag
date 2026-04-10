// ─── Types ────────────────────────────────────────────────────────────────────

export interface Artist {
  id: number;
  name: string;
}

export interface Album {
  id: number;
  name: string;
  picUrl: string;
}

export interface SongResult {
  song_id: number;
  name: string;
  artists: Artist[];
  album: Album;
  style: string[] | null;
  mood: string[] | null;
  keywords: string[] | null;
  popularity?: number;
}

export interface SearchResponse {
  query: string;
  results: SongResult[];
}

export interface QRResponse {
  qr_img: string;   // base64 data URL
  key: string;
}

export interface LoginStatusResponse {
  code: number;    // 800=expired 801=waiting 802=scanned 803=success
  message: string;
}

// ─── API Calls ────────────────────────────────────────────────────────────────

const BASE = '';

export async function searchSongs(query: string, limit: number): Promise<SearchResponse> {
  const res = await fetch(`${BASE}/api/search?q=${encodeURIComponent(query)}&l=${limit}`);
  if (!res.ok) throw new Error(`Search failed (${res.status})`);
  return res.json();
}

export async function generateQR(): Promise<QRResponse> {
  const res = await fetch(`${BASE}/api/login/qr`, { method: 'POST' });
  if (!res.ok) throw new Error(`QR generation failed (${res.status})`);
  return res.json();
}

export async function pollLoginStatus(key: string): Promise<LoginStatusResponse> {
  const res = await fetch(`${BASE}/api/login/status?key=${encodeURIComponent(key)}`);
  if (!res.ok) throw new Error(`Status check failed (${res.status})`);
  return res.json();
}

export async function triggerDailyJob(): Promise<void> {
  const res = await fetch(`${BASE}/api/trigger-job`, { method: 'POST' });
  if (!res.ok) throw new Error(`Trigger job failed (${res.status})`);
}

export async function triggerEmbeddingJob(): Promise<void> {
  const res = await fetch(`${BASE}/api/trigger-embedding`, { method: 'POST' });
  if (!res.ok) throw new Error(`Trigger embedding failed (${res.status})`);
}
