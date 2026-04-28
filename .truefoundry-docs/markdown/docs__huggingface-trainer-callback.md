> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Huggingface Trainer Callback

> Callback for Huggingface Trainer to automatically log metrics and checkpoints to TrueFoundry

The following example demonstrates how to use the TrueFoundry Huggingface `TrainerCallback` to automatically log metrics and checkpoints to TrueFoundry.
As an example, we will train a small language model on a small dataset but you can use this callback for any model. We also log the model at the end of the training.

## Training Script with Huggingface Trainer Callback

<Note type="info">
  `TrueFoundryMLCallback` is available in `truefoundry>=0.11.4`
</Note>

<Note type="info">
  Please choose `logging_steps` and `save_steps` in `TrainingArguments` carefully.

  * `logging_steps` controls how often the metrics are logged
  * `save_steps` controls how often the checkpoints are uploaded
</Note>

```python train.py {12-15,37-48,93-103,112,121-132} lines theme={"dark"}
import logging
from datetime import datetime

from datasets import Dataset
from transformers import (
    AutoModelForCausalLM,
    AutoTokenizer,
)
from transformers.trainer import Trainer
from transformers.training_args import TrainingArguments

from truefoundry import ml
from truefoundry.ml.integrations.huggingface.trainer_callback import (
    TrueFoundryMLCallback,
)

logger = logging.getLogger(__name__)


def tokenize_fn(batch, tokenizer):
    out = tokenizer(batch["text"], truncation=True, max_length=128, padding="longest")
    out["labels"] = out["input_ids"].copy()
    return out


def main():
    import argparse

    parser = argparse.ArgumentParser()
    parser.add_argument("--ml-repo", type=str, required=True)
    parser.add_argument("--run-name", type=str, required=False)
    args = parser.parse_args()

    dt = datetime.now().strftime("%Y-%m-%d-%H-%M-%S")
    run_name = args.run_name or f"test-run-{dt}"

    client = ml.get_client()
    run = client.create_run(
        ml_repo=args.ml_repo,
        run_name=run_name,
        auto_end=False,
    )
    run_name = run.run_name
    callback = TrueFoundryMLCallback(
        run=run,
        log_checkpoints=True,
        checkpoint_artifact_name=f"ckpt-{run_name}",
    )

    # 1. Create minimal dataset
    examples = [
        {"text": "Hello, how are you? I am fine."},
        {"text": "Translate: cat -> chat"},
        {"text": "Q: What is 2+2? A: 4"},
        {"text": "Say hello in Spanish: Hola"},
        {"text": "Who wrote 'Hamlet'? William Shakespeare."},
    ]
    ds = Dataset.from_list(examples)

    # 2. Load tokenizer and model
    model_id = "llamafactory/tiny-random-Llama-3"
    tokenizer = AutoTokenizer.from_pretrained(model_id)
    tokenizer.pad_token = tokenizer.eos_token
    model = AutoModelForCausalLM.from_pretrained(model_id)

    # 3. Tokenize as causal LM
    ds = ds.map(
        tokenize_fn,
        batched=True,
        remove_columns=["text"],
        fn_kwargs={"tokenizer": tokenizer},
    )

    # 4. Setup training arguments
    training_args = TrainingArguments(
        output_dir="./output",
        per_device_train_batch_size=len(examples),
        num_train_epochs=15,
        logging_strategy="steps",
        logging_steps=5,
        eval_strategy="steps",
        eval_steps=5,
        save_strategy="steps",
        save_steps=5,
        save_total_limit=2,
        learning_rate=5e-5,
        fp16=False,
        bf16=False,
        report_to="none",
        load_best_model_at_end=True,
    )

    run.log_params(
        {
            "model_id": model_id,
            "batch_size": training_args.per_device_train_batch_size,
            "num_train_epochs": training_args.num_train_epochs,
            "learning_rate": training_args.learning_rate,
            "fp16": training_args.fp16,
            "bf16": training_args.bf16,
            "report_to": training_args.report_to,
        }
    )

    # 5. Trainer
    trainer = Trainer(
        model=model,
        tokenizer=tokenizer,
        args=training_args,
        train_dataset=ds,
        eval_dataset=ds,
        callbacks=[callback],
    )

    # 6. Train and save
    trainer.train()
    model_dir = "./model"
    trainer.save_model(model_dir)

    # 7. Log the final model
    run.log_model(
        name=f"model-{run_name}",
        model_file_or_folder=model_dir,
        framework=ml.TransformersFramework(
            library_name="transformers",  # type: ignore
            pipeline_tag="text-generation",
        ),
        metadata={
            "base_model": model_id,
        },
        step=trainer.state.best_global_step or trainer.state.global_step,
    )


if __name__ == "__main__":
    main()
```

## Running the script

1. [Create a ML Repo](/docs/ml-repo-quickstart#create-a-ml-repo) if not already created
2. [Setup TrueFoundry CLI](/docs/ml-repo-quickstart#setup-the-truefoundry-cli)
3. Run the script

```shell lines theme={"dark"}
python train.py --ml-repo <ml-repo-name> --run-name <run-name>
# E.g. python train.py  --ml-repo llm-finetuning --run-name test-hf-cb-gxerg
```

## Viewing the Run on UI

<Tip>
  If you ran the script as a [TrueFoundry Job](/docs/introduction-to-a-job), you can access the run in the Job Details column of job run

  <Frame caption="Click on the Run Name to view the run details">
    <img src="https://mintcdn.com/truefoundry/5ZMNUF311tgGO_yT/images/hf-trainer-callback-5.png?fit=max&auto=format&n=5ZMNUF311tgGO_yT&q=85&s=f607a2b388d60ee5b82bf78eeaae613e" width="2916" height="926" data-path="images/hf-trainer-callback-5.png" />
  </Frame>
</Tip>

Navigate to the ML Repo, you should see the run in the Runs Tab.

All metrics, checkpoints and models should be available in this run.

<Frame caption="Run Details">
  <img src="https://mintcdn.com/truefoundry/5ZMNUF311tgGO_yT/images/hf-trainer-callback-1.png?fit=max&auto=format&n=5ZMNUF311tgGO_yT&q=85&s=8ea74c68d5efefd86cdceba17eb91bd6" width="2910" height="1554" data-path="images/hf-trainer-callback-1.png" />
</Frame>

<Frame caption="Metrics">
  <img src="https://mintcdn.com/truefoundry/5ZMNUF311tgGO_yT/images/hf-trainer-callback-2.png?fit=max&auto=format&n=5ZMNUF311tgGO_yT&q=85&s=1ad4d2310cd34fea86f2d519a6cb050d" width="2924" height="1578" data-path="images/hf-trainer-callback-2.png" />
</Frame>

<Frame caption="Checkpoints">
  <img src="https://mintcdn.com/truefoundry/5ZMNUF311tgGO_yT/images/hf-trainer-callback-3.png?fit=max&auto=format&n=5ZMNUF311tgGO_yT&q=85&s=c38efab47bae05e91211758af8f636d0" width="2902" height="1544" data-path="images/hf-trainer-callback-3.png" />
</Frame>

<Frame caption="Models">
  <img src="https://mintcdn.com/truefoundry/5ZMNUF311tgGO_yT/images/hf-trainer-callback-4.png?fit=max&auto=format&n=5ZMNUF311tgGO_yT&q=85&s=507c013d71dd02db1c3becbbdacaa2f2" width="2884" height="1350" data-path="images/hf-trainer-callback-4.png" />
</Frame>
