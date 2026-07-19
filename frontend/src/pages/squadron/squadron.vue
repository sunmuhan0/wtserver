<template>
  <view class="container">
    <view class="search-bar">
      <input class="input" v-model="name" placeholder="输入联队名称" />
      <button class="btn" @tap="search">查询</button>
    </view>

    <view v-if="squadron" class="detail">
      <text class="sname">{{ squadron.name }}</text>
      <text class="tag">{{ squadron.tag }}</text>
      <view class="info-grid">
        <view class="info-item">
          <text class="label">成员数</text>
          <text class="value">{{ squadron.members }}</text>
        </view>
        <view class="info-item">
          <text class="label">队长</text>
          <text class="value">{{ squadron.leader }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { getSquadron, type Squadron } from '@/api/vehicle'

const name = ref('')
const squadron = ref<Squadron | null>(null)

async function search() {
  if (!name.value) return
  try {
    squadron.value = await getSquadron(name.value)
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
.detail {
  text-align: center;
}
.sname {
  font-size: 40rpx;
  font-weight: bold;
  display: block;
}
.tag {
  display: block;
  color: #f1c40f;
  margin: 10rpx 0 40rpx;
}
.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20rpx;
}
.info-item {
  background: rgba(255,255,255,0.08);
  border-radius: 16rpx;
  padding: 30rpx;
}
.label {
  display: block;
  font-size: 24rpx;
  color: #999;
}
.value {
  display: block;
  font-size: 32rpx;
  margin-top: 8rpx;
}
</style>
