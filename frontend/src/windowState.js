export function getWindowItemCounts(windows, playlistsByWindowId = {}) {
  return windows.reduce((counts, window) => {
    const playlist = Array.isArray(playlistsByWindowId[window.id]) ? playlistsByWindowId[window.id] : []
    counts[window.id] = playlist.length
    return counts
  }, {})
}
