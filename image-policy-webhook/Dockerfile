FROM python:3.9-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY image_policy_webhook.py .


CMD ["python", "image_policy_webhook.py"]