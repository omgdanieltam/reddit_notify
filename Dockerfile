# syntax=docker/dockerfile:1

FROM python:3.12-alpine

WORKDIR /app
COPY . .

RUN pip install -r requirements.txt

ENTRYPOINT ["python", "main.py"]