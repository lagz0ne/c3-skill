---
id: ref-file-handling
c3-seal: c54cd6d41e70ef50586b9222a5ab690bc4e24ff606c19faede047b9cc8ab44e7
title: File Handling Pattern
type: ref
goal: Store and retrieve files (attachments, invoices) with deduplication and proper content type handling.
---

# File Handling Pattern

## Goal

Store and retrieve files (attachments, invoices) with deduplication and proper content type handling.

## Choice

- Store files as BYTEA in PostgreSQL with content-hash-based deduplication
- Naming convention encodes prefix, timestamp, and hash for uniqueness; returning name preserves original filename
- Result type uses `success | failure | skipped` tri-state for explicit error handling

## Why

- PostgreSQL storage keeps files transactional with the rest of the data (no external object store to sync)
- Content hashing prevents duplicate uploads without caller awareness
- Tri-state result avoids exceptions for expected conditions like duplicates

## Architecture

Files are stored in PostgreSQL as binary data, with metadata for retrieval:

```mermaid
graph LR
    U[Upload] --> S[fileStorage.store]
    S --> H[Hash Content]
    H --> D[(files table)]
    D --> G[fileStorage.get]
    G --> R[Response]
```

## File Storage Service

```typescript
export const fileStorage = service({
  deps: { logger, fileQueries },
  factory: (_, { logger, fileQueries }) => ({
    store: async (ctx, content: Blob, prefix: string),
    has: async (ctx, hashValue: string),
    get: async (ctx, filename: string),
    remove: async (ctx, name: string),
    removeByPrefix: async (ctx, prefix: string),
  }),
});
```

## Naming Convention

| Name Type | Format | Example |
| --- | --- | --- |
| storingName | {prefix}.{timestamp}.{hash}{ext} | 451.1705312000.a1b2c3d4.pdf |
| returningName | {prefix}.{timestamp}.{hash}_{originalName}{ext} | 451.1705312000.a1b2c3d4_invoice.pdf |

The `returningName` preserves the original filename for display, while `storingName` is used internally.

## Supported File Types

| Type | MIME | Extension |
| --- | --- | --- |
| XML | application/xml | .xml |
| PDF | application/pdf | .pdf |
| XLS | application/vnd.ms-excel | .xls |
| XLSX | application/vnd.openxmlformats-officedocument.spreadsheetml.sheet | .xlsx |
| ZIP | application/zip | .zip |

## File Utilities

```typescript
export const fileUtils = {
  whatType(contentType: string): 'xml' | 'zip' | 'pdf' | 'xls' | 'xlsx' | 'unknown',
  extension(contentType: string): string,
  contentType(ext: string): string,
};
```

## Store Operation

```typescript
async store(ctx, content: Blob, prefix: string = ''): Promise<OperationResult<{
  returningName: string;
  storingFile: string;
}>>
```

1. Validate prefix (no underscores)
2. Determine file type from MIME
3. Hash content (MD5, first 8 chars)
4. Generate unique names
5. Insert into database

```typescript
const hashValue = createHash('md5')
  .update(Buffer.from(ab))
  .digest('hex')
  .substring(0, 8);

const storingFile = `${prefix}.${timestamp}.${hashValue}${ext}`;
const returningName = `${prefix}.${timestamp}.${hashValue}_${filename}${ext}`;
```

## Get Operation

```typescript
async get(ctx, filename: string): Promise<OperationResult<{
  content: Blob;
  contentType: string;
  storingName: string;
}>>
```

1. Parse filename to extract storingName
2. Query database by storingName
3. Return Blob with correct content type

## Database Schema

```sql
CREATE TABLE files (
  id SERIAL PRIMARY KEY,
  storing_name TEXT UNIQUE NOT NULL,
  original_name TEXT,
  content_type TEXT NOT NULL,
  content BYTEA NOT NULL,
  size_bytes INTEGER NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);
```

## Result Types

```typescript
type OperationResult<P> =
  | { state: 'success'; result: P }
  | { state: 'failure'; error: any }
  | { state: 'skipped'; detail: string };
```

## File Queries

| Query | Purpose |
| --- | --- |
| insertFile | Store new file |
| getFileByStoringName | Retrieve file |
| fileExistsByStoringName | Check existence |
| deleteFileByStoringName | Remove file |
| deleteFilesByPrefix | Bulk remove |

## Usage in Invoice Import

```typescript
// Store uploaded file
const storedFileResult = await fileStorage.store(ctx, attachment);

if (storedFileResult.state === 'failure') {
  return { state: 'failure', error: storedFileResult.error };
}

// Get file for processing
const getFileResult = await fileStorage.get(ctx, filename);
const { content: file } = getFileResult.result;

// Parse based on type
if (fileUtils.whatType(file.type) === 'xml') {
  const content = await file.text();
  const invoice = parseInvoice(content);
}
```

## Serving Files via Route

```typescript
export const Route = createFileRoute('/files/$hash')({
  loader: async ({ params }) => {
    const result = await fileStorage.get(ctx, params.hash);
    if (result.state === 'failure') {
      throw new Error('File not found');
    }
    return new Response(result.result.content, {
      headers: { 'Content-Type': result.result.contentType }
    });
  }
});
```

## Edge Cases

| Scenario | Behavior |
| --- | --- |
| Duplicate upload | Hash collision detected, skip |
| Unknown file type | Store as application/octet-stream |
| Missing file | Return failure state |
| Prefix with underscore | Reject (reserved for name parsing) |

## Cited By

- c3-2-api (File Storage)
