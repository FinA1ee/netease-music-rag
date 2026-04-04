# API Documentation

## `GET /api/search`
Search songs using natural language via LLM embedding and pgvector similarity.

**Query Parameters:**
- `q` (string, required): Natural language string (e.g. "Upbeat road trip song").
- `l` (int, optional): Limit results count. Default is 5.

**Response:**
```json
{
  "query": "the input query",
  "songs": [
    {
      "id": 123456,
      "name": "Song Name",
      "artist": "Artist Name",
      "album": "Album Name",
      "cover_url": "http://...",
      "lyrics": "...",
      "description": "AI generated text",
      "style_tags": ["Pop"],
      "mood_tags": ["Happy"],
      "scene_tags": ["Driving"],
      "created_at": "2026-04-03T10:00:00Z",
      "updated_at": "2026-04-03T10:00:00Z"
    }
  ]
}
```

---

## `POST /api/trigger-job`
Manually trigger the background cron job to fetch Netease daily recommendations, query the LLM to get tags + descriptions, vectorize, and insert into the database.

**Response:**
```json
{
  "message": "Job started in background"
}
```
