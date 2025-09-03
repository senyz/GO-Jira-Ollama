go build -o jira-ai
docker run -it --rm -v "%cd%:/app" -w /app golang:alpine sh