import { useQuery } from '@tanstack/react-query'
import { getAssets } from '../api/assetsApi.js'

export function useAssets({ enabled = false } = {}) {
  return useQuery({
    queryKey: ['assets'],
    queryFn: getAssets,
    enabled,
  })
}
