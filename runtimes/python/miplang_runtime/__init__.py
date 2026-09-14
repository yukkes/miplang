from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any


@dataclass(frozen=True)
class SetIR:
    name: str


@dataclass(frozen=True)
class ModelIR:
    schema_version: str
    name: str
    sets: tuple[str, ...]
    raw: dict[str, Any]


def load_ir(path: str | Path) -> ModelIR:
    data = json.loads(Path(path).read_text(encoding="utf-8"))
    return ModelIR(
        schema_version=data["schemaVersion"],
        name=data["name"],
        sets=tuple(item["name"] for item in data.get("sets", [])),
        raw=data,
    )


__all__ = ["ModelIR", "SetIR", "load_ir"]
