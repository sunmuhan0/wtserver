const BASE_URL = 'http://localhost:8080/api/v1'

export async function request<T>(path: string): Promise<T> {
  const res = await uni.request({
    url: `${BASE_URL}${path}`,
    method: 'GET',
  })
  return res.data as T
}
