# Build locally first: mise run //modules/account:build
FROM gcr.io/distroless/base-debian12:nonroot
COPY dist/account /account
EXPOSE 8081 8082
ENTRYPOINT ["/account"]
CMD ["service"]
