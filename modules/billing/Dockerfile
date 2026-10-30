# Build locally first: mise run //modules/billing:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/billing /billing
EXPOSE 8081 8082
ENTRYPOINT ["/billing"]
CMD ["service"]
