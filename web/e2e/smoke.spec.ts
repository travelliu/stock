import { test, expect } from '@playwright/test'

test.describe('smoke', () => {
  test('login -> add stock -> analysis -> draft', async ({ page }) => {
    await page.goto('/login')
    await page.fill('input[placeholder="用户名"]', 'admin')
    const adminPassword = process.env.ADMIN_PASSWORD || 'changeme-see-server-logs'
    await page.fill('input[placeholder="密码"]', adminPassword)
    await page.click('button:has-text("登录")')
    await page.waitForURL('/portfolio')

    await page.click('button:has-text("添加股票")')
    await page.click('.el-select-v2__wrapper')
    await page.fill('.el-select-v2__wrapper input', '600519')
    await page.click('.el-select-dropdown__item:has-text("600519")')
    await page.click('button:has-text("添加")')

    await page.click('text=详情')
    await page.waitForURL(/\/stock\//)

    await page.click('text=分析')
    await expect(page.locator('text=价差模型')).toBeVisible()

    await page.click('text=草稿')
    await page.fill('input', '1600')
    await page.click('button:has-text("保存")')
  })
})
