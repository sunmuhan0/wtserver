<template>
  <view class="container">
    <view class="search-bar">
      <input class="input" v-model="nickname" placeholder="输入玩家昵称" />
      <button class="btn" @tap="search">查询</button>
    </view>

    <view v-if="player" class="profile">
      <text class="nickname">{{ player.nick }}</text>
      <text class="rank">{{ player.rank }}</text>

      <view class="mode-tabs">
        <text v-for="mode in modes" :key="mode.key" class="mode-tab" :class="{ active: activeMode === mode.key }" @tap="activeMode = mode.key">{{ mode.label }}</text>
      </view>

      <view class="stats-grid">
        <view class="stat-item">
          <text class="stat-value">{{ current.battles }}</text>
          <text class="stat-label">场次</text>
        </view>
        <view class="stat-item">
          <text class="stat-value">{{ current.win_rate }}%</text>
          <text class="stat-label">胜率</text>
        </view>
        <view class="stat-item">
          <text class="stat-value">{{ current.kd }}</text>
          <text class="stat-label">KD</text>
        </view>
        <view class="stat-item">
          <text class="stat-value">{{ current.kills }}</text>
          <text class="stat-label">击杀</text>
        </view>
        <view class="stat-item">
          <text class="stat-value">{{ current.deaths }}</text>
          <text class="stat-label">死亡</text>
        </view>
        <view class="stat-item">
          <text class="stat-value">{{ current.kills_per_battle }}</text>
          <text class="stat-label">场均击杀</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { getPlayer, type Player, type ModeStats } from '@/api/player'

const nickname = ref('')
const player = ref<Player | null>(null)
const activeMode = ref('arcade')

const modes = [
  { key: 'arcade', label: '街机' },
  { key: 'realistic', label: '历史' },
  { key: 'simulator', label: '全真' },
]

const current = computed<ModeStats>(() => {
  if (!player.value) return {} as ModeStats
  return player.value[activeMode.value as keyof Player] as ModeStats || {} as ModeStats
})

async function search() {
  if (!nickname.value) return
  try {
    player.value = await getPlayer(nickname.value)
    activeMode.value = 'arcade'
  } catch {
    uni.showToast({ title: '查询失败', icon: 'none' })
  }
}
</script>

<style lang="scss">
.container {
  padding: 30rpx;
  min-height: 100vh;
  background: #1a1a2e;
  color: #fff;
}
.search-bar {
  display: flex;
  gap: 20rpx;
  margin-bottom: 40rpx;
}
.input {
  flex: 1;
  height: 80rpx;
  background: rgba(255,255,255,0.1);
  border-radius: 16rpx;
  padding: 0 30rpx;
  color: #fff;
}
.btn {
  width: 140rpx;
  height: 80rpx;
  line-height: 80rpx;
  background: #e74c3c;
  color: #fff;
  border-radius: 16rpx;
  text-align: center;
}
.profile {
  text-align: center;
}
.nickname {
  font-size: 40rpx;
  font-weight: bold;
  display: block;
}
.rank {
  display: block;
  color: #f1c40f;
  margin: 10rpx 0 40rpx;
}
.mode-tabs {
  display: flex;
  gap: 10rpx;
  margin-bottom: 30rpx;
  justify-content: center;
}
.mode-tab {
  padding: 10rpx 30rpx;
  background: rgba(255,255,255,0.08);
  border-radius: 30rpx;
  font-size: 26rpx;
}
.mode-tab.active {
  background: #e74c3c;
  color: #fff;
}
.stats-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20rpx;
}
.stat-item {
  background: rgba(255,255,255,0.08);
  border-radius: 16rpx;
  padding: 30rpx;
}
.stat-value {
  display: block;
  font-size: 40rpx;
  font-weight: bold;
  color: #e74c3c;
}
.stat-label {
  display: block;
  font-size: 24rpx;
  color: #999;
  margin-top: 8rpx;
}
</style>
