> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# DSPy

> Learn how to use dspy with TrueFoundry AI Gateway, including setup steps, use cases, and production-ready examples.

This guide provides instructions for integrating [DSPy](https://github.com/stanfordnlp/dspy) with the Truefoundry AI Gateway.

## What is DSPy?

DSPy is a framework for algorithmically optimizing language model prompts and weights through a programming-first approach. It enables developers to build and optimize LM-based systems by treating prompts as learnable parameters rather than manually crafted text, using a declarative programming model that separates logic from optimization.

## Prerequisites

Before integrating DSPy with TrueFoundry, ensure you have:

1. **TrueFoundry Account**: Create a [Truefoundry account](https://www.truefoundry.com/register) and follow the instructions in our [Gateway Quick Start Guide](https://docs.truefoundry.com/gateway/quick-start)
2. **DSPy Installation**: Install DSPy using pip: `pip install -U dspy`

## Setup Process

### 1. Configure DSPy with TrueFoundry Gateway

DSPy integrates seamlessly with TrueFoundry's gateway through the `LM` interface. Configure the LM with TrueFoundry's gateway URL and your API key:

```python lines theme={"dark"}
import dspy

# Set up your TrueFoundry-enabled LM client
lm = dspy.LM(
    model="openai/anthropic-account/claude-4",  # Use openai/ prefix with TrueFoundry model name
    api_key="your-truefoundry-api-key",
    api_base="{GATEWAY_BASE_URL}"
)

# Configure DSPy to use the TrueFoundry-enabled client
dspy.configure(lm=lm)
```

You will get your base URL and model name directly from the unified code snippet:

<Frame>
  <img src="https://mintcdn.com/truefoundry/n3EuZuJ0K8wBFp1G/images/new-code-snippet.png?fit=max&auto=format&n=n3EuZuJ0K8wBFp1G&q=85&s=3634c2dc8c3565fd77ab896d3fd07ed9" alt="TrueFoundry playground showing unified code snippet with base URL and model name" width="2940" height="1664" data-path="images/new-code-snippet.png" />
</Frame>

Replace:

* `your-truefoundry-api-key` with your actual TrueFoundry API key (if required)
* `{GATEWAY_BASE_URL}` with your TrueFoundry Gateway Base URL
* `openai/anthropic-account/claude-4` with your desired model using the `openai/` prefix

### 2. Environment Variables Configuration

For persistent configuration across your DSPy applications, set these environment variables:

```bash lines theme={"dark"}
export OPENAI_API_KEY="your-truefoundry-api-key"
export OPENAI_BASE_URL="{GATEWAY_BASE_URL}"
```

## Usage Examples

### Basic DSPy with TrueFoundry Gateway

Here's a simple example demonstrating DSPy with TrueFoundry integration:

```python lines theme={"dark"}
import dspy

# Set up the TrueFoundry-enabled client
lm = dspy.LM(
    model="openai/anthropic-account/claude-4",
    api_key="your-truefoundry-api-key",
    api_base="{GATEWAY_BASE_URL}"
)
dspy.configure(lm=lm)

# Simple completion
response = lm("Say this is a test!", temperature=0.7)
print(response)  # => ['This is a test!']
```

### DSPy Signatures and Modules

Create more sophisticated DSPy programs using signatures and modules:

```python lines theme={"dark"}
import dspy

# Configure TrueFoundry LM
lm = dspy.LM(
    model="openai/anthropic-account/claude-4",
    api_key="your-truefoundry-api-key",
    api_base="{GATEWAY_BASE_URL}"
)
dspy.configure(lm=lm)

# Define a signature for question answering
class GenerateAnswer(dspy.Signature):
    """Answer questions with short factoid answers."""
    question = dspy.InputField()
    answer = dspy.OutputField(desc="often between 1 and 5 words")

# Create a module using Chain of Thought
generate_answer = dspy.ChainOfThought(GenerateAnswer)

# Ask a question
question = "What is the capital of Brazil?"
pred = generate_answer(question=question)
print(f"Question: {question}")
print(f"Answer: {pred.answer}")
```

### Advanced RAG System with DSPy

Build a complete RAG (Retrieval-Augmented Generation) system:

```python lines theme={"dark"}
import dspy

# Configure TrueFoundry-enabled LM
lm = dspy.LM(
    model="openai/anthropic-account/claude-4",
    api_key="your-truefoundry-api-key",
    api_base="{GATEWAY_BASE_URL}"
)
dspy.configure(lm=lm)

# Define signatures for RAG pipeline
class GenerateSearchQuery(dspy.Signature):
    """Write a simple search query that will help answer a complex question."""
    context = dspy.InputField(desc="may contain relevant facts")
    question = dspy.InputField()
    query = dspy.OutputField()

class GenerateAnswer(dspy.Signature):
    """Answer questions with short factoid answers."""
    context = dspy.InputField(desc="may contain relevant facts")
    question = dspy.InputField()
    answer = dspy.OutputField(desc="often between 1 and 5 words")

# Mock retrieval function (replace with your actual retrieval system)
def retrieve(query: str) -> list[str]:
    # This would typically connect to your vector database or search API
    return [
        "Brazil is a country in South America.",
        "The capital of Brazil is Brasília.",
        "Brasília was founded in 1960 and became the capital."
    ]

# Create RAG module
class RAG(dspy.Module):
    def __init__(self):
        super().__init__()
        self.generate_query = dspy.ChainOfThought(GenerateSearchQuery)
        self.generate_answer = dspy.ChainOfThought(GenerateAnswer)
    
    def forward(self, question):
        # Generate search query
        search_query = self.generate_query(context="", question=question).query
        
        # Retrieve relevant documents
        passages = retrieve(search_query)
        context = "\n".join(passages)
        
        # Generate final answer
        pred = self.generate_answer(context=context, question=question)
        return dspy.Prediction(context=context, answer=pred.answer)

# Use the RAG system
rag = RAG()
question = "What is the capital of Brazil?"
result = rag(question)

print(f"Question: {question}")
print(f"Context: {result.context}")
print(f"Answer: {result.answer}")
```

### DSPy Optimization with TrueFoundry

Optimize your DSPy programs using the built-in optimizers:

```python lines theme={"dark"}
import dspy
from dspy.evaluate import Evaluate

# Configure TrueFoundry LM
lm = dspy.LM(
    model="openai/anthropic-account/claude-4",
    api_key="your-truefoundry-api-key",
    api_base="{GATEWAY_BASE_URL}"
)
dspy.configure(lm=lm)

# Define a simple QA module
class QA(dspy.Module):
    def __init__(self):
        super().__init__()
        self.generate_answer = dspy.ChainOfThought("question -> answer")
    
    def forward(self, question):
        prediction = self.generate_answer(question=question)
        return dspy.Prediction(answer=prediction.answer)

# Create training examples
trainset = [
    dspy.Example(question="What is the capital of France?", answer="Paris"),
    dspy.Example(question="What is the capital of Japan?", answer="Tokyo"),
    dspy.Example(question="What is the capital of Brazil?", answer="Brasília"),
]

# Define evaluation metric
def validate_answer(example, pred, trace=None):
    return example.answer.lower() in pred.answer.lower()

# Initialize module and optimizer
qa_module = QA()
optimizer = dspy.BootstrapFewShot(metric=validate_answer)

# Optimize the module
optimized_qa = optimizer.compile(qa_module, trainset=trainset)

# Test the optimized module
question = "What is the capital of Brazil?"
result = optimized_qa(question)
print(f"Optimized Answer: {result.answer}")
```

## Benefits of Using TrueFoundry Gateway with DSPy

1. **Cost Tracking**: Monitor and track costs across all your DSPy operations with detailed metrics
2. **Security**: Enhanced security with centralized API key management
3. **Access Controls**: Implement fine-grained access controls for different teams
4. **Rate Limiting**: Prevent API quota exhaustion with intelligent rate limiting
5. **Fallback Support**: Automatic failover to alternative providers when needed
6. **Analytics**: Detailed analytics and monitoring for all LLM calls in your DSPy pipelines
7. **Multi-Provider Support**: Seamlessly switch between different model providers (OpenAI, Anthropic, Google, etc.)
8. **Performance Optimization**: Track and optimize the performance of your DSPy modules
