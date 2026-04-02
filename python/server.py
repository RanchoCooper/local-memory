"""
LocalMemory AI Server
Provides Embedding and information extraction services
"""

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import Optional
import uvicorn

from ai.embedding import EmbeddingService
from ai.extractor import MemoryExtractor

app = FastAPI(title="LocalMemory AI", version="1.0.0")

# CORS configuration
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Global service instances
embedding_service: Optional[EmbeddingService] = None
extractor: Optional[MemoryExtractor] = None


def init_services():
    """Initialize AI services"""
    global embedding_service, extractor
    try:
        embedding_service = EmbeddingService()
        extractor = MemoryExtractor()
        print("AI services initialized successfully")
    except Exception as e:
        print(f"Warning: Failed to initialize AI services: {e}")
        print("Running in mock mode")


# Request/Response models
class EmbedRequest(BaseModel):
    text: str


class EmbedBatchRequest(BaseModel):
    texts: list[str]


class EmbedResponse(BaseModel):
    embedding: list[float]
    error: Optional[str] = None


class EmbedBatchResponse(BaseModel):
    embeddings: list[list[float]]
    error: Optional[str] = None


class ExtractRequest(BaseModel):
    text: str


class ExtractResponse(BaseModel):
    type: str
    key: str
    value: str
    confidence: float
    error: Optional[str] = None


class HealthResponse(BaseModel):
    status: str
    embedding_model: str
    mock_mode: bool


@app.get("/health")
async def health():
    """Health check"""
    if embedding_service is None:
        init_services()

    mock_mode = embedding_service is None or embedding_service.mock
    return HealthResponse(
        status="ok",
        embedding_model=embedding_service.model_name if embedding_service else "unknown",
        mock_mode=mock_mode
    )


@app.post("/embed", response_model=EmbedResponse)
async def embed(request: EmbedRequest):
    """Generate vector embedding for a single text"""
    try:
        if embedding_service is None:
            init_services()

        if embedding_service is None or embedding_service.mock:
            # Mock mode: return random vector
            import random
            embedding = [random.random() for _ in range(384)]
            return EmbedResponse(embedding=embedding)

        embedding = await embedding_service.embed(request.text)
        return EmbedResponse(embedding=embedding)
    except Exception as e:
        return EmbedResponse(embedding=[], error=str(e))


@app.post("/embed/batch", response_model=EmbedBatchResponse)
async def embed_batch(request: EmbedBatchRequest):
    """Batch generate vector embeddings"""
    try:
        if embedding_service is None:
            init_services()

        if embedding_service is None or embedding_service.mock:
            # Mock mode
            import random
            embeddings = [
                [random.random() for _ in range(384)]
                for _ in request.texts
            ]
            return EmbedBatchResponse(embeddings=embeddings)

        embeddings = await embedding_service.embed_batch(request.texts)
        return EmbedBatchResponse(embeddings=embeddings)
    except Exception as e:
        return EmbedBatchResponse(embeddings=[], error=str(e))


@app.post("/extract", response_model=ExtractResponse)
async def extract(request: ExtractRequest):
    """Extract structured memory from text"""
    try:
        if extractor is None:
            init_services()

        if extractor is None:
            return ExtractResponse(
                type="fact",
                key="unknown",
                value=request.text,
                confidence=0.5,
                error="Extractor not initialized"
            )

        result = await extractor.extract(request.text)
        return ExtractResponse(**result)
    except Exception as e:
        return ExtractResponse(
            type="fact",
            key="unknown",
            value=request.text,
            confidence=0.0,
            error=str(e)
        )


@app.get("/")
async def root():
    """Root path"""
    return {"message": "LocalMemory AI Server", "version": "1.0.0"}


if __name__ == "__main__":
    init_services()
    uvicorn.run(app, host="127.0.0.1", port=8081)
