import { BrowserHelper } from '../helpers/browser.js';
import { AuthHelper } from '../helpers/auth.js';

/**
 * 用户门户测试
 */
describe('用户门户测试', () => {
  let browser;
  let auth;

  beforeAll(async () => {
    browser = new BrowserHelper();
    await browser.launch();
    auth = new AuthHelper(browser);
  });

  afterAll(async () => {
    await browser.close();
  });

  test('应该能够访问用户门户首页', async () => {
    await browser.goto('/user');
    await browser.screenshot('user-portal-home');
    
    // 验证跳转到用户门户
    const url = browser.page.url();
    expect(url).toContain('/user');
  });

  test('应该显示注册页面', async () => {
    await browser.goto('/user/register');
    await browser.screenshot('user-register-page');
    
    // 验证注册表单存在
    const hasForm = await browser.exists('form');
    expect(hasForm).toBe(true);
  });

  test('应该显示登录页面', async () => {
    await browser.goto('/user/login');
    await browser.screenshot('user-login-page');
    
    // 验证登录表单
    const hasUsername = await browser.exists('input[type="text"]');
    const hasPassword = await browser.exists('input[type="password"]');
    
    expect(hasUsername).toBe(true);
    expect(hasPassword).toBe(true);
  });
});
