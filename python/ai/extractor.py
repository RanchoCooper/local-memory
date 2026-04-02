"""
Information Extractor
Extracts structured memories from text
"""

from typing import Optional
import re


class MemoryExtractor:
    """
    Memory information extractor.
    MVP uses rule-based matching, can be extended with LLM in the future.
    """

    def __init__(self):
        """Initialize the extractor"""
        # Preference patterns
        self.preference_patterns = [
            r"like[sz]?\s+(.+)",
            r"prefer[s]?\s+(.+)",
            r"rather\s+(?:than\s+)?(.+)",
            r"favorite\s+(?:thing\s+)?(.+)",
            r"enjoys?\s+(.+)",
            r"using\s+(.+?)\s+for\s+programming",
        ]

        # Skill patterns
        self.skill_patterns = [
            r"good\s+at\s+(.+)",
            r"skille[dr]\s+(.+)",
            r"knows?\s+(.+)",
            r"experienced\s+in\s+(.+)",
        ]

        # Goal patterns
        self.goal_patterns = [
            r"want[s]?\s+to\s+(.+)",
            r"plains?\s+to\s+(.+)",
            r"going\s+to\s+(.+)",
            r"goal\s+(?:is\s+)?(.+)",
        ]

    async def extract(self, text: str) -> dict:
        """
        Extract structured memory from text.

        Args:
            text: Input text

        Returns:
            Dictionary containing type, key, value, confidence
        """
        text = text.strip()

        # Try to extract preference
        for pattern in self.preference_patterns:
            match = re.search(pattern, text, re.IGNORECASE)
            if match:
                return {
                    "type": "preference",
                    "key": self._extract_key(text),
                    "value": match.group(1).strip(),
                    "confidence": 0.9
                }

        # Try to extract skill
        for pattern in self.skill_patterns:
            match = re.search(pattern, text, re.IGNORECASE)
            if match:
                return {
                    "type": "skill",
                    "key": self._extract_key(text),
                    "value": match.group(1).strip(),
                    "confidence": 0.85
                }

        # Try to extract goal
        for pattern in self.goal_patterns:
            match = re.search(pattern, text, re.IGNORECASE)
            if match:
                return {
                    "type": "goal",
                    "key": self._extract_key(text),
                    "value": match.group(1).strip(),
                    "confidence": 0.8
                }

        # Default to fact
        return {
            "type": "fact",
            "key": self._extract_key(text),
            "value": text,
            "confidence": 0.6
        }

    def _extract_key(self, text: str) -> str:
        """
        Extract key from text.
        Uses the first N characters or the first sentence.
        """
        # Use first 50 characters as key
        if len(text) > 50:
            # Try to truncate at sentence boundary
            for sep in [".", ",", ";", ":", "!", "?"]:
                idx = text[:50].rfind(sep)
                if idx > 0:
                    return text[:idx+1].strip()
            return text[:50] + "..."

        return text
