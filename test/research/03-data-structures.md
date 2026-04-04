# PageIndex — Data Structures

## TOCNode (Tree Node)

```json
{
  "title": "Section Name",
  "node_id": "scene-1-wall",
  "summary": "Single sentence describing section content (15-25 words)",
  "nodes": [
    {
      "title": "Subsection Name",
      "node_id": "scene-1-wall-details",
      "summary": "Single sentence summary of subsection",
      "nodes": []
    }
  ]
}
```

### Required Fields (exactly 4):
| Field      | Type     | Description                                        |
|------------|----------|----------------------------------------------------|
| `title`    | string   | Section heading                                    |
| `node_id`  | string   | Unique kebab-case identifier                       |
| `summary`  | string   | One sentence, 15-25 words                          |
| `nodes`    | array    | Child nodes (recursive, max depth 2)               |

### Extended Fields (from some implementations):
| Field         | Type     | Description                                     |
|---------------|----------|-------------------------------------------------|
| `start_index` | int      | Start page/section reference                    |
| `end_index`   | int      | End page/section reference                      |
| `description` | string   | Optional detailed explanation                   |
| `metadata`    | object   | Key-value pairs for contextual info             |
| `sub_nodes`   | array    | Alternative name for `nodes` in some versions   |

## Full Index Structure

```json
[
  {
    "title": "Chapter 1: Introduction",
    "node_id": "chapter-1-introduction",
    "summary": "Introduces the core concepts and background of the topic.",
    "nodes": [
      {
        "title": "1.1 Background",
        "node_id": "background",
        "summary": "Provides historical context and motivation for the work.",
        "nodes": []
      },
      {
        "title": "1.2 Problem Statement",
        "node_id": "problem-statement",
        "summary": "Defines the specific problem being addressed.",
        "nodes": []
      }
    ]
  },
  {
    "title": "Chapter 2: Methodology",
    "node_id": "chapter-2-methodology",
    "summary": "Describes the approach and techniques used in the study.",
    "nodes": []
  }
]
```

## Node-to-Content Mapping

The `node_id` maps directly to raw content (text, images, tables).
During Phase 2B (extraction), the system scans the document and extracts
content between section boundaries identified by matching headings.

## Navigation Output

When the LLM navigates (Phase 2A), it returns:
```
node-id-1, node-id-2, node-id-3
```
Just comma-separated IDs, no explanation. Maximum 1-3 nodes per query.
