package mcp

import mcplib "trpc.group/trpc-go/trpc-mcp-go"

// tools returns the MCP tool definitions.
func tools() []*mcplib.Tool {
	return []*mcplib.Tool{
		mcplib.NewTool("logidoc_list_documents",
			mcplib.WithDescription("List all indexed documents with their IDs, names, descriptions and indexation status."),
		),
		mcplib.NewTool("logidoc_get_toc",
			mcplib.WithDescription("Get the table of contents of a document: hierarchical tree of section titles and summaries. Use this to understand what a document contains before retrieving specific sections."),
			mcplib.WithString("doc_id", mcplib.Required(), mcplib.Description("Document ID")),
		),
		mcplib.NewTool("logidoc_get_sections",
			mcplib.WithDescription("Get the full text content of specific sections from a document. Use node IDs from the table of contents."),
			mcplib.WithString("doc_id", mcplib.Required(), mcplib.Description("Document ID")),
			mcplib.WithString("node_ids", mcplib.Required(), mcplib.Description("Comma-separated node IDs to retrieve")),
		),
	}
}
