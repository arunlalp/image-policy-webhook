FROM python:3.9-slim

WORKDIR /app

# Upgrade pip to the latest version for better dependency resolution
RUN pip install --upgrade pip

# Ensure compatible urllib3 version is installed before others
COPY requirements.txt .
RUN pip install --no-cache-dir "urllib3>=1.25.4,<1.27" && \
    pip install --no-cache-dir -r requirements.txt

COPY image_policy_webhook.py .

CMD ["python", "image_policy_webhook.py"]
