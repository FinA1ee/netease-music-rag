import React, { useState } from 'react';
import { searchSongs, triggerDailyJob, Song } from './api';
import SongCard from './components/SongCard';

const App: React.FC = () => {
    const [query, setQuery] = useState('');
    const [songs, setSongs] = useState<Song[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const handleSearch = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!query.trim()) return;
        setLoading(true);
        setError('');
        try {
            const results = await searchSongs(query);
            setSongs(results);
        } catch (err: any) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    const handleTriggerJob = async () => {
        try {
            await triggerDailyJob();
            alert("Background job triggered! Check server logs.");
        } catch (err: any) {
            alert("Failed: " + err.message);
        }
    }

    return (
        <div className="app-container">
            <header className="app-header">
                <h1>🎵 NetEase Music RAG AI</h1>
                <p>Natural language music recommendation via Gemini & pgvector</p>
            </header>

            <main className="main-content">
                <form className="search-form" onSubmit={handleSearch}>
                    <input 
                        type="text" 
                        value={query} 
                        onChange={(e) => setQuery(e.target.value)} 
                        placeholder="E.g. I want an upbeat song for a road trip..."
                        disabled={loading}
                    />
                    <button type="submit" disabled={loading}>
                        {loading ? 'Searching...' : 'Search'}
                    </button>
                </form>

                <div className="controls">
                    <button className="secondary-btn" onClick={handleTriggerJob}>
                        Manual Trigger Daily Crawl
                    </button>
                </div>

                {error && <div className="error">{error}</div>}

                <div className="results-container">
                    {songs.map(song => (
                        <SongCard key={song.id} song={song} />
                    ))}
                    {songs.length === 0 && !loading && !error && (
                        <div className="empty-state">
                            Try searching for a song based on your mood, scene, or style!
                        </div>
                    )}
                </div>
            </main>
        </div>
    );
};

export default App;
