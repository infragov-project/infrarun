FROM python:3.11-slim

ARG TARGETARCH

RUN apt-get update && \
    apt-get install -y --no-install-recommends git curl && \
    rm -rf /var/lib/apt/lists/*

RUN curl -sSL https://install.python-poetry.org | python -

ENV PATH="/root/.local/bin:$PATH"

WORKDIR /app

# Ensure that changes in the repo lead to cache invalidation
ADD https://api.github.com/repos/sr-lab/GLITCH/git/refs/heads/interactive_repair version.json
RUN git clone https://github.com/sr-lab/GLITCH.git

WORKDIR /app/GLITCH

RUN git checkout interactive_repair

WORKDIR /app/GLITCH/glitch/rego/rego_python/src/rego_python/bin/

RUN set -eux; \
    echo "Detected architecture: $TARGETARCH"; \
    if [ "$TARGETARCH" = "amd64" ]; then \
        curl -fL -o librego-linux-amd64.so https://github.com/sr-lab/GLITCH/releases/download/rego_python-v0.1.0/librego-linux-amd64.so; \
    elif [ "$TARGETARCH" = "arm64" ]; then \
        curl -fL -o librego-linux-arm64.so https://github.com/sr-lab/GLITCH/releases/download/rego_python-v0.1.0/librego-linux-arm64.so; \
    else \
        echo "Unsupported architecture: $TARGETARCH" >&2; \
        exit 1; \
    fi

WORKDIR /app/GLITCH/

RUN poetry lock

RUN poetry install

ENTRYPOINT ["poetry", "run", "glitch"]
