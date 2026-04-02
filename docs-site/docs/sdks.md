# SDKs

Install the SDK for your language and start integrating logidoc in a few lines.

## Install

::: code-group

```bash [TypeScript]
npm install logidoc
```

```bash [Python]
pip install logidoc
```

```bash [Go]
go get github.com/grignolalouis/logidoc-sdk-go
```

:::

## Initialize the client

::: code-group

```typescript [TypeScript]
import { GrignolalouisApiClient } from "logidoc";

const client = new GrignolalouisApiClient({
  baseUrl: "http://localhost:7042",
  headers: { Authorization: "Bearer sk-logidoc-your-key" }, // optional
});
```

```python [Python]
from logidoc import LogidocClient

client = LogidocClient(
    base_url="http://localhost:7042",
    headers={"Authorization": "Bearer sk-logidoc-your-key"},  # optional
)
```

```go [Go]
import logidoc "github.com/grignolalouis/logidoc-sdk-go"

client := logidoc.NewClient(
    logidoc.WithBaseURL("http://localhost:7042"),
)
```

:::

## Upload a document

::: code-group

```typescript [TypeScript]
import fs from "fs";

const doc = await client.uploadDocument({
  file: fs.createReadStream("report.pdf"),
});

console.log(doc.id);     // "5d04eeed-076d-..."
console.log(doc.status); // "uploaded"
```

```python [Python]
doc = client.upload_document(
    file=open("report.pdf", "rb"),
)

print(doc.id)     # "5d04eeed-076d-..."
print(doc.status) # "uploaded"
```

```go [Go]
f, _ := os.Open("report.pdf")
defer f.Close()

doc, _ := client.UploadDocument(ctx, f)

fmt.Println(doc.ID)     // "5d04eeed-076d-..."
fmt.Println(doc.Status) // "uploaded"
```

:::

## Trigger indexation

::: code-group

```typescript [TypeScript]
await client.indexDocument({ id: doc.id });
```

```python [Python]
client.index_document(id=doc.id)
```

```go [Go]
client.IndexDocument(ctx, doc.ID)
```

:::

Poll until ready:

::: code-group

```typescript [TypeScript]
let status = "indexing";
while (status === "indexing") {
  await new Promise((r) => setTimeout(r, 3000));
  const d = await client.getDocument({ id: doc.id });
  status = d.status;
}
```

```python [Python]
import time

status = "indexing"
while status == "indexing":
    time.sleep(3)
    d = client.get_document(id=doc.id)
    status = d.status
```

```go [Go]
for {
    d, _ := client.GetDocument(ctx, doc.ID)
    if d.Status != "indexing" {
        break
    }
    time.Sleep(3 * time.Second)
}
```

:::

## Get the table of contents

::: code-group

```typescript [TypeScript]
const toc = await client.getDocumentToc({ id: doc.id });

for (const section of toc.toc) {
  console.log(`${section.title} (p.${section.startPage}-${section.endPage})`);
  for (const child of section.children) {
    console.log(`  ${child.title}`);
  }
}
```

```python [Python]
toc = client.get_document_toc(id=doc.id)

for section in toc.toc:
    print(f"{section.title} (p.{section.start_page}-{section.end_page})")
    for child in section.children:
        print(f"  {child.title}")
```

```go [Go]
toc, _ := client.GetDocumentToc(ctx, doc.ID)

for _, section := range toc.Toc {
    fmt.Printf("%s (p.%d-%d)\n", section.Title, section.StartPage, section.EndPage)
    for _, child := range section.Children {
        fmt.Printf("  %s\n", child.Title)
    }
}
```

:::

## Retrieve sections

::: code-group

```typescript [TypeScript]
const sections = await client.getDocumentSections({
  id: doc.id,
  ids: "chapter-1-introduction,section-1-2-vision",
});

for (const s of sections.sections) {
  console.log(`## ${s.title} (p.${s.startPage})`);
  console.log(s.text);
}
```

```python [Python]
sections = client.get_document_sections(
    id=doc.id,
    ids="chapter-1-introduction,section-1-2-vision",
)

for s in sections.sections:
    print(f"## {s.title} (p.{s.start_page})")
    print(s.text)
```

```go [Go]
sections, _ := client.GetDocumentSections(ctx, doc.ID, "chapter-1-introduction,section-1-2-vision")

for _, s := range sections.Sections {
    fmt.Printf("## %s (p.%d)\n", s.Title, s.StartPage)
    fmt.Println(s.Text)
}
```

:::

## Delete a document

::: code-group

```typescript [TypeScript]
await client.deleteDocument({ id: doc.id });
```

```python [Python]
client.delete_document(id=doc.id)
```

```go [Go]
client.DeleteDocument(ctx, doc.ID)
```

:::

## Full example

::: code-group

```typescript [TypeScript]
import fs from "fs";
import { GrignolalouisApiClient } from "logidoc";

const client = new GrignolalouisApiClient({
  baseUrl: "http://localhost:7042",
});

// Upload
const doc = await client.uploadDocument({
  file: fs.createReadStream("report.pdf"),
});

// Index and wait
await client.indexDocument({ id: doc.id });
let d = await client.getDocument({ id: doc.id });
while (d.status === "indexing") {
  await new Promise((r) => setTimeout(r, 2000));
  d = await client.getDocument({ id: doc.id });
}

// Browse TOC
const toc = await client.getDocumentToc({ id: doc.id });
console.log(`${toc.toc.length} top-level sections`);

// Get specific sections
const sections = await client.getDocumentSections({
  id: doc.id,
  ids: toc.toc[0].id,
});
console.log(sections.sections[0].text);
```

```python [Python]
import time
from logidoc import LogidocClient

client = LogidocClient(base_url="http://localhost:7042")

# Upload
doc = client.upload_document(file=open("report.pdf", "rb"))

# Index and wait
client.index_document(id=doc.id)
while True:
    d = client.get_document(id=doc.id)
    if d.status != "indexing":
        break
    time.sleep(2)

# Browse TOC
toc = client.get_document_toc(id=doc.id)
print(f"{len(toc.toc)} top-level sections")

# Get specific sections
sections = client.get_document_sections(
    id=doc.id,
    ids=toc.toc[0].id,
)
print(sections.sections[0].text)
```

```go [Go]
package main

import (
    "context"
    "fmt"
    "os"
    "time"

    logidoc "github.com/grignolalouis/logidoc-sdk-go"
)

func main() {
    ctx := context.Background()
    client := logidoc.NewClient(logidoc.WithBaseURL("http://localhost:7042"))

    // Upload
    f, _ := os.Open("report.pdf")
    doc, _ := client.UploadDocument(ctx, f)
    f.Close()

    // Index and wait
    client.IndexDocument(ctx, doc.ID)
    for {
        d, _ := client.GetDocument(ctx, doc.ID)
        if d.Status != "indexing" {
            break
        }
        time.Sleep(2 * time.Second)
    }

    // Browse TOC
    toc, _ := client.GetDocumentToc(ctx, doc.ID)
    fmt.Printf("%d top-level sections\n", len(toc.Toc))

    // Get specific sections
    sections, _ := client.GetDocumentSections(ctx, doc.ID, toc.Toc[0].ID)
    fmt.Println(sections.Sections[0].Text)
}
```

:::

## Repositories

| Language | Repository |
|----------|-----------|
| TypeScript | [grignolalouis/logidoc-sdk-ts](https://github.com/grignolalouis/logidoc-sdk-ts) |
| Python | [grignolalouis/logidoc-sdk-python](https://github.com/grignolalouis/logidoc-sdk-python) |
| Go | [grignolalouis/logidoc-sdk-go](https://github.com/grignolalouis/logidoc-sdk-go) |
