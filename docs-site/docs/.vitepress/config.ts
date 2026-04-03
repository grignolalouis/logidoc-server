import { defineConfig } from 'vitepress'

const badge = (method: string, text: string) =>
  `<span class="sidebar-badge ${method.toLowerCase()}">${method}</span> ${text}`

export default defineConfig({
  title: 'logidoc',
  description: 'Document indexing for AI agents',
  base: '/logidoc-server/',

  head: [
    ['link', { rel: 'icon', href: '/logidoc-server/images/logidocLogo.png' }]
  ],

  themeConfig: {
    logo: '/images/logidocLogo.png',

    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Reference', link: '/reference/upload-document' },
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
        text: 'Reference',
        items: [
          { text: badge('POST', 'Upload Document'), link: '/reference/upload-document' },
          { text: badge('GET', 'List Documents'), link: '/reference/list-documents' },
          { text: badge('GET', 'Get Document'), link: '/reference/get-document' },
          { text: badge('POST', 'Index Document'), link: '/reference/index-document' },
          { text: badge('GET', 'Get TOC'), link: '/reference/get-toc' },
          { text: badge('GET', 'Get Sections'), link: '/reference/get-sections' },
          { text: badge('GET', 'Download File'), link: '/reference/download-file' },
          { text: badge('DELETE', 'Delete Document'), link: '/reference/delete-document' },
          { text: badge('GET', 'Health'), link: '/reference/health' },
          { text: 'MCP Tools', link: '/reference/mcp-tools' },
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
