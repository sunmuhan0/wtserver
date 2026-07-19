<template>
  <view class="container">
    <view class="controls">
      <view class="stat-selector">
        <text
          v-for="s in statOptions"
          :key="s.key"
          class="stat-tab"
          :class="{ active: selectedStat === s.key }"
          @tap="selectedStat = s.key"
        >{{ s.label }}</text>
      </view>
      <view class="legend">
        <text class="legend-label">低</text>
        <view class="legend-gradient" :style="{ background: legendGradient }" />
        <text class="legend-label">高</text>
      </view>
    </view>

    <scroll-view class="table-wrapper" scroll-x>
      <view class="heatmap-table">
        <view class="table-header">
          <view class="corner-cell">国家 \ 类型</view>
          <view class="type-cell" v-for="t in data?.types" :key="t">{{ typeLabel(t) }}</view>
        </view>
        <view class="table-row" v-for="n in data?.nations" :key="n">
          <view class="nation-cell">
            <text>{{ nationLabel(n) }}</text>
          </view>
          <view
            class="data-cell"
            v-for="t in data?.types"
            :key="t"
            :style="{ backgroundColor: cellColor(getCell(n, t)) }"
            @tap="showCellDetail(n, t)"
          >
            <text class="cell-value">{{ formatCellValue(getCell(n, t)) }}</text>
          </view>
        </view>
      </view>
    </scroll-view>

    <view v-if="selectedCell" class="detail-panel">
      <text class="detail-title">{{ nationLabel(selectedCell.nation) }} - {{ typeLabel(selectedCell.type) }}</text>
      <view class="detail-row"><text>载具数量: {{ selectedCell.count }}</text></view>
      <view class="detail-row"><text>平均BR: {{ selectedCell.avg_br || '-' }}</text></view>
      <view class="detail-row"><text>胜率: {{ selectedCell.win_rate ? selectedCell.win_rate + '%' : '-' }}</text></view>
      <view class="detail-row"><text>总场次: {{ selectedCell.games_played || '-' }}</text></view>
      <view class="detail-row"><text>玩家数: {{ selectedCell.player_count || '-' }}</text></view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getGlobalStats, type GlobalStats, type GlobalStatCell } from '@/api/globalstats'

const data = ref<GlobalStats | null>(null)
const selectedStat = ref<'win_rate' | 'count' | 'games_played'>('win_rate')
const selectedCell = ref<GlobalStatCell | null>(null)

const statOptions = [
  { key: 'win_rate' as const, label: '胜率' },
  { key: 'count' as const, label: '载具数' },
  { key: 'games_played' as const, label: '总场次' },
]

const legendGradient = computed(() => {
  if (selectedStat.value === 'win_rate') {
    return 'linear-gradient(to right, #ff4444, #ffaa00, #44ff44)'
  }
  return 'linear-gradient(to right, #1a1a2e, #e74c3c)'
})

function getCell(nation: string, type: string): GlobalStatCell | null {
  return data.value?.cells.find(c => c.nation === nation && c.type === type) || null
}

function cellColor(cell: GlobalStatCell | null): string {
  if (!cell) return 'rgba(255,255,255,0.05)'
  let val = 0
  let min = 0
  let max = 1

  if (selectedStat.value === 'win_rate') {
    val = cell.win_rate || 0
    min = 30
    max = 70
    if (val === 0) return 'rgba(255,255,255,0.05)'
  } else if (selectedStat.value === 'count') {
    val = cell.count || 0
    max = Math.max(...(data.value?.cells.map(c => c.count) || [1]))
  } else {
    val = cell.games_played || 0
    max = Math.max(...(data.value?.cells.map(c => c.games_played) || [1]))
  }

  const ratio = Math.min(Math.max((val - min) / (max - min), 0), 1)

  if (selectedStat.value === 'win_rate') {
    const r = Math.round(255 * (1 - ratio))
    const g = Math.round(255 * ratio)
    return `rgba(${r}, ${g}, 50, 0.7)`
  }
  const intensity = Math.round(ratio * 200 + 30)
  return `rgba(231, 76, 60, ${0.2 + ratio * 0.6})`
}

function formatCellValue(cell: GlobalStatCell | null): string {
  if (!cell) return '-'
  if (selectedStat.value === 'win_rate') {
    return cell.win_rate ? cell.win_rate.toFixed(1) + '%' : '-'
  }
  if (selectedStat.value === 'count') return String(cell.count)
  if (selectedStat.value === 'games_played') {
    return cell.games_played ? formatNumber(cell.games_played) : '-'
  }
  return '-'
}

function formatNumber(n: number): string {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
  return String(n)
}

function showCellDetail(nation: string, type: string) {
  selectedCell.value = getCell(nation, type)
}

function nationLabel(n: string): string {
  const labels: Record<string, string> = {
    usa: '美国', germany: '德国', ussr: '苏联', britain: '英国',
    japan: '日本', china: '中国', italy: '意大利', france: '法国',
    sweden: '瑞典', israel: '以色列',
  }
  return labels[n] || n
}

function typeLabel(t: string): string {
  const labels: Record<string, string> = {
    aircraft: '飞机', tanks: '坦克', helicopters: '直升机',
    ships: '舰船', coastal: '近海舰艇',
  }
  return labels[t] || t
}

onMounted(async () => {
  try {
    data.value = await getGlobalStats()
  } catch {
    uni.showToast({ title: '获取数据失败', icon: 'none' })
  }
})
</script>

<style lang="scss">
.container {
  padding: 20rpx;
  min-height: 100vh;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  color: #fff;
}

.controls {
  display: flex;
  flex-direction: column;
  gap: 20rpx;
  margin-bottom: 30rpx;
}

.stat-selector {
  display: flex;
  gap: 10rpx;
}

.stat-tab {
  padding: 12rpx 24rpx;
  border-radius: 12rpx;
  font-size: 26rpx;
  background: rgba(255, 255, 255, 0.08);
  color: #aaa;
  flex: 1;
  text-align: center;

  &.active {
    background: #e74c3c;
    color: #fff;
  }
}

.legend {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.legend-label {
  font-size: 22rpx;
  color: #999;
}

.legend-gradient {
  flex: 1;
  height: 20rpx;
  border-radius: 10rpx;
}

.table-wrapper {
  width: 100%;
  overflow-x: auto;
}

.heatmap-table {
  display: table;
  border-collapse: collapse;
  min-width: 100%;
}

.table-header {
  display: table-row;
}

.corner-cell {
  display: table-cell;
  padding: 16rpx 12rpx;
  font-size: 22rpx;
  color: #999;
  min-width: 140rpx;
  white-space: nowrap;
}

.type-cell {
  display: table-cell;
  padding: 16rpx 12rpx;
  font-size: 24rpx;
  font-weight: bold;
  text-align: center;
  min-width: 120rpx;
  color: #e74c3c;
}

.table-row {
  display: table-row;
}

.nation-cell {
  display: table-cell;
  padding: 16rpx 12rpx;
  font-size: 24rpx;
  font-weight: bold;
  min-width: 140rpx;
  white-space: nowrap;
}

.data-cell {
  display: table-cell;
  padding: 20rpx 12rpx;
  text-align: center;
  border: 1rpx solid rgba(255, 255, 255, 0.06);
  border-radius: 4rpx;
}

.cell-value {
  font-size: 24rpx;
  font-weight: 500;
}

.detail-panel {
  margin-top: 30rpx;
  padding: 24rpx;
  background: rgba(255, 255, 255, 0.08);
  border-radius: 16rpx;
}

.detail-title {
  font-size: 30rpx;
  font-weight: bold;
  margin-bottom: 16rpx;
  display: block;
}

.detail-row {
  font-size: 26rpx;
  padding: 8rpx 0;
  color: #ccc;
}
</style>
