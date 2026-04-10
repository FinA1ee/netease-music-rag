import { useEffect, useRef } from 'react';

export type EventType =
  | 'connected'
  | 'job_started'
  | 'playlist_processing'
  | 'song_analysed'
  | 'song_skipped'
  | 'job_completed'
  | 'embedding_started'
  | 'embedding_done';

export interface JobEvent {
  type: EventType;
  payload?: Record<string, any>;
}

/**
 * Opens a persistent SSE connection to /api/jobs/stream.
 * Calls onEvent each time an event arrives.
 * Automatically closes on unmount.
 */
export function useJobEvents(onEvent: (event: JobEvent) => void): void {
  const onEventRef = useRef(onEvent);
  onEventRef.current = onEvent; // always call the latest version

  useEffect(() => {
    const es = new EventSource('/api/jobs/stream');

    es.onmessage = (e: MessageEvent) => {
      try {
        const event: JobEvent = JSON.parse(e.data);
        onEventRef.current(event);
      } catch {
        // ignore malformed events
      }
    };

    es.onerror = () => {
      // EventSource auto-reconnects — no action needed
    };

    return () => es.close();
  }, []); // mount once; onEvent changes handled via ref
}
