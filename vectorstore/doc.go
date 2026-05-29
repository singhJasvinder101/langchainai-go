// Package vectorstore defines a provider-agnostic interface for retrieval storage.
//
// Each backend lives in its own subpackage and accepts an embedder.Embedder at
// construction time. Import only the backends you need:
//
//   - vectorstore/memory
//   - vectorstore/chroma
//   - vectorstore/qdrant
//   - vectorstore/pinecone
//   - vectorstore/weaviate
package vectorstore
