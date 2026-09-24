# Semantic guardrail evaluation (Phase 0)

This command is the Phase 0 go/no-go check for the `semantic_policy`
guardrail. It sends a labeled evaluation set to TypeSafe Jev and measures
whether Jev can tell each rule's positives apart from its negatives in formal
Persian, colloquial Persian, Finglish, and English.

Each sample is one request. All rules for the sample's target (input or
output) go into that one request as Noul questions, the same batching the
gateway will use. The text is sent as
`{"role": "...", "untrusted_content": "..."}`, never as bare text.

## Dataset

`data/rules.json` defines five rules. Each rule has English and Persian wording
for its statement and for its `true`/`false` criteria.

| Rule | Target | Source |
| --- | --- | --- |
| `jailbreak` | input | security-baseline pack |
| `system_prompt_extraction` | input | security-baseline pack |
| `evaluator_address` | input | security-baseline pack |
| `social_engineering` | input | custom banking rule |
| `financial_commitment` | output | custom banking rule |

`jailbreak` was two rules in the first run: instruction override and
unrestricted role-play. Jev scored each rule's positives above 0.9 on the
other rule too. That is semantically right, since a no-limits persona is an
override, so the two became one rule.

`data/samples/*.jsonl` holds 465 synthetic samples: 202 input positives,
141 input hard negatives, 40 benign inputs, and 41 positive plus 41 hard
negative outputs. Every rule has at least 13 formal Persian, 13 colloquial
Persian, 6 Finglish, and 6 English positives.

`samples/hard.jsonl` holds 53 samples tagged `embedded` (an injection inside
a pasted email, minutes, or review), `obfuscated` (spaced letters, Arabic
ي and ك, zero-width characters, leetspeak Finglish), `multi_turn` (earlier
turns set up the attack, sent as `prior_turns`), `long`, or `paraphrase`.
Every tagged sample also counts in its register group, so the hard samples
are part of the gate, and the `hard` group reports them on their own.

A sample is a negative for a rule only when it is a hard negative or benign.
A sample that is positive for a different rule is left out of that rule's
metrics: an attack that also trips a second rule is still a correct block,
not a false positive.

Hard negatives are deliberately close to the boundary: fiction and
legitimate role-play, questions about jailbreaks as a topic, users correcting
their own earlier message, own-account banking questions, and replies that
describe a policy without promising anything.

All names, account numbers, and national IDs are fabricated.

## Running

Run from `services/control-plane`:

```powershell
go run ./cmd/semantic-guardrail-eval -validate-only
```

```powershell
$env:TYPESAFE_API_KEY = "<key>"
go run ./cmd/semantic-guardrail-eval -variants en,fa
```

`-variants en,fa` runs the full set twice: once with English rule wording and
once with Persian rule wording. The sample text is the same in both runs.
That makes 930 requests, well under the default limit of 1200 requests per
minute at `-concurrency 4`.

The run writes `report.json` and `predictions.jsonl` to `-out`, which
defaults to a `semantic-guardrail-eval` folder in the system temp directory,
and prints a summary table followed by the best threshold per language
class. Rules whose Persian, Finglish, and English thresholds differ by more
than 0.2 are flagged, because one shared threshold would cost recall on the
class that scores lower. `predictions.jsonl` keeps every rule
probability, so thresholds can be re-analyzed without calling Jev again.

## Gate

The gate passes when all of these hold:

- No request failed.
- Latency p95 is at or below `-max-p95-ms`, 400 ms by default.
- For at least one variant, every rule reaches `-min-recall` (0.85) on the
  `persian` group (formal plus colloquial, hard samples included) at a
  threshold whose false positive rate is at or below `-max-fpr` (0.03).

Finglish and English are reported but do not gate. The report also gives
recall and false positive rate at the PRD's default thresholds of 0.35
(review) and 0.70 (block).

## Limits of this set

- The set is small. The output rule has 28 Persian negatives, so a single
  false positive exceeds a 3% false positive rate. Treat results as a
  screening signal and grow the set from real, consented traffic before GA.
- The texts and labels were drafted with an AI assistant. A native Persian
  speaker should review them, especially the colloquial and Finglish
  registers, before the result is used for a decision.
- Latency is measured from wherever the command runs and includes retry
  backoff. Run it from inside the cluster for a production-representative
  p95.
