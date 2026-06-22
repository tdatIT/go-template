---
name: writing-dockerfile

description: "Writes Dockerfiles for containerizing applications. Use when the user asks to create a Dockerfile, optimize image size, or add multi-stage builds for Go projects."
---

# Writing Dockerfiles for Go Projects
Create Dockerfiles to containerize Go applications. Use multi-stage builds to keep the image small and the runtime minimal.

## Workflow
1. Start with a generic Go builder image, such as `golang:<version>-alpine`, and use `/app` as the working directory.
2. Copy `go.mod` and `go.sum` first, then run `go mod download` to cache dependencies.
3. Copy the remaining source code after dependencies are ready.
4. Build a static Linux binary with the needed flags, for example:

   ```dockerfile
   RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
       -tags musl \
       -ldflags="-w -s -extldflags '-static'" \
       -o /app/app-bin \
       ./cmd/main.go
   ```

5. Use a minimal runtime image such as `alpine:<version>`, install only required packages like `ca-certificates`, and avoid extra tools.
6. Create a non-root user and group, then set the binary ownership and permissions before switching users.
7. Copy the built binary **and the `config/` directory** into the runtime image. Set `ENV CONFIG_PATH=/app/config/config` so Viper can find the default config file without a bind mount:

   ```dockerfile
   COPY --from=builder /app/app-bin   /app/app-bin
   COPY --from=builder /app/config    /app/config
   ENV CONFIG_PATH=/app/config/config
   ```

8. Expose only the needed port and set the binary as the entrypoint.


## Rules
- Use multi-stage builds to separate build-time and runtime dependencies.
- Keep version placeholders generic unless the user asks for exact versions.
- Avoid including source files, compilers, package managers, or other unnecessary artifacts in the final image.
- Prefer a static binary and a minimal runtime base image when CGO is not needed.
- Prefer a non-root runtime user when elevated privileges are not required.