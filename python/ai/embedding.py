"""
Embedding Service
Uses sentence-transformers to generate vector embeddings
"""

from sentence_transformers import SentenceTransformer
from typing import Optional
import numpy as np


class EmbeddingService:
    """
    Vector embedding service.
    Supports local models (sentence-transformers).
    """

    def __init__(self, model_name: str = "all-MiniLM-L6-v2"):
        """
        Initialize the embedding service.

        Args:
            model_name: Model name, defaults to lightweight and fast all-MiniLM-L6-v2
        """
        self.model_name = model_name
        self.mock = False
        self._model: Optional[SentenceTransformer] = None

        try:
            self._model = SentenceTransformer(model_name)
            print(f"Loaded embedding model: {model_name}")
        except Exception as e:
            print(f"Warning: Failed to load model '{model_name}': {e}")
            print("Running in mock mode")
            self.mock = True

    @property
    def dimension(self) -> int:
        """Returns the vector dimension"""
        if self.mock:
            return 384  # Default dimension for MiniLM
        return self._model.get_sentence_embedding_dimension()

    async def embed(self, text: str) -> list[float]:
        """
        Generate vector embedding for a single text.

        Args:
            text: Input text

        Returns:
            List of floats representing the vector
        """
        if self.mock:
            # Return random vector
            import random
            return [random.random() for _ in range(384)]

        embedding = self._model.encode(text)
        return embedding.tolist()

    async def embed_batch(self, texts: list[str]) -> list[list[float]]:
        """
        Batch generate vector embeddings.

        Args:
            texts: List of texts

        Returns:
            List of vectors
        """
        if self.mock:
            import random
            return [
                [random.random() for _ in range(384)]
                for _ in texts
            ]

        embeddings = self._model.encode(texts)
        return embeddings.tolist()

    def encode(self, text: str) -> np.ndarray:
        """
        Synchronous encoding interface (for internal use).

        Args:
            text: Input text

        Returns:
            Numpy vector
        """
        if self.mock:
            import random
            return np.array([random.random() for _ in range(384)])

        return self._model.encode(text)
