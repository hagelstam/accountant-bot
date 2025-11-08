FROM public.ecr.aws/lambda/python:3.14

COPY --from=ghcr.io/astral-sh/uv:latest /uv /usr/local/bin/uv

WORKDIR ${LAMBDA_TASK_ROOT}

COPY uv.lock pyproject.toml ./

RUN uv export --frozen --no-hashes > requirements.txt && \
    pip install --no-cache-dir -r requirements.txt && \
    rm requirements.txt

COPY src/accountant_bot ./accountant_bot

CMD ["accountant_bot.lambda_handler.lambda_handler"]
