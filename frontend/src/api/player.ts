import { request } from './index'

export interface Player {
  nick: string
  rank: string
  arcade: ModeStats
  realistic: ModeStats
  simulator: ModeStats
}

export interface ModeStats {
  battles: number
  wins: number
  win_rate: number
  kills: number
  deaths: number
  kd: number
  kills_per_battle: number
}

export function getPlayer(nickname: string) {
  return request<Player>(`/player-ts/${nickname}`)
}
