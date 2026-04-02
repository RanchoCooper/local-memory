"""
Embedding 服务
使用 sentence-transformers 生成向量嵌入
"""

from sentence_transformers import SentenceTransformer
from typing import Optional
import numpy as np


class EmbeddingService:
    """
    向量嵌入服务
    支持本地模型（sentence-transformers）
    """

    def __init__(self, model_name: str = "all-MiniLM-L6-v2"):
        """
        初始化 Embedding 服务

        Args:
            model_name: 模型名称，默认使用轻量快速的 all-MiniLM-L6-v2
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
        """返回向量维度"""
        if self.mock:
            return 384  # MiniLM 的默认维度
        return self._model.get_sentence_embedding_dimension()

    async def embed(self, text: str) -> list[float]:
        """
        生成单个文本的向量嵌入

        Args:
            text: 输入文本

        Returns:
            float 列表，向量
        """
        if self.mock:
            # 返回随机向量
            import random
            return [random.random() for _ in range(384)]

        embedding = self._model.encode(text)
        return embedding.tolist()

    async def embed_batch(self, texts: list[str]) -> list[list[float]]:
        """
        批量生成向量嵌入

        Args:
            texts: 文本列表

        Returns:
            向量列表
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
        同步编码接口（供内部使用）

        Args:
            text: 输入文本

        Returns:
            numpy 向量
        """
        if self.mock:
            import random
            return np.array([random.random() for _ in range(384)])

        return self._model.encode(text)
