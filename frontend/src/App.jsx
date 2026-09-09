import { useCallback, useEffect, useMemo, useState } from 'react'
import './App.css'

const API_BASE = import.meta.env.VITE_API_URL || 'https://media-scheduler-backend.onrender.com/'

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
    ...options,
  })

  const text = await response.text()
  const payload = text ? JSON.parse(text) : null

  if (!response.ok) {
    throw new Error(payload?.error || 'Request failed')
  }

  return payload
}

function App() {
  const [windows, setWindows] = useState([])
  const [mediaLibrary, setMediaLibrary] = useState([])
  const [selectedWindowId, setSelectedWindowId] = useState(null)
  const [playlist, setPlaylist] = useState([])
  const [syncEvent, setSyncEvent] = useState(null)
  const [newWindowName, setNewWindowName] = useState('')
  const [status, setStatus] = useState('Loading scheduler…')
  const [playbackIndex, setPlaybackIndex] = useState(0)

  const refreshWindows = useCallback(async () => {
    const payload = await request('/api/windows')
    const items = payload.windows || []
    setWindows(items)
    if (!selectedWindowId && items[0]) {
      setSelectedWindowId(items[0].id)
    }
    return items
  }, [selectedWindowId])

  const refreshMedia = useCallback(async () => {
    const payload = await request('/api/media')
    setMediaLibrary(payload.media || [])
  }, [])

  const refreshSync = useCallback(async () => {
    const payload = await request('/api/sync/active')
    setSyncEvent(payload.sync || null)
  }, [])

  const refreshPlaylist = useCallback(async (windowId) => {
    if (!windowId) {
      setPlaylist([])
      return
    }
    const response = await request(`/api/windows/${windowId}/playlist`)
    setPlaylist(response.playlist || [])
  }, [])

  const loadDashboard = useCallback(async () => {
    setStatus('Refreshing scheduler…')
    try {
      const windowsRes = await refreshWindows()
      await refreshMedia()
      await refreshSync()
      if (selectedWindowId) {
        await refreshPlaylist(selectedWindowId)
      } else if (windowsRes[0]) {
        setSelectedWindowId(windowsRes[0].id)
      }
      setStatus('Scheduler online')
    } catch (error) {
      setStatus(error.message)
    }
  }, [refreshMedia, refreshPlaylist, refreshSync, refreshWindows, selectedWindowId])

  useEffect(() => {
    loadDashboard()
  }, [loadDashboard])

  useEffect(() => {
    if (!selectedWindowId) {
      setPlaylist([])
      return
    }
    refreshPlaylist(selectedWindowId)
  }, [refreshPlaylist, selectedWindowId])

  useEffect(() => {
    const timer = setInterval(() => {
      refreshSync()
    }, 3000)
    return () => clearInterval(timer)
  }, [refreshSync])

  useEffect(() => {
    if (!playlist.length) {
      setPlaybackIndex(0)
      return
    }

    const currentItem = playlist[playbackIndex] || playlist[0]
    const delayMs = Math.max((currentItem?.duration || 5) * 1000, 1500)
    const timer = setTimeout(() => {
      setPlaybackIndex((previous) => (previous + 1) % playlist.length)
    }, delayMs)

    return () => clearTimeout(timer)
  }, [playbackIndex, playlist])

  const selectedWindow = useMemo(
    () => windows.find((window) => window.id === selectedWindowId) || null,
    [selectedWindowId, windows],
  )

  const previewItem = playlist[playbackIndex % Math.max(playlist.length, 1)] || null

  const addWindow = async () => {
    if (!newWindowName.trim()) return
    try {
      await request('/api/windows', {
        method: 'POST',
        body: JSON.stringify({ name: newWindowName.trim() }),
      })
      setNewWindowName('')
      await loadDashboard()
    } catch (error) {
      setStatus(error.message)
    }
  }

  const addMediaToPlaylist = async (mediaId) => {
    if (!selectedWindowId) return
    try {
      await request(`/api/windows/${selectedWindowId}/playlist`, {
        method: 'POST',
        body: JSON.stringify({ media_id: mediaId, duration: 5, position: playlist.length + 1 }),
      })
      await refreshPlaylist(selectedWindowId)
      setStatus('Playlist updated')
    } catch (error) {
      setStatus(error.message)
    }
  }

  const removePlaylistItem = async (itemId) => {
    if (!selectedWindowId) return
    try {
      await request(`/api/windows/${selectedWindowId}/playlist/${itemId}`, {
        method: 'DELETE',
      })
      await refreshPlaylist(selectedWindowId)
      setStatus('Item removed')
    } catch (error) {
      setStatus(error.message)
    }
  }

  const triggerSync = async (mediaId) => {
    try {
      const result = await request('/api/sync', {
        method: 'POST',
        body: JSON.stringify({ media_id: mediaId, duration: 20 }),
      })
      setSyncEvent(result.sync || null)
      setStatus('Global sync started')
    } catch (error) {
      setStatus(error.message)
    }
  }

  const stopSync = async () => {
    try {
      await request('/api/sync/stop', { method: 'POST' })
      await refreshSync()
      setStatus('Sync stopped')
    } catch (error) {
      setStatus(error.message)
    }
  }

  const renderMediaPreview = (item) => {
    if (!item) return <div className="media-placeholder">No media queued</div>

    if (item.media?.type === 'blank' || item.media_type === 'blank' || item.type === 'blank') {
      return <div className="media-placeholder blank">Blank screen</div>
    }

    if (item.media?.type === 'video' || item.media_type === 'video' || item.type === 'video') {
      const url = item.media?.url || item.media_url || item.url
      return <video src={url} autoPlay loop muted playsInline className="live-media" />
    }

    const url = item.media?.url || item.media_url || item.url
    return <img src={url} alt={item.name || 'Media preview'} className="live-media" />
  }

  return (
    <div className="app-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">Multi-window media scheduler</p>
          <h1>Sync Playback Console</h1>
        </div>
        <div className="status-pill">{status}</div>
      </header>

      <section className="stats-grid">
        <article className="stat-card">
          <span>Windows</span>
          <strong>{windows.length}</strong>
        </article>
        <article className="stat-card">
          <span>Library assets</span>
          <strong>{mediaLibrary.length}</strong>
        </article>
        <article className="stat-card">
          <span>Queued items</span>
          <strong>{playlist.length}</strong>
        </article>
        <article className="stat-card">
          <span>Global sync</span>
          <strong>{syncEvent ? 'Active' : 'Idle'}</strong>
        </article>
      </section>

      <main className="layout-grid">
        <aside className="panel windows-panel">
          <div className="panel-header">
            <h2>Display windows</h2>
          </div>

          <div className="window-input">
            <input
              value={newWindowName}
              onChange={(event) => setNewWindowName(event.target.value)}
              placeholder="Add new window"
            />
            <button onClick={addWindow}>Create</button>
          </div>

          <div className="window-list">
            {windows.map((window) => (
              <button
                key={window.id}
                type="button"
                className={`window-item ${selectedWindowId === window.id ? 'selected' : ''}`}
                onClick={() => setSelectedWindowId(window.id)}
              >
                <span>{window.name}</span>
                <small>{playlist.length} items</small>
              </button>
            ))}
          </div>
        </aside>

        <section className="panel main-panel">
          <div className="panel-header">
            <h2>{selectedWindow ? selectedWindow.name : 'No window selected'}</h2>
          </div>

          <div className="preview-stage">
            {renderMediaPreview(previewItem)}
            {previewItem && (
              <div className="preview-meta">
                <span>{previewItem.media?.name || previewItem.name || 'Item'}</span>
                <strong>{previewItem.duration}s</strong>
              </div>
            )}
          </div>

          <div className="playlist-section">
            <div className="section-title-row">
              <h3>Playlist sequence</h3>
              <span>{playlist.length} queued</span>
            </div>

            <div className="playlist-list">
              {playlist.length === 0 ? (
                <div className="empty-state">No media assigned yet.</div>
              ) : (
                playlist.map((item, index) => (
                  <div
                    key={item.id}
                    className={`playlist-row ${index === playbackIndex ? 'playing' : ''}`}
                  >
                    <div className="playlist-index">{index + 1}</div>
                    <div className="playlist-info">
                      <strong>{item.media?.name || item.name || 'Media'}</strong>
                      <small>{item.media?.type || item.media_type || item.type}</small>
                    </div>
                    <div className="playlist-duration">{item.duration}s</div>
                    <button type="button" className="text-button" onClick={() => removePlaylistItem(item.id)}>
                      Remove
                    </button>
                  </div>
                ))
              )}
            </div>
          </div>
        </section>

        <aside className="panel library-panel">
          <div className="panel-header">
            <h2>Media library</h2>
          </div>

          <div className="library-list">
            {mediaLibrary.map((media) => (
              <div key={media.id} className="library-card">
                <div className="thumb-wrap">
                  {media.type === 'video' ? (
                    <video src={media.url} muted playsInline className="thumb" />
                  ) : media.type === 'blank' ? (
                    <div className="thumb blank-thumb">Blank</div>
                  ) : (
                    <img src={media.url} alt={media.name} className="thumb" />
                  )}
                </div>
                <div className="library-meta">
                  <strong>{media.name}</strong>
                  <span>{media.type}</span>
                </div>
                <div className="library-actions">
                  <button type="button" onClick={() => addMediaToPlaylist(media.id)}>Add to playlist</button>
                  <button type="button" className="secondary" onClick={() => triggerSync(media.id)}>Sync all</button>
                </div>
              </div>
            ))}
          </div>
        </aside>
      </main>

      <section className="panel sync-panel">
        <div className="panel-header">
          <h2>Global sync playback</h2>
          <button type="button" className="secondary" onClick={stopSync}>Stop sync</button>
        </div>

        <div className="sync-content">
          {syncEvent ? (
            <>
              <div className="sync-preview">
                {renderMediaPreview({
                  media: { type: syncEvent.media?.type || syncEvent.media_type, url: syncEvent.media?.url || syncEvent.media_url },
                  name: syncEvent.media?.name || 'Global sync',
                  duration: syncEvent.duration,
                })}
              </div>
              <div className="sync-meta">
                <h3>{syncEvent.media?.name || 'Global sync asset'}</h3>
                <p>Status: {syncEvent.status}</p>
                <p>Duration: {syncEvent.duration}s</p>
                <p>Ends: {new Date(syncEvent.end_time).toLocaleTimeString()}</p>
              </div>
            </>
          ) : (
            <div className="empty-state">No active global sync event.</div>
          )}
        </div>
      </section>
    </div>
  )
}

export default App
