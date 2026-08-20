#!/usr/bin/env python3
"""Run a reproducible TTS/STT comparison across OpenAI-compatible gateways."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import statistics
import subprocess
import time
import unicodedata
import urllib.error
import urllib.request
import uuid
from dataclasses import asdict, dataclass
from datetime import UTC, datetime
from pathlib import Path
from typing import Any


TEST_CASES = (
    (
        "formal",
        "گدار ای‌آی یک درگاه یکپارچه برای دسترسی، پایش و مدیریت مدل‌های هوش مصنوعی است.",
    ),
    (
        "numbers",
        "سفارش شماره ۱۴۰۵، ساعت ۸ و ۴۵ دقیقه ثبت شد و مبلغ آن ۲ میلیون و ۳۵۰ هزار تومان بود.",
    ),
    (
        "technical",
        "در این آزمایش، دقت تبدیل گفتار به متن و تأخیر تولید گفتار از متن را برای API بررسی می‌کنیم.",
    ),
)


@dataclass(frozen=True)
class Platform:
    name: str
    base_url: str
    api_key: str
    tts_model: str
    stt_model: str


def utc_now() -> str:
    return datetime.now(UTC).isoformat()


def normalize_persian(text: str) -> str:
    text = unicodedata.normalize("NFKC", text).lower()
    text = text.translate(str.maketrans({"ي": "ی", "ى": "ی", "ك": "ک", "ۀ": "ه"}))
    text = re.sub(r"[\u064b-\u065f\u0670\u06d6-\u06ed]", "", text)
    text = re.sub(r"[^0-9۰-۹a-zآ-ی]+", " ", text)

    return " ".join(text.split())


def edit_distance(left: list[str], right: list[str]) -> int:
    previous = list(range(len(right) + 1))
    for row, left_item in enumerate(left, start=1):
        current = [row]
        for column, right_item in enumerate(right, start=1):
            current.append(
                min(
                    current[-1] + 1,
                    previous[column] + 1,
                    previous[column - 1] + (left_item != right_item),
                )
            )
        previous = current

    return previous[-1]


def error_rates(reference: str, hypothesis: str) -> dict[str, float]:
    normalized_reference = normalize_persian(reference)
    normalized_hypothesis = normalize_persian(hypothesis)
    reference_words = normalized_reference.split()
    hypothesis_words = normalized_hypothesis.split()
    reference_characters = list(normalized_reference.replace(" ", ""))
    hypothesis_characters = list(normalized_hypothesis.replace(" ", ""))

    return {
        "wer": edit_distance(reference_words, hypothesis_words) / max(1, len(reference_words)),
        "cer": edit_distance(reference_characters, hypothesis_characters)
        / max(1, len(reference_characters)),
    }


def request(
    url: str,
    api_key: str,
    body: bytes,
    content_type: str,
    timeout_seconds: float,
) -> tuple[int, dict[str, str], bytes, float]:
    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": content_type,
        "User-Agent": "godarai-audio-platform-benchmark/1.0",
    }
    http_request = urllib.request.Request(url, data=body, headers=headers, method="POST")
    started = time.perf_counter()
    try:
        with urllib.request.urlopen(http_request, timeout=timeout_seconds) as response:
            payload = response.read()
            status = response.status
            response_headers = dict(response.headers.items())
    except urllib.error.HTTPError as error:
        payload = error.read()
        status = error.code
        response_headers = dict(error.headers.items())
    latency_ms = (time.perf_counter() - started) * 1000

    return status, response_headers, payload, latency_ms


def safe_headers(headers: dict[str, str]) -> dict[str, str]:
    allowed = {
        "content-length",
        "content-type",
        "openai-processing-ms",
        "x-request-id",
        "x-godarai-request-id",
        "x-ratelimit-limit-requests",
        "x-ratelimit-remaining-requests",
        "x-ratelimit-reset-requests",
    }

    return {key.lower(): value for key, value in headers.items() if key.lower() in allowed}


def audio_duration(path: Path) -> float | None:
    try:
        result = subprocess.run(
            [
                "ffprobe",
                "-v",
                "error",
                "-show_entries",
                "format=duration",
                "-of",
                "default=noprint_wrappers=1:nokey=1",
                str(path),
            ],
            check=True,
            capture_output=True,
            text=True,
        )
    except (FileNotFoundError, subprocess.CalledProcessError):
        return None

    return float(result.stdout.strip())


def json_error(payload: bytes) -> str:
    try:
        parsed = json.loads(payload)
    except (UnicodeDecodeError, json.JSONDecodeError):
        return payload[:300].decode("utf-8", errors="replace")

    if isinstance(parsed, dict):
        error = parsed.get("error")
        if isinstance(error, dict):
            return str(error.get("message") or error)
        if error:
            return str(error)

    return str(parsed)[:300]


def synthesize(
    platform: Platform,
    case_name: str,
    text: str,
    output_dir: Path,
    timeout_seconds: float,
) -> tuple[dict[str, Any], Path | None]:
    body = json.dumps(
        {
            "model": platform.tts_model,
            "voice": "alloy",
            "input": text,
            "response_format": "mp3",
            "speed": 1.0,
        },
        ensure_ascii=False,
    ).encode("utf-8")
    status, headers, payload, latency_ms = request(
        f"{platform.base_url}/audio/speech",
        platform.api_key,
        body,
        "application/json; charset=utf-8",
        timeout_seconds,
    )
    result: dict[str, Any] = {
        "platform": platform.name,
        "case": case_name,
        "model": platform.tts_model,
        "status": status,
        "latency_ms": round(latency_ms, 2),
        "headers": safe_headers(headers),
    }
    if status < 200 or status >= 300:
        result["error"] = json_error(payload)

        return result, None

    path = output_dir / f"tts-{platform.name}-{case_name}.mp3"
    path.write_bytes(payload)
    duration = audio_duration(path)
    result.update(
        {
            "bytes": len(payload),
            "sha256": hashlib.sha256(payload).hexdigest(),
            "audio_duration_seconds": round(duration, 3) if duration is not None else None,
            "realtime_factor": round(latency_ms / 1000 / duration, 3)
            if duration
            else None,
            "artifact": path.name,
        }
    )

    return result, path


def multipart_body(fields: dict[str, str], file_path: Path) -> tuple[bytes, str]:
    boundary = f"----godarai-benchmark-{uuid.uuid4().hex}"
    chunks: list[bytes] = []
    for key, value in fields.items():
        chunks.extend(
            [
                f"--{boundary}\r\n".encode(),
                f'Content-Disposition: form-data; name="{key}"\r\n\r\n'.encode(),
                value.encode("utf-8"),
                b"\r\n",
            ]
        )
    chunks.extend(
        [
            f"--{boundary}\r\n".encode(),
            (
                f'Content-Disposition: form-data; name="file"; filename="{file_path.name}"\r\n'
            ).encode(),
            b"Content-Type: audio/mpeg\r\n\r\n",
            file_path.read_bytes(),
            b"\r\n",
            f"--{boundary}--\r\n".encode(),
        ]
    )

    return b"".join(chunks), f"multipart/form-data; boundary={boundary}"


def transcribe(
    platform: Platform,
    source_platform: str,
    case_name: str,
    reference: str,
    audio_path: Path,
    timeout_seconds: float,
) -> dict[str, Any]:
    body, content_type = multipart_body(
        {"model": platform.stt_model, "language": "fa", "response_format": "json"},
        audio_path,
    )
    status, headers, payload, latency_ms = request(
        f"{platform.base_url}/audio/transcriptions",
        platform.api_key,
        body,
        content_type,
        timeout_seconds,
    )
    result: dict[str, Any] = {
        "platform": platform.name,
        "source_platform": source_platform,
        "case": case_name,
        "model": platform.stt_model,
        "status": status,
        "latency_ms": round(latency_ms, 2),
        "headers": safe_headers(headers),
    }
    if status < 200 or status >= 300:
        result["error"] = json_error(payload)

        return result

    try:
        parsed = json.loads(payload)
        transcript = parsed.get("text", "") if isinstance(parsed, dict) else ""
    except (UnicodeDecodeError, json.JSONDecodeError):
        transcript = payload.decode("utf-8", errors="replace")
    result["transcript"] = transcript
    result.update(error_rates(reference, transcript))

    return result


def load_platforms() -> list[Platform]:
    definitions = (
        (
            "godarai",
            "GODARAI_API_KEY",
            "GODARAI_BASE_URL",
            "https://delpak.godarai.ir/v1",
            "GODARAI_TTS_MODEL",
            "openai:openai-calm-oasis:tts-1",
            "GODARAI_STT_MODEL",
            "openai:openai-calm-oasis:gpt-4o-mini-transcribe",
        ),
        (
            "avalai",
            "AVALAI_API_KEY",
            "AVALAI_BASE_URL",
            "https://api.avalai.ir/v1",
            "AVALAI_TTS_MODEL",
            "tts-1",
            "AVALAI_STT_MODEL",
            "gpt-4o-mini-transcribe",
        ),
    )
    platforms: list[Platform] = []
    for (
        name,
        key_variable,
        base_url_variable,
        default_base_url,
        tts_model_variable,
        default_tts_model,
        stt_model_variable,
        default_stt_model,
    ) in definitions:
        api_key = os.getenv(key_variable, "").strip()
        if not api_key:
            continue
        platforms.append(
            Platform(
                name=name,
                base_url=os.getenv(base_url_variable, default_base_url).rstrip("/"),
                api_key=api_key,
                tts_model=os.getenv(tts_model_variable, default_tts_model),
                stt_model=os.getenv(stt_model_variable, default_stt_model),
            )
        )

    return platforms


def summarize(records: list[dict[str, Any]], operation: str) -> dict[str, Any]:
    selected = [record for record in records if record["operation"] == operation]
    by_platform: dict[str, Any] = {}
    for platform in sorted({record["platform"] for record in selected}):
        platform_records = [record for record in selected if record["platform"] == platform]
        successful = [record for record in platform_records if 200 <= record["status"] < 300]
        summary: dict[str, Any] = {
            "requests": len(platform_records),
            "successes": len(successful),
            "success_rate": len(successful) / max(1, len(platform_records)),
        }
        if successful:
            summary["median_latency_ms"] = round(
                statistics.median(record["latency_ms"] for record in successful), 2
            )
        if operation == "tts" and successful:
            realtime_factors = [
                record["realtime_factor"]
                for record in successful
                if record.get("realtime_factor") is not None
            ]
            if realtime_factors:
                summary["median_realtime_factor"] = round(
                    statistics.median(realtime_factors), 3
                )
        if operation == "stt" and successful:
            summary["mean_wer"] = round(
                statistics.mean(record["wer"] for record in successful), 4
            )
            summary["mean_cer"] = round(
                statistics.mean(record["cer"] for record in successful), 4
            )
        by_platform[platform] = summary

    return by_platform


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--timeout", type=float, default=120)
    args = parser.parse_args()
    platforms = load_platforms()
    if not platforms:
        parser.error("Set GODARAI_API_KEY and/or AVALAI_API_KEY.")

    args.output.mkdir(parents=True, exist_ok=True)
    records: list[dict[str, Any]] = []
    audio_sources: list[tuple[str, str, str, Path]] = []
    for platform in platforms:
        for case_name, text in TEST_CASES:
            result, path = synthesize(platform, case_name, text, args.output, args.timeout)
            result["operation"] = "tts"
            records.append(result)
            if path is not None:
                audio_sources.append((platform.name, case_name, text, path))

    for platform in platforms:
        for source_platform, case_name, reference, audio_path in audio_sources:
            result = transcribe(
                platform,
                source_platform,
                case_name,
                reference,
                audio_path,
                args.timeout,
            )
            result["operation"] = "stt"
            records.append(result)

    report = {
        "schema_version": 1,
        "generated_at": utc_now(),
        "method": {
            "tts_voice": "alloy",
            "tts_speed": 1.0,
            "stt_language": "fa",
            "test_cases": [{"name": name, "text": text} for name, text in TEST_CASES],
            "platforms": [
                {
                    "name": platform.name,
                    "base_url": platform.base_url,
                    "tts_model": platform.tts_model,
                    "stt_model": platform.stt_model,
                }
                for platform in platforms
            ],
        },
        "summary": {
            "tts": summarize(records, "tts"),
            "stt": summarize(records, "stt"),
        },
        "records": records,
    }
    report_path = args.output / "results.json"
    report_path.write_text(json.dumps(report, ensure_ascii=False, indent=2), encoding="utf-8")
    print(json.dumps(report["summary"], ensure_ascii=False, indent=2))
    print(f"Report: {report_path}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
