# Local ONNX Semantic Search

`c3x` uses `sentence-transformers/all-MiniLM-L6-v2` through the repository's ONNX export pinned at Hugging Face revision `1110a243fdf4706b3f48f1d95db1a4f5529b4d41`. Query and entity text are tokenized with the model `vocab.txt`, run through ONNX Runtime, mean-pooled over the attention mask, and L2-normalized into 384-dimensional vectors.

Distribution choice: download on first use. The base Go binary stays small and does not embed the roughly 90 MB model or the platform-specific ONNX Runtime native library. `c3x index` and `c3x search --semantic` populate the user cache (`C3_SEMANTIC_CACHE_DIR`, or the OS user cache under `c3/semantic`) and work offline afterward. The ONNX Go binding itself is CGO-backed; `CGO_ENABLED=0` builds compile with a semantic-unavailable stub, while CGO-enabled platform builds get real local embeddings. Plain keyword/graph search never downloads assets and remains usable with no model present.

Search fusion is additive. FTS5/content/graph results remain the primary offline path; when the semantic index and cached ONNX runtime are present, semantic hits are combined with keyword/graph ranks using reciprocal-rank fusion.
