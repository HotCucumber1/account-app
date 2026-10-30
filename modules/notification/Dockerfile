# Build locally first: mise run //modules/notification:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/notification /notification
EXPOSE 8081 8082
ENTRYPOINT ["/notification"]
CMD ["service"]
