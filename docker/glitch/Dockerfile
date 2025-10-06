FROM python:3.11-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends git curl && \
    rm -rf /var/lib/apt/lists/*

RUN curl -sSL https://install.python-poetry.org | python -

ENV PATH="/root/.local/bin:$PATH"

WORKDIR /app

RUN git clone https://github.com/sr-lab/GLITCH.git

WORKDIR /app/GLITCH

RUN git checkout interactive_repair

RUN poetry install

ENTRYPOINT ["poetry", "run", "glitch"]
