<template>
  <view class="container">
    <view class="header">
      <image class="logo" src="/static/logo.png" mode="aspectFit" />
      <text class="title">战争雷霆查询</text>
    </view>

    <view class="search-box">
      <input class="search-input" v-model="query" placeholder="输入玩家昵称 / 载具名称" />
      <button class="search-btn" @tap="handleSearch">查询</button>
    </view>

    <view class="quick-links">
      <view class="link-card" @tap="navigateTo('player')">
        <text class="link-icon">👤</text>
        <text class="link-label">玩家查询</text>
      </view>
      <view class="link-card" @tap="navigateTo('vehicle')">
        <text class="link-icon">🚁</text>
        <text class="link-label">载具百科</text>
      </view>
      <view class="link-card" @tap="loadNews">
        <text class="link-icon">📰</text>
        <text class="link-label">最新资讯</text>
      </view>
    </view>

    <view v-if="news.length" class="news-section">
      <text class="section-title">游戏资讯</text>
      <view class="news-item" v-for="(item, i) in news" :key="i">
        <text class="news-text">{{ item }}</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { getNews } from '@/api/vehicle'

const query = ref('')
const news = ref<string[]>([])

function handleSearch() {
  if (!query.value) return
  uni.navigateTo({ url: `/pages/player/player?nickname=${query.value}` })
}

function navigateTo(page: string) {
  uni.navigateTo({ url: `/pages/${page}/${page}` })
}

async function loadNews() {
  try {
    const res = await getNews()
    news.value = res.news
  } catch {
    uni.showToast({ title: '获取资讯失败', icon: 'none' })
  }
}
</script>

<style lang="scss">
.container {
  padding: 30rpx;
  min-height: 100vh;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  color: #fff;
}
.header {
  text-align: center;
  padding: 60rpx 0;
}
.logo {
  width: 160rpx;
  height: 160rpx;
}
.title {
  display: block;
  margin-top: 20rpx;
  font-size: 40rpx;
  font-weight: bold;
}
.search-box {
  display: flex;
  gap: 20rpx;
  margin: 20rpx 0;
}
.search-input {
  flex: 1;
  height: 80rpx;
  background: rgba(255,255,255,0.1);
  border-radius: 16rpx;
  padding: 0 30rpx;
  color: #fff;
  font-size: 28rpx;
}
.search-btn {
  width: 160rpx;
  height: 80rpx;
  line-height: 80rpx;
  background: #e74c3c;
  color: #fff;
  border-radius: 16rpx;
  font-size: 28rpx;
  text-align: center;
}
.quick-links {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20rpx;
  margin: 40rpx 0;
}
.link-card {
  background: rgba(255,255,255,0.08);
  border-radius: 20rpx;
  padding: 40rpx;
  text-align: center;
}
.link-icon {
  font-size: 60rpx;
}
.link-label {
  display: block;
  margin-top: 16rpx;
  font-size: 28rpx;
}
.news-section {
  margin-top: 40rpx;
}
.section-title {
  font-size: 32rpx;
  font-weight: bold;
  margin-bottom: 20rpx;
  display: block;
}
.news-item {
  background: rgba(255,255,255,0.06);
  border-radius: 12rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}
.news-text {
  font-size: 26rpx;
  color: #ccc;
}
</style>
