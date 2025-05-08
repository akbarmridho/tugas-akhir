FROM node:20.19.1-bookworm-slim AS builder

WORKDIR /app

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

# Copy just package.json and package-lock.json
# to speed up the build using Docker layer cache.
COPY package*.json ./
COPY pnpm-lock.yaml ./

ENV NODE_ENV=development

# Install all dependencies. Don't audit to speed up the installation.
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile

# Next, copy the source files using the user set
# in the base image.
COPY . ./

# Install all dependencies and build the project.
# Don't audit to speed up the installation.
RUN pnpm run build

# Create final image
FROM node:20.19.1-bookworm-slim

WORKDIR /app

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

# Copy only built JS files from builder image
COPY --from=builder /app /app

# Copy just package.json and package-lock.json
# to speed up the build using Docker layer cache.
COPY package*.json ./
COPY pnpm-lock.yaml ./

ENV NODE_ENV=production

# Install NPM packages, skip optional and development dependencies to
# keep the image small. Avoid logging too much and print the dependency
# tree for debugging
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm config set loglevel error \
    && pnpm install --prod --frozen-lockfile \
    && echo "Installed PNPM packages:" \
    && (pnpm list --prod --depth Infinity || true) \
    && echo "Node.js version:" \
    && node --version \
    && echo "NPM version:" \
    && pnpm --version

# Next, copy the remaining files and directories with the source code.
# Since we do this after NPM install, quick build will be really fast
# for most source file changes.
COPY . ./
