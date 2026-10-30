# Build locally first: mise run //modules/storage:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/storage /storage
EXPOSE 8081 8082
ENTRYPOINT ["/storage"]
CMD ["service"]
