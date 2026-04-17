import { chromium } from 'playwright';
import fs from 'fs';

const BASE = 'http://localhost:5173';
const API  = 'http://localhost:8080/api/v1';

async function main() {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1400, height: 900 } });

  // 收集控制台日志
  const logs = [];
  page.on('console', msg => logs.push(`[${msg.type()}] ${msg.text()}`));
  page.on('response', res => {
    if (res.url().includes('/api/v1/')) {
      console.log(`  API: ${res.request().method()} ${res.url().replace(API, '')} -> ${res.status()}`);
    }
  });

  console.log('=== 1. 打开首页 ===');
  await page.goto(BASE, { waitUntil: 'networkidle', timeout: 10000 });
  await page.screenshot({ path: 'e2e-01-home.png', fullPage: true });
  console.log('  截图: e2e-01-home.png');

  // 查看当前页面URL和标题
  console.log(`  URL: ${page.url()}`);
  console.log(`  标题: ${await page.title()}`);

  // 查看页面上有什么文字
  const bodyText = await page.locator('body').innerText();
  console.log(`  页面内容片段: ${bodyText.substring(0, 200)}`);

  console.log('\n=== 2. 查看导航菜单 ===');
  const navItems = await page.locator('nav a, .sidebar a, .el-menu a, [class*="sidebar"] a').allTextContents();
  console.log(`  导航项: ${JSON.stringify(navItems)}`);

  // 查找所有可点击的菜单/链接
  const allLinks = await page.locator('a, [role="menuitem"], .nav-item, .menu-item').allTextContents();
  console.log(`  所有链接: ${JSON.stringify(allLinks.slice(0, 20))}`);

  console.log('\n=== 3. 尝试导航到设备管理/测量/校准页面 ===');
  const routes = ['/measurement', '/calibration', '/device-mgmt', '/devices'];
  for (const route of routes) {
    const url = `${BASE}${route}`;
    const resp = await page.goto(url, { waitUntil: 'networkidle', timeout: 8000 }).catch(() => null);
    if (resp && resp.ok()) {
      console.log(`  ${route}: OK (${resp.status()})`);
      await page.screenshot({ path: `e2e-${route.replace(/\//g, '')}.png`, fullPage: true });
    } else {
      console.log(`  ${route}: 失败或不存在`);
    }
  }

  console.log('\n=== 4. 查找设备相关页面 ===');
  // 回到首页重新开始
  await page.goto(BASE, { waitUntil: 'networkidle', timeout: 8000 });

  // 尝试查找并点击侧边栏中的设备相关菜单
  const sidebarTexts = await page.locator('[class*="sidebar"], [class*="side"], nav, .el-menu').first().innerText().catch(() => '');
  console.log(`  侧边栏文本: ${sidebarTexts.substring(0, 300)}`);

  // 截图当前页面状态
  await page.screenshot({ path: 'e2e-current.png', fullPage: true });

  console.log('\n=== 5. 模拟API调用流程 ===');
  // 直接在浏览器中调用API，模拟前端操作
  const result = await page.evaluate(async (apiBase) => {
    const results = {};

    // 获取设备列表
    const devResp = await fetch(`${apiBase}/devices`);
    const devData = await devResp.json();
    results.devices = devData.data?.map(d => ({ id: d.id, name: d.name, type: d.type, model: d.model, status: d.status }));

    // 查找已连接的811A设备
    const pressureDev = devData.data?.find(d => d.model?.includes('811A') && d.status === 'connected');
    if (pressureDev) {
      results.pressureDevice = pressureDev.id;

      // 设置校准设备
      const measureDev = devData.data?.find(d => d.type === 'measure');
      if (measureDev) {
        results.measureDevice = measureDev.id;
        const setResp = await fetch(`${apiBase}/calibration/devices`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ measureDeviceId: measureDev.id, pressureDeviceId: pressureDev.id })
        });
        const setData = await setResp.json();
        results.setDevices = setData;

        // 读取压力
        const pressResp = await fetch(`${apiBase}/calibration/pressure`);
        const pressData = await pressResp.json();
        results.pressure = pressData;

        // 读取稳定状态
        const stabResp = await fetch(`${apiBase}/calibration/stability`);
        const stabData = await stabResp.json();
        results.stability = stabData;
      }
    }

    return results;
  }, API);

  console.log('  API调用结果:', JSON.stringify(result, null, 2));

  // 输出收集的控制台日志
  if (logs.length > 0) {
    console.log('\n=== 浏览器控制台日志 ===');
    logs.forEach(l => console.log(`  ${l}`));
  }

  console.log('\n=== 测试完成 ===');
  await browser.close();
}

main().catch(err => {
  console.error('测试失败:', err.message);
  process.exit(1);
});
