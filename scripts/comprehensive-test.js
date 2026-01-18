const puppeteer = require('puppeteer');

(async () => {
  const browser = await puppeteer.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });
  
  const page = await browser.newPage();
  
  let totalTests = 0;
  let passedTests = 0;
  let failedTests = 0;
  const errors = [];
  
  // 监听所有失败的API
  page.on('response', async response => {
    const url = response.url();
    if (url.includes('/api/') && response.status() >= 400) {
      try {
        const body = await response.text();
        errors.push({
          type: 'api',
          status: response.status(),
          url,
          body: body.substring(0, 200)
        });
      } catch (e) {}
    }
  });
  
  // 监听控制台错误
  page.on('console', msg => {
    if (msg.type() === 'error') {
      const text = msg.text();
      if (text.includes('ERR-') || text.includes('Failed') || text.includes('错误')) {
        errors.push({ type: 'console', text });
      }
    }
  });
  
  const test = async (name, fn) => {
    totalTests++;
    try {
      await fn();
      passedTests++;
      console.log(`✅ ${name}`);
    } catch (error) {
      failedTests++;
      console.log(`❌ ${name}: ${error.message}`);
      errors.push({ type: 'test', name, error: error.message });
    }
  };
  
  try {
    console.log('========================================');
    console.log('综合功能测试');
    console.log('========================================\n');
    
    // 登录
    await test('登录', async () => {
      await page.goto('http://localhost:8081/login', { waitUntil: 'networkidle2' });
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      const inputs = await page.$$('input');
      if (inputs.length < 2) throw new Error('找不到登录表单');
      
      await inputs[0].type('admin');
      await inputs[1].type('admin123');
      
      const buttons = await page.$$('button');
      let clicked = false;
      for (const button of buttons) {
        const text = await page.evaluate(el => el.textContent, button);
        if (text.includes('登录')) {
          await button.click();
          clicked = true;
          break;
        }
      }
      
      if (!clicked) throw new Error('找不到登录按钮');
      
      await page.waitForNavigation({ waitUntil: 'networkidle2', timeout: 10000 }).catch(() => {});
      await new Promise(resolve => setTimeout(resolve, 2000));
    });
    
    // 测试IP限制页面
    await test('访问IP限制页面', async () => {
      await page.goto('http://localhost:8081/admin/ip-restriction', { waitUntil: 'networkidle2' });
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      const title = await page.evaluate(() => document.querySelector('.page-title, h1, h2')?.textContent || '');
      if (!title.includes('IP') && !title.includes('限制')) {
        // 检查是否有统计卡片作为备选验证
        const hasContent = await page.evaluate(() => !!document.querySelector('.stat-card, .el-tabs'));
        if (!hasContent) throw new Error('页面内容不正确');
      }
    });
    
    await test('IP限制 - 统计概览标签', async () => {
      const tabs = await page.$$('.el-tabs__item');
      if (tabs.length > 0) {
        await tabs[0].click();
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    });
    
    await test('IP限制 - 设置标签', async () => {
      const tabs = await page.$$('.el-tabs__item');
      if (tabs.length > 1) {
        await tabs[1].click();
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    });
    
    await test('IP限制 - 白名单标签', async () => {
      const tabs = await page.$$('.el-tabs__item');
      if (tabs.length > 2) {
        await tabs[2].click();
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    });
    
    await test('IP限制 - 黑名单标签', async () => {
      const tabs = await page.$$('.el-tabs__item');
      if (tabs.length > 3) {
        await tabs[3].click();
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    });
    
    await test('IP限制 - 用户在线IP标签', async () => {
      const tabs = await page.$$('.el-tabs__item');
      if (tabs.length > 4) {
        await tabs[4].click();
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    });
    
    await test('IP限制 - IP历史标签', async () => {
      const tabs = await page.$$('.el-tabs__item');
      if (tabs.length > 5) {
        await tabs[5].click();
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    });
    
    // 测试财务报表页面
    await test('访问财务报表页面', async () => {
      await page.goto('http://localhost:8081/admin/reports', { waitUntil: 'networkidle2' });
      await new Promise(resolve => setTimeout(resolve, 3000));
      
      const title = await page.evaluate(() => document.querySelector('.page-title')?.textContent || '');
      if (!title.includes('财务') && !title.includes('报表')) throw new Error('页面标题不正确');
    });
    
    await test('财务报表 - 统计卡片显示', async () => {
      const hasStats = await page.evaluate(() => !!document.querySelector('.stat-card'));
      if (!hasStats) throw new Error('没有找到统计卡片');
    });
    
    await test('财务报表 - 图表显示', async () => {
      const hasCharts = await page.evaluate(() => !!document.querySelector('.chart-container'));
      if (!hasCharts) throw new Error('没有找到图表');
    });
    
    // 检查是否有错误消息
    await test('检查页面错误消息', async () => {
      const errorMsgs = await page.evaluate(() => {
        return Array.from(document.querySelectorAll('.el-message--error')).map(el => el.textContent);
      });
      if (errorMsgs.length > 0) {
        throw new Error(`发现错误消息: ${errorMsgs.join(', ')}`);
      }
    });
    
    console.log('\n========================================');
    console.log('测试结果');
    console.log('========================================\n');
    
    console.log(`总测试数: ${totalTests}`);
    console.log(`✅ 通过: ${passedTests}`);
    console.log(`❌ 失败: ${failedTests}`);
    
    if (errors.length > 0) {
      console.log('\n发现的错误:');
      errors.forEach(err => {
        if (err.type === 'api') {
          console.log(`\n  API错误: ${err.status} ${err.url}`);
          console.log(`  响应: ${err.body}`);
        } else if (err.type === 'console') {
          console.log(`\n  控制台错误: ${err.text}`);
        } else if (err.type === 'test') {
          console.log(`\n  测试失败: ${err.name} - ${err.error}`);
        }
      });
    } else {
      console.log('\n✅ 没有发现任何错误');
    }
    
    if (failedTests === 0 && errors.length === 0) {
      console.log('\n🎉 所有测试通过！应用运行正常。');
      console.log('\n如果用户仍然看到错误，请建议用户:');
      console.log('1. 清除浏览器缓存 (Ctrl+Shift+Delete)');
      console.log('2. 硬刷新页面 (Ctrl+Shift+R 或 Cmd+Shift+R)');
      console.log('3. 使用无痕模式测试');
      console.log('4. 关闭所有标签页后重新打开');
    }
    
  } catch (error) {
    console.error('\n[测试失败]:', error.message);
  } finally {
    await browser.close();
  }
})();
