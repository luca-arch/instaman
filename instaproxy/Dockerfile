FROM python:3.12.5-bookworm

EXPOSE 15000

ENV PYTHONUNBUFFERED=1
ENV PYTHONDONTWRITEBYTECODE=1

# Install dependencies first.
COPY requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir -r /tmp/requirements.txt && \
    rm /tmp/requirements.txt

# Copy custom module. This folder should be overridden by the Docker Compose file when developing locally.
ENV PYTHONPATH="/mnt/modules"
COPY ./instaproxy /mnt/modules/instaproxy

ENTRYPOINT [ "fastapi", "run", "--port", "15000", "/mnt/modules/instaproxy/server.py" ]