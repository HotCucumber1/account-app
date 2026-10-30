# Build locally first: mise run //modules/access:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/access /access
EXPOSE 8081 8082
ENTRYPOINT ["/access"]
CMD ["service"]
