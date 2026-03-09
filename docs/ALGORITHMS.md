# Algorithms Used

## 1) Parallel Incremental Scan

- Walk configured roots recursively.
- Skip ignored directories (`.git`, `node_modules`, etc.).
- Process files in parallel with `runtime.NumCPU()` workers.
- Reuse old record when `size` and `modtime` are unchanged.

## 2) Tokenization and Labeling

- Tokenize filename/path/content using regex `[A-Za-z0-9_]{2,}`.
- Remove stop words.
- Keep top-frequency tokens.
- Generate labels:
  - extension labels (`ext:md`)
  - folder labels (`folder:documents`)
  - topic labels (`topic:toefl`)

## 3) Ranking Function

For each candidate record:

`score = 0.45 * token_match + 0.30 * recency + 0.15 * path_boost + 0.10 * label_boost`

Where:

- `token_match`: matched query token frequency
- `recency`: `exp(-age_days / 45)`
- `path_boost`: preference for Desktop/Documents/Downloads
- `label_boost`: overlap between query tokens and labels

## 4) Inverted Index

- Build `token -> [file_id...]` map.
- Search starts from token postings, then ranks candidates.
- Avoid full scan of all files per query.

## 5) Privacy Gate

- Global content reading is OFF by default.
- If OFF, index uses filename/path only.
- Content indexing requires explicit user enable (`fileatlas content on`) or file-level `full`.
