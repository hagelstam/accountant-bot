FROM public.ecr.aws/lambda/python:3.14 AS builder

COPY --from=ghcr.io/astral-sh/uv:0.9.8 /uv /usr/local/bin/uv

WORKDIR /build

COPY uv.lock pyproject.toml ./

RUN uv export --frozen --no-hashes > requirements.txt && \
    pip install --no-cache-dir -r requirements.txt --target /build/python

FROM public.ecr.aws/lambda/python:3.14

WORKDIR ${LAMBDA_TASK_ROOT}

COPY --from=builder /build/python /var/lang/lib/python3.14/site-packages

COPY src/accountant_bot ./accountant_bot

CMD ["accountant_bot.lambda_handler.lambda_handler"]
