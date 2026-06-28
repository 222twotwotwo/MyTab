export interface MarketStyle {
  id: string
  name: string
  description: string
  tags: string[]
  swatches: string[]
  html: string
}

const clockScript = `
<script>
const clock = document.querySelector('[data-clock]');
const date = document.querySelector('[data-date]');
function pad(value) {
  return String(value).padStart(2, '0');
}
function tick() {
  const now = new Date();
  if (clock) clock.textContent = pad(now.getHours()) + ':' + pad(now.getMinutes());
  if (date) date.textContent = now.toLocaleDateString('zh-CN', { weekday: 'long', month: 'long', day: 'numeric' });
}
tick();
setInterval(tick, 1000);
</script>`

function page(title: string, css: string, body: string): string {
  return `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>${title}</title>
  <style>
    * { box-sizing: border-box; }
    html, body { min-height: 100%; margin: 0; }
    body {
      display: grid;
      place-items: center;
      padding: 42px;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      letter-spacing: 0;
    }
    main { width: min(1040px, 100%); }
    h1, p { margin: 0; }
    .time { font-size: clamp(58px, 10vw, 132px); line-height: .92; font-weight: 850; }
    .date { margin-top: 14px; font-size: 18px; font-weight: 750; }
    .grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 14px; margin-top: 34px; }
    .tile { min-height: 122px; border-radius: 18px; padding: 18px; }
    .tile strong { display: block; margin-bottom: 10px; font-size: 16px; }
    .tile span { display: block; font-size: 13px; line-height: 1.6; }
    @media (max-width: 780px) {
      body { padding: 24px; place-items: start center; }
      .grid { grid-template-columns: 1fr; }
    }
    ${css}
  </style>
</head>
<body>
  ${body}
  ${clockScript}
</body>
</html>`
}

export const marketStyles: MarketStyle[] = [
  {
    id: 'morning-focus',
    name: '晨间专注',
    description: '明亮、清爽，适合把新标签页变成每日开始的工作台。',
    tags: ['清爽', '专注', '浅色'],
    swatches: ['#f7fbf7', '#0f766e', '#f59e0b', '#263242'],
    html: page(
      '晨间专注',
      `
      body { color: #1d2b2a; background: linear-gradient(135deg, #f7fbf7 0%, #e8f4ee 52%, #fff6df 100%); }
      .date { color: #52706d; }
      .grid { grid-template-columns: 1.2fr 1fr 1fr; }
      .tile { border: 1px solid rgb(28 81 75 / 13%); background: rgb(255 255 255 / 72%); box-shadow: 0 18px 45px rgb(38 50 66 / 10%); }
      .tile:nth-child(1) { background: #0f766e; color: white; }
      .tile:nth-child(2) { background: #fff3c4; }
      .tile:nth-child(3) { background: #ffffff; }
      `,
      `
      <main>
        <p class="date" data-date></p>
        <h1 class="time" data-clock></h1>
        <section class="grid">
          <div class="tile"><strong>今日主线</strong><span>把最重要的一件事放到最显眼的位置。</span></div>
          <div class="tile"><strong>轻量记录</strong><span>适合搭配便签、日程、待办类扩展使用。</span></div>
          <div class="tile"><strong>呼吸感</strong><span>浅色背景和柔和阴影，打开页面不刺眼。</span></div>
        </section>
      </main>`
    )
  },
  {
    id: 'orbit-night',
    name: '轨道夜航',
    description: '深色仪表盘风格，适合夜间和沉浸式工作场景。',
    tags: ['深色', '仪表盘', '科技'],
    swatches: ['#101820', '#43d9ad', '#f4d35e', '#e76f51'],
    html: page(
      '轨道夜航',
      `
      body { color: #f4f7fb; background: radial-gradient(circle at 18% 18%, #264653 0, transparent 30%), linear-gradient(145deg, #101820 0%, #18242e 48%, #0d1017 100%); }
      main { position: relative; }
      main::before { content: ""; position: absolute; inset: -34px; border: 1px solid rgb(67 217 173 / 25%); border-radius: 34px; pointer-events: none; }
      .date { color: #a9c8c0; }
      .time { color: #ffffff; text-shadow: 0 0 34px rgb(67 217 173 / 28%); }
      .tile { border: 1px solid rgb(255 255 255 / 12%); background: rgb(255 255 255 / 7%); box-shadow: inset 0 1px 0 rgb(255 255 255 / 8%); }
      .tile strong { color: #43d9ad; }
      .tile:nth-child(2) strong { color: #f4d35e; }
      .tile:nth-child(3) strong { color: #e76f51; }
      `,
      `
      <main>
        <p class="date" data-date></p>
        <h1 class="time" data-clock></h1>
        <section class="grid">
          <div class="tile"><strong>任务轨道</strong><span>保持少量目标，避免新标签页变成噪声入口。</span></div>
          <div class="tile"><strong>系统状态</strong><span>适合放置服务监控、构建状态或常用工作入口。</span></div>
          <div class="tile"><strong>深夜模式</strong><span>低亮度背景，长时间打开更舒适。</span></div>
        </section>
      </main>`
    )
  },
  {
    id: 'paper-desk',
    name: '纸张书桌',
    description: '带有文档感和阅读节奏的静态标签页样式。',
    tags: ['阅读', '文档', '安静'],
    swatches: ['#fffaf0', '#2f3a32', '#b7791f', '#3f7cac'],
    html: page(
      '纸张书桌',
      `
      body { color: #2f3a32; background: #f2eee3; }
      main { border: 1px solid #d8cfbd; border-radius: 4px; padding: 42px; background: #fffaf0; box-shadow: 0 24px 60px rgb(52 45 35 / 14%); }
      .date { color: #796d5d; }
      .time { font-family: Georgia, "Times New Roman", serif; font-weight: 700; }
      .tile { border: 1px solid #e4dbc8; border-radius: 4px; background: #fffdf6; }
      .tile strong { color: #b7791f; }
      .tile:nth-child(2) strong { color: #3f7cac; }
      .tile:nth-child(3) strong { color: #2f855a; }
      `,
      `
      <main>
        <p class="date" data-date></p>
        <h1 class="time" data-clock></h1>
        <section class="grid">
          <div class="tile"><strong>阅读计划</strong><span>适合知识管理、文档入口和学习型标签页。</span></div>
          <div class="tile"><strong>项目笔记</strong><span>清晰的纸张边界让内容层次更稳定。</span></div>
          <div class="tile"><strong>轻量桌面</strong><span>弱装饰、强可读，适合每天反复打开。</span></div>
        </section>
      </main>`
    )
  },
  {
    id: 'pulse-board',
    name: '脉冲看板',
    description: '高对比、强节奏，适合把常用入口放到醒目的页面里。',
    tags: ['活力', '看板', '高对比'],
    swatches: ['#151515', '#ff4d6d', '#2ec4b6', '#ffbe0b'],
    html: page(
      '脉冲看板',
      `
      body { color: #f9fafb; background: linear-gradient(135deg, #151515 0%, #262626 45%, #0b2b2b 100%); }
      .date { color: #b7f6ed; }
      .time { color: #ffffff; }
      .grid { grid-template-columns: repeat(3, minmax(0, 1fr)); }
      .tile { border: 2px solid #f9fafb; box-shadow: 8px 8px 0 #000000; }
      .tile:nth-child(1) { background: #ff4d6d; }
      .tile:nth-child(2) { background: #2ec4b6; color: #071716; }
      .tile:nth-child(3) { background: #ffbe0b; color: #1f1300; }
      `,
      `
      <main>
        <p class="date" data-date></p>
        <h1 class="time" data-clock></h1>
        <section class="grid">
          <div class="tile"><strong>快速开始</strong><span>醒目的色块适合承载高频入口和固定动作。</span></div>
          <div class="tile"><strong>状态提醒</strong><span>更强的视觉锚点，适合目标、习惯和提醒。</span></div>
          <div class="tile"><strong>今日能量</strong><span>打开浏览器时给自己一个明确启动信号。</span></div>
        </section>
      </main>`
    )
  }
]
