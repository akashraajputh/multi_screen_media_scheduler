import test from 'node:test'
import assert from 'node:assert/strict'
import { getWindowItemCounts } from './windowState.js'

test('window item counts reflect each window independently', () => {
  const windows = [
    { id: 1, name: 'Windows 1' },
    { id: 2, name: 'Windows 2' },
  ]

  const playlistsByWindowId = {
    1: [{ id: 1 }, { id: 2 }, { id: 3 }, { id: 4 }],
    2: [],
  }

  assert.deepStrictEqual(getWindowItemCounts(windows, playlistsByWindowId), {
    1: 4,
    2: 0,
  })
})
