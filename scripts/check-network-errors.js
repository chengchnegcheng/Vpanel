const puppeteer = require('puppeteer');

(async () => {
  const browser = await puppeteer.launch({
    headless: false,  // 显示浏览器
    args: ['--no-sandbox', '--disable-setuid-sandbox'],
    devtools: true  // 打开开发者工具
  });
  
  const page = await browser.newPage();
  
  // 监听所有请求和响应
  page.on('request', request => {
    const url = request.url();
    if (url.includes('/api/admin/ip')) {
      console.log(`\n→ 请求: ${request.method()} ${url}`);
    }
  });
  
  page.on('response', async response => {
    const url = response.url();
    if (url.includes('/api/admin/ip')) {
      const status = response.status();
      console.log(`← 响应: ${status} ${url}`);
      
      if (status >= 400) {
        try {
          const body = await response.text();
          console.log(`   错误内容: ${body}`);
        } catch (e) {}
      }
    }
  });
  
  page.on('console', msg => {
    const text = msg.text();
    if (text.includes('ERR-') || text.includes('Failed') || text.includes('错误')) {
      console.log(`\n[浏览器] ${text}`);
    }
  });
  
  try {
    console.log('登录...');
    await page.goto('http://localhost:8081/login', { waitUntil: 'networkidle2' });
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    const inputs = await page.$$('input');
    await inputs[0].type('admin');
    await inputs[1].type('admin123');
    
    const buttons = await page.$$('button');
    for (const button of buttons) {
      const text = await page.evaluate(el => el.textContent, button);
      if (text.includes('登录')) {
        await button.click();
        break;
      }
    }
    
    await page.waitForNavigation({ waitUntil: 'networkidle2', timeout: 10000 }).catch(() => {});
    await new Promise(resolve => setTimeout(resolve, 2000));
    
    console.log('\n访问IP限制页面...\n');
    await page.goto('http://localhost:8081/admin/ip-restriction', { waitUntil: 'networkidle2' });
    
    console.log('\n等待30秒观察错误...');
    await new Promise(resolve => setTimeout(resolve, 30000));
    
  } catch (error) {
    console.error('错误:', error.message);
  }
  
  // 不关闭浏览器，让用户手动检查
  console.log('\n浏览器保持打开，请手动检查Network标签');
})();
