#!/usr/bin/env python3
"""Decode Cursor Agent stream-json NDJSON into a compact live console stream."""

from __future__ import annotations

import json
import sys
from typing import Any, Iterable


def nested_values(value: Any) -> Iterable[tuple[str, Any]]:
    if isinstance(value, dict):
        for key, child in value.items():
            yield key, child
            yield from nested_values(child)
    elif isinstance(value, list):
        for child in value:
            yield from nested_values(child)


def first_value(event: dict[str, Any], names: set[str]) -> Any:
    for key, value in nested_values(event):
        if key in names and value not in (None, "", [], {}):
            return value
    return None


def text_parts(value: Any) -> list[str]:
    parts: list[str] = []
    if isinstance(value, dict):
        if value.get("type") in {"text", "output_text"} and isinstance(value.get("text"), str):
            parts.append(value["text"])
        for key, child in value.items():
            if key not in {"thinking", "reasoning", "text"}:
                parts.extend(text_parts(child))
    elif isinstance(value, list):
        for child in value:
            parts.extend(text_parts(child))
    return parts


def tool_name(event: dict[str, Any]) -> str:
    direct = event.get("name") or event.get("tool_name")
    if isinstance(direct, str):
        return direct
    call = event.get("tool_call")
    if isinstance(call, dict):
        direct = call.get("name") or call.get("tool_name")
        if isinstance(direct, str):
            return direct
        for key in call:
            if key.endswith("ToolCall"):
                return key[: -len("ToolCall")]
    return "tool"


def tool_brief(event: dict[str, Any]) -> str:
    preferred = {"path", "command", "cmd", "url", "query", "pattern", "file_path"}
    value = first_value(event, preferred)
    if value is None:
        return ""
    if not isinstance(value, str):
        value = json.dumps(value, ensure_ascii=False, separators=(",", ":"))
    value = " ".join(value.split())
    return value[:117] + "..." if len(value) > 120 else value


partial_calls: set[str] = set()


def emit(event: dict[str, Any]) -> None:
    event_type = event.get("type")
    subtype = event.get("subtype")

    if event_type == "system" and subtype == "init":
        model = first_value(event, {"model"}) or "unknown"
        session = first_value(event, {"session_id", "sessionId"}) or "unknown"
        print(f"agent initialized (model={model}, session={session})", flush=True)
        return

    if event_type == "assistant":
        parts = text_parts(event.get("message", event))
        text = "".join(parts)
        if not text:
            return
        call_id = str(event.get("model_call_id") or event.get("modelCallId") or "default")
        partial = "timestamp_ms" in event or subtype in {"delta", "partial"}
        if partial:
            partial_calls.add(call_id)
            sys.stdout.write(text)
        elif call_id in partial_calls:
            partial_calls.remove(call_id)
            if not text.endswith("\n"):
                sys.stdout.write("\n")
        else:
            sys.stdout.write(text)
            if not text.endswith("\n"):
                sys.stdout.write("\n")
        sys.stdout.flush()
        return

    if event_type == "tool_call" and subtype == "started":
        brief = tool_brief(event)
        suffix = f" {brief}" if brief else ""
        print(f"→ {tool_name(event)}{suffix}", flush=True)
        return

    if event_type == "result":
        duration = event.get("duration_ms") or event.get("durationMs")
        suffix = f" ({duration} ms)" if duration is not None else ""
        print(f"agent finished{suffix}", flush=True)


def main() -> int:
    for raw_line in sys.stdin:
        line = raw_line.rstrip("\n")
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            print(line, flush=True)
            continue
        if isinstance(event, dict):
            emit(event)
        else:
            print(line, flush=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
