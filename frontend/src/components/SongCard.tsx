import React from 'react';
import { Song } from '../api';

interface SongCardProps {
    song: Song;
}

const SongCard: React.FC<SongCardProps> = ({ song }) => {
    return (
        <div className="song-card">
            <div className="song-header">
                <img src={song.cover_url} alt={song.album} className="song-cover" />
                <div className="song-info">
                    <h3>{song.name}</h3>
                    <p className="artist">{song.artist} • {song.album}</p>
                </div>
            </div>
            
            <div className="song-description">
                <p><strong>AI Description:</strong> {song.description}</p>
                <div className="tags-container">
                    {song.style_tags && song.style_tags.map((tag, idx) => (
                        <span key={`style-${idx}`} className="tag style-tag">{tag}</span>
                    ))}
                    {song.mood_tags && song.mood_tags.map((tag, idx) => (
                        <span key={`mood-${idx}`} className="tag mood-tag">{tag}</span>
                    ))}
                    {song.scene_tags && song.scene_tags.map((tag, idx) => (
                        <span key={`scene-${idx}`} className="tag scene-tag">{tag}</span>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default SongCard;
