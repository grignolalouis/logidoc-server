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
      { text: 'Guide', link: '/guide/setup' },
      { text: 'MCP', link: '/mcp/setup' },
      { text: 'Reference', link: '/reference/upload-document' },
      { text: 'GitHub', link: 'https://github.com/grignolalouis/logidoc-server' }
    ],

    sidebar: [
      {
        text: 'Guide',
        items: [
          { text: 'Setup', link: '/guide/setup' },
          { text: 'How It Works', link: '/guide/how-it-works' },
          { text: 'Deployment', link: '/guide/deployment' },
        ]
      },
      {
        text: 'MCP',
        items: [
          { text: 'Setup', link: '/mcp/setup' },
          { text: 'list_documents', link: '/mcp/list-documents' },
          { text: 'search', link: '/mcp/search' },
          { text: 'get_toc', link: '/mcp/get-toc' },
          { text: 'get_sections', link: '/mcp/get-sections' },
        ]
      },
      {
        text: 'HTTP API',
        items: [
          { text: badge('POST', 'Upload Document'), link: '/reference/upload-document' },
          { text: badge('GET', 'List Documents'), link: '/reference/list-documents' },
          { text: badge('GET', 'Get Document'), link: '/reference/get-document' },
          { text: badge('POST', 'Index Document'), link: '/reference/index-document' },
          { text: badge('GET', 'Get TOC'), link: '/reference/get-toc' },
          { text: badge('GET', 'Get Sections'), link: '/reference/get-sections' },
          { text: badge('GET', 'Search'), link: '/reference/search' },
          { text: badge('GET', 'Download File'), link: '/reference/download-file' },
          { text: badge('DELETE', 'Delete Document'), link: '/reference/delete-document' },
          { text: badge('GET', 'Health'), link: '/reference/health' },
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
