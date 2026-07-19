import { request } from './index'

export interface Vehicle {
  name: string
  country: string
  type: string
  rank: number
  br: string
  is_premium: boolean
}

export interface Squadron {
  name: string
  tag: string
  members: number
  leader: string
}

export function getVehicle(name: string) {
  return request<Vehicle>(`/vehicle/${name}`)
}

export function getSquadron(name: string) {
  return request<Squadron>(`/squadron/${name}`)
}

export function getNews() {
  return request<{ news: string[] }>('/news')
}
