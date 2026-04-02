import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'logidoc',
  description: 'Document indexing for AI agents',
  base: '/logidoc-server/',

  head: [
    ['link', { rel: 'icon', href: '/logidoc-server/logo.svg' }]
  ],

  themeConfig: {
    logo: '/logo.svg',

    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'API', link: '/api/endpoints' },
      { text: 'SDKs', link: '/sdks' },
      { text: 'GitHub', link: 'https://github.com/grignolalouis/logidoc-server' }
    ],

    sidebar: [
      {
        text: 'Introduction',
        items: [
          { text: 'What is logidoc?', link: '/guide/what-is-logidoc' },
          { text: 'Getting Started', link: '/guide/getting-started' },
        ]
      },
      {
        text: 'Guide',
        items: [
          { text: 'Configuration', link: '/guide/configuration' },
          { text: 'LLM Providers', link: '/guide/providers' },
          { text: 'MCP Integration', link: '/guide/mcp' },
          { text: 'Authentication', link: '/guide/authentication' },
          { text: 'Deployment', link: '/guide/deployment' },
        ]
      },
      {
        text: 'API Reference',
        items: [
          { text: 'Endpoints', link: '/api/endpoints' },
          { text: 'MCP Tools', link: '/api/mcp-tools' },
        ]
      },
      {
        text: 'SDKs',
        items: [
          { text: 'Overview', link: '/sdks' },
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/grignolalouis/logidoc-server' }
    ],

    footer: {
      message: 'Released under the MIT License.',
    }
  }
})
