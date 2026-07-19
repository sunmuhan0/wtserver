<template>
  <view class="splash">
    <view class="brand">
      <image class="logo" src="/static/logo.png" mode="aspectFit" />
      <text class="title">战争雷霆查询</text>
    </view>

    <view class="action">
      <text class="countdown">{{ countdown }}s</text>
      <button class="skip-btn" @tap="goHome">跳过</button>
    </view>

    <view class="ad-wrapper" v-if="showAd">
      <ad :unit-id="adUnitId" ad-intervals="30"></ad>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { AD_UNITS } from '@/api/ad'

const countdown = ref(5)
const showAd = ref(false)
const adUnitId = AD_UNITS.splash

let timer: ReturnType<typeof setInterval> | null = null

function goHome() {
  if (timer) clearInterval(timer)
  uni.reLaunch({ url: '/pages/index/index' })
}

onMounted(() => {
  showAd.value = true

  timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      goHome()
    }
  }, 1000)
})
</script>

<style>
.splash {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  color: #fff;
  position: relative;
}
.brand {
  text-align: center;
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}
.logo {
  width: 200rpx;
  height: 200rpx;
  margin-bottom: 30rpx;
}
.title {
  font-size: 48rpx;
  font-weight: bold;
}
.action {
  display: flex;
  align-items: center;
  gap: 20rpx;
  margin-bottom: 40rpx;
}
.countdown {
  font-size: 28rpx;
  color: #aaa;
}
.skip-btn {
  width: 140rpx;
  height: 60rpx;
  line-height: 60rpx;
  background: rgba(255,255,255,0.15);
  color: #fff;
  border-radius: 30rpx;
  font-size: 26rpx;
  text-align: center;
  border: none;
}
.ad-wrapper {
  width: 100%;
  padding: 20rpx 0;
  display: flex;
  justify-content: center;
}
</style>
