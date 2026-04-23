import { beforeEach, describe, expect, it, vi } from 'vitest'

import { resolveAlarm } from '../calibration'
import * as clientApi from '../client'

vi.mock('../client', () => ({
  requestJSON: vi.fn()
}))

describe('calibration api resolveAlarm', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(clientApi.requestJSON).mockResolvedValue({ data: { status: 'ok' } })
  })

  it('posts recollect decision to resolve-alarm endpoint', async () => {
    await resolveAlarm('recollect')

    expect(clientApi.requestJSON).toHaveBeenCalledWith('/calibration/resolve-alarm', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ decision: 'recollect' })
    })
  })

  it('posts skip decision to resolve-alarm endpoint', async () => {
    await resolveAlarm('skip')

    expect(clientApi.requestJSON).toHaveBeenCalledWith('/calibration/resolve-alarm', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ decision: 'skip' })
    })
  })
})
