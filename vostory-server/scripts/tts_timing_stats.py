"""汇总 TTS worker 的「tts segment timing」日志，看瓶颈在准备、等 GPU、推理还是上传。

用法: python scripts/tts_timing_stats.py [日志文件，默认 storage/logs/server.log] [--since "2026-10-27 10:00"]
"""
import json
import sys

FIELDS = ["pipe_wait_ms", "prepare_ms", "gpu_wait_ms", "gpu_idle_ms", "infer_ms", "persist_ms", "finish_ms", "total_ms"]


def pct(values, p):
    if not values:
        return 0
    values = sorted(values)
    return values[min(len(values) - 1, int(len(values) * p))]


def main():
    args = sys.argv[1:]
    since = ""
    if "--since" in args:
        i = args.index("--since")
        since = args[i + 1]
        del args[i : i + 2]
    path = args[0] if args else "storage/logs/server.log"

    rows, results = [], {}
    with open(path, encoding="utf-8", errors="replace") as f:
        for line in f:
            if "tts segment timing" not in line or (since and line[: len(since)] < since):
                continue
            try:
                row = json.loads(line.rsplit("\t", 1)[1])
            except (ValueError, IndexError):
                continue
            results[row.get("result")] = results.get(row.get("result"), 0) + 1
            if row.get("result") == "generated":
                rows.append(row)

    print(f"片段数: {sum(results.values())}  按结果: {results}")
    if not rows:
        return
    print(f"\n已生成 {len(rows)} 条（单位 ms）")
    print(f"{'阶段':<14}{'平均':>8}{'P50':>8}{'P90':>8}{'最大':>8}")
    for name in FIELDS:
        vals = [r[name] for r in rows if name in r]
        if vals:
            print(f"{name:<14}{sum(vals) // len(vals):>8}{pct(vals, 0.5):>8}{pct(vals, 0.9):>8}{max(vals):>8}")

    # GPU 并发为 1 时才有 gpu_idle_ms；空闲 < 5 秒才算流水线没跟上（更长的多半是队列本来就空）
    idles = [r["gpu_idle_ms"] for r in rows if "gpu_idle_ms" in r]
    starved = [v for v in idles if 50 < v < 5000]
    if idles:
        infer = sum(r["infer_ms"] for r in rows)
        print(f"\nGPU 空等（50ms~5s，流水线没跟上）: {len(starved)}/{len(idles)} 次，"
              f"共 {sum(starved) / 1000:.1f}s，占推理总时长 {sum(starved) * 100 / max(infer, 1):.1f}%")


if __name__ == "__main__":
    main()
