"""
信息提取服务
从文本中提取结构化记忆
"""

from typing import Optional
import re


class MemoryExtractor:
    """
    记忆信息提取器
    MVP 阶段使用规则匹配，后续可接入 LLM
    """

    def __init__(self):
        """初始化提取器"""
        # 偏好模式
        self.preference_patterns = [
            r"喜欢\s+(.+)",
            r"偏好\s+(.+)",
            r"倾向于\s+(.+)",
            r"更喜欢\s+(.+)",
            r"喜欢用\s+(.+)",
            r"使用\s+(.+?)\s+编程",
        ]

        # 技能模式
        self.skill_patterns = [
            r"擅长\s+(.+)",
            r"会用\s+(.+)",
            r"熟悉\s+(.+)",
            r"掌握\s+(.+)",
        ]

        # 目标模式
        self.goal_patterns = [
            r"想要\s+(.+)",
            r"打算\s+(.+)",
            r"计划\s+(.+)",
            r"目标\s+(.+)",
        ]

    async def extract(self, text: str) -> dict:
        """
        从文本中提取结构化记忆

        Args:
            text: 输入文本

        Returns:
            包含 type, key, value, confidence 的字典
        """
        text = text.strip()

        # 尝试提取偏好
        for pattern in self.preference_patterns:
            match = re.search(pattern, text)
            if match:
                return {
                    "type": "preference",
                    "key": self._extract_key(text),
                    "value": match.group(1).strip(),
                    "confidence": 0.9
                }

        # 尝试提取技能
        for pattern in self.skill_patterns:
            match = re.search(pattern, text)
            if match:
                return {
                    "type": "skill",
                    "key": self._extract_key(text),
                    "value": match.group(1).strip(),
                    "confidence": 0.85
                }

        # 尝试提取目标
        for pattern in self.goal_patterns:
            match = re.search(pattern, text)
            if match:
                return {
                    "type": "goal",
                    "key": self._extract_key(text),
                    "value": match.group(1).strip(),
                    "confidence": 0.8
                }

        # 默认作为事实
        return {
            "type": "fact",
            "key": self._extract_key(text),
            "value": text,
            "confidence": 0.6
        }

    def _extract_key(self, text: str) -> str:
        """
        从文本中提取 key
        使用文本的前 N 个字符或第一个句子
        """
        # 取前 50 个字符作为 key
        if len(text) > 50:
            # 尝试在句号或逗号处截断
            for sep in ["。", "，", ".", ","]:
                idx = text[:50].rfind(sep)
                if idx > 0:
                    return text[:idx+1].strip()
            return text[:50] + "..."

        return text
