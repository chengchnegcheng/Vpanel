const puppeteer = require('puppeteer');

(async () => {
  console.log('开始检查 IP 限制错误...\n');

  const browser = await puppeteer.launch({
    headless: false,
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });

  const page = await browser.newPage();
  
  // 监听所有请求和响应
  page.on('response', async response => {
    const url = response.url();
    const status = response.status();
    
    if (url.includes('/api/admin/ip-restrictions')) {
      console.log(`\n📡 API: ${status} ${url}`);
      
      try {
        const body = await response.text();
        console.log('响应:', body);
      } catch (e) {
        console.log('无法读取响应');
      }
    }
  });

  // 监听控制台错误
  page.on('console', msg => {
    if (msg.type() === 'error') {
      console.log('❌ 浏览器错误:', msg.text());
    }
  });

  try {
    // 访问登录页面
    console.log('1. 访问登录页面...');
    await page.goto('http://localhost:8081/admin/login', { 
      waitUntil: 'networkidle2',
      timeout: 30000 
    });
    await new Promise(resolve => setTimeout(resolve, 2000));

    // 查找输入框
    const inputs = await page.$$('input');
    console.log(`找到 ${inputs.length} 个输入框`);

    if (inputs.length >= 2) {
      // 输入用户名和密码
      await inputs[0].type('admin');
      await inputs[1].type('admin123');
      
      // 点击登录按钮
      const button = await page.$('button');
      if (button) {
        await button.click();
        console.log('2. 点击登录按钮...');
        await new Promise(resolve => setTimeout(resolve, 3000));
      }
    }

    // 检查是否有 token
    const token = await page.evaluate(() => localStorage.getItem('token'));
    console.log('Token:', token ? '存在' : '不存在');

    if (!token) {
      console.log('\n❌ 登录失败，尝试直接访问 IP 限制页面...');
    }

    // 访问 IP 限制页面
    console.log('\n3. 访问 IP 限制页面...');
    await page.goto('http://localhost:8081/admin/ip-restriction', { 
      waitUntil: 'networkidle2',
      timeout: 30000 
    });
    await new Promise(resolve => setTimeout(resolve, 3000));

    // 检查页面上的错误
    const pageContent = await page.content();
    
    if (pageContent.includes('ERR-') || pageContent.includes('应用错误')) {
      console.log('\n❌ 页面上发现错误！');
      
      // 查找错误元素
      const errorText = await page.evaluate(() => {
        const errorEl = document.querySelector('.error-message, .el-message__content, [class*="error"]');
        return errorEl ? errorEl.textContent : null;
      });
      
      if (errorText) {
        console.log('错误内容:', errorText);
      }
    }

    // 手动调用 API
    console.log('\n4. 手动测试 API...');
    const apiResult = await page.evaluate(async (token) => {
      try {
        const response = await fetch('http://localhost:8081/api/admin/ip-restrictions/stats', {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        });
        
        const text = await response.text();
        return {
          status: response.status,
          statusText: response.statusText,
          body: text
        };
      } catch (error) {
        return {
          error: error.message
        };
      }
    }, token);

    console.log('API 结果:');
    console.log(JSON.stringify(apiResult, null, 2));

    // 截图
    await page.screenshot({ path: 'ip-error-screenshot.png', fullPage: true });
    console.log('\n✓ 截图已保存: ip-error-screenshot.png');

  } catch (error) {
    console.error('\n❌ 错误:', error.message);
  } finally {
    await browser.close();
  }
})();
