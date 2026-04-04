export interface Song {
    id: number;
    name: string;
    artist: string;
    album: string;
    cover_url: string;
    lyrics: string;
    description: string;
    style_tags: string[];
    mood_tags: string[];
    scene_tags: string[];
}

export const searchSongs = async (query: string): Promise<Song[]> => {
    const res = await fetch(`/api/search?q=${encodeURIComponent(query)}&l=10`);
    if (!res.ok) {
        throw new Error('Search failed');
    }
    const data = await res.json();
    return data.songs || [];
};

export const triggerDailyJob = async (): Promise<void> => {
    const res = await fetch(`/api/trigger-job`, { method: 'POST' });
    if (!res.ok) {
        throw new Error('Trigger job failed');
    }
};
