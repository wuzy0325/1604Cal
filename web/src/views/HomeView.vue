<template>
  <div class="home-view">
    <header class="page-header">
      <h1>欢迎使用1604校准系统</h1>
      <p class="subtitle">
        请选择要进入的模块
      </p>
    </header>
    
    <div class="module-cards">
      <router-link
        to="/measurement"
        class="module-card"
      >
        <div class="card-icon">
          <el-icon><Tools /></el-icon>
        </div>
        <h3>计量工作台</h3>
        <p>单设备校准和数据采集</p>
        <ul class="features">
          <li>自动/手动控制模式</li>
          <li>16通道实时数据</li>
          <li>压力表自动生成</li>
        </ul>
      </router-link>
      
      <router-link
        to="/multi-pressure"
        class="module-card"
      >
        <div class="card-icon">
          <el-icon><CircleCheck /></el-icon>
        </div>
        <h3>多设备打压</h3>
        <p>并行控制多个打压设备</p>
        <ul class="features">
          <li>多设备同时监控</li>
          <li>独立压力设定</li>
          <li>实时状态显示</li>
        </ul>
      </router-link>
      
      <router-link
        to="/calibration"
        class="module-card"
      >
        <div class="card-icon">
          <el-icon><SetUp /></el-icon>
        </div>
        <h3>标定工作台</h3>
        <p>完整的1604标定流程</p>
        <ul class="features">
          <li>6步向导式操作</li>
          <li>通道选择矩阵</li>
          <li>数据拟合与导出</li>
        </ul>
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Tools, CircleCheck, SetUp } from '@element-plus/icons-vue'
</script>

<style scoped lang="scss">
.home-view {
  padding: var(--spacing-xl);
  
  .page-header {
    margin-bottom: var(--spacing-xl);
    
    h1 {
      color: var(--text-primary);
      font-size: 28px;
      margin: 0 0 var(--spacing-sm) 0;
    }
    
    .subtitle {
      color: var(--text-secondary);
      font-size: 16px;
      margin: 0;
    }
  }
  
  .module-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: var(--spacing-lg);
    
    .module-card {
      background: var(--bg-secondary);
      border: 1px solid var(--border-color);
      border-radius: var(--radius-lg);
      padding: var(--spacing-xl);
      text-decoration: none;
      transition: all 0.3s;
      
      &:hover {
        border-color: var(--accent-primary);
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(233, 69, 96, 0.2);
      }
      
      .card-icon {
        width: 60px;
        height: 60px;
        background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
        border-radius: var(--radius-md);
        display: flex;
        align-items: center;
        justify-content: center;
        margin-bottom: var(--spacing-md);
        
        .el-icon {
          font-size: 28px;
          color: white;
        }
      }
      
      h3 {
        color: var(--text-primary);
        font-size: 20px;
        margin: 0 0 var(--spacing-sm) 0;
      }
      
      p {
        color: var(--text-secondary);
        margin: 0 0 var(--spacing-md) 0;
      }
      
      .features {
        list-style: none;
        padding: 0;
        margin: 0;
        
        li {
          color: var(--text-muted);
          font-size: 14px;
          padding: var(--spacing-xs) 0;
          padding-left: 16px;
          position: relative;
          
          &::before {
            content: '•';
            position: absolute;
            left: 0;
            color: var(--accent-primary);
          }
        }
      }
    }
  }
}
</style>