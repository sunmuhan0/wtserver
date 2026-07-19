import { request } from './index'

export interface GlobalStatCell {
  nation: string
  type: string
  count: number
  avg_br: number
  win_rate: number
  games_played: number
  player_count: number
}

export interface GlobalStats {
  nations: string[]
  types: string[]
  cells: GlobalStatCell[]
}

export function getGlobalStats() {
  return request<GlobalStats>('/globalstats')
}
